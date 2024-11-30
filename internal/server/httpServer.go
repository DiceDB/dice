package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/dicedb/dice/internal/eval"
	"github.com/dicedb/dice/internal/server/abstractserver"
	"github.com/dicedb/dice/internal/wal"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/internal/clientio"
	"github.com/dicedb/dice/internal/cmd"
	"github.com/dicedb/dice/internal/comm"
	derrors "github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/ops"
	"github.com/dicedb/dice/internal/server/utils"
	"github.com/dicedb/dice/internal/shard"
)

const (
	Abort     = "ABORT"
	stringNil = "(nil)"
)

var unimplementedCommands = map[string]bool{
	"Q.UNWATCH": true,
}

type HTTPServer struct {
	abstractserver.AbstractServer
	shardManager       *shard.ShardManager
	ioChan             chan *ops.StoreResponse
	httpServer         *http.Server
	qwatchResponseChan chan comm.QwatchResponse
	shutdownChan       chan struct{}
}

type HTTPQwatchResponse struct {
	Cmd   string `json:"cmd"`
	Query string `json:"query"`
	Data  []any  `json:"data"`
}

// CaseInsensitiveMux wraps ServeMux and forces REST paths to lowecase
type CaseInsensitiveMux struct {
	mux *http.ServeMux
}

func (cim *CaseInsensitiveMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Convert the path to lowercase before passing to the underlying mux.
	r.URL.Path = strings.ToLower(r.URL.Path)
	cim.mux.ServeHTTP(w, r)
}

func NewHTTPServer(shardManager *shard.ShardManager, wl wal.AbstractWAL) *HTTPServer {
	mux := http.NewServeMux()
	caseInsensitiveMux := &CaseInsensitiveMux{mux: mux}
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", config.DiceConfig.HTTP.Port),
		Handler:           caseInsensitiveMux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	httpServer := &HTTPServer{
		shardManager:       shardManager,
		ioChan:             make(chan *ops.StoreResponse, 1000),
		httpServer:         srv,
		qwatchResponseChan: make(chan comm.QwatchResponse),
		shutdownChan:       make(chan struct{}),
	}

	mux.HandleFunc("/", httpServer.DiceHTTPHandler)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("ok"))
		if err != nil {
			return
		}
	})

	return httpServer
}

func (s *HTTPServer) Run(ctx context.Context) error {
	var wg sync.WaitGroup
	var shutdownErr, listenErr error

	httpCtx, cancelHTTP := context.WithCancel(ctx)
	defer cancelHTTP()

	s.shardManager.RegisterIOThread("httpServer", s.ioChan, nil)

	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case <-ctx.Done():
		case <-s.shutdownChan:
			shutdownErr = derrors.ErrAborted
			slog.Debug("Shutting down HTTP Server")
		}

		err := s.httpServer.Shutdown(httpCtx)
		if err != nil {
			slog.Error("HTTP Server Shutdown Failed", slog.Any("error", err))
			shutdownErr = err
			return
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		listenErr = s.httpServer.ListenAndServe()
	}()

	wg.Wait()
	// Return the appropriate error
	if shutdownErr != nil {
		return shutdownErr
	}

	return listenErr
}

func (s *HTTPServer) DiceHTTPHandler(writer http.ResponseWriter, request *http.Request) {
	// convert to REDIS cmd
	diceDBCmd, err := utils.ParseHTTPRequest(request)
	if err != nil {
		responseJSON, _ := json.Marshal(utils.HTTPResponse{Status: utils.HTTPStatusError, Data: "Invalid HTTP request format"})
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusBadRequest) // Set HTTP status code to 500
		_, err = writer.Write(responseJSON)
		if err != nil {
			slog.Error("Error writing response", "error", err)
		}
		slog.Error("Error parsing HTTP request", slog.Any("error", err))
		return
	}

	if diceDBCmd.Cmd == Abort {
		slog.Debug("ABORT command received")
		slog.Debug("Shutting down HTTP Server")
		close(s.shutdownChan)
		return
	}

	if unimplementedCommands[diceDBCmd.Cmd] {
		responseJSON, _ := json.Marshal(utils.HTTPResponse{Status: utils.HTTPStatusError, Data: fmt.Sprintf("Command %s is not implemented with HTTP", diceDBCmd.Cmd)})
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusBadRequest) // Set HTTP status code to 500
		_, err = writer.Write(responseJSON)
		if err != nil {
			slog.Error("Error writing response", "error", err)
		}
		slog.Error("Command %s is not implemented", slog.String("cmd", diceDBCmd.Cmd))
		return
	}

	// send request to Shard Manager
	s.shardManager.GetShard(0).ReqChan <- &ops.StoreOp{
		Cmd:        diceDBCmd,
		IOThreadID: "httpServer",
		ShardID:    0,
		HTTPOp:     true,
	}

	// Wait for response
	resp := <-s.ioChan

	s.writeResponse(writer, resp, diceDBCmd)
}

func (s *HTTPServer) DiceHTTPQwatchHandler(writer http.ResponseWriter, request *http.Request) {
	// convert to REDIS cmd
	diceDBCmd, err := utils.ParseHTTPRequest(request)
	if err != nil {
		http.Error(writer, "Error parsing HTTP request", http.StatusBadRequest)
		slog.Error("Error parsing HTTP request", slog.Any("error", err))
		return
	}

	if len(diceDBCmd.Args) < 1 {
		slog.Error("Invalid request for QWATCH")
		http.Error(writer, "Invalid request for QWATCH", http.StatusBadRequest)
		return
	}

	// Check if the qwatchResponse channel exists
	if s.qwatchResponseChan == nil {
		http.Error(writer, "Internal error", http.StatusInternalServerError)
		return
	}

	// Check if the connection supports flushing
	flusher, ok := writer.(http.Flusher)
	if !ok {
		http.Error(writer, "Streaming unsupported", http.StatusBadRequest)
		return
	}

	// Set SSE headers
	writer.Header().Set("Content-Type", "text/event-stream")
	writer.Header().Set("Cache-Control", "no-cache")
	writer.Header().Set("Connection", "keep-alive")
	writer.WriteHeader(http.StatusOK)
	// We're a generating a unique client id, to keep track in core of requests from registered clients
	clientIdentifierID := generateUniqueInt32(request)
	qwatchQuery := diceDBCmd.Args[0]
	qwatchClient := comm.NewHTTPQwatchClient(s.qwatchResponseChan, clientIdentifierID)
	// Prepare the store operation
	storeOp := &ops.StoreOp{
		Cmd:        diceDBCmd,
		IOThreadID: "httpServer",
		ShardID:    0,
		Client:     qwatchClient,
		HTTPOp:     true,
	}

	slog.Info("Registered client for watching query", slog.Any("clientID", clientIdentifierID),
		slog.Any("query", qwatchQuery))
	s.shardManager.GetShard(0).ReqChan <- storeOp

	// Wait for 1st sync response from server for QWATCH and flush it to client
	resp := <-s.ioChan
	s.writeQWatchResponse(writer, resp)
	flusher.Flush()
	// Keep listening for context cancellation (client disconnect) and continuous responses
	doneChan := request.Context().Done()
	for {
		select {
		case resp := <-s.qwatchResponseChan:
			// Since we're reusing
			if resp.ClientIdentifierID == clientIdentifierID {
				s.writeQWatchResponse(writer, resp)
			}
		case <-s.shutdownChan:
			return
		case <-doneChan:
			// Client disconnected or request finished
			slog.Info("Client disconnected")
			unWatchCmd := &cmd.DiceDBCmd{
				Cmd:  "Q.UNWATCH",
				Args: []string{qwatchQuery},
			}
			storeOp.Cmd = unWatchCmd
			s.shardManager.GetShard(0).ReqChan <- storeOp
			resp := <-s.ioChan
			s.writeResponse(writer, resp, diceDBCmd)
			return
		}
	}
}

func (s *HTTPServer) writeQWatchResponse(writer http.ResponseWriter, response interface{}) {
	var result interface{}
	var err error

	// Use type assertion to handle both types of responses
	switch resp := response.(type) {
	case comm.QwatchResponse:
		result = resp.Result
		err = resp.Error
	case *ops.StoreResponse:
		result = resp.EvalResponse.Result
		err = resp.EvalResponse.Error
	default:
		slog.Error("Unsupported response type")
		http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var rp *clientio.RESPParser
	if err != nil {
		rp = clientio.NewRESPParser(bytes.NewBuffer([]byte(err.Error())))
	} else {
		rp = clientio.NewRESPParser(bytes.NewBuffer(result.([]byte)))
	}

	val, err := rp.DecodeOne()
	if err != nil {
		slog.Error("Error decoding response: %v", slog.Any("error", err))
		http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var responseJSON []byte
	// Convert the decoded response to the HTTPQwatchResponse struct
	// which will be sent to the HTTP SSE client if not of expected type response just encode and send
	switch v := val.(type) {
	case []interface{}:
		if len(v) >= 3 {
			qwatchResp := HTTPQwatchResponse{
				Cmd:   v[0].(string),
				Query: v[1].(string),
				Data:  v[2].([]interface{}),
			}
			responseJSON, err = json.Marshal(qwatchResp)
		}
	default:
		responseJSON, err = json.Marshal(val)
	}

	if err != nil {
		slog.Error("Error marshaling QueryData to JSON: %v", slog.Any("error", err))
		http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Format the response as SSE event
	_, err = writer.Write(responseJSON)
	if err != nil {
		slog.Error("Error writing SSE data: %v", slog.Any("error", err))
		http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	flusher, ok := writer.(http.Flusher)
	if !ok {
		http.Error(writer, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	flusher.Flush() // Flush the response to send it to the client
}

func (s *HTTPServer) writeResponse(writer http.ResponseWriter, result *ops.StoreResponse, diceDBCmd *cmd.DiceDBCmd) {
	var (
		responseValue interface{}
		err           error
		httpResponse  utils.HTTPResponse
		isDiceErr     bool
	)

	// Check if the command is migrated, if it is we use EvalResponse values
	// else we use RESPParser to decode the response
	_, ok := CmdMetaMap[diceDBCmd.Cmd]
	// TODO: Remove this conditional check and if (true) condition when all commands are migrated
	if !ok {
		responseValue, err = DecodeEvalResponse(result.EvalResponse)
		if err != nil {
			slog.Error("Error decoding response", "error", err)
			httpResponse = utils.HTTPResponse{Status: utils.HTTPStatusError, Data: "Internal Server Error"}
			writeJSONResponse(writer, httpResponse, http.StatusInternalServerError)
			return
		}
	} else {
		if result.EvalResponse.Error != nil {
			isDiceErr = true
			responseValue = result.EvalResponse.Error.Error()
		} else {
			responseValue = result.EvalResponse.Result
		}
	}

	// Create the HTTP response
	httpResponse = utils.HTTPResponse{Data: ResponseParser(responseValue)}
	if isDiceErr {
		httpResponse.Status = utils.HTTPStatusError
	} else {
		httpResponse.Status = utils.HTTPStatusSuccess
	}

	// Write the response back to the client
	writeJSONResponse(writer, httpResponse, http.StatusOK)
}

// Helper function to write the JSON response
func writeJSONResponse(writer http.ResponseWriter, response utils.HTTPResponse, statusCode int) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(statusCode)

	responseJSON, err := json.Marshal(response)
	if err != nil {
		slog.Error("Error marshaling response", "error", err)
		http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	_, err = writer.Write(responseJSON)
	if err != nil {
		slog.Error("Error writing response", "error", err)
	}
}

// ResponseParser parses the response value for both migrated and non-migrated cmds and
// returns response to be rendered for HTTP/WS response
func ResponseParser(responseValue interface{}) interface{} {
	switch v := responseValue.(type) {
	case []interface{}:
		// Parses []interface{} as part of EvalResponse e.g. JSON.ARRPOP
		// and adds to httpResponse. Also replaces "(nil)" with null JSON value
		// in response array.
		r := make([]interface{}, 0, len(v))
		for _, resp := range v {
			if val, ok := resp.(clientio.RespType); ok {
				if stringNil == RespTypeToValue(val) {
					r = append(r, nil)
				} else {
					r = append(r, RespTypeToValue(val))
				}
			} else {
				r = append(r, resp)
			}
		}
		return r

	case []byte:
		return string(v)

	case clientio.RespType:
		responseValue = RespTypeToValue(v)
		if responseValue == stringNil {
			responseValue = nil // in order to convert it in json null
		}

		return responseValue

	case interface{}:
		if val, ok := v.(clientio.RespType); ok {
			return RespTypeToValue(val)
		}
	}

	return responseValue
}

func generateUniqueInt32(r *http.Request) uint32 {
	var sb strings.Builder
	sb.WriteString(r.RemoteAddr)
	sb.WriteString(r.UserAgent())
	sb.WriteString(r.Method)
	sb.WriteString(r.URL.Path)

	// Hash the string using CRC32 and cast it to an int32
	return crc32.ChecksumIEEE([]byte(sb.String()))
}

// DecodeEvalResponse Helper function to decode EvalResponse based on the error or result
func DecodeEvalResponse(evalResp *eval.EvalResponse) (interface{}, error) {
	var rp *clientio.RESPParser

	if evalResp.Error != nil {
		rp = clientio.NewRESPParser(bytes.NewBuffer([]byte(evalResp.Error.Error())))
	} else {
		rp = clientio.NewRESPParser(bytes.NewBuffer(evalResp.Result.([]byte)))
	}

	res, err := rp.DecodeOne()
	if err != nil {
		return nil, err
	}

	return replaceNilInInterface(res), nil
}

func replaceNilInInterface(data interface{}) interface{} {
	switch v := data.(type) {
	case string:
		if v == stringNil {
			return nil
		}
		return v
	case []interface{}:
		// Process each element in the slice
		for i, elem := range v {
			v[i] = replaceNilInInterface(elem)
		}
		return v
	case map[string]interface{}:
		// Process each value in the map
		for key, value := range v {
			v[key] = replaceNilInInterface(value)
		}
		return v
	default:
		// For other types, return as is
		return data
	}
}

func RespTypeToValue(respType clientio.RespType) interface{} {
	var respArrString = map[clientio.RespType]string{
		clientio.NIL:           "(nil)",  // Represents a RESP Nil Bulk String, which indicates a null value.
		clientio.OK:            "OK",     // Represents a RESP Simple String with value "OK".
		clientio.CommandQueued: "QUEUED", // Represents a Simple String indicating that a command has been queued.
		clientio.EmptyArray:    "*0",     // Represents an empty RESP Array.
	}

	var respArrInt = map[clientio.RespType]float64{
		clientio.IntegerZero:        0,  // Represents a RESP Integer with value 0.
		clientio.IntegerOne:         1,  // Represents a RESP Integer with value 1.
		clientio.IntegerNegativeOne: -1, // Represents a RESP Integer with value -1.
		clientio.IntegerNegativeTwo: -2, // Represents a RESP Integer with value -2.
	}

	if val, exists := respArrString[respType]; exists {
		return val
	}
	// Check if respType exists in respArrInt map
	if val, exists := respArrInt[respType]; exists {
		return val
	}

	// Default to nil if respType is not recognized
	return nil
}

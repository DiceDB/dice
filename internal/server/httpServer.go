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

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/internal/clientio"
	"github.com/dicedb/dice/internal/cmd"
	"github.com/dicedb/dice/internal/comm"
	derrors "github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/ops"
	"github.com/dicedb/dice/internal/server/utils"
	"github.com/dicedb/dice/internal/shard"
)

const Abort = "ABORT"

var unimplementedCommands = map[string]bool{
	"QUNWATCH": true,
}

type HTTPServer struct {
	shardManager       *shard.ShardManager
	ioChan             chan *ops.StoreResponse
	httpServer         *http.Server
	logger             *slog.Logger
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

func NewHTTPServer(shardManager *shard.ShardManager, logger *slog.Logger) *HTTPServer {
	mux := http.NewServeMux()
	caseInsensitiveMux := &CaseInsensitiveMux{mux: mux}
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", config.HTTPPort),
		Handler:           caseInsensitiveMux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	httpServer := &HTTPServer{
		shardManager:       shardManager,
		ioChan:             make(chan *ops.StoreResponse, 1000),
		httpServer:         srv,
		logger:             logger,
		qwatchResponseChan: make(chan comm.QwatchResponse),
		shutdownChan:       make(chan struct{}),
	}

	mux.HandleFunc("/", httpServer.DiceHTTPHandler)
	mux.HandleFunc("/qwatch", httpServer.DiceHTTPQwatchHandler)
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
	var err error

	httpCtx, cancelHTTP := context.WithCancel(ctx)
	defer cancelHTTP()

	s.shardManager.RegisterWorker("httpServer", s.ioChan)

	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case <-ctx.Done():
		case <-s.shutdownChan:
			err = derrors.ErrAborted
			s.logger.Debug("Shutting down HTTP Server")
		}

		shutdownErr := s.httpServer.Shutdown(httpCtx)
		if shutdownErr != nil {
			s.logger.Error("HTTP Server Shutdown Failed", slog.Any("error", err))
			err = shutdownErr
			return
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.logger.Info("HTTP Server running", slog.String("addr", s.httpServer.Addr))
		err = s.httpServer.ListenAndServe()
	}()

	wg.Wait()
	return err
}

func (s *HTTPServer) DiceHTTPHandler(writer http.ResponseWriter, request *http.Request) {
	// convert to REDIS cmd
	diceDBCmd, err := utils.ParseHTTPRequest(request)
	if err != nil {
		http.Error(writer, "Error parsing HTTP request", http.StatusBadRequest)
		s.logger.Error("Error parsing HTTP request", slog.Any("error", err))
		return
	}

	if diceDBCmd.Cmd == Abort {
		s.logger.Debug("ABORT command received")
		s.logger.Debug("Shutting down HTTP Server")
		close(s.shutdownChan)
		return
	}

	if unimplementedCommands[diceDBCmd.Cmd] {
		http.Error(writer, "Command is not implemented with HTTP", http.StatusBadRequest)
		s.logger.Error("Command %s is not implemented", slog.String("cmd", diceDBCmd.Cmd))
		_, err := writer.Write([]byte("Command is not implemented with HTTP"))
		if err != nil {
			s.logger.Error("Error writing response", slog.Any("error", err))
			return
		}
		return
	}

	// send request to Shard Manager
	s.shardManager.GetShard(0).ReqChan <- &ops.StoreOp{
		Cmd:      diceDBCmd,
		WorkerID: "httpServer",
		ShardID:  0,
		HTTPOp:   true,
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
		s.logger.Error("Error parsing HTTP request", slog.Any("error", err))
		return
	}

	if len(diceDBCmd.Args) < 1 {
		s.logger.Error("Invalid request for QWATCH")
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
		Cmd:      diceDBCmd,
		WorkerID: "httpServer",
		ShardID:  0,
		Client:   qwatchClient,
		HTTPOp:   true,
	}

	s.logger.Info("Registered client for watching query", slog.Any("clientID", clientIdentifierID),
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
			s.logger.Info("Client disconnected")
			unWatchCmd := &cmd.DiceDBCmd{
				Cmd:  "QUNWATCH",
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
		s.logger.Error("Unsupported response type")
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
		s.logger.Error("Error decoding response: %v", slog.Any("error", err))
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
		s.logger.Error("Error marshaling QueryData to JSON: %v", slog.Any("error", err))
		http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Format the response as SSE event
	_, err = writer.Write(responseJSON)
	if err != nil {
		s.logger.Error("Error writing SSE data: %v", slog.Any("error", err))
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
	_, ok := WorkerCmdsMeta[diceDBCmd.Cmd]
	var rp *clientio.RESPParser

	var responseValue interface{}
	// TODO: Remove this conditional check and if (true) condition when all commands are migrated
	if !ok {
		var err error
		if result.EvalResponse.Error != nil {
			rp = clientio.NewRESPParser(bytes.NewBuffer([]byte(result.EvalResponse.Error.Error())))
		} else {
			rp = clientio.NewRESPParser(bytes.NewBuffer(result.EvalResponse.Result.([]byte)))
		}

		responseValue, err = rp.DecodeOne()
		if err != nil {
			s.logger.Error("Error decoding response", "error", err)
			http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	} else {
		if result.EvalResponse.Error != nil {
			responseValue = result.EvalResponse.Error.Error()
		} else {
			responseValue = result.EvalResponse.Result
		}
	}

	// func HandlePredefinedResponse(response interface{}) []byte {
	respArr := []string{
		"(nil)",  // Represents a RESP Nil Bulk String, which indicates a null value.
		"OK",     // Represents a RESP Simple String with value "OK".
		"QUEUED", // Represents a Simple String indicating that a command has been queued.
		"0",      // Represents a RESP Integer with value 0.
		"1",      // Represents a RESP Integer with value 1.
		"-1",     // Represents a RESP Integer with value -1.
		"-2",     // Represents a RESP Integer with value -2.
		"*0",     // Represents an empty RESP Array.
	}

	if val, ok := responseValue.(clientio.RespType); ok {
		responseValue = respArr[val]
	}

	if bt, ok := responseValue.([]byte); ok {
		responseValue = string(bt)
	}
	httpResponse := utils.HTTPResponse{Data: responseValue}

	responseJSON, err := json.Marshal(httpResponse)
	if err != nil {
		s.logger.Error("Error marshaling response", "error", err)
		http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(responseJSON)
	if err != nil {
		s.logger.Error("Error writing response", "error", err)
	}
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

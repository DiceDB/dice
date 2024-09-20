package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/dicedb/dice/internal/clientio"
	"github.com/dicedb/dice/internal/cmd"
	"github.com/dicedb/dice/internal/server/utils"

	"github.com/charmbracelet/log"
	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/internal/ops"
	"github.com/dicedb/dice/internal/querywatcher"
	"github.com/dicedb/dice/internal/shard"
	dstore "github.com/dicedb/dice/internal/store"
)

var unimplementedCommands = map[string]bool{
	"QUNWATCH":  true,
	"SUBSCRIBE": true,
	"ABORT":     true,
}

const QWATCH = "QWATCH"

type HTTPServer struct {
	queryWatcher *querywatcher.QueryManager
	shardManager *shard.ShardManager
	ioChan       chan *ops.StoreResponse
	watchChan    chan dstore.WatchEvent
	httpServer   *http.Server
}

type HTTPQwatchResponse struct {
	Cmd   string `json:"cmd"`
	Query string `json:"query"`
	Data  []any  `json:"data"`
}

type HTTPQwatchWriter struct {
	Writer http.ResponseWriter
	Query  string
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

func NewHTTPServer(shardManager *shard.ShardManager, watchChan chan dstore.WatchEvent) *HTTPServer {
	mux := http.NewServeMux()
	caseInsensitiveMux := &CaseInsensitiveMux{mux: mux}
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", config.HTTPPort),
		Handler:           caseInsensitiveMux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	httpServer := &HTTPServer{
		shardManager: shardManager,
		queryWatcher: querywatcher.NewQueryManager(),
		ioChan:       make(chan *ops.StoreResponse, 1000),
		watchChan:    watchChan,
		httpServer:   srv,
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
		<-ctx.Done()
		err = s.httpServer.Shutdown(httpCtx)
		// TODO: Check for clean connection close in case a QWATCH client still subscribed
		if err != nil && !errors.Is(err, context.Canceled) {
			log.Errorf("HTTP Server Shutdown Failed: %v", err)
			return
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Infof("HTTP Server running on Port%s", s.httpServer.Addr)
		err = s.httpServer.ListenAndServe()
	}()

	wg.Wait()
	return err
}

func (s *HTTPServer) DiceHTTPHandler(writer http.ResponseWriter, request *http.Request) {
	// convert to REDIS cmd
	redisCmd, err := utils.ParseHTTPRequest(request)
	if err != nil {
		log.Errorf("Error parsing HTTP request: %v", err)
		http.Error(writer, "Error parsing HTTP request", http.StatusBadRequest)
		return
	}

	if unimplementedCommands[redisCmd.Cmd] {
		log.Errorf("Command %s is not implemented", redisCmd.Cmd)
		http.Error(writer, "Command is not implemented with HTTP", http.StatusBadRequest)
		return
	}

	// send request to Shard Manager
	s.shardManager.GetShard(0).ReqChan <- &ops.StoreOp{
		Cmd:      redisCmd,
		WorkerID: "httpServer",
		ShardID:  0,
		HTTPOp:   true,
	}

	// Wait for response
	resp := <-s.ioChan
	s.writeResponse(writer, resp.Result)
}

func (s *HTTPServer) DiceHTTPQwatchHandler(writer http.ResponseWriter, request *http.Request) {
	// convert to REDIS cmd
	redisCmd, err := utils.ParseHTTPRequest(request)
	if err != nil {
		log.Errorf("Error parsing HTTP request: %v", err)
		http.Error(writer, "Error parsing HTTP request", http.StatusBadRequest)
		return
	}

	if len(redisCmd.Args) < 1 {
		log.Errorf("Invalid request for QWATCH")
		http.Error(writer, "Invalid request for QWATCH", http.StatusBadRequest)
		return
	}

	// Check if the connection supports flushing
	flusher, ok := writer.(http.Flusher)
	if !ok {
		http.Error(writer, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Set SSE headers
	writer.Header().Set("Content-Type", "text/event-stream")
	writer.Header().Set("Cache-Control", "no-cache")
	writer.Header().Set("Connection", "keep-alive")
	writer.WriteHeader(http.StatusOK)
	// We're a generating a unique client id, to keep track in core of requests from registered clients
	clientID := generateUniqueInt32(request)
	clientRequestID := make(map[uint32]HTTPQwatchWriter)
	var clientWriter HTTPQwatchWriter
	clientWriter.Writer = writer
	clientWriter.Query = redisCmd.Args[0]
	clientRequestID[clientID] = clientWriter

	// Prepare the store operation
	storeOp := &ops.StoreOp{
		Cmd:                  redisCmd,
		WorkerID:             "httpServer",
		ShardID:              0,
		HTTPOp:               true,
		HTTPClientRespWriter: clientWriter.Writer,
		RequestID:            clientID,
	}

	log.Infof("Registered client with id %d for watching query %s", clientID, clientWriter.Query)
	s.shardManager.GetShard(0).ReqChan <- storeOp

	// Wait for 1st sync response from server for QWATCH and flush it to client
	resp := <-s.ioChan
	s.writeResponse(writer, resp.Result)
	flusher.Flush()
	// Keep listening for context cancellation (client disconnect) and continuous responses
	doneChan := request.Context().Done()
	for {
		<-doneChan
		// Client disconnected or request finished
		log.Infof("Client disconnected")
		unWatchCmd := &cmd.RedisCmd{
			Cmd:  "QUNWATCH",
			Args: []string{clientWriter.Query},
		}
		storeOp.Cmd = unWatchCmd
		s.shardManager.GetShard(0).ReqChan <- storeOp
		resp := <-s.ioChan
		s.writeResponse(writer, resp.Result)
		return
	}
}

func (s *HTTPServer) writeResponse(writer http.ResponseWriter, result []byte) {
	rp := clientio.NewRESPParser(bytes.NewBuffer(result))
	val, err := rp.DecodeOne()
	if err != nil {
		log.Error("Error decoding response", "error", err)
		http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	responseJSON, err := json.Marshal(val)
	if err != nil {
		log.Error("Error marshaling response", "error", err)
		http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(responseJSON)
	if err != nil {
		log.Error("Error writing response", "error", err)
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

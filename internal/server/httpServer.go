package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/dicedb/dice/internal/cmd"
	"github.com/dicedb/dice/internal/id"

	"github.com/dicedb/dice/internal/clientio"
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

func NewHTTPServer(shardManager *shard.ShardManager, watchChan chan dstore.WatchEvent) *HTTPServer {
	mux := http.NewServeMux()
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", config.HTTPPort),
		Handler:           mux,
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
		if err != nil {
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
		return
	}

	if unimplementedCommands[redisCmd.Cmd] {
		log.Errorf("Command %s is not implemented", redisCmd.Cmd)
		_, err := writer.Write([]byte("Command is not implemented with HTTP"))
		if err != nil {
			log.Errorf("Error writing response: %v", err)
			return
		}
		return
	}

	if redisCmd.Cmd == QWATCH {
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
		return
	}

	if redisCmd.Cmd != QWATCH || len(redisCmd.Args) < 1 {
		http.Error(writer, "Invalid params", http.StatusBadRequest)
		return
	}

	// Set SSE headers
	writer.Header().Set("Content-Type", "text/event-stream")
	writer.Header().Set("Cache-Control", "no-cache")
	writer.Header().Set("Connection", "keep-alive")
	writer.WriteHeader(http.StatusOK)
	uniqueID := id.NextID()
	clientRequestID := make(map[uint32]HTTPQwatchWriter)
	var clientWriter HTTPQwatchWriter
	clientWriter.Writer = writer
	clientWriter.Query = redisCmd.Args[0]
	clientRequestID[uniqueID] = clientWriter

	// Prepare the store operation
	storeOp := &ops.StoreOp{
		Cmd:                  redisCmd,
		WorkerID:             "httpServer",
		ShardID:              0,
		HTTPOp:               true,
		HTTPClientRespWriter: clientWriter.Writer,
		RequestID:            uniqueID,
	}

	s.shardManager.GetShard(0).ReqChan <- storeOp

	// Wait for response
	resp := <-s.ioChan
	s.writeResponse(writer, resp.Result)
	flusher, ok := writer.(http.Flusher)
	if !ok {
		http.Error(writer, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	flusher.Flush() // Flush the response to send it to the client
	doneChan := request.Context().Done()
	for range doneChan {
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

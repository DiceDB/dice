package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/dicedb/dice/internal/clientio"
	"github.com/dicedb/dice/internal/cmd"
	"github.com/dicedb/dice/internal/id"
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

	if redisCmd.Cmd == "QWATCH" {
		s.handleQWATCH(writer, redisCmd, request)
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

	rp := clientio.NewRESPParser(bytes.NewBuffer(resp.Result))
	val, err := rp.DecodeOne()
	if err != nil {
		log.Errorf("Error decoding response: %v", err)
		return
	}

	// Write response
	responseJSON, err := json.Marshal(val)
	if err != nil {
		log.Errorf("Error marshaling response: %v", err)
		return
	}
	_, err = writer.Write(responseJSON)
	if err != nil {
		log.Errorf("Error writing response: %v", err)
		return
	}
}

func (s *HTTPServer) handleQWATCH(writer http.ResponseWriter, redisCmd *cmd.RedisCmd, request *http.Request) {
	// Set SSE headers
	writer.Header().Set("Content-Type", "text/event-stream")
	writer.Header().Set("Cache-Control", "no-cache")
	writer.Header().Set("Connection", "keep-alive")
	writer.WriteHeader(http.StatusOK)
	flusher, ok := writer.(http.Flusher)
	if !ok {
		http.Error(writer, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Create a dedicated response channel for this request
	qwatchResponseChan := make(chan *ops.StoreResponse)
	uniqueID := id.NextID()
	clientRequestID := make(map[http.ResponseWriter]uint32)
	clientRequestID[writer] = uniqueID

	// Currently reusing the qwatchResponseChan
	s.shardManager.GetShard(0).ReqChan <- &ops.StoreOp{
		Cmd:              redisCmd,
		WorkerID:         "httpServer",
		ShardID:          0,
		HTTPOp:           true,
		HTTPResponseChan: qwatchResponseChan,
		RequestID:        uniqueID,
	}

	for {
		select {
		case resp := <-qwatchResponseChan:
			if resp == nil || resp.Result == nil {
				log.Errorf("Error from shard")
				http.Error(writer, "Error processing request", http.StatusInternalServerError)
				return
			}

			rp := clientio.NewRESPParser(bytes.NewBuffer(resp.Result))
			val, err := rp.DecodeOne()
			if err != nil {
				log.Errorf("Error decoding response: %v", err)
				return
			}

			var responseJSON []byte
			// Convert the decoded response to the HTTPQwatchResponse struct
			// Else just encode and send
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
				log.Errorf("Error marshaling QueryData to JSON: %v", err)
				return
			}

			// Format the response as SSE event
			_, err = writer.Write(responseJSON)
			if err != nil {
				log.Errorf("Error writing SSE data: %v", err)
				return
			}
			flusher.Flush() // Flush the response to send it to the client

		case <-request.Context().Done():
			// Client disconnected or request finished
			// TODO: We need to create a way to remove watcher once client disconnected
			log.Infof("Client disconnected")
			return
		}
	}
}

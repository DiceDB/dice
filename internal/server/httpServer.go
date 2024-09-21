package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/internal/clientio"
	"github.com/dicedb/dice/internal/ops"
	"github.com/dicedb/dice/internal/querywatcher"
	"github.com/dicedb/dice/internal/server/utils"
	"github.com/dicedb/dice/internal/shard"
	dstore "github.com/dicedb/dice/internal/store"
)

var unimplementedCommands map[string]bool = map[string]bool{
	"QWATCH":    true,
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
	logger       *slog.Logger
}

func NewHTTPServer(
	shardManager *shard.ShardManager,
	watchChan chan dstore.WatchEvent,
	logger *slog.Logger,
) *HTTPServer {
	mux := http.NewServeMux()
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", config.HTTPPort),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	httpServer := &HTTPServer{
		shardManager: shardManager,
		queryWatcher: querywatcher.NewQueryManager(logger),
		ioChan:       make(chan *ops.StoreResponse, 1000),
		watchChan:    watchChan,
		httpServer:   srv,
		logger:       logger,
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
			s.logger.Error("HTTP Server Shutdown Failed", slog.Any("error", err))
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
	redisCmd, err := utils.ParseHTTPRequest(request)
	if err != nil {
		s.logger.Error("Error parsing HTTP request", slog.Any("error", err))
		return
	}

	if unimplementedCommands[redisCmd.Cmd] {
		s.logger.Error("Command %s is not implemented", slog.String("cmd", redisCmd.Cmd))
		_, err := writer.Write([]byte("Command is not implemented with HTTP"))
		if err != nil {
			s.logger.Error("Error writing response", slog.Any("error", err))
			return
		}
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
		s.logger.Error("Error decoding response", slog.Any("error", err))
		return
	}

	// Write response
	responseJSON, err := json.Marshal(val)
	if err != nil {
		s.logger.Error("Error marshaling response", slog.Any("error", err))
		return
	}
	_, err = writer.Write(responseJSON)
	if err != nil {
		s.logger.Error("Error writing response", slog.Any("error", err))
		return
	}
}

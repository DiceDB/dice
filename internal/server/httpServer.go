package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/charmbracelet/log"
	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/internal/clientio"
	"github.com/dicedb/dice/internal/ops"
	"github.com/dicedb/dice/internal/querywatcher"
	"github.com/dicedb/dice/internal/server/utils"
	"github.com/dicedb/dice/internal/shard"
	dstore "github.com/dicedb/dice/internal/store"
	"net/http"
	"sync"
)

type HttpServer struct {
	queryWatcher *querywatcher.QueryManager
	shardManager *shard.ShardManager
	ioChan       chan *ops.StoreResponse
	watchChan    chan dstore.WatchEvent
	httpServer   *http.Server
}

// NewHttpServer
// TODO: This isn't ideal as we create a separate instance of ShardManager, which we want to be 1, which all servers share
func NewHttpServer(shardManager *shard.ShardManager, watchChan chan dstore.WatchEvent) *HttpServer {
	mux := http.NewServeMux()
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.HTTPPort),
		Handler: mux,
	}

	httpServer := &HttpServer{
		shardManager: shardManager,
		queryWatcher: querywatcher.NewQueryManager(),
		ioChan:       make(chan *ops.StoreResponse, 1000),
		watchChan:    watchChan,
		httpServer:   srv,
	}

	mux.HandleFunc("/", httpServer.DiceHttpHandler)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("ok"))
		if err != nil {
			return
		}
	})

	return httpServer
}

func (s *HttpServer) Run(ctx context.Context) error {
	var wg sync.WaitGroup
	var err error

	_, cancelHttp := context.WithCancel(ctx)
	defer cancelHttp()

	s.shardManager.RegisterWorker("httpServer", s.ioChan)

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		err = s.httpServer.Shutdown(ctx)
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

func (s *HttpServer) DiceHttpHandler(writer http.ResponseWriter, request *http.Request) {
	// convert to REDIS cmd
	redisCmd, err := utils.ParseHTTPRequest(request)
	if err != nil {
		log.Errorf("Error parsing HTTP request: %v", err)
		return
	}

	// send request to Shard Manager
	s.shardManager.GetShard(0).ReqChan <- &ops.StoreOp{
		Cmd:      redisCmd,
		WorkerID: "httpServer",
		ShardID:  0,
		HttpOp:   true,
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
	responseJson, err := json.Marshal(val)
	if err != nil {
		log.Errorf("Error marshalling response: %v", err)
		return
	}
	_, err = writer.Write(responseJson)
	if err != nil {
		log.Errorf("Error writing response: %v", err)
		return
	}
}

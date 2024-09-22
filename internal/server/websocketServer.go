package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/internal/clientio"
	"github.com/dicedb/dice/internal/ops"
	"github.com/dicedb/dice/internal/querywatcher"
	"github.com/dicedb/dice/internal/server/utils"
	"github.com/dicedb/dice/internal/shard"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/gorilla/websocket"
)

type WebsocketServer struct {
	querywatcher    *querywatcher.QueryManager
	shardManager    *shard.ShardManager
	ioChan          chan *ops.StoreResponse
	watchChan       chan dstore.WatchEvent
	websocketServer *http.Server
	upgrader        websocket.Upgrader
}

var unimplementedCommandsWebsocket map[string]bool = map[string]bool{
	"QWATCH":    true,
	"QUNWATCH":  true,
	"SUBSCRIBE": true,
	"ABORT":     true,
}

func NewWebSocketServer(shardManager *shard.ShardManager, watchChan chan dstore.WatchEvent) *WebsocketServer {
	mux := http.NewServeMux()
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", config.WebsocketPort),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	websocketServer := &WebsocketServer{
		shardManager:    shardManager,
		querywatcher:    querywatcher.NewQueryManager(),
		ioChan:          make(chan *ops.StoreResponse, 1000),
		watchChan:       watchChan,
		websocketServer: srv,
		upgrader:        upgrader,
	}

	mux.HandleFunc("/", websocketServer.WebsocketHandler)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("ok"))
		if err != nil {
			return
		}
	})
	return websocketServer
}

func (s *WebsocketServer) Run(ctx context.Context) error {
	var wg sync.WaitGroup
	var err error

	websocketCtx, cancelWebsocket := context.WithCancel(ctx)
	defer cancelWebsocket()

	s.shardManager.RegisterWorker("wsServer", s.ioChan)

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		err = s.websocketServer.Shutdown(websocketCtx)
		if err != nil {
			log.Errorf("Websocket Server shutdown failed: %v", err)
			return
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Infof("Websocket Server running on port %v", s.websocketServer.Addr[1:])
		err = s.websocketServer.ListenAndServe()
	}()

	wg.Wait()
	return err
}

func (s *WebsocketServer) WebsocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("Websocket upgrade failed: %v", err)
	}
	defer conn.Close()

	for {
		// read incoming message
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Errorf("Websocket read failed: %v", err)
			break
		}
		fmt.Println(msg)

		// parse message to dice command
		redisCmd, err := utils.ParseWebsocketMessage(msg)
		if err != nil {
			log.Errorf("Error parsing Websocket request: %v", err)
		}

		if unimplementedCommandsWebsocket[redisCmd.Cmd] {
			log.Errorf("Command %s is not implemented", redisCmd.Cmd)
			_, err := w.Write([]byte("Command is not implemented with Websocket"))
			if err != nil {
				log.Errorf("Error writing response: %v", err)
				return
			}
			return
		}

		// send request to Shard Manager
		s.shardManager.GetShard(0).ReqChan <- &ops.StoreOp{
			Cmd:         redisCmd,
			WorkerID:    "wsServer",
			ShardID:     0,
			WebsocketOp: true,
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
		err = conn.WriteMessage(websocket.TextMessage, responseJSON)
		if err != nil {
			log.Errorf("Error writing response: %v", err)
			return
		}
	}
}

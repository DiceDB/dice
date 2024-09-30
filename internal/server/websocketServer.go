package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/internal/clientio"
	diceerrors "github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/ops"
	"github.com/dicedb/dice/internal/querymanager"
	"github.com/dicedb/dice/internal/server/utils"
	"github.com/dicedb/dice/internal/shard"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/gorilla/websocket"
)

const Qwatch = "QWATCH"
const Qunwatch = "QUNWATCH"
const Subscribe = "SUBSCRIBE"

var unimplementedCommandsWebsocket = map[string]bool{
	Qwatch:    true,
	Qunwatch:  true,
	Subscribe: true,
}

type WebsocketServer struct {
	querymanager    *querymanager.Manager
	shardManager    *shard.ShardManager
	ioChan          chan *ops.StoreResponse
	watchChan       chan dstore.QueryWatchEvent
	websocketServer *http.Server
	upgrader        websocket.Upgrader
	logger          *slog.Logger
	shutdownChan    chan struct{}
}

func NewWebSocketServer(shardManager *shard.ShardManager, watchChan chan dstore.QueryWatchEvent, logger *slog.Logger) *WebsocketServer {
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
		querymanager:    querymanager.NewQueryManager(logger),
		ioChan:          make(chan *ops.StoreResponse, 1000),
		watchChan:       watchChan,
		websocketServer: srv,
		upgrader:        upgrader,
		logger:          logger,
		shutdownChan:    make(chan struct{}),
	}

	mux.HandleFunc("/", websocketServer.WebsocketHandler)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("OK"))
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
		select {
		case <-ctx.Done():
		case <-s.shutdownChan:
			err = diceerrors.ErrAborted
			s.logger.Debug("Shutting down Websocket Server")
		}

		shutdownErr := s.websocketServer.Shutdown(websocketCtx)
		if shutdownErr != nil {
			s.logger.Error("Websocket Server shutdown failed:", slog.Any("error", err))
			return
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.logger.Info("Websocket Server running", slog.String("port", s.websocketServer.Addr[1:]))
		err = s.websocketServer.ListenAndServe()
	}()

	wg.Wait()
	return err
}

func (s *WebsocketServer) WebsocketHandler(w http.ResponseWriter, r *http.Request) {
	// upgrade http connection to websocket
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	// closing handshake
	defer func() {
		_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "close 1000 (normal)"))
		conn.Close()
	}()

	for {
		// read incoming message
		_, msg, err := conn.ReadMessage()
		if err != nil {
			writeResponse(conn, []byte("error: command reading failed"))
			continue
		}

		// parse message to dice command
		redisCmd, err := utils.ParseWebsocketMessage(msg)
		if errors.Is(err, diceerrors.ErrEmptyCommand) {
			continue
		} else if err != nil {
			writeResponse(conn, []byte("error: parsing failed"))
			continue
		}

		if redisCmd.Cmd == Abort {
			close(s.shutdownChan)
			break
		}

		if unimplementedCommandsWebsocket[redisCmd.Cmd] {
			writeResponse(conn, []byte("Command is not implemented with Websocket"))
			continue
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
		var rp *clientio.RESPParser
		if resp.EvalResponse.Error != nil {
			rp = clientio.NewRESPParser(bytes.NewBuffer([]byte(resp.EvalResponse.Error.Error())))
		} else {
			rp = clientio.NewRESPParser(bytes.NewBuffer(resp.EvalResponse.Result.([]byte)))
		}

		val, err := rp.DecodeOne()
		if err != nil {
			writeResponse(conn, []byte("error: decoding response"))
			continue
		}

		respBytes, err := json.Marshal(val)
		if err != nil {
			writeResponse(conn, []byte("error: marshaling json response"))
			continue
		}

		// Write response
		writeResponse(conn, respBytes)
	}
}

func writeResponse(conn *websocket.Conn, text []byte) {
	_ = conn.WriteMessage(websocket.TextMessage, text)
}

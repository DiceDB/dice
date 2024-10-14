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
	"github.com/dicedb/dice/internal/cmd"
	"github.com/dicedb/dice/internal/comm"
	diceerrors "github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/ops"
	"github.com/dicedb/dice/internal/server/utils"
	"github.com/dicedb/dice/internal/shard"
	"github.com/gorilla/websocket"
)

const QWatch = "QWATCH"
const Subscribe = "SUBSCRIBE"
const Qunwatch = "QUNWATCH"

var unimplementedCommandsWebsocket = map[string]bool{
	Qunwatch: true,
}

type WebsocketServer struct {
	shardManager       *shard.ShardManager
	ioChan             chan *ops.StoreResponse
	websocketServer    *http.Server
	upgrader           websocket.Upgrader
	qwatchResponseChan chan comm.QwatchResponse
	shutdownChan       chan struct{}
	logger             *slog.Logger
}

func NewWebSocketServer(shardManager *shard.ShardManager, logger *slog.Logger) *WebsocketServer {
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
		shardManager:       shardManager,
		ioChan:             make(chan *ops.StoreResponse, 1000),
		websocketServer:    srv,
		upgrader:           upgrader,
		qwatchResponseChan: make(chan comm.QwatchResponse),
		shutdownChan:       make(chan struct{}),
		logger:             logger,
	}

	mux.HandleFunc("/", websocketServer.WebsocketHandler)
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
		diceDBCmd, err := utils.ParseWebsocketMessage(msg)
		if errors.Is(err, diceerrors.ErrEmptyCommand) {
			continue
		} else if err != nil {
			writeResponse(conn, []byte("error: parsing failed"))
			continue
		}

		if diceDBCmd.Cmd == Abort {
			close(s.shutdownChan)
			break
		}

		if unimplementedCommandsWebsocket[diceDBCmd.Cmd] {
			writeResponse(conn, []byte("Command is not implemented with Websocket"))
			continue
		}

		// create request
		sp := &ops.StoreOp{
			Cmd:         diceDBCmd,
			WorkerID:    "wsServer",
			ShardID:     0,
			WebsocketOp: true,
		}

		// handle qwatch and qunwatch commands
		if diceDBCmd.Cmd == QWatch || diceDBCmd.Cmd == Subscribe {
			clientIdentifierID := generateUniqueInt32(r)
			sp.Client = comm.NewHTTPQwatchClient(s.qwatchResponseChan, clientIdentifierID)

			// start a goroutine for subsequent updates
			go s.processQwatchUpdates(clientIdentifierID, conn, diceDBCmd)
		}

		s.shardManager.GetShard(0).ReqChan <- sp
		resp := <-s.ioChan
		s.processResponse(conn, diceDBCmd, resp)
	}
}

func (s *WebsocketServer) processQwatchUpdates(clientIdentifierID uint32, conn *websocket.Conn, dicDBCmd *cmd.DiceDBCmd) {
	for {
		select {
		case resp := <-s.qwatchResponseChan:
			if resp.ClientIdentifierID == clientIdentifierID {
				s.processResponse(conn, dicDBCmd, resp)
			}
		case <-s.shutdownChan:
			return
		}
	}
}

func (s *WebsocketServer) processResponse(conn *websocket.Conn, diceDBCmd *cmd.DiceDBCmd, response interface{}) {
	var result interface{}
	var err error

	// check response type
	switch resp := response.(type) {
	case comm.QwatchResponse:
		result = resp.Result
		err = resp.Error
	case *ops.StoreResponse:
		result = resp.EvalResponse.Result
		err = resp.EvalResponse.Error
	default:
		s.logger.Error("Unsupported response type")
		writeResponse(conn, []byte("error: 500 Internal Server Error"))
		return
	}

	_, ok := WorkerCmdsMeta[diceDBCmd.Cmd]
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

	var responseValue interface{}
	// TODO: Remove this conditional check and if (true) condition when all commands are migrated
	if !ok {
		var rp *clientio.RESPParser
		if err != nil {
			rp = clientio.NewRESPParser(bytes.NewBuffer([]byte(err.Error())))
		} else {
			rp = clientio.NewRESPParser(bytes.NewBuffer(result.([]byte)))
		}

		responseValue, err = rp.DecodeOne()
		if err != nil {
			s.logger.Error("Error decoding response", "error", err)
			writeResponse(conn, []byte("error: 500 Internal Server Error"))
			return
		}
	} else {
		if err != nil {
			responseValue = err.Error()
		} else {
			responseValue = result
		}
	}

	if val, ok := responseValue.(clientio.RespType); ok {
		responseValue = respArr[val]
	}

	if bt, ok := responseValue.([]byte); ok {
		responseValue = string(bt)
	}

	respBytes, err := json.Marshal(responseValue)
	if err != nil {
		s.logger.Error("Error marshaling json", "error", err)
		writeResponse(conn, []byte("error: marshaling json"))
		return
	}

	// success
	writeResponse(conn, respBytes)
}

func writeResponse(conn *websocket.Conn, text []byte) {
	_ = conn.WriteMessage(websocket.TextMessage, text)
}

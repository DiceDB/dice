package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"sync"
	"syscall"
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
	"golang.org/x/exp/rand"
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

		// send request to Shard Manager
		s.shardManager.GetShard(0).ReqChan <- &ops.StoreOp{
			Cmd:         diceDBCmd,
			WorkerID:    "wsServer",
			ShardID:     0,
			WebsocketOp: true,
		}

		// Wait for response
		resp := <-s.ioChan

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
		var rp *clientio.RESPParser

		var responseValue interface{}
		// TODO: Remove this conditional check and if (true) condition when all commands are migrated
		if !ok {
			var err error
			if resp.EvalResponse.Error != nil {
				rp = clientio.NewRESPParser(bytes.NewBuffer([]byte(resp.EvalResponse.Error.Error())))
			} else {
				rp = clientio.NewRESPParser(bytes.NewBuffer(resp.EvalResponse.Result.([]byte)))
			}

			responseValue, err = rp.DecodeOne()
			if err != nil {
				s.logger.Error("Error decoding response", "error", err)
				writeResponse(conn, []byte("error: Internal Server Error"))
				return
			}
		} else {
			if resp.EvalResponse.Error != nil {
				responseValue = resp.EvalResponse.Error.Error()
			} else {
				responseValue = resp.EvalResponse.Result
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
			writeResponse(conn, []byte("error: marshaling json response"))
			continue
		}

		// Write response with retries for transient errors
		if err := WriteResponseWithRetries(conn, respBytes, config.DiceConfig.WebSocket.MaxWriteResponseRetries); err != nil {
			s.logger.Error(fmt.Sprintf("Error reading message: %v", err))
			break // Exit the loop on write error
		}
	}
}

func WriteResponseWithRetries(conn *websocket.Conn, text []byte, maxRetries int) error {
	for attempts := 0; attempts < maxRetries; attempts++ {
		// Set a write deadline
		if err := conn.SetWriteDeadline(time.Now().Add(config.DiceConfig.WebSocket.WriteResponseTimeout)); err != nil {
			slog.Error(fmt.Sprintf("Error setting write deadline: %v", err))
			return err
		}

		// Attempt to write message
		err := conn.WriteMessage(websocket.TextMessage, text)
		if err == nil {
			break // Exit loop if write succeeds
		}

		// Handle network errors
		var netErr *net.OpError
		if !errors.As(err, &netErr) {
			return fmt.Errorf("error writing message: %w", err)
		}

		opErr, ok := netErr.Err.(*os.SyscallError)
		if !ok {
			return fmt.Errorf("network operation error: %w", err)
		}

		if opErr.Syscall != "write" {
			return fmt.Errorf("unexpected syscall error: %w", err)
		}

		switch opErr.Err {
		case syscall.EPIPE:
			return fmt.Errorf("broken pipe: %w", err)
		case syscall.ECONNRESET:
			return fmt.Errorf("connection reset by peer: %w", err)
		case syscall.ENOBUFS:
			return fmt.Errorf("no buffer space available: %w", err)
		case syscall.EAGAIN:
			// Exponential backoff with jitter
			backoffDuration := time.Duration(attempts+1)*100*time.Millisecond + time.Duration(rand.Intn(50))*time.Millisecond 

			slog.Warn(fmt.Sprintf(
				"Temporary issue (EAGAIN) on attempt %d. Retrying in %v...", 
				attempts+1, backoffDuration,
			))

			time.Sleep(backoffDuration)
			continue
		default:
			return fmt.Errorf("write error: %w", err)
		}
	}

	return nil
}

func writeResponse(conn *websocket.Conn, text []byte) {
	// Set a write deadline to prevent hanging
	if err := conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
		slog.Error(fmt.Sprintf("Error setting write deadline: %v", err))
		return
	}

	err := conn.WriteMessage(websocket.TextMessage, text)
	if err != nil {
		slog.Error(fmt.Sprintf("Error writing response: %v", err))
	}
}

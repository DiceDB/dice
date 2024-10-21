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
	"github.com/dicedb/dice/internal/cmd"
	"github.com/dicedb/dice/internal/comm"
	diceerrors "github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/ops"
	"github.com/dicedb/dice/internal/server/utils"
	"github.com/dicedb/dice/internal/shard"
	"github.com/gorilla/websocket"
	"golang.org/x/exp/rand"
)

const Qwatch = "Q.WATCH"
const Qunwatch = "Q.UNWATCH"
const Subscribe = "SUBSCRIBE"

var unimplementedCommandsWebsocket = map[string]bool{
	Qunwatch: true,
}

type QuerySubscription struct {
	Subscribe          bool // true for subscribe, false for unsubscribe
	Cmd                *cmd.DiceDBCmd
	ClientIdentifierID uint32
	Client             *websocket.Conn
}

type WebsocketServer struct {
	shardManager       *shard.ShardManager
	ioChan             chan *ops.StoreResponse
	websocketServer    *http.Server
	upgrader           websocket.Upgrader
	subscriptionChan   chan QuerySubscription // to subscribe clients
	subscribedClients  sync.Map               // to maintain records of subscribed clients
	qwatchResponseChan chan comm.QwatchResponse
	shutdownChan       chan struct{}
	mu                 sync.Mutex
	logger             *slog.Logger
}

func NewWebSocketServer(shardManager *shard.ShardManager, port int, logger *slog.Logger) *WebsocketServer {
	mux := http.NewServeMux()
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
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
		subscriptionChan:   make(chan QuerySubscription),
		subscribedClients:  sync.Map{},
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

	wsCtx, cancelWS := context.WithCancel(ctx)
	defer cancelWS() // TODO move cancel below wg.WAIT()

	s.shardManager.RegisterWorker("wsServer", s.ioChan, nil)

	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case <-ctx.Done():
		case <-s.shutdownChan:
			err = diceerrors.ErrAborted
			s.logger.Debug("Shutting down Websocket Server", slog.Any("time", time.Now()))
		}

		shutdownErr := s.websocketServer.Shutdown(wsCtx)
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
		if err != nil {
			s.logger.Debug("Error in Websocket Server", slog.Any("time", time.Now()), slog.Any("error", err))
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.listenForSubscriptions(wsCtx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.processQwatchUpdates(wsCtx)
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
		s.mu.Lock()
		defer s.mu.Unlock()
		_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "close 1000 (normal)"))
		conn.Close()
	}()

	maxRetries := config.DiceConfig.WebSocket.MaxWriteResponseRetries
	for {
		// read incoming message
		_, msg, err := conn.ReadMessage()
		if err != nil {
			// acceptable close errors
			errs := []int{websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseAbnormalClosure}
			if !websocket.IsCloseError(err, errs...) {
				s.logger.Warn("failed to read message from client", slog.Any("error", err))
			}
			break
		}

		// parse message to dice command
		diceDBCmd, err := utils.ParseWebsocketMessage(msg)
		if errors.Is(err, diceerrors.ErrEmptyCommand) {
			continue
		} else if err != nil {
			if err := s.writeResponseWithRetries(conn, []byte("error: parsing failed"), maxRetries); err != nil {
				s.logger.Debug(fmt.Sprintf("Error writing message: %v", err))
			}
			continue
		}

		// TODO - on abort, close client connection instead of closing server?
		if diceDBCmd.Cmd == Abort {
			close(s.shutdownChan)
			break
		}

		if unimplementedCommandsWebsocket[diceDBCmd.Cmd] {
			if err := s.writeResponseWithRetries(conn, []byte("Command is not implemented with Websocket"), maxRetries); err != nil {
				s.logger.Debug(fmt.Sprintf("Error writing message: %v", err))
			}
			continue
		}

		// create request
		sp := &ops.StoreOp{
			Cmd:         diceDBCmd,
			WorkerID:    "wsServer",
			ShardID:     0,
			WebsocketOp: true,
		}

		// handle q.watch commands
		if diceDBCmd.Cmd == Qwatch || diceDBCmd.Cmd == Subscribe {
			clientIdentifierID := generateUniqueInt32(r)
			sp.Client = comm.NewHTTPQwatchClient(s.qwatchResponseChan, clientIdentifierID)

			// start a goroutine for subsequent updates
			// go s.processQwatchUpdates(clientIdentifierID, conn, diceDBCmd)

			// subscribe client for updates
			event := QuerySubscription{
				Subscribe:          true,
				Cmd:                diceDBCmd,
				ClientIdentifierID: clientIdentifierID,
				Client:             conn,
			}
			s.subscriptionChan <- event
		}

		s.shardManager.GetShard(0).ReqChan <- sp
		resp := <-s.ioChan
		if err := s.processResponse(conn, diceDBCmd, resp); err != nil {
			break
		}
	}
}

func (s *WebsocketServer) listenForSubscriptions(ctx context.Context) {
	for {
		select {
		case event := <-s.subscriptionChan:
			if event.Subscribe {
				s.subscribedClients.LoadOrStore(event.ClientIdentifierID, event.Client)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *WebsocketServer) processQwatchUpdates(ctx context.Context) {
	for {
		select {
		case resp := <-s.qwatchResponseChan:
			client, ok := s.subscribedClients.Load(resp.ClientIdentifierID)
			if !ok {
				s.logger.Error("message received but client not found, clientIdentifierID: %v", resp.ClientIdentifierID)
			}
			conn, ok := client.(*websocket.Conn)
			if !ok {
				s.logger.Error("error typecasting client to *websocket.Conn")
			}

			dicDBCmd := &cmd.DiceDBCmd{
				Cmd:  Qwatch,
				Args: []string{},
			}

			if err := s.processResponse(conn, dicDBCmd, resp); err != nil {
				s.logger.Debug("Error writing qwatch update to client", slog.Any("clientIdentifierID", resp.ClientIdentifierID), slog.Any("error", err))
				continue
			}
		case <-s.shutdownChan:
			return
		case <-ctx.Done():
			return
		}
	}
}

func (s *WebsocketServer) processResponse(conn *websocket.Conn, diceDBCmd *cmd.DiceDBCmd, response interface{}) error {
	var result interface{}
	var err error
	maxRetries := config.DiceConfig.WebSocket.MaxWriteResponseRetries

	// check response type
	switch resp := response.(type) {
	case comm.QwatchResponse:
		result = resp.Result
		err = resp.Error
	case *ops.StoreResponse:
		result = resp.EvalResponse.Result
		err = resp.EvalResponse.Error
	default:
		s.logger.Debug("Unsupported response type")
		if err := s.writeResponseWithRetries(conn, []byte("error: 500 Internal Server Error"), maxRetries); err != nil {
			s.logger.Debug(fmt.Sprintf("Error writing message: %v", err))
			return fmt.Errorf("error writing response: %v", err)
		}
		return nil
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
			s.logger.Debug("Error decoding response", "error", err)
			if err := s.writeResponseWithRetries(conn, []byte("error: 500 Internal Server Error"), maxRetries); err != nil {
				s.logger.Debug(fmt.Sprintf("Error writing message: %v", err))
				return fmt.Errorf("error writing response: %v", err)
			}
			return nil
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
		s.logger.Debug("Error marshaling json", "error", err)
		if err := s.writeResponseWithRetries(conn, []byte("error: marshaling json"), maxRetries); err != nil {
			s.logger.Debug(fmt.Sprintf("Error writing message: %v", err))
			return fmt.Errorf("error writing response: %v", err)
		}
		return nil
	}

	// success
	if err := s.writeResponseWithRetries(conn, respBytes, config.DiceConfig.WebSocket.MaxWriteResponseRetries); err != nil {
		s.logger.Debug(fmt.Sprintf("Error writing message: %v", err))
		return fmt.Errorf("error writing response: %v", err)
	}

	return nil
}

func (s *WebsocketServer) writeResponseWithRetries(conn *websocket.Conn, text []byte, maxRetries int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return WriteResponseWithRetries(conn, text, maxRetries)
}

// WriteResponseWithRetries wrties response with retries for transient errors
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

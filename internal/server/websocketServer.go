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

	"github.com/dicedb/dice/internal/server/abstractserver"
	"github.com/dicedb/dice/internal/wal"

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
	abstractserver.AbstractServer
	shardManager       *shard.ShardManager
	ioChan             chan *ops.StoreResponse
	websocketServer    *http.Server
	upgrader           websocket.Upgrader
	subscriptionChan   chan QuerySubscription // to subscribe clients
	subscribedClients  sync.Map               // to maintain records of subscribed clients
	qwatchResponseChan chan comm.QwatchResponse
	shutdownChan       chan struct{}
	mu                 sync.Mutex
}

func NewWebSocketServer(shardManager *shard.ShardManager, port int, wl wal.AbstractWAL) *WebsocketServer {
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
	}

	mux.HandleFunc("/", websocketServer.WebsocketHandler)
	return websocketServer
}

func (s *WebsocketServer) Run(ctx context.Context) error {
	var wg sync.WaitGroup
	var err error

	wsCtx, cancelWS := context.WithCancel(ctx)
	defer cancelWS()

	s.shardManager.RegisterIOThread("wsServer", s.ioChan, nil)

	// start server
	wg.Add(1)
	go func() {
		defer wg.Done()
		slog.Info("also listening WebSocket on", slog.String("port", s.websocketServer.Addr[1:]))
		err = s.websocketServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("error while listening on WebSocket", slog.Any("error", err))
		}
	}()

	// shutdown server gracefully
	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case <-ctx.Done():
		case <-s.shutdownChan:
			err = diceerrors.ErrAborted
			slog.Debug("Shutting down Websocket Server", slog.Any("time", time.Now()))
		}

		shutdownErr := s.websocketServer.Shutdown(wsCtx)
		if shutdownErr != nil {
			slog.Error("Websocket Server shutdown failed:", slog.Any("error", shutdownErr))
			return
		}
	}()

	// process Q.WATCH updates
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
		closeErr := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "close 1000 (normal)"))
		if closeErr != nil {
			slog.Debug("Error during closing handshake", slog.Any("error", closeErr))
		}
		conn.Close()
	}()

	maxRetries := config.DiceConfig.WebSocket.MaxWriteResponseRetries
	for {
		// read incoming message
		_, msg, err := conn.ReadMessage()
		if err != nil {
			// acceptable close errors
			errs := []int{websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseAbnormalClosure}
			if websocket.IsCloseError(err, errs...) {
				break
			}
			slog.Error("Error reading message", slog.Any("error", err))
			break
		}

		// parse message to dice command
		diceDBCmd, err := utils.ParseWebsocketMessage(msg)
		if errors.Is(err, diceerrors.ErrEmptyCommand) {
			continue
		} else if err != nil {
			if err := s.writeResponseWithRetries(conn, []byte("error: parsing failed"), maxRetries); err != nil {
				slog.Debug(fmt.Sprintf("Error writing message: %v", err))
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
				slog.Debug(fmt.Sprintf("Error writing message: %v", err))
			}
			continue
		}

		// create request
		sp := &ops.StoreOp{
			Cmd:         diceDBCmd,
			IOThreadID:  "wsServer",
			ShardID:     0,
			WebsocketOp: true,
		}

		// handle q.watch commands
		if diceDBCmd.Cmd == Qwatch || diceDBCmd.Cmd == Subscribe {
			clientIdentifierID := generateUniqueInt32(r)
			sp.Client = comm.NewHTTPQwatchClient(s.qwatchResponseChan, clientIdentifierID)

			// subscribe client for updates if watch is enabled
			if config.DiceConfig.Performance.EnableWatch {
				s.subscribedClients.LoadOrStore(clientIdentifierID, conn)
			} else {
				if err := s.writeResponseWithRetries(conn, []byte("error: watch is not enabled"), maxRetries); err != nil {
					slog.Debug(fmt.Sprintf("Error writing message: %v", err))
				}
				return
			}
		}

		s.shardManager.GetShard(0).ReqChan <- sp
		resp := <-s.ioChan
		if err := s.processResponse(conn, diceDBCmd, resp); err != nil {
			break
		}
	}
}

func (s *WebsocketServer) processQwatchUpdates(ctx context.Context) {
	for {
		select {
		case resp := <-s.qwatchResponseChan:
			client, ok := s.subscribedClients.Load(resp.ClientIdentifierID)
			if !ok {
				slog.Debug("message received but client not found", slog.Any("clientIdentifierID", resp.ClientIdentifierID))
				continue
			}
			conn, ok := client.(*websocket.Conn)
			if !ok {
				slog.Debug("error typecasting client to *websocket.Conn")
				continue
			}

			if err := s.processQwatchResponse(conn, resp); err != nil {
				slog.Debug("Error writing response to client", slog.Any("clientIdentifierID", resp.ClientIdentifierID), slog.Any("error", err))
				continue
			}
		case <-s.shutdownChan:
			return
		case <-ctx.Done():
			return
		}
	}
}

func (s *WebsocketServer) processQwatchResponse(conn *websocket.Conn, response interface{}) error {
	var result interface{}
	var err error
	maxRetries := config.DiceConfig.WebSocket.MaxWriteResponseRetries

	// check response type
	switch resp := response.(type) {
	case comm.QwatchResponse:
		result = resp.Result
		err = resp.Error
	default:
		slog.Debug("Unsupported response type")
		if err := s.writeResponseWithRetries(conn, []byte("error: 500 Internal Server Error"), maxRetries); err != nil {
			slog.Debug(fmt.Sprintf("Error writing message: %v", err))
			return fmt.Errorf("error writing response: %v", err)
		}
		return nil
	}

	var responseValue interface{}
	var rp *clientio.RESPParser
	if err != nil {
		rp = clientio.NewRESPParser(bytes.NewBuffer([]byte(err.Error())))
	} else {
		rp = clientio.NewRESPParser(bytes.NewBuffer(result.([]byte)))
	}

	responseValue, err = rp.DecodeOne()
	if err != nil {
		slog.Debug("Error decoding response", "error", err)
		if err := s.writeResponseWithRetries(conn, []byte("error: 500 Internal Server Error"), maxRetries); err != nil {
			slog.Debug(fmt.Sprintf("Error writing message: %v", err))
			return fmt.Errorf("error writing response: %v", err)
		}
		return nil
	}

	respBytes, err := json.Marshal(responseValue)
	if err != nil {
		slog.Debug("Error marshaling json", "error", err)
		if err := s.writeResponseWithRetries(conn, []byte("error: marshaling json"), maxRetries); err != nil {
			slog.Debug(fmt.Sprintf("Error writing message: %v", err))
			return fmt.Errorf("error writing response: %v", err)
		}
		return nil
	}

	// success
	// Write response with retries for transient errors
	if err := s.writeResponseWithRetries(conn, respBytes, config.DiceConfig.WebSocket.MaxWriteResponseRetries); err != nil {
		slog.Debug(fmt.Sprintf("Error writing message: %v", err))
		return fmt.Errorf("error writing response: %v", err)
	}

	return nil
}

func (s *WebsocketServer) processResponse(conn *websocket.Conn, diceDBCmd *cmd.DiceDBCmd, response *ops.StoreResponse) error {
	var err error
	maxRetries := config.DiceConfig.WebSocket.MaxWriteResponseRetries

	var responseValue interface{}
	// Check if the command is migrated, if it is we use EvalResponse values
	// else we use RESPParser to decode the response
	_, ok := CmdMetaMap[diceDBCmd.Cmd]
	// TODO: Remove this conditional check and if (true) condition when all commands are migrated
	if !ok {
		responseValue, err = DecodeEvalResponse(response.EvalResponse)
		if err != nil {
			slog.Debug("Error decoding response", "error", err)
			if err := s.writeResponseWithRetries(conn, []byte("error: 500 Internal Server Error"), maxRetries); err != nil {
				slog.Debug(fmt.Sprintf("Error writing message: %v", err))
				return fmt.Errorf("error writing response: %v", err)
			}
			return nil
		}
	} else {
		if response.EvalResponse.Error != nil {
			responseValue = response.EvalResponse.Error.Error()
		} else {
			responseValue = response.EvalResponse.Result
		}
	}

	// Create websocket response
	wsResponse := ResponseParser(responseValue)
	respBytes, err := json.Marshal(wsResponse)
	if err != nil {
		slog.Debug("Error marshaling json", "error", err)
		if err := s.writeResponseWithRetries(conn, []byte("error: marshaling json"), maxRetries); err != nil {
			slog.Debug(fmt.Sprintf("Error writing message: %v", err))
			return fmt.Errorf("error writing response: %v", err)
		}
		return nil
	}

	// success
	if err := s.writeResponseWithRetries(conn, respBytes, config.DiceConfig.WebSocket.MaxWriteResponseRetries); err != nil {
		slog.Debug(fmt.Sprintf("Error writing message: %v", err))
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

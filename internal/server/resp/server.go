package resp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dice/internal/watchmanager"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/internal/clientio/iohandler/netconn"
	respparser "github.com/dicedb/dice/internal/clientio/requestparser/resp"
	"github.com/dicedb/dice/internal/ops"
	"github.com/dicedb/dice/internal/shard"
	"github.com/dicedb/dice/internal/worker"
)

var (
	workerCounter uint64
	startTime     = time.Now().UnixNano() / int64(time.Millisecond)
)

var (
	ErrInvalidIPAddress = errors.New("invalid IP address")
)

const (
	DefaultConnBacklogSize = 128
)

type Server struct {
	Host            string
	Port            int
	serverFD        int
	connBacklogSize int
	workerManager   *worker.WorkerManager
	shardManager    *shard.ShardManager
	watchManager    *watchmanager.Manager
	cmdWatchChan    chan dstore.CmdWatchEvent
	globalErrorChan chan error
	logger          *slog.Logger
}

func NewServer(shardManager *shard.ShardManager, workerManager *worker.WorkerManager, cmdWatchChan chan dstore.CmdWatchEvent, globalErrChan chan error, l *slog.Logger) *Server {
	return &Server{
		Host:            config.DiceConfig.Server.Addr,
		Port:            config.DiceConfig.Server.Port,
		connBacklogSize: DefaultConnBacklogSize,
		workerManager:   workerManager,
		shardManager:    shardManager,
		watchManager:    watchmanager.NewManager(l),
		cmdWatchChan:    cmdWatchChan,
		globalErrorChan: globalErrChan,
		logger:          l,
	}
}

func (s *Server) Run(ctx context.Context) (err error) {
	// BindAndListen the desired port to the server
	if err = s.BindAndListen(); err != nil {
		s.logger.Error("failed to bind server", slog.Any("error", err))
		return err
	}

	defer s.ReleasePort()

	// Start a go routine to accept connections
	errChan := make(chan error, 1)
	wg := &sync.WaitGroup{}

	wg.Add(2)
	go func() {
		defer wg.Done()
		s.watchManager.Run(ctx, s.cmdWatchChan)
	}()

	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		if err := s.AcceptConnectionRequests(ctx, wg); err != nil {
			errChan <- fmt.Errorf("failed to accept connections %w", err)
		}
	}(wg)

	s.logger.Info("DiceDB ready to accept connections on port", slog.Int("resp-port", config.Port))

	select {
	case <-ctx.Done():
		s.logger.Info("Context canceled, initiating shutdown")
	case err = <-errChan:
		s.logger.Error("Error while accepting connections, initiating shutdown", slog.Any("error", err))
	}

	s.Shutdown()

	wg.Wait() // Wait for the go routines to finish
	s.logger.Info("All connections are closed, RESP server exiting gracefully.")

	return err
}

func (s *Server) BindAndListen() error {
	serverFD, socketErr := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if socketErr != nil {
		return fmt.Errorf("failed to create socket: %w", socketErr)
	}

	// Close the socket on exit if an error occurs
	var err error
	defer func() {
		if err != nil {
			if closeErr := syscall.Close(serverFD); closeErr != nil {
				// Wrap the close error with the original bind/listen error
				s.logger.Error("Error occurred", slog.Any("error", err), "additionally, failed to close socket", slog.Any("close-err", closeErr))
			} else {
				s.logger.Error("Error occurred", slog.Any("error", err))
			}
		}
	}()

	if err = syscall.SetsockoptInt(serverFD, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); err != nil {
		return fmt.Errorf("failed to set SO_REUSEADDR: %w", err)
	}

	if err = syscall.SetNonblock(serverFD, true); err != nil {
		return fmt.Errorf("failed to set socket to non-blocking: %w", err)
	}

	ip4 := net.ParseIP(s.Host)
	if ip4 == nil {
		return ErrInvalidIPAddress
	}

	sockAddr := &syscall.SockaddrInet4{
		Port: s.Port,
		Addr: [4]byte{ip4[0], ip4[1], ip4[2], ip4[3]},
	}
	if err = syscall.Bind(serverFD, sockAddr); err != nil {
		return fmt.Errorf("failed to bind socket: %w", err)
	}

	if err = syscall.Listen(serverFD, s.connBacklogSize); err != nil {
		return fmt.Errorf("failed to listen on socket: %w", err)
	}

	s.serverFD = serverFD
	s.logger.Info("RESP Server successfully bound", slog.String("Host", s.Host), slog.Int("Port", s.Port))
	return nil
}

// ReleasePort closes the server socket.
func (s *Server) ReleasePort() {
	if err := syscall.Close(s.serverFD); err != nil {
		s.logger.Error("Failed to close server socket", slog.Any("error", err))
	} else {
		s.logger.Debug("Server socket closed successfully")
	}
}

// AcceptConnectionRequests accepts new client connections
func (s *Server) AcceptConnectionRequests(ctx context.Context, wg *sync.WaitGroup) error {
	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Context canceled, initiating RESP server shutdown")

			return ctx.Err()
		default:
			clientFD, _, err := syscall.Accept(s.serverFD)
			if err != nil {
				if errors.Is(err, syscall.EAGAIN) || errors.Is(err, syscall.EWOULDBLOCK) {
					continue // No more connections to accept at this time
				}

				return fmt.Errorf("error accepting connection: %w", err)
			}

			// Register a new worker for the client
			ioHandler, err := netconn.NewIOHandler(clientFD, s.logger)
			if err != nil {
				s.logger.Error("Failed to create new IOHandler for clientFD", slog.Int("client-fd", clientFD), slog.Any("error", err))
				return err
			}

			parser := respparser.NewParser(s.logger)
			respChan := make(chan *ops.StoreResponse)
			wID := GenerateUniqueWorkerID()
			w := worker.NewWorker(wID, respChan, ioHandler, parser, s.shardManager, s.globalErrorChan, s.logger)
			if err != nil {
				s.logger.Error("Failed to create new worker for clientFD", slog.Int("client-fd", clientFD), slog.Any("error", err))
				return err
			}

			// Register the worker with the worker manager
			err = s.workerManager.RegisterWorker(w)
			if err != nil {
				return err
			}

			wg.Add(1)
			go func(wID string) {
				wg.Done()
				defer func(wm *worker.WorkerManager, workerID string) {
					err := wm.UnregisterWorker(workerID)
					if err != nil {
						s.logger.Warn("Failed to unregister worker", slog.String("worker-id", wID), slog.Any("error", err))
					}
				}(s.workerManager, wID)
				wctx, cwctx := context.WithCancel(ctx)
				defer cwctx()
				err := w.Start(wctx)
				if err != nil {
					s.logger.Debug("Worker stopped", slog.String("worker-id", wID), slog.Any("error", err))
				}
			}(wID)
		}
	}
}

func GenerateUniqueWorkerID() string {
	count := atomic.AddUint64(&workerCounter, 1)
	timestamp := time.Now().UnixNano()/int64(time.Millisecond) - startTime
	return fmt.Sprintf("W-%d-%d", timestamp, count)
}

func (s *Server) Shutdown() {
	// Not implemented
}

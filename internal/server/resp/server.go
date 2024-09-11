package resp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/dicedb/dice/internal/clientio/iohandler/netconn"
	respparser "github.com/dicedb/dice/internal/clientio/requestparser/resp"
	"github.com/dicedb/dice/internal/ops"
	"github.com/dicedb/dice/internal/shard"

	"github.com/charmbracelet/log"
	"github.com/dicedb/dice/config"
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
	wm              *worker.WorkerManager
	sm              *shard.ShardManager
	syncClose       sync.Once
}

func NewServer(sm *shard.ShardManager, wm *worker.WorkerManager) (*Server, error) {
	return &Server{
		Host:            config.DiceConfig.Server.Addr,
		Port:            config.DiceConfig.Server.Port,
		connBacklogSize: DefaultConnBacklogSize,
		wm:              wm,
		sm:              sm,
	}, nil
}

func (s *Server) Run(ctx context.Context) (err error) {
	// BindAndListen the desired port to the server
	if err = s.BindAndListen(); err != nil {
		log.Error("failed to bind server", "error", err)
		return err
	}

	defer s.ReleasePort()

	// Start a go routine to accept connections
	errChan := make(chan error, 1)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		if err := s.AcceptConnectionRequests(ctx, wg); err != nil {
			errChan <- fmt.Errorf("failed to accept connections %w", err)
		}
	}(wg)

	log.Infof("DiceDB ready to accept connections on port %d", config.Port)

	select {
	case <-ctx.Done():
		log.Info("Context cancelled, initiating shutdown")
	case err = <-errChan:
		log.Errorf("Error while accepting connections, initiating shutdown: %v", err)
	}

	if shutdownErr := s.Shutdown(); err != nil {
		log.Errorf("Failed to shut down RESP server: %v", shutdownErr)
	}

	wg.Wait() // Wait for the go routines to finish
	log.Info("All connections are closed, RESP server exiting gracefully.")

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
				log.Errorf("Error occurred: %v; additionally, failed to close socket: %v", err, closeErr)
			} else {
				log.Errorf("Error occurred: %v", err)
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
	log.Infof("Server successfully bound to %s:%d", s.Host, s.Port)
	return nil
}

// ReleasePort closes the server socket.
func (s *Server) ReleasePort() {
	s.syncClose.Do(func() {
		if err := syscall.Close(s.serverFD); err != nil {
			log.Errorf("Failed to close server socket: %v", err)
		} else {
			log.Info("Server socket closed successfully.")
		}
	})
}

// AcceptConnectionRequests accepts new client connections
func (s *Server) AcceptConnectionRequests(ctx context.Context, wg *sync.WaitGroup) error {
	for {
		select {
		case <-ctx.Done():
			log.Info("Context canceled, initiating RESP server shutdown")
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
			ioHandler, err := netconn.NewIOHandler(clientFD)
			if err != nil {
				log.Errorf("Failed to create new IOHandler for clientFD %d: %v", clientFD, err)
				// TODO: Return appropriate error to the client
			}

			wID := GenerateUniqueWorkerID()
			parser := respparser.NewParser()
			respChan := make(chan *ops.StoreResponse)

			w := worker.NewWorker(wID, respChan, ioHandler, parser, s.sm)
			if err != nil {
				log.Errorf("Failed to create new worker for clientFD %d: %v", clientFD, err)
				// TODO: Return appropriate error to the client
				return err
			}

			// Register the worker with the worker manager
			err = s.wm.RegisterWorker(w)
			if err != nil {
				return err
			}

			wg.Add(1)
			go func(wID string) {
				wg.Done()
				defer func(wm *worker.WorkerManager, workerID string) {
					err := wm.UnregisterWorker(workerID)
					if err != nil {
						log.Warnf("Failed to unregister worker %s: %v", workerID, err)
					}
				}(s.wm, wID)
				err := w.Start(ctx)
				if err != nil {
					log.Warnf("Failed to start worker %s: %v", wID, err)
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

func (s *Server) Shutdown() (err error) {
	return err
}

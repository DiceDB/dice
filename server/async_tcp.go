package server

import (
	"context"
	"errors"
	"net"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/charmbracelet/log"
	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/core"
	"github.com/dicedb/dice/core/iomultiplexer"
	"github.com/dicedb/dice/internal/constants"
	"github.com/dicedb/dice/server/utils"
)

var AbortedErr = errors.New("server received ABORT command")

type AsyncServer struct {
	serverFD               int
	maxClients             int
	multiplexer            iomultiplexer.IOMultiplexer
	multiplexerPollTimeout time.Duration
	connectedClients       map[int]*core.Client
	store                  *core.Store
	lastCronExecTime       time.Time
	cronFrequency          time.Duration
}

// NewAsyncServer initializes a new AsyncServer
func NewAsyncServer() *AsyncServer {

	return &AsyncServer{
		maxClients:             20000,
		connectedClients:       make(map[int]*core.Client),
		store:                  core.NewStore(),
		multiplexerPollTimeout: 100 * time.Millisecond,
		lastCronExecTime:       utils.GetCurrentTime(),
		cronFrequency:          1 * time.Second,
	}
}

// SetupUsers initializes the default user for the server
func (s *AsyncServer) SetupUsers() error {
	user, err := core.UserStore.Add(core.DefaultUserName)
	if err != nil {
		return err
	}
	if err := user.SetPassword(config.RequirePass); err != nil {
		return err
	}
	log.Info("default user set up", "password required", config.RequirePass != constants.EmptyStr)
	return nil
}

// WatchKeys watches for changes in keys and notifies clients
func (s *AsyncServer) WatchKeys(ctx context.Context) {
	for {
		select {
		case event := <-core.WatchChannel:
			s.store.WatchList.Range(func(key, value interface{}) bool {
				query := key.(core.DSQLQuery)
				clients := value.(*sync.Map)

				if core.WildCardMatch(query.KeyRegex, event.Key) {
					queryResult, err := core.ExecuteQuery(query, s.store)
					if err != nil {
						log.Error(err)
						return true
					}

					encodedResult := core.Encode(core.CreatePushResponse(&query, &queryResult), false)
					clients.Range(func(clientKey, _ interface{}) bool {
						clientFd := clientKey.(int)
						_, err := syscall.Write(clientFd, encodedResult)
						if err != nil {
							s.store.RemoveWatcher(query, clientFd)
						}
						return true
					})
				}
				return true
			})
		case <-ctx.Done():
			return
		}
	}
}

// FindPortAndBind binds the server to the given host and port
func (s *AsyncServer) FindPortAndBind() error {
	serverFD, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		return err
	}
	s.serverFD = serverFD

	if err := syscall.SetsockoptInt(serverFD, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); err != nil {
		return err
	}

	if err := syscall.SetNonblock(serverFD, true); err != nil {
		return err
	}

	ip4 := net.ParseIP(config.Host)

	return syscall.Bind(serverFD, &syscall.SockaddrInet4{
		Port: config.Port,
		Addr: [4]byte{ip4[0], ip4[1], ip4[2], ip4[3]},
	})
}

// Run starts the server, accepts connections, and handles client requests
func (s *AsyncServer) Run(ctx context.Context) error {
	defer syscall.Close(s.serverFD)

	watchCtx, cancelWatch := context.WithCancel(ctx)
	defer cancelWatch()

	if err := s.SetupUsers(); err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.WatchKeys(watchCtx)
	}()

	if err := syscall.Listen(s.serverFD, s.maxClients); err != nil {
		return err
	}

	var err error
	s.multiplexer, err = iomultiplexer.New(s.maxClients)
	if err != nil {
		return err
	}

	defer func(multiplexer iomultiplexer.IOMultiplexer) {
		err := multiplexer.Close()
		if err != nil {
			log.Warn("failed to close multiplexer", "error", err)
		}
	}(s.multiplexer)

	if err := s.multiplexer.Subscribe(iomultiplexer.Event{
		Fd: s.serverFD,
		Op: iomultiplexer.OpRead,
	}); err != nil {
		return err
	}

	err = s.eventLoop(ctx)

	cancelWatch()
	wg.Wait()

	return err
}

// eventLoop listens for events and handles client requests. It also runs a cron job to delete expired keys
func (s *AsyncServer) eventLoop(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			// TODO: Remove this log
			log.Info("received context cancellation, stopping event loop")
			return ctx.Err()
		default:
			if time.Now().After(s.lastCronExecTime.Add(s.cronFrequency)) {
				core.DeleteExpiredKeys(s.store)
				s.lastCronExecTime = utils.GetCurrentTime()
			}

			events, err := s.multiplexer.Poll(s.multiplexerPollTimeout)
			if err != nil {
				if errors.Is(err, syscall.EINTR) {
					continue
				}
				return err
			}

			for _, event := range events {
				if event.Fd == s.serverFD {
					if err := s.acceptConnection(); err != nil {
						log.Warn(err)
					}
				} else {
					if err := s.handleClientEvent(event); err != nil {
						if errors.Is(err, AbortedErr) {
							log.Info("Received abort command, initiating graceful shutdown")
							return err
						}
						log.Warn(err)
					}
				}
			}
		}
	}
}

// acceptConnection accepts a new client connection and subscribes to read events on the connection.
func (s *AsyncServer) acceptConnection() error {
	fd, _, err := syscall.Accept(s.serverFD)
	if err != nil {
		return err
	}

	s.connectedClients[fd] = core.NewClient(fd)
	if err := syscall.SetNonblock(fd, true); err != nil {
		return err
	}

	return s.multiplexer.Subscribe(iomultiplexer.Event{
		Fd: fd,
		Op: iomultiplexer.OpRead,
	})
}

// handleClientEvent reads commands from the client connection and responds to the client. It also handles disconnections.
func (s *AsyncServer) handleClientEvent(event iomultiplexer.Event) error {
	comm := s.connectedClients[event.Fd]
	if comm == nil {
		return nil
	}

	commands, hasAbort, err := readCommands(comm)
	if err != nil {
		if err := syscall.Close(event.Fd); err != nil {
			log.Printf("error closing client connection: %v", err)
		}
		delete(s.connectedClients, event.Fd)
		return err
	}

	respond(commands, comm, s.store)
	if hasAbort {
		return AbortedErr
	}

	return nil
}

// WaitForSignal listens for OS signals and triggers shutdown
func (s *AsyncServer) WaitForSignal(cancel context.CancelFunc, sigs chan os.Signal) {
	<-sigs
	cancel()
}

// InitiateShutdown gracefully shuts down the server
func (s *AsyncServer) InitiateShutdown() {
	log.Info("initiating shutdown")

	// Close all client connections
	for fd, _ := range s.connectedClients {
		// Close the client socket
		if err := syscall.Close(fd); err != nil {
			log.Warn("failed to close client connection", "error", err)
		}

		delete(s.connectedClients, fd)
	}

	core.Shutdown(s.store)
}

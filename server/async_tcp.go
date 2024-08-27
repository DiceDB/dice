package server

import (
	"context"
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

type AsyncServer struct {
	serverFD         int
	maxClients       int
	multiplexer      iomultiplexer.IOMultiplexer
	connectedClients map[int]*core.Client
	store            *core.Store
	lastCronExecTime time.Time
	cronFrequency    time.Duration
}

// NewAsyncServer initializes a new AsyncServer
func NewAsyncServer() *AsyncServer {

	return &AsyncServer{
		maxClients:       20000,
		connectedClients: make(map[int]*core.Client),
		store:            core.NewStore(),
		lastCronExecTime: utils.GetCurrentTime(),
		cronFrequency:    1 * time.Second,
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
func (s *AsyncServer) Run(parentCtx context.Context, wg *sync.WaitGroup) error {
	defer syscall.Close(s.serverFD)

	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	const pollTimeout = 100 * time.Millisecond

	if err := s.SetupUsers(); err != nil {
		return err
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.WatchKeys(ctx)
	}()

	if err := syscall.Listen(s.serverFD, s.maxClients); err != nil {
		return err
	}

	var err error
	s.multiplexer, err = iomultiplexer.New(s.maxClients)
	if err != nil {
		return err
	}
	defer s.multiplexer.Close()

	if err := s.multiplexer.Subscribe(iomultiplexer.Event{
		Fd: s.serverFD,
		Op: iomultiplexer.OpRead,
	}); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			// Check if context is canceled
			s.initiateShutdown()
			return nil

		default:
			if time.Now().After(s.lastCronExecTime.Add(s.cronFrequency)) {
				core.DeleteExpiredKeys(s.store)
				s.lastCronExecTime = utils.GetCurrentTime()
			}

			events, err := s.multiplexer.Poll(pollTimeout)
			if err != nil {
				// Check for context cancellation on error
				if ctx.Err() != nil {
					s.initiateShutdown()
					return nil
				}
				continue
			}

			for _, event := range events {
				if event.Fd == s.serverFD {
					fd, _, err := syscall.Accept(s.serverFD)
					if err != nil {
						log.Warn(err)
						continue
					}

					s.connectedClients[fd] = core.NewClient(fd)
					if err := syscall.SetNonblock(fd, true); err != nil {
						//nolint:gocritic
						log.Fatal(err)
					}

					if err := s.multiplexer.Subscribe(iomultiplexer.Event{
						Fd: fd,
						Op: iomultiplexer.OpRead,
					}); err != nil {
						log.Fatal(err)
					}
				} else {
					comm := s.connectedClients[event.Fd]
					if comm == nil {
						continue
					}
					cmds, hasAbort, err := readCommands(comm)
					if err != nil {
						syscall.Close(event.Fd)
						delete(s.connectedClients, event.Fd)
						continue
					}
					respond(cmds, comm, s.store)
					if hasAbort {
						// context will be cancelled using defer.
						return nil
					}
				}
			}
		}
	}
}

// WaitForSignal listens for OS signals and triggers shutdown
func (s *AsyncServer) WaitForSignal(cancel context.CancelFunc, sigs chan os.Signal) {
	<-sigs
	cancel()
}

// initiateShutdown gracefully shuts down the server
func (s *AsyncServer) initiateShutdown() {
	log.Info("initiating shutdown")
	core.Shutdown(s.store)
}

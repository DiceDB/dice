package server

import (
	"context"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/charmbracelet/log"
	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/core"
	"github.com/dicedb/dice/core/iomultiplexer"
	"github.com/dicedb/dice/internal/constants"
	"github.com/dicedb/dice/server/utils"
)

var cronFrequency time.Duration = 1 * time.Second
var lastCronExecTime time.Time = utils.GetCurrentTime()

const EngineStatusWAITING int32 = 1 << 1
const EngineStatusBUSY int32 = 1 << 2
const EngineStatusSHUTTINGDOWN int32 = 1 << 3
const EngineStatusTRANSACTION int32 = 1 << 4

var eStatus int32 = EngineStatusWAITING

var connectedClients map[int]*core.Client

var asyncStore = core.NewStore()

func init() {
	connectedClients = make(map[int]*core.Client)
}

func setupUsers() {
	var (
		user *core.User
		err  error
	)
	log.Info("setting up default user.", "password required", config.RequirePass != constants.EmptyStr)
	if user, err = core.UserStore.Add(core.DefaultUserName); err != nil {
		log.Fatal(err)
	}
	if err = user.SetPassword(config.RequirePass); err != nil {
		log.Fatal(err)
	}
}

// Waits on `core.WatchChannel` to receive updates about keys. Sends the update
// to all the clients that are watching the key.
// The message sent to the client will contain the new value and the operation
// that was performed on the key.
func WatchKeys(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case event := <-core.WatchChannel:
			asyncStore.WatchList.Range(func(key, value interface{}) bool {
				query := key.(core.DSQLQuery)
				clients := value.(*sync.Map)

				if core.WildCardMatch(query.KeyRegex, event.Key) {
					queryResult, err := core.ExecuteQuery(query, asyncStore)
					if err != nil {
						log.Error(err)
						return true // continue to next item
					}

					encodedResult := core.Encode(core.CreatePushResponse(&query, &queryResult), false)
					clients.Range(func(clientKey, _ interface{}) bool {
						clientFd := clientKey.(int)
						_, err := syscall.Write(clientFd, encodedResult)
						if err != nil {
							clients.Delete(clientFd)
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

func WaitForSignal(wg *sync.WaitGroup, sigs chan os.Signal) {
	defer wg.Done()
	<-sigs

	// if server is busy continue to wait
	for atomic.LoadInt32(&eStatus) == EngineStatusBUSY { //nolint:revive
	}

	// CRITICAL TO HANDLE
	// We do not want server to ever go back to BUSY state
	// when control flow is here ->

	// immediately set the status to be SHUTTING DOWN
	// the only place where we can set the status to be SHUTTING DOWN
	atomic.StoreInt32(&eStatus, EngineStatusSHUTTINGDOWN)

	// if server is in any other state, initiate a shutdown
	core.Shutdown(asyncStore)
	os.Exit(0) //nolint:gocritic
}

func FindPortAndBind() (int, error) {
	serverFD, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		return 0, err
	}

	if err := syscall.SetsockoptInt(serverFD, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); err != nil {
		return 0, err
	}

	// Set the Socket operate in a non-blocking mode
	if err := syscall.SetNonblock(serverFD, true); err != nil {
		return 0, err
	}

	// Bind the IP and the port
	ip4 := net.ParseIP(config.Host)

	if err := syscall.Bind(serverFD, &syscall.SockaddrInet4{
		Port: config.Port,
		Addr: [4]byte{ip4[0], ip4[1], ip4[2], ip4[3]},
	}); err != nil {
		return 0, err
	}

	return serverFD, nil
}

func RunAsyncTCPServer(serverFD int, wg *sync.WaitGroup) {
	defer wg.Done()
	defer syscall.Close(serverFD)

	log.Info("starting an asynchronous TCP server on", config.Host, config.Port)

	setupUsers()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	defer func() {
		atomic.StoreInt32(&eStatus, EngineStatusSHUTTINGDOWN)
	}()
	maxClients := 20000

	wg.Add(1)
	go WatchKeys(ctx, wg)

	// Start listening
	if err := syscall.Listen(serverFD, maxClients); err != nil {
		log.Fatal("error while listening", err) //nolint:gocritic
	}

	log.Info("ready to accept connections")

	// AsyncIO starts here!!

	// creating multiplexer instance
	var multiplexer iomultiplexer.IOMultiplexer
	multiplexer, err := iomultiplexer.New(maxClients)
	if err != nil {
		log.Fatal(err)
	}
	defer multiplexer.Close()

	// Listen to read events on the Server itself
	if err := multiplexer.Subscribe(iomultiplexer.Event{
		Fd: serverFD,
		Op: iomultiplexer.OpRead,
	}); err != nil {
		log.Fatal(err)
	}

	// loop until the server is not shutting down
	for atomic.LoadInt32(&eStatus) != EngineStatusSHUTTINGDOWN {
		if time.Now().After(lastCronExecTime.Add(cronFrequency)) {
			core.DeleteExpiredKeys(asyncStore)
			lastCronExecTime = utils.GetCurrentTime()
		}

		// Say, the Engine triggered SHUTTING down when the control flow is here ->
		// Current: Engine status == WAITING
		// Update: Engine status = SHUTTING_DOWN
		// Then we have to exit (handled in Signal Handler)

		// poll for events that are ready for IO
		events, err := multiplexer.Poll(-1)
		if err != nil {
			continue
		}

		// Here, we do not want server to go back from SHUTTING DOWN
		// to BUSY
		// If the engine status == SHUTTING_DOWN over here ->
		// We have to exit
		// hence the only legal transitiion is from WAITING to BUSY
		// if that does not happen then we can exit.

		// mark engine as BUSY only when it is in the waiting state
		if !atomic.CompareAndSwapInt32(&eStatus, EngineStatusWAITING, EngineStatusBUSY) {
			// if swap unsuccessful then the existing status is not WAITING, but something else
			if eStatus == EngineStatusSHUTTINGDOWN {
				return
			}
		}

		for _, event := range events {
			// if the socket server itself is ready for an IO
			if event.Fd == serverFD {
				// accept the incoming connection from a client
				fd, _, err := syscall.Accept(serverFD)
				if err != nil {
					log.Warn(err)
					continue
				}

				connectedClients[fd] = core.NewClient(fd)
				if err := syscall.SetNonblock(fd, true); err != nil {
					log.Fatal(err)
				}

				// add this new TCP connection to be monitored
				if err := multiplexer.Subscribe(iomultiplexer.Event{
					Fd: fd,
					Op: iomultiplexer.OpRead,
				}); err != nil {
					log.Fatal(err)
				}
			} else {
				comm := connectedClients[event.Fd]
				if comm == nil {
					continue
				}
				cmds, hasABORT, err := readCommands(comm)

				if err != nil {
					syscall.Close(event.Fd)
					delete(connectedClients, event.Fd)
					continue
				}
				respond(cmds, comm, asyncStore)
				if hasABORT {
					cancel()
					return
				}
			}
		}

		// mark engine as WAITING
		// no contention as the signal handler is blocked until
		// the engine is BUSY
		atomic.StoreInt32(&eStatus, EngineStatusWAITING)
	}
}

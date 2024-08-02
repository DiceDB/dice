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
)

const (
	EngineStatus_WAITING       int32 = 1 << 1
	EngineStatus_BUSY          int32 = 1 << 2
	EngineStatus_SHUTTING_DOWN int32 = 1 << 3
	EngineStatus_TRANSACTION   int32 = 1 << 4

	MaxConnClients int = 20000
)

var (
	cronFrequency    time.Duration = 1 * time.Second
	lastCronExecTime time.Time     = time.Now()
	eStatus          int32         = EngineStatus_WAITING

	connectedClients map[int]*core.Client
)

type (
	AsyncServer struct {
		serverFD int
		mux      iomultiplexer.IOMultiplexer

		wg         *sync.WaitGroup
		ctx        context.Context
		cancelFunc context.CancelFunc
	}
)

func init() {
	connectedClients = make(map[int]*core.Client)
}

// Waits on `core.WatchChannel` to receive updates about keys. Sends the update
// to all the clients that are watching the key.
// The message sent to the client will contain the new value and the operation
// that was performed on the key.
func (asyncServer *AsyncServer) WatchKeys() {
	defer asyncServer.wg.Done()
	for {
		select {
		case event := <-core.WatchChannel:
			core.WatchList.Range(func(key, value interface{}) bool {
				query := key.(core.DSQLQuery)
				clients := value.(*sync.Map)

				if core.WildCardMatch(query.KeyRegex, event.Key) {
					result, err := core.ExecuteQuery(query)
					if err != nil {
						log.Error(err)
						return true // continue to next item
					}

					encodedResult := core.Encode(result, false)
					clients.Range(func(clientKey, _ interface{}) bool {
						clientFd := clientKey.(int)
						_, err := syscall.Write(clientFd, encodedResult)
						if err != nil {
							core.RemoveWatcher(query, clientFd)
						}
						return true
					})
				}
				return true
			})
		case <-asyncServer.ctx.Done():
			return
		}
	}
}

func WaitForSignal(wg *sync.WaitGroup, sigs chan os.Signal) {
	defer wg.Done()
	<-sigs

	// CRITICAL TO HANDLE
	// We do not want server to ever go back to BUSY state
	// when control flow is here ->

	// immediately set the status to be SHUTTING DOWN
	// the only place where we can set the status to be SHUTTING DOWN
	atomic.StoreInt32(&eStatus, EngineStatus_SHUTTING_DOWN)

	// if server is in any other state, initiate a shutdown
	core.Shutdown()
	os.Exit(0)
}

func FindPortAndBind() (int, error) {
	serverFD, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		return 0, err
	}

	if err = syscall.SetsockoptInt(serverFD, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); err != nil {
		return 0, err
	}

	// Set the Socket operate in a non-blocking mode
	if err = syscall.SetNonblock(serverFD, true); err != nil {
		return 0, err
	}

	// Bind the IP and the port
	ip4 := net.ParseIP(config.Host)

	if err = syscall.Bind(serverFD, &syscall.SockaddrInet4{
		Port: config.Port,
		Addr: [4]byte{ip4[0], ip4[1], ip4[2], ip4[3]},
	}); err != nil {
		return 0, err
	}

	return serverFD, nil
}

/*
----------------------------------
Async Server
----------------------------------
*/
func NewAsyncServer(wg *sync.WaitGroup, serverFD int) (asyncServer *AsyncServer, err error) {
	asyncServer = &AsyncServer{
		serverFD: serverFD,
		wg:       wg,
	}
	if asyncServer.ctx, asyncServer.cancelFunc = context.WithCancel(context.Background()); err != nil {
		return
	}
	return
}

func (asyncServer *AsyncServer) Listen() (err error) {
	log.Info("starting an asynchronous TCP server on", config.Host, config.Port)

	asyncServer.wg.Add(1)
	if err = syscall.Listen(asyncServer.serverFD, MaxConnClients); err != nil {
		log.Fatal("error while listening", err)
		return
	}
	log.Info("ready to accept connections")

	if asyncServer.mux, err = iomultiplexer.New(MaxConnClients); err != nil {
		log.Fatal(err)
		return
	}
	// Listen to read events on the Server itself
	if err = asyncServer.mux.Subscribe(iomultiplexer.Event{
		Fd: asyncServer.serverFD,
		Op: iomultiplexer.OP_READ,
	}); err != nil {
		log.Fatal(err)
	}

	return
}

func (asyncServer *AsyncServer) Close() {
	defer syscall.Close(asyncServer.serverFD)
	defer asyncServer.wg.Done()
	defer asyncServer.cancelFunc()
	defer func() {
		atomic.StoreInt32(&eStatus, EngineStatus_SHUTTING_DOWN)
	}()
	defer asyncServer.mux.Close()
}

func (asyncServer *AsyncServer) deleteExpiredKeys() (err error) {
	if time.Now().After(lastCronExecTime.Add(cronFrequency)) {
		core.DeleteExpiredKeys()
		lastCronExecTime = time.Now()
	}
	return
}

func (asyncServer *AsyncServer) acceptConn(serverFD int) (err error) {
	// accept the incoming connection from a client
	fd, _, err := syscall.Accept(serverFD)
	if err != nil {
		log.Warn(err)
		return
	}

	connectedClients[fd] = core.NewClient(fd)
	if err = syscall.SetNonblock(fd, true); err != nil {
		log.Error(err)
		return
	}

	// add this new TCP connection to be monitored
	if err = asyncServer.mux.Subscribe(iomultiplexer.Event{
		Fd: fd,
		Op: iomultiplexer.OP_READ,
	}); err != nil {
		log.Error(err)
		return
	}
	return
}

func (asyncServer *AsyncServer) acceptBytesFromConn(serverFD int) (err error) {
	comm := connectedClients[serverFD]
	if comm == nil {
		return
	}
	cmds, hasABORT, err := readCommands(comm)

	if err != nil {
		syscall.Close(serverFD)
		delete(connectedClients, serverFD)
		return
	}
	respond(cmds, comm)
	if hasABORT {
		asyncServer.cancelFunc()
		return
	}
	return
}

func RunAsyncTCPServer(serverFD int, wg *sync.WaitGroup) {
	var (
		asyncServer *AsyncServer
		err         error
	)
	if asyncServer, err = NewAsyncServer(wg, serverFD); err != nil {
		log.Fatal(err)
		return
	}
	defer asyncServer.Close()
	if err = asyncServer.Listen(); err != nil {
		log.Fatal(err)
	}

	go asyncServer.WatchKeys()

	// AsyncIO starts here!!

	// loop until the server is not shutting down
	for atomic.LoadInt32(&eStatus) != EngineStatus_SHUTTING_DOWN {
		asyncServer.deleteExpiredKeys()

		// Say, the Engine triggered SHUTTING down when the control flow is here ->
		// Current: Engine status == WAITING
		// Update: Engine status = SHUTTING_DOWN
		// Then we have to exit (handled in Signal Handler)

		// poll for events that are ready for IO
		events, err := asyncServer.mux.Poll(-1)
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
		if !atomic.CompareAndSwapInt32(&eStatus, EngineStatus_WAITING, EngineStatus_BUSY) {
			// if swap unsuccessful then the existing status is not WAITING, but something else
			switch eStatus {
			case EngineStatus_SHUTTING_DOWN:
				return
			}
		}

		for _, event := range events {
			// if the socket server itself is ready for an IO
			if event.Fd == serverFD {
				if err = asyncServer.acceptConn(serverFD); err != nil {
					continue
				}
			} else {
				if err = asyncServer.acceptBytesFromConn(event.Fd); err != nil {
					continue
				}
			}
		}
	}
}

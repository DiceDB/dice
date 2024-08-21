package nitroserver

import (
	"context"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/charmbracelet/log"
	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/core"
	"github.com/dicedb/dice/core/iomultiplexer"
	"github.com/dicedb/dice/server"
)

const EngineStatusWAITING int32 = 1 << 1
const EngineStatusBUSY int32 = 1 << 2
const EngineStatusSHUTTINGDOWN int32 = 1 << 3
const EngineStatusTRANSACTION int32 = 1 << 4

var eStatus int32 = EngineStatusWAITING

var connectedClients map[int]*core.Client

func init() {
	connectedClients = make(map[int]*core.Client)
}

func RunNitroServer(serverFD, cores int, wg *sync.WaitGroup) {
	defer wg.Done()
	defer syscall.Close(serverFD)

	log.Info("Starting Dice Nitro server on", config.Host, config.Port)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	defer func() {
		atomic.StoreInt32(&eStatus, EngineStatusSHUTTINGDOWN)
	}()

	maxClients := 20000
	wg.Add(1)
	InitShards(ctx, wg, cores)

	// Start listening
	if err := syscall.Listen(serverFD, maxClients); err != nil {
		log.Fatal("error while listening", err) //nolint:gocritic
	}

	log.Info("Nitro server is ready to accept connections")

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
		// Todo: Find concerns with disabling this
		//if time.Now().After(lastCronExecTime.Add(cronFrequency)) {
		//	core.DeleteExpiredKeys(asyncStore)
		//	lastCronExecTime = utils.GetCurrentTime()
		//}

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
				cmds, hasABORT, err := server.ReadCommands(comm)

				if err != nil {
					syscall.Close(event.Fd)
					delete(connectedClients, event.Fd)
					continue
				}

				SubmitAndListenClientOperation(comm, cmds)

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

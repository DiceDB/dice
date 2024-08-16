package server

import (
	"context"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
	"fmt"
	"bytes"

	"github.com/charmbracelet/log"
	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/core"
	"github.com/dicedb/dice/core/iomultiplexer"
)

func RunThreadedServer(serverFD int, wg *sync.WaitGroup) {
	defer wg.Done()
	defer syscall.Close(serverFD)

	log.Info("starting multi-threaded TCP server on", config.Host, config.Port)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	defer func() {
		atomic.StoreInt32(&eStatus, EngineStatus_SHUTTING_DOWN)
	}()
	maxClients := 20000

	wg.Add(1)
	go WatchKeys(ctx, wg)

	// Start listening
	if err := syscall.Listen(serverFD, maxClients); err != nil {
		log.Fatal("error while listening", err)
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
		Op: iomultiplexer.OP_READ,
	}); err != nil {
		log.Fatal(err)
	}

	// loop until the server is not shutting down
	for atomic.LoadInt32(&eStatus) != EngineStatus_SHUTTING_DOWN {
		if time.Now().After(lastCronExecTime.Add(cronFrequency)) {
			core.DeleteExpiredKeys(storeSyncTcp)
			lastCronExecTime = time.Now()
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
					Op: iomultiplexer.OP_READ,
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
				// respond(cmds, comm, storeSyncTcp)

				go func() {
					for {
						ioresult, ok := <-ipool.Get().ioresch
						if !ok {
							fmt.Println("ioresch channel closed. Exiting goroutine.")
							return
						}
						fmt.Println("Message to return from Threaded tcp:", ioresult.message + "<")
				
						var response []byte
						buf := bytes.NewBuffer(response)
						buf.Write([]byte(ioresult.message))

						if _, err := comm.Write(buf.Bytes()); err != nil {
							log.Info("Error writing to client")
							log.Error(err)
						}
					}
				}()
				
				ipool.Get().ioreqch <- &IORequest{conn: comm, cmds: &cmds, keys: core.GetKeyForOperation(cmds)}					
				

				if hasABORT {
					cancel()
					return
				}
			}
		}

		// mark engine as WAITING
		// no contention as the signal handler is blocked until
		// the engine is BUSY
		atomic.StoreInt32(&eStatus, EngineStatus_WAITING)
	}
}
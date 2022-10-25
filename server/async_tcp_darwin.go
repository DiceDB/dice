package server

import (
	"log"
	"net"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/core"
)

/*
Implementation Reference for OSX Support:
i) https://overengineered.dev/writing-an-event-loop-in-go
ii) https://www.freebsd.org/cgi/man.cgi?query=kqueue&sektion=2
*/
func RunAsyncTCPServer(wg *sync.WaitGroup) error {
	defer wg.Done()
	defer func() {
		atomic.StoreInt32(&eStatus, EngineStatus_SHUTTING_DOWN)
	}()

	log.Println("starting an asynchronous TCP server on", config.Host, config.Port)

	max_clients := 20000

	// Create K Event Objects to hold events
	var events []syscall.Kevent_t = make([]syscall.Kevent_t, max_clients)

	// Create a socket
	serverFD, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		return err
	}
	defer syscall.Close(serverFD)

	// Set the Socket operate in a non-blocking mode
	if err = syscall.SetNonblock(serverFD, true); err != nil {
		return err
	}

	// Bind the IP and the port
	ip4 := net.ParseIP(config.Host)

	if err = syscall.Bind(serverFD, &syscall.SockaddrInet4{
		Port: config.Port,
		Addr: [4]byte{ip4[0], ip4[1], ip4[2], ip4[3]},
	}); err != nil {
		return err
	}

	// Start listening
	if err = syscall.Listen(serverFD, max_clients); err != nil {
		return err
	}

	// AsyncIO starts here!!

	// creating KQueue instance
	kqueueFD, err := syscall.Kqueue()
	if err != nil {
		log.Fatal(err)
	}
	defer syscall.Close(kqueueFD)

	// Specify the events we want to get hints about
	// and set the socket on which
	var socketServerEvent syscall.Kevent_t = syscall.Kevent_t{
		Ident:  uint64(serverFD),
		Filter: syscall.EVFILT_READ,
		Flags:  syscall.EV_ADD | syscall.EV_ENABLE,
	}

	// Listen to read events on the Server itself
	if changeEventRegistered, err := syscall.Kevent(kqueueFD, []syscall.Kevent_t{socketServerEvent}, nil, nil); err != nil || changeEventRegistered == -1 {
		return err
	}

	// loop until the server is not shutting down
	for atomic.LoadInt32(&eStatus) != EngineStatus_SHUTTING_DOWN {
		if time.Now().After(lastCronExecTime.Add(cronFrequency)) {
			core.DeleteExpiredKeys()
			lastCronExecTime = time.Now()
		}

		// Say, the Engine triggered SHUTTING down when the control flow is here ->
		// Current: Engine status == WAITING
		// Update: Engine status = SHUTTING_DOWN
		// Then we have to exit (handled in Signal Handler)

		// see if any FD is ready for an IO
		nevents, e := syscall.Kevent(kqueueFD, nil, events, nil)
		if e != nil {
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
				return nil
			}
		}

		for i := 0; i < nevents; i++ {
			// if the socket server itself is ready for an IO
			if int(events[i].Ident) == serverFD {
				// accept the incoming connection from a client
				fd, _, err := syscall.Accept(serverFD)
				if err != nil {
					log.Println("err", err)
					continue
				}

				connectedClients[fd] = core.NewClient(fd)
				syscall.SetNonblock(fd, true)

				// add this new TCP connection to be monitored
				var socketClientEvent syscall.Kevent_t = syscall.Kevent_t{
					Ident:  uint64(fd),
					Filter: syscall.EVFILT_READ,
					Flags:  syscall.EV_ADD | syscall.EV_ENABLE,
				}

				if changeEventRegistered, err := syscall.Kevent(kqueueFD, []syscall.Kevent_t{socketClientEvent}, nil, nil); err != nil || changeEventRegistered == -1 {
					return err
				}

			} else {
				comm := connectedClients[int(events[i].Ident)]
				if comm == nil {
					continue
				}
				cmds, hasABORT, err := readCommands(comm)
				if err != nil {
					syscall.Close(int(events[i].Ident))
					delete(connectedClients, int(events[i].Ident))
					continue
				}
				respond(cmds, comm)
				if hasABORT {
					return
				}
			}
		}

		// mark engine as WAITING
		// no contention as the signal handler is blocked until
		// the engine is BUSY
		atomic.StoreInt32(&eStatus, EngineStatus_WAITING)
	}

	return nil
}

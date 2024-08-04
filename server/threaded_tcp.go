package server

import (
	"sync"
	"syscall"
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/dicedb/dice/config"
	"github.com/charmbracelet/log"
	"github.com/dicedb/dice/core"
)

func RunThreadedServer(serverFD int, wg *sync.WaitGroup) {
	defer wg.Done()
	defer syscall.Close(serverFD)

	log.Info("starting an threaded TCP server on", config.Host, config.Port)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ln, err := net.Listen("tcp", config.Host + ":" + strconv.Itoa(config.Port))
	if err != nil {
		panic(err)
	}
	defer ln.Close()

	wg.Add(1)
	go WatchKeys(ctx, wg)

	fmt.Println("server listening on port ", config.Port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("error accepting connection, moving on ...", err)
			continue
		}

		cmds, hasABORT, err := readCommands(conn)

		for _, cmd := range cmds {
			var key = core.getKeyForOperation(cmd)
			ipool.Get().reqch <- &Request(cmd, key, conn) // core.NewClient(fd)
			// Cann run from shardpool
		}

		if hasABORT {
			cancel()
			return
		}
	}
}
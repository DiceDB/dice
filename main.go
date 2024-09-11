package main

import (
	"context"
	"errors"
	"flag"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/dicedb/dice/internal/server"

	"github.com/charmbracelet/log"
	"github.com/dicedb/dice/config"
)

func setupFlags() {
	flag.StringVar(&config.Address, "address", "0.0.0.0:7379", "address for the dice server")
	// flag.IntVar(&config.Port, "port", 7379, "port for the dice server")
	flag.StringVar(&config.RequirePass, "requirepass", config.RequirePass, "enable authentication for the default user")
	flag.Parse()
}

func main() {
	setupFlags()

	ctx, cancel := context.WithCancel(context.Background())

	// Handle SIGTERM and SIGINT
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)

	// Initialize the AsyncServer
	asyncServer := server.NewAsyncServer()

	// Find a port and bind it
	if err := asyncServer.Bind(); err != nil {
		cancel()
		log.Fatal("Error finding and binding port:", "error", err)
	}

	wg := sync.WaitGroup{}
	// Goroutine to handle shutdown signals

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-sigs
		asyncServer.InitiateShutdown()
		cancel()
	}()

	// Run the server
	err := asyncServer.Run(ctx)

	// Handling different server errors
	if err != nil {
		if errors.Is(err, context.Canceled) {
			log.Info("Server was canceled")
		} else if errors.Is(err, server.ErrAborted) {
			log.Info("Server received abort command")
		} else {
			log.Error("Server error", "error", err)
		}
	} else {
		log.Info("Server stopped without error")
	}

	close(sigs)
	wg.Wait()
	log.Info("Server has shut down gracefully")
}

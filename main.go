package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/server"
)

func setupFlags() {
	flag.StringVar(&config.Host, "host", "0.0.0.0", "host for the dice server")
	flag.IntVar(&config.Port, "port", 7379, "port for the dice server")
	flag.Parse()
}

func main() {
	setupFlags()

	// Handle SIGTERM and SIGINT
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	// Find a port and bind
	serverFD, err := server.FindPortAndBind()
	if err != nil {
		log.Fatalf("failed to find and bind port: %v", err)
	}

	// Run the server, listen to incoming connections and handle them
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go server.RunAsyncTCPServer(serverFD, wg)

	// Listen to signals, but not a hard blocker to shutdown
	go server.WaitForSignal(wg, sigChan)

	// Wait for all goroutines to finish
	wg.Wait()
}

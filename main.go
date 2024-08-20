package main

import (
	"flag"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/charmbracelet/log"
	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/server"
)

func setupFlags() {
	flag.StringVar(&config.Host, "host", "0.0.0.0", "host for the dice server")
	flag.IntVar(&config.Port, "port", 7379, "port for the dice server")
	flag.StringVar(&config.RequirePass, "requirepass", config.RequirePass, "enable authentication for the default user")
	flag.Parse()

	log.Info("Password", config.RequirePass)
}

func main() {
	setupFlags()

	// Handle SIGTERM and SIGINT
	var sigs chan os.Signal = make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)

	// Find a port and bind
	// If port not available, raise FATAL error
	serverFD, err := server.FindPortAndBind()
	if err != nil {
		log.Fatal(err)
		return
	}

	var wg sync.WaitGroup

	// Run the server, listen to incoming connections and handle them
	wg.Add(1)

	log.Info("Starting Classic Async TCP Server")
	go server.RunAsyncTCPServer(serverFD, &wg)

	// Listen to signals, but not a hardblocker to shutdown
	go server.WaitForSignal(&wg, sigs)

	// Wait for all goroutines to finish
	wg.Wait()
}

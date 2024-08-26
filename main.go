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
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)

	// Create a wait group to manage goroutines
	var wg sync.WaitGroup

	// Initialize the AsyncServer
	asyncServer := server.NewAsyncServer(&wg)

	// Find a port and bind it
	if err := asyncServer.FindPortAndBind(); err != nil {
		log.Fatal("Error finding and binding port:", err)
		return
	}

	// Start the server in a goroutine
	wg.Add(1)
	go func() {
		if err := asyncServer.Run(); err != nil {
			log.Fatal("Error running the server:", err)
		}
	}()

	// Start signal handling to listen for shutdown signals in a separate goroutine
	go asyncServer.WaitForSignal(sigs)

	// Wait for all goroutines to complete
	wg.Wait()

	log.Info("Server has shut down gracefully")
}

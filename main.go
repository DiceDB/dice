package main

import (
	"flag"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync"
	"syscall"

	"github.com/charmbracelet/log"
	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/server"
	"github.com/dicedb/dice/server/nitroserver"
)

func getDiceCores(requestedCores int) int {
	var diceNumCores int
	if requestedCores < 1 {
		diceNumCores = 1
		log.Info("Cannot use " + strconv.Itoa(requestedCores) + " Cores. Setting Cores to Max usable cores: " + strconv.Itoa(diceNumCores))
	} else if requestedCores >= runtime.NumCPU() {
		diceNumCores = runtime.NumCPU() - 1
		log.Info("Cannot use " + strconv.Itoa(requestedCores) + " Cores. Setting Cores to Max usable cores: " + strconv.Itoa(diceNumCores))
	} else {
		diceNumCores = requestedCores
	}
	return diceNumCores
}

func setupFlags() {
	flag.StringVar(&config.Host, "host", "0.0.0.0", "host for the dice server")
	flag.IntVar(&config.Port, "port", 7379, "port for the dice server")
	flag.StringVar(&config.RequirePass, "requirepass", config.RequirePass, "enable authentication for the default user")
	flag.IntVar(&config.Cores, "cores", 1, "Number of CPU cores for dice server")
	flag.Parse()

	// Set appropriate num of Dice Cores for Nitro Mode
	config.Cores = getDiceCores(config.Cores)
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

	if config.Cores == 1 {
		log.Info("Starting Classic Async TCP Server")
		go server.RunAsyncTCPServer(serverFD, &wg)
	} else {
		log.Info("Starting Dice Nitro Server with " + strconv.Itoa(config.Cores) + " Cores")
		go nitroserver.RunNitroServer(serverFD, config.Cores, &wg)
	}

	// Listen to signals, but not a hardblocker to shutdown
	go server.WaitForSignal(&wg, sigs)

	// Wait for all goroutines to finish
	wg.Wait()
}

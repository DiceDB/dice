package main

import (
	"context"
	"errors"
	"flag"
	"github.com/dicedb/dice/internal/server/resp"
	"github.com/dicedb/dice/internal/worker"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/dicedb/dice/internal/shard"
	dstore "github.com/dicedb/dice/internal/store"

	"github.com/dicedb/dice/internal/server"

	"github.com/charmbracelet/log"
	"github.com/dicedb/dice/config"
)

func setupFlags() {
	flag.StringVar(&config.Host, "host", "0.0.0.0", "host for the dice server")
	flag.IntVar(&config.Port, "port", 7379, "port for the dice server")
	flag.BoolVar(&config.EnableHTTP, "enable-http", true, "run server in HTTP mode as well")
	flag.IntVar(&config.HTTPPort, "http-port", 8082, "HTTP port for the dice server")
	flag.StringVar(&config.RequirePass, "requirepass", config.RequirePass, "enable authentication for the default user")
	flag.StringVar(&config.CustomConfigFilePath, "o", config.CustomConfigFilePath, "dir path to create the config file")
	flag.StringVar(&config.ConfigFileLocation, "c", config.ConfigFileLocation, "file path of the config file")
	flag.BoolVar(&config.InitConfigCmd, "init-config", false, "initialize a new config file")
	flag.Parse()
}

func main() {
	setupFlags()
	config.SetupConfig()
	ctx, cancel := context.WithCancel(context.Background())

	// Handle SIGTERM and SIGINT
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)

	watchChan := make(chan dstore.WatchEvent, config.DiceConfig.Server.KeysLimit)
	shardManager := shard.NewShardManager(1, watchChan)
	workerManager := worker.NewWorkerManager(config.DiceConfig.Server.MaxClients, shardManager)

	// Initialize the RESP Server
	respServer, err := resp.NewServer(shardManager, workerManager)
	if err != nil {
		log.Fatal("Error initializing server:", err)
	}

	// Initialize the HTTP server
	httpServer := server.NewHTTPServer(shardManager, watchChan)

	wg := sync.WaitGroup{}

	// Goroutine to handle shutdown signals
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-sigs
		cancel()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		shardManager.Run(ctx)
	}()

	serverErrCh := make(chan error, 2)
	var serverWg sync.WaitGroup
	serverWg.Add(1)
	go func() {
		defer serverWg.Done()
		// Run the server
		err := respServer.Run(ctx)

		// Handling different server errors
		if err != nil {
			if errors.Is(err, context.Canceled) {
				log.Debug("Server was canceled")
			} else if errors.Is(err, server.ErrAborted) {
				log.Debug("Server received abort command")
			} else {
				log.Error("Server error", "error", err)
			}
			serverErrCh <- err
		} else {
			log.Debug("Server stopped without error")
		}
	}()

	serverWg.Add(1)
	go func() {
		defer serverWg.Done()
		// Run the HTTP server
		err := httpServer.Run(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				log.Debug("HTTP Server was canceled")
			} else if errors.Is(err, server.ErrAborted) {
				log.Debug("HTTP received abort command")
			} else {
				log.Error("HTTP Server error", "error", err)
			}
			serverErrCh <- err
		} else {
			log.Debug("HTTP Server stopped without error")
		}
	}()

	go func() {
		serverWg.Wait()
		close(serverErrCh) // Close the channel when both servers are done
	}()

	for err := range serverErrCh {
		if err != nil && errors.Is(err, server.ErrAborted) {
			// if either the AsyncServer or the HTTPServer received an abort command,
			// cancel the context, helping gracefully exiting all servers
			cancel()
		}
	}

	close(sigs)
	cancel()

	wg.Wait()
	log.Debug("Server has shut down gracefully")
}

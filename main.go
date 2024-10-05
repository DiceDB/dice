package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"

	"github.com/dicedb/dice/config"
	diceerrors "github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/logger"
	"github.com/dicedb/dice/internal/server"
	"github.com/dicedb/dice/internal/server/resp"
	"github.com/dicedb/dice/internal/shard"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dice/internal/worker"
)

func init() {
	flag.StringVar(&config.Host, "host", "0.0.0.0", "host for the dice server")
	flag.IntVar(&config.Port, "port", 7379, "port for the dice server")
	flag.BoolVar(&config.EnableHTTP, "enable-http", true, "run server in HTTP mode as well")
	flag.BoolVar(&config.EnableMultiThreading, "enable-multithreading", false, "run server in multithreading mode")
	flag.IntVar(&config.HTTPPort, "http-port", 8082, "HTTP port for the dice server")
	flag.IntVar(&config.WebsocketPort, "websocket-port", 8379, "Websocket port for the dice server")
	flag.StringVar(&config.RequirePass, "requirepass", config.RequirePass, "enable authentication for the default user")
	flag.StringVar(&config.CustomConfigFilePath, "o", config.CustomConfigFilePath, "dir path to create the config file")
	flag.StringVar(&config.FileLocation, "c", config.FileLocation, "file path of the config file")
	flag.BoolVar(&config.InitConfigCmd, "init-config", false, "initialize a new config file")
	flag.IntVar(&config.KeysLimit, "keys-limit", config.KeysLimit, "keys limit for the dice server. "+
		"This flag controls the number of keys each shard holds at startup. You can multiply this number with the "+
		"total number of shard threads to estimate how much memory will be required at system start up.")
	flag.Parse()

	config.SetupConfig()
}

func main() {
	logr := logger.New(logger.Opts{WithTimestamp: true})
	slog.SetDefault(logr)

	ctx, cancel := context.WithCancel(context.Background())

	// Handle SIGTERM and SIGINT
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)

	watchChan := make(chan dstore.QueryWatchEvent, config.DiceConfig.Server.KeysLimit)
	var serverErrCh chan error

	// Get the number of available CPU cores on the machine using runtime.NumCPU().
	// This determines the total number of logical processors that can be utilized
	// for parallel execution. Setting the maximum number of CPUs to the available
	// core count ensures the application can make full use of all available hardware.
	// If multithreading is not enabled, server will run on a single core.
	var numCores int
	if config.EnableMultiThreading {
		serverErrCh = make(chan error, 1)
		numCores = runtime.NumCPU()
		logr.Debug("The DiceDB server has started in multi-threaded mode.", slog.Int("number of cores", numCores))
	} else {
		serverErrCh = make(chan error, 2)
		logr.Debug("The DiceDB server has started in single-threaded mode.")
		numCores = 1
	}

	// The runtime.GOMAXPROCS(numCores) call limits the number of operating system
	// threads that can execute Go code simultaneously to the number of CPU cores.
	// This enables Go to run more efficiently, maximizing CPU utilization and
	// improving concurrency performance across multiple goroutines.
	runtime.GOMAXPROCS(numCores)

	// Initialize the ShardManager
	shardManager := shard.NewShardManager(uint8(numCores), watchChan, serverErrCh, logr)

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		shardManager.Run(ctx)
	}()

	var serverWg sync.WaitGroup

	// Initialize the AsyncServer server
	// Find a port and bind it
	if !config.EnableMultiThreading {
		asyncServer := server.NewAsyncServer(shardManager, watchChan, logr)
		if err := asyncServer.FindPortAndBind(); err != nil {
			cancel()
			logr.Error("Error finding and binding port", slog.Any("error", err))
			os.Exit(1)
		}

		serverWg.Add(1)
		go func() {
			defer serverWg.Done()
			// Run the server
			err := asyncServer.Run(ctx)

			// Handling different server errors
			if err != nil {
				if errors.Is(err, context.Canceled) {
					logr.Debug("Server was canceled")
				} else if errors.Is(err, diceerrors.ErrAborted) {
					logr.Debug("Server received abort command")
				} else {
					logr.Error(
						"Server error",
						slog.Any("error", err),
					)
				}
				serverErrCh <- err
			} else {
				logr.Debug("Server stopped without error")
			}
		}()

		// Goroutine to handle shutdown signals
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-sigs
			asyncServer.InitiateShutdown()
			cancel()
		}()

		// Initialize the HTTP server
		httpServer := server.NewHTTPServer(shardManager, logr)
		serverWg.Add(1)
		go func() {
			defer serverWg.Done()
			// Run the HTTP server
			err := httpServer.Run(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					logr.Debug("HTTP Server was canceled")
				} else if errors.Is(err, diceerrors.ErrAborted) {
					logr.Debug("HTTP received abort command")
				} else {
					logr.Error("HTTP Server error", slog.Any("error", err))
				}
				serverErrCh <- err
			} else {
				logr.Debug("HTTP Server stopped without error")
			}
		}()
	} else {
		workerManager := worker.NewWorkerManager(config.DiceConfig.Server.MaxClients, shardManager)
		// Initialize the RESP Server
		respServer := resp.NewServer(shardManager, workerManager, serverErrCh, logr)
		serverWg.Add(1)
		go func() {
			defer serverWg.Done()
			// Run the server
			err := respServer.Run(ctx)

			// Handling different server errors
			if err != nil {
				if errors.Is(err, context.Canceled) {
					logr.Debug("Server was canceled")
				} else if errors.Is(err, diceerrors.ErrAborted) {
					logr.Debug("Server received abort command")
				} else {
					logr.Error("Server error", "error", err)
				}
				serverErrCh <- err
			} else {
				logr.Debug("Server stopped without error")
			}
		}()

		// Goroutine to handle shutdown signals
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-sigs
			respServer.Shutdown()
			cancel()
		}()
	}

	websocketServer := server.NewWebSocketServer(shardManager, watchChan, logr)
	serverWg.Add(1)
	go func() {
		defer serverWg.Done()
		// Run the Websocket server
		err := websocketServer.Run(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				logr.Debug("Websocket Server was canceled")
			} else if errors.Is(err, diceerrors.ErrAborted) {
				logr.Debug("Websocket received abort command")
			} else {
				logr.Error("Websocket Server error", "error", err)
			}
			serverErrCh <- err
		} else {
			logr.Debug("Websocket Server stopped without error")
		}
	}()

	go func() {
		serverWg.Wait()
		close(serverErrCh) // Close the channel when both servers are done
	}()

	for err := range serverErrCh {
		if err != nil && errors.Is(err, diceerrors.ErrAborted) {
			// if either the AsyncServer/RESPServer or the HTTPServer received an abort command,
			// cancel the context, helping gracefully exiting all servers
			cancel()
		}
	}

	close(sigs)
	cancel()

	wg.Wait()
	logr.Debug("Server has shut down gracefully")
}

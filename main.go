package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"sync"
	"syscall"

	"github.com/dicedb/dice/internal/logger"
	"github.com/dicedb/dice/internal/server/abstractserver"

	"github.com/dicedb/dice/config"
	diceerrors "github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/observability"
	"github.com/dicedb/dice/internal/server"
	"github.com/dicedb/dice/internal/server/resp"
	"github.com/dicedb/dice/internal/shard"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dice/internal/worker"
)

func init() {
	flag.StringVar(&config.Host, "host", "0.0.0.0", "host for the DiceDB server")

	flag.IntVar(&config.Port, "port", 7379, "port for the DiceDB server")

	flag.IntVar(&config.HTTPPort, "http-port", 7380, "port for accepting requets over HTTP")
	flag.BoolVar(&config.EnableHTTP, "enable-http", false, "enable DiceDB to listen, accept, and process HTTP")

	flag.IntVar(&config.WebsocketPort, "websocket-port", 7381, "port for accepting requets over WebSocket")
	flag.BoolVar(&config.EnableWebsocket, "enable-websocket", false, "enable DiceDB to listen, accept, and process WebSocket")

	flag.BoolVar(&config.EnableMultiThreading, "enable-multithreading", false, "enable multithreading execution and leverage multiple CPU cores")
	flag.IntVar(&config.NumShards, "num-shards", -1, "number shards to create. defaults to number of cores")

	flag.BoolVar(&config.EnableWatch, "enable-watch", false, "enable support for .WATCH commands and real-time reactivity")
	flag.BoolVar(&config.EnableProfiling, "enable-profiling", false, "enable profiling and capture critical metrics and traces in .prof files")

	flag.StringVar(&config.DiceConfig.Logging.LogLevel, "log-level", "info", "log level, values: info, debug")

	flag.StringVar(&config.RequirePass, "requirepass", config.RequirePass, "enable authentication for the default user")
	flag.StringVar(&config.CustomConfigFilePath, "o", config.CustomConfigFilePath, "dir path to create the config file")
	flag.StringVar(&config.FileLocation, "c", config.FileLocation, "file path of the config file")
	flag.BoolVar(&config.InitConfigCmd, "init-config", false, "initialize a new config file")
	flag.IntVar(&config.KeysLimit, "keys-limit", config.KeysLimit, "keys limit for the DiceDB server. "+
		"This flag controls the number of keys each shard holds at startup. You can multiply this number with the "+
		"total number of shard threads to estimate how much memory will be required at system start up.")

	flag.Parse()

	config.SetupConfig()

	iid := observability.GetOrCreateInstanceID()
	config.DiceConfig.InstanceID = iid

	slog.SetDefault(logger.New())
}

func main() {
	fmt.Print(`
██████╗ ██╗ ██████╗███████╗██████╗ ██████╗ 
██╔══██╗██║██╔════╝██╔════╝██╔══██╗██╔══██╗
██║  ██║██║██║     █████╗  ██║  ██║██████╔╝
██║  ██║██║██║     ██╔══╝  ██║  ██║██╔══██╗
██████╔╝██║╚██████╗███████╗██████╔╝██████╔╝
╚═════╝ ╚═╝ ╚═════╝╚══════╝╚═════╝ ╚═════╝

`)
	slog.Info("starting DiceDB", slog.String("version", config.DiceDBVersion))
	slog.Info("running with", slog.Int("port", config.Port))
	slog.Info("running with", slog.Bool("enable-watch", config.EnableWatch))

	if config.EnableProfiling {
		slog.Info("running with", slog.Bool("enable-profiling", config.EnableProfiling))
	}

	go observability.Ping()

	ctx, cancel := context.WithCancel(context.Background())

	// Handle SIGTERM and SIGINT
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)

	var (
		queryWatchChan chan dstore.QueryWatchEvent
		cmdWatchChan   chan dstore.CmdWatchEvent
		serverErrCh    = make(chan error, 2)
	)

	if config.EnableWatch {
		bufSize := config.DiceConfig.Performance.WatchChanBufSize
		queryWatchChan = make(chan dstore.QueryWatchEvent, bufSize)
		cmdWatchChan = make(chan dstore.CmdWatchEvent, bufSize)
	}

	// Get the number of available CPU cores on the machine using runtime.NumCPU().
	// This determines the total number of logical processors that can be utilized
	// for parallel execution. Setting the maximum number of CPUs to the available
	// core count ensures the application can make full use of all available hardware.
	// If multithreading is not enabled, server will run on a single core.
	var numShards int
	if config.EnableMultiThreading {
		numShards = runtime.NumCPU()
		if config.NumShards > 0 {
			numShards = config.NumShards
		}
		slog.Info("running with", slog.String("mode", "multi-threaded"), slog.Int("num-shards", numShards))
	} else {
		numShards = 1
		slog.Info("running with", slog.String("mode", "single-threaded"))
	}

	// The runtime.GOMAXPROCS(numShards) call limits the number of operating system
	// threads that can execute Go code simultaneously to the number of CPU cores.
	// This enables Go to run more efficiently, maximizing CPU utilization and
	// improving concurrency performance across multiple goroutines.
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Initialize the ShardManager
	shardManager := shard.NewShardManager(uint8(numShards), queryWatchChan, cmdWatchChan, serverErrCh)

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		shardManager.Run(ctx)
	}()

	var serverWg sync.WaitGroup

	if config.EnableMultiThreading {
		if config.EnableProfiling {
			stopProfiling, err := startProfiling()
			if err != nil {
				slog.Error("Profiling could not be started", slog.Any("error", err))
				sigs <- syscall.SIGKILL
			}
			defer stopProfiling()
		}

		workerManager := worker.NewWorkerManager(config.DiceConfig.Performance.MaxClients, shardManager)
		respServer := resp.NewServer(shardManager, workerManager, cmdWatchChan, serverErrCh)
		serverWg.Add(1)
		go runServer(ctx, &serverWg, respServer, serverErrCh)
	} else {
		asyncServer := server.NewAsyncServer(shardManager, queryWatchChan)
		if err := asyncServer.FindPortAndBind(); err != nil {
			slog.Error("Error finding and binding port", slog.Any("error", err))
			sigs <- syscall.SIGKILL
		}

		serverWg.Add(1)
		go runServer(ctx, &serverWg, asyncServer, serverErrCh)

		if config.EnableHTTP {
			httpServer := server.NewHTTPServer(shardManager)
			serverWg.Add(1)
			go runServer(ctx, &serverWg, httpServer, serverErrCh)
		}
	}

	if config.EnableWebsocket {
		websocketServer := server.NewWebSocketServer(shardManager, config.WebsocketPort)
		serverWg.Add(1)
		go runServer(ctx, &serverWg, websocketServer, serverErrCh)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-sigs
		cancel()
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
}

func runServer(ctx context.Context, wg *sync.WaitGroup, srv abstractserver.AbstractServer, errCh chan<- error) {
	defer wg.Done()
	if err := srv.Run(ctx); err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			slog.Debug(fmt.Sprintf("%T was canceled", srv))
		case errors.Is(err, diceerrors.ErrAborted):
			slog.Debug(fmt.Sprintf("%T received abort command", srv))
		case errors.Is(err, http.ErrServerClosed):
			slog.Debug(fmt.Sprintf("%T received abort command", srv))
		default:
			slog.Error(fmt.Sprintf("%T error", srv), slog.Any("error", err))
		}
		errCh <- err
	} else {
		slog.Debug("bye.")
	}
}
func startProfiling() (func(), error) {
	// Start CPU profiling
	cpuFile, err := os.Create("cpu.prof")
	if err != nil {
		return nil, fmt.Errorf("could not create cpu.prof: %w", err)
	}

	if err = pprof.StartCPUProfile(cpuFile); err != nil {
		cpuFile.Close()
		return nil, fmt.Errorf("could not start CPU profile: %w", err)
	}

	// Start memory profiling
	memFile, err := os.Create("mem.prof")
	if err != nil {
		pprof.StopCPUProfile()
		cpuFile.Close()
		return nil, fmt.Errorf("could not create mem.prof: %w", err)
	}

	// Start block profiling
	runtime.SetBlockProfileRate(1)

	// Start execution trace
	traceFile, err := os.Create("trace.out")
	if err != nil {
		runtime.SetBlockProfileRate(0)
		memFile.Close()
		pprof.StopCPUProfile()
		cpuFile.Close()
		return nil, fmt.Errorf("could not create trace.out: %w", err)
	}

	if err := trace.Start(traceFile); err != nil {
		traceFile.Close()
		runtime.SetBlockProfileRate(0)
		memFile.Close()
		pprof.StopCPUProfile()
		cpuFile.Close()
		return nil, fmt.Errorf("could not start trace: %w", err)
	}

	// Return a cleanup function
	return func() {
		// Stop the CPU profiling and close cpuFile
		pprof.StopCPUProfile()
		cpuFile.Close()

		// Write heap profile
		runtime.GC()
		if err := pprof.WriteHeapProfile(memFile); err != nil {
			slog.Warn("could not write memory profile", slog.Any("error", err))
		}

		memFile.Close()

		// Write block profile
		blockFile, err := os.Create("block.prof")
		if err != nil {
			slog.Warn("could not create block profile", slog.Any("error", err))
		} else {
			if err := pprof.Lookup("block").WriteTo(blockFile, 0); err != nil {
				slog.Warn("could not write block profile", slog.Any("error", err))
			}
			blockFile.Close()
		}

		runtime.SetBlockProfileRate(0)

		// Stop trace and close traceFile
		trace.Stop()
		traceFile.Close()
	}, nil
}

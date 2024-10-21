package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/dicedb/dice/internal/server/abstractserver"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"sync"
	"syscall"

	"github.com/dicedb/dice/config"
	diceerrors "github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/logger"
	"github.com/dicedb/dice/internal/observability"
	"github.com/dicedb/dice/internal/server"
	"github.com/dicedb/dice/internal/server/resp"
	"github.com/dicedb/dice/internal/shard"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dice/internal/worker"
)

func init() {
	flag.StringVar(&config.Host, "host", "0.0.0.0", "host for the dicedb server")
	flag.IntVar(&config.Port, "port", 7379, "port for the dicedb server")
	flag.BoolVar(&config.EnableHTTP, "enable-http", true, "run server in HTTP mode as well")
	flag.BoolVar(&config.EnableMultiThreading, "enable-multithreading", false, "run server in multithreading mode")
	flag.IntVar(&config.HTTPPort, "http-port", 8082, "HTTP port for the dicedb server")
	flag.IntVar(&config.WebsocketPort, "websocket-port", 8379, "Websocket port for the dicedb server")
	flag.StringVar(&config.RequirePass, "requirepass", config.RequirePass, "enable authentication for the default user")
	flag.StringVar(&config.CustomConfigFilePath, "o", config.CustomConfigFilePath, "dir path to create the config file")
	flag.StringVar(&config.FileLocation, "c", config.FileLocation, "file path of the config file")
	flag.BoolVar(&config.InitConfigCmd, "init-config", false, "initialize a new config file")
	flag.IntVar(&config.KeysLimit, "keys-limit", config.KeysLimit, "keys limit for the dicedb server. "+
		"This flag controls the number of keys each shard holds at startup. You can multiply this number with the "+
		"total number of shard threads to estimate how much memory will be required at system start up.")
	flag.BoolVar(&config.EnableProfiling, "enable-profiling", false, "enable profiling for the dicedb server")
	flag.BoolVar(&config.EnableWatch, "enable-watch", false, "enable reactivity features which power the .WATCH commands")

	flag.Parse()

	config.SetupConfig()

	iid := observability.GetOrCreateInstanceID()
	config.DiceConfig.InstanceID = iid
}

func main() {
	logr := logger.New(logger.Opts{WithTimestamp: true})
	slog.SetDefault(logr)

	go observability.Ping(logr)

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
	var numCores int
	if config.EnableMultiThreading {
		numCores = runtime.NumCPU()
		logr.Debug("The DiceDB server has started in multi-threaded mode.", slog.Int("number of cores", numCores))
	} else {
		logr.Debug("The DiceDB server has started in single-threaded mode.")
	}

	// The runtime.GOMAXPROCS(numCores) call limits the number of operating system
	// threads that can execute Go code simultaneously to the number of CPU cores.
	// This enables Go to run more efficiently, maximizing CPU utilization and
	// improving concurrency performance across multiple goroutines.
	runtime.GOMAXPROCS(numCores)

	// Initialize the ShardManager
	shardManager := shard.NewShardManager(uint8(numCores), queryWatchChan, cmdWatchChan, serverErrCh, logr)

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		shardManager.Run(ctx)
	}()

	var serverWg sync.WaitGroup

	if config.EnableMultiThreading {
		if config.EnableProfiling {
			stopProfiling, err := startProfiling(logr)
			if err != nil {
				logr.Error("Profiling could not be started", slog.Any("error", err))
				sigs <- syscall.SIGKILL
			}
			defer stopProfiling()
		}

		workerManager := worker.NewWorkerManager(config.DiceConfig.Performance.MaxClients, shardManager)
		respServer := resp.NewServer(shardManager, workerManager, cmdWatchChan, serverErrCh, logr)
		serverWg.Add(1)
		go runServer(ctx, &serverWg, respServer, logr, serverErrCh)
	} else {
		asyncServer := server.NewAsyncServer(shardManager, queryWatchChan, logr)
		if err := asyncServer.FindPortAndBind(); err != nil {
			logr.Error("Error finding and binding port", slog.Any("error", err))
			sigs <- syscall.SIGKILL
		}

		serverWg.Add(1)
		go runServer(ctx, &serverWg, asyncServer, logr, serverErrCh)

		httpServer := server.NewHTTPServer(shardManager, logr)
		serverWg.Add(1)
		go runServer(ctx, &serverWg, httpServer, logr, serverErrCh)
	}

	websocketServer := server.NewWebSocketServer(shardManager, config.WebsocketPort, logr)
	serverWg.Add(1)
	go runServer(ctx, &serverWg, websocketServer, logr, serverErrCh)

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
	logr.Debug("Server has shut down gracefully")
}

func runServer(ctx context.Context, wg *sync.WaitGroup, srv abstractserver.AbstractServer, logr *slog.Logger, errCh chan<- error) {
	defer wg.Done()
	if err := srv.Run(ctx); err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			logr.Debug(fmt.Sprintf("%T was canceled", srv))
		case errors.Is(err, diceerrors.ErrAborted):
			logr.Debug(fmt.Sprintf("%T received abort command", srv))
		default:
			logr.Error(fmt.Sprintf("%T error", srv), slog.Any("error", err))
		}
		errCh <- err
	} else {
		logr.Debug(fmt.Sprintf("%T stopped without error", srv))
	}
}
func startProfiling(logr *slog.Logger) (func(), error) {
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
			logr.Warn("could not write memory profile", slog.Any("error", err))
		}

		memFile.Close()

		// Write block profile
		blockFile, err := os.Create("block.prof")
		if err != nil {
			logr.Warn("could not create block profile", slog.Any("error", err))
		} else {
			if err := pprof.Lookup("block").WriteTo(blockFile, 0); err != nil {
				logr.Warn("could not write block profile", slog.Any("error", err))
			}
			blockFile.Close()
		}

		runtime.SetBlockProfileRate(0)

		// Stop trace and close traceFile
		trace.Stop()
		traceFile.Close()
	}, nil
}

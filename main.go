package main

import (
	"context"
	"errors"
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
	"time"

	"github.com/dicedb/dice/internal/cli"
	"github.com/dicedb/dice/internal/logger"
	"github.com/dicedb/dice/internal/server/abstractserver"
	"github.com/dicedb/dice/internal/wal"
	"github.com/dicedb/dice/internal/watchmanager"

	"github.com/dicedb/dice/config"
	diceerrors "github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/iothread"
	"github.com/dicedb/dice/internal/observability"
	"github.com/dicedb/dice/internal/server"
	"github.com/dicedb/dice/internal/server/resp"
	"github.com/dicedb/dice/internal/shard"
	dstore "github.com/dicedb/dice/internal/store"
)

func main() {
	iid := observability.GetOrCreateInstanceID()
	config.DiceConfig.InstanceID = iid
	slog.SetDefault(logger.New())
	cli.Execute()
	go observability.Ping()

	ctx, cancel := context.WithCancel(context.Background())

	// Handle SIGTERM and SIGINT
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)

	var (
		queryWatchChan           chan dstore.QueryWatchEvent
		cmdWatchChan             chan dstore.CmdWatchEvent
		serverErrCh              = make(chan error, 2)
		cmdWatchSubscriptionChan = make(chan watchmanager.WatchSubscription)
		wl                       wal.AbstractWAL
	)

	wl, _ = wal.NewNullWAL()
	if config.DiceConfig.Persistence.Enabled {
		if config.DiceConfig.Persistence.WALEngine == "sqlite" {
			_wl, err := wal.NewSQLiteWAL(config.DiceConfig.Persistence.WALDir)
			if err != nil {
				slog.Warn("could not create WAL with", slog.String("wal-engine", config.DiceConfig.Persistence.WALEngine), slog.Any("error", err))
				sigs <- syscall.SIGKILL
				return
			}
			wl = _wl
		} else if config.DiceConfig.Persistence.WALEngine == "aof" {
			_wl, err := wal.NewAOFWAL(config.DiceConfig.Persistence.WALDir)
			if err != nil {
				slog.Warn("could not create WAL with", slog.String("wal-engine", config.DiceConfig.Persistence.WALEngine), slog.Any("error", err))
				sigs <- syscall.SIGKILL
				return
			}
			wl = _wl
		} else {
			slog.Error("unsupported WAL engine", slog.String("engine", config.DiceConfig.Persistence.WALEngine))
			sigs <- syscall.SIGKILL
			return
		}

		if err := wl.Init(time.Now()); err != nil {
			slog.Error("could not initialize WAL", slog.Any("error", err))
		} else {
			go wal.InitBG(wl)
		}

		slog.Debug("WAL initialization complete")

		if config.DiceConfig.Persistence.RestoreFromWAL {
			slog.Info("restoring database from WAL")
			wal.ReplayWAL(wl)
			slog.Info("database restored from WAL")
		}
	}

	if config.DiceConfig.Performance.EnableWatch {
		bufSize := config.DiceConfig.Performance.WatchChanBufSize
		queryWatchChan = make(chan dstore.QueryWatchEvent, bufSize)
		cmdWatchChan = make(chan dstore.CmdWatchEvent, bufSize)
	}

	// Get the number of available CPU cores on the machine using runtime.NumCPU().
	// This determines the total number of logical processors that can be utilized
	// for parallel execution. Setting the maximum number of CPUs to the available
	// core count ensures the application can make full use of all available hardware.
	var numShards int
	numShards = runtime.NumCPU()
	if config.DiceConfig.Performance.NumShards > 0 {
		numShards = config.DiceConfig.Performance.NumShards
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

	if config.DiceConfig.Performance.EnableProfiling {
		stopProfiling, err := startProfiling()
		if err != nil {
			slog.Error("Profiling could not be started", slog.Any("error", err))
			sigs <- syscall.SIGKILL
		}
		defer stopProfiling()
	}
	ioThreadManager := iothread.NewManager(config.DiceConfig.Performance.MaxClients, shardManager)
	respServer := resp.NewServer(shardManager, ioThreadManager, cmdWatchSubscriptionChan, cmdWatchChan, serverErrCh, wl)
	serverWg.Add(1)
	go runServer(ctx, &serverWg, respServer, serverErrCh)

	if config.DiceConfig.HTTP.Enabled {
		httpServer := server.NewHTTPServer(shardManager, wl)
		serverWg.Add(1)
		go runServer(ctx, &serverWg, httpServer, serverErrCh)
	}

	if config.DiceConfig.WebSocket.Enabled {
		websocketServer := server.NewWebSocketServer(shardManager, config.DiceConfig.WebSocket.Port, wl)
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

	if config.DiceConfig.Persistence.Enabled {
		wal.ShutdownBG()
	}

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

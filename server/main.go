// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package server

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
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/dicedb/dice/internal/auth"
	"github.com/dicedb/dice/internal/cmd"
	"github.com/dicedb/dice/internal/server/ironhawk"
	"github.com/dicedb/dice/internal/shardmanager"
	"github.com/dicedb/dicedb-go/wire"

	"github.com/dicedb/dice/internal/wal"

	"github.com/dicedb/dice/config"
	diceerrors "github.com/dicedb/dice/internal/errors"
)

func printConfiguration() {
	slog.Info("starting DiceDB", slog.String("version", config.DiceDBVersion))
	slog.Info("running with", slog.Int("total_commands", cmd.Total()))
	slog.Info("running with", slog.String("engine", config.Config.Engine))
	slog.Info("running with", slog.Int("port", config.Config.Port))
	slog.Info("running on", slog.Int("cores", runtime.NumCPU()))

	// Conditionally add the number of shards to be used for DiceDB
	numShards := runtime.NumCPU()
	if config.Config.NumShards > 0 {
		numShards = config.Config.NumShards
	}
	slog.Info("running with", slog.Int("shards", numShards))
}

func printBanner() {
	fmt.Print(`
	██████╗ ██╗ ██████╗███████╗██████╗ ██████╗ 
	██╔══██╗██║██╔════╝██╔════╝██╔══██╗██╔══██╗
	██║  ██║██║██║     █████╗  ██║  ██║██████╔╝
	██║  ██║██║██║     ██╔══╝  ██║  ██║██╔══██╗
	██████╔╝██║╚██████╗███████╗██████╔╝██████╔╝
	╚═════╝ ╚═╝ ╚═════╝╚══════╝╚═════╝ ╚═════╝
			
`)
}

const EngineRESP = "resp"
const EngineIRONHAWK = "ironhawk"
const EngineSILVERPINE = "silverpine"

func Start() {
	printBanner()
	printConfiguration()

	user, err := auth.UserStore.Add(config.Config.Username)
	if err != nil {

		//log errors like for example : if modified to prevent duplicates later

		slog.Error("Failed to add default user to user store",
			slog.String("username", config.Config.Username),
			slog.Any("error", err))

		// Consider exiting if the default user cannot be created:
		// os.Exit(1)
	} else {
		//  set the password only if one is provided.
		if config.Config.Password != "" {
			if err := user.SetPassword(config.Config.Password); err != nil {
				// Log an error if password hashing/setting fails.
				slog.Error("Failed to set password for default user",
					slog.String("username", config.Config.Username),
					slog.Any("error", err))
				// Consider exiting if password setting is critical and fails:
				// os.Exit(1)
			}
		} else {
			// log a warning if starting without a password for the default user.
			// clear security implication clear.
			slog.Warn("Starting server without a password configured for the default user.",
				slog.String("username", config.Config.Username),
				slog.String("security_implication", "Authentication may not be required for this user."),
				slog.String("recommendation", "Consider setting a password using the --password flag or config file."))
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Handle SIGTERM and SIGINT
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)

	var (
		serverErrCh = make(chan error, 2)
		wl          wal.AbstractWAL
	)

	wl, _ = wal.NewNullWAL()
	if config.Config.EnableWAL {
		_wl, err := wal.NewAOFWAL(config.Config.WALDir)
		if err != nil {
			slog.Warn("could not create WAL at", slog.String("wal-dir", config.Config.WALDir), slog.Any("error", err))
			sigs <- syscall.SIGKILL
			cancel()
			return
		}
		wl = _wl

		if err := wl.Init(time.Now()); err != nil {
			slog.Error("could not initialize WAL", slog.Any("error", err))
		} else {
			go wal.InitBG(wl)
		}

		slog.Debug("WAL initialization complete")
	}

	// Get the number of available CPU cores on the machine using runtime.NumCPU().
	// This determines the total number of logical processors that can be utilized
	// for parallel execution. Setting the maximum number of CPUs to the available
	// core count ensures the application can make full use of all available hardware.
	var numShards int
	numShards = runtime.NumCPU()
	if config.Config.NumShards > 0 {
		numShards = config.Config.NumShards
	}

	// The runtime.GOMAXPROCS(numShards) call limits the number of operating system
	// threads that can execute Go code simultaneously to the number of CPU cores.
	// This enables Go to run more efficiently, maximizing CPU utilization and
	// improving concurrency performance across multiple goroutines.
	runtime.GOMAXPROCS(runtime.NumCPU())

	shardManager := shardmanager.NewShardManager(numShards, serverErrCh)
	watchManager := ironhawk.NewWatchManager()

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		shardManager.Run(ctx)
	}()

	var serverWg sync.WaitGroup

	if config.EnableProfile {
		stopProfiling, err := startProfiling()
		if err != nil {
			slog.Error("Profiling could not be started", slog.Any("error", err))
			sigs <- syscall.SIGKILL
		}
		defer stopProfiling()
	}

	ioThreadManager := ironhawk.NewIOThreadManager()
	ironhawkServer := ironhawk.NewServer(shardManager, ioThreadManager, watchManager)

	serverWg.Add(1)
	go runServer(ctx, &serverWg, ironhawkServer, serverErrCh)

	// Recovery from WAL logs
	if config.Config.EnableWAL {
		slog.Info("restoring database from WAL")
		callback := func(entry *wal.WALEntry) error {
			command := strings.Split(string(entry.Data), " ")
			cmdTemp := cmd.Cmd{
				C: &wire.Command{
					Cmd:  command[0],
					Args: command[1:],
				},
				IsReplay: true,
			}
			_, err := cmdTemp.Execute(shardManager)
			if err != nil {
				return fmt.Errorf("error handling WAL replay: %w", err)
			}
			return nil
		}
		if err := wl.Replay(callback); err != nil {
			slog.Error("error restoring from WAL", slog.Any("error", err))
		}
		slog.Info("database restored from WAL")
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

	if config.Config.EnableWAL {
		wal.ShutdownBG()
	}

	cancel()

	wg.Wait()
}

func runServer(ctx context.Context, wg *sync.WaitGroup, srv *ironhawk.Server, errCh chan<- error) {
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

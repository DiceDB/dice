package resp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"sync"
	"time"

	"github.com/dicedb/dice/internal/server/resp"
	"github.com/dicedb/dice/internal/worker"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/internal/clientio"
	derrors "github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/logger"
	"github.com/dicedb/dice/internal/shard"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dice/testutils"
	dicedb "github.com/dicedb/dicedb-go"
)

type TestServerOptions struct {
	Port int
}

//nolint:unused
func getLocalConnection() net.Conn {
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", config.DiceConfig.AsyncServer.Port))
	if err != nil {
		panic(err)
	}
	return conn
}

// deleteTestKeys is a utility to delete a list of keys before running a test
//
//nolint:unused
func deleteTestKeys(keysToDelete []string, store *dstore.Store) {
	for _, key := range keysToDelete {
		store.Del(key)
	}
}

//nolint:unused
func getLocalSdk() *dicedb.Client {
	return dicedb.NewClient(&dicedb.Options{
		Addr: fmt.Sprintf(":%d", config.DiceConfig.AsyncServer.Port),

		DialTimeout:           10 * time.Second,
		ReadTimeout:           30 * time.Second,
		WriteTimeout:          30 * time.Second,
		ContextTimeoutEnabled: true,

		MaxRetries: -1,

		PoolSize:        10,
		PoolTimeout:     30 * time.Second,
		ConnMaxIdleTime: time.Minute,
	})
}

func FireCommand(conn net.Conn, cmd string) interface{} {
	var err error
	args := testutils.ParseCommand(cmd)
	_, err = conn.Write(clientio.Encode(args, false))
	if err != nil {
		slog.Error(
			"error while firing command",
			slog.Any("error", err),
			slog.String("command", cmd),
		)
		os.Exit(1)
	}

	rp := clientio.NewRESPParser(conn)
	v, err := rp.DecodeOne()
	if err != nil {
		if err == io.EOF {
			return nil
		}
		slog.Error(
			"error while firing command",
			slog.Any("error", err),
			slog.String("command", cmd),
		)
		os.Exit(1)
	}

	return v
}

//nolint:unused
func fireCommandAndGetRESPParser(conn net.Conn, cmd string) *clientio.RESPParser {
	args := testutils.ParseCommand(cmd)
	_, err := conn.Write(clientio.Encode(args, false))
	if err != nil {
		slog.Error(
			"error while firing command",
			slog.Any("error", err),
			slog.String("command", cmd),
		)
		os.Exit(1)
	}

	return clientio.NewRESPParser(conn)
}

func RunTestServer(wg *sync.WaitGroup, opt TestServerOptions) {
	slog.SetDefault(logger.New())
	config.DiceConfig.Network.IOBufferLength = 16
	config.DiceConfig.Persistence.WriteAOFOnCleanup = false
	if opt.Port != 0 {
		config.DiceConfig.AsyncServer.Port = opt.Port
	} else {
		config.DiceConfig.AsyncServer.Port = 9739
	}

	queryWatchChan := make(chan dstore.QueryWatchEvent, config.DiceConfig.Memory.KeysLimit)
	cmdWatchChan := make(chan dstore.CmdWatchEvent, config.DiceConfig.Memory.KeysLimit)
	gec := make(chan error)
	shardManager := shard.NewShardManager(1, queryWatchChan, cmdWatchChan, gec)
	workerManager := worker.NewWorkerManager(20000, shardManager)
	// Initialize the RESP Server
	testServer := resp.NewServer(shardManager, workerManager, cmdWatchChan, gec)

	ctx, cancel := context.WithCancel(context.Background())
	fmt.Println("Starting the test server on port", config.DiceConfig.AsyncServer.Port)

	shardManagerCtx, cancelShardManager := context.WithCancel(ctx)
	wg.Add(1)
	go func() {
		defer wg.Done()
		shardManager.Run(shardManagerCtx)
	}()

	// Start the server in a goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := testServer.Run(ctx); err != nil {
			if errors.Is(err, derrors.ErrAborted) {
				cancelShardManager()
				return
			}
			slog.Error("Test server encountered an error", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	go func() {
		for err := range gec {
			if err != nil && errors.Is(err, derrors.ErrAborted) {
				// if either the AsyncServer/RESPServer or the HTTPServer received an abort command,
				// cancel the context, helping gracefully exiting all servers
				cancel()
			}
		}
	}()
}

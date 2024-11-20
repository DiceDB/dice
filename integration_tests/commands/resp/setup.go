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
	"github.com/dicedb/dice/internal/wal"
	"github.com/dicedb/dice/internal/watchmanager"
	"github.com/dicedb/dice/internal/worker"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/internal/clientio"
	derrors "github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/shard"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dice/testutils"
	dicedb "github.com/dicedb/dicedb-go"
)

type TestServerOptions struct {
	Port int
}

// getLocalConnection returns a local TCP connection to the database
//
//nolint:unused
func getLocalConnection() net.Conn {
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", config.DiceConfig.AsyncServer.Port))
	if err != nil {
		panic(err)
	}
	return conn
}

func ClosePublisherSubscribers(publisher net.Conn, subscribers []net.Conn) error {
	if err := publisher.Close(); err != nil {
		return fmt.Errorf("error closing publisher connection: %v", err)
	}
	for _, sub := range subscribers {
		time.Sleep(100 * time.Millisecond) // [TODO] why is this needed?
		if err := sub.Close(); err != nil {
			return fmt.Errorf("error closing subscriber connection: %v", err)
		}
	}
	return nil
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

type WatchSubscriber struct {
	client *dicedb.Client
	watch  *dicedb.WatchConn
}

func ClosePublisherSubscribersSDK(publisher *dicedb.Client, subscribers []WatchSubscriber) error {
	if err := publisher.Close(); err != nil {
		return fmt.Errorf("error closing publisher connection: %v", err)
	}
	for _, sub := range subscribers {
		if err := sub.watch.Close(); err != nil {
			return fmt.Errorf("error closing subscriber watch connection: %v", err)
		}
		if err := sub.client.Close(); err != nil {
			return fmt.Errorf("error closing subscriber connection: %v", err)
		}
	}
	return nil
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
	config.DiceConfig.Network.IOBufferLength = 16
	config.DiceConfig.Persistence.WriteAOFOnCleanup = false

	// #1261: Added here to prevent resp integration tests from failing on lower-spec machines
	config.DiceConfig.Memory.KeysLimit = 2000
	if opt.Port != 0 {
		config.DiceConfig.AsyncServer.Port = opt.Port
	} else {
		config.DiceConfig.AsyncServer.Port = 9739
	}

	queryWatchChan := make(chan dstore.QueryWatchEvent, config.DiceConfig.Performance.WatchChanBufSize)
	cmdWatchChan := make(chan dstore.CmdWatchEvent, config.DiceConfig.Performance.WatchChanBufSize)
	cmdWatchSubscriptionChan := make(chan watchmanager.WatchSubscription)
	gec := make(chan error)
	shardManager := shard.NewShardManager(1, queryWatchChan, cmdWatchChan, gec)
	workerManager := worker.NewWorkerManager(20000, shardManager)
	// Initialize the RESP Server
	wl, _ := wal.NewNullWAL()
	testServer := resp.NewServer(shardManager, workerManager, cmdWatchSubscriptionChan, cmdWatchChan, gec, wl)

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

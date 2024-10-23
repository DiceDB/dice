package async

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

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/internal/clientio"
	derrors "github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/server"
	"github.com/dicedb/dice/internal/shard"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dice/testutils"
	dicedb "github.com/dicedb/dicedb-go"
)

type TestServerOptions struct {
	Port       int
	MaxClients int32
}

//nolint:unused
func getLocalConnection() net.Conn {
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", config.DiceConfig.AsyncServer.Port))
	if err != nil {
		panic(err)
	}
	return conn
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

func RunTestServer(ctx context.Context, wg *sync.WaitGroup, opt TestServerOptions) {
	// Override configs to test server config, this is enabled to handle test env runs
	// as those envs are resource constrained
	config.DiceConfig.Network.IOBufferLength = 16
	config.DiceConfig.Persistence.WriteAOFOnCleanup = false
	config.DiceConfig.Performance.StoreMapInitSize = 1024
	config.DiceConfig.Memory.EvictionRatio = 0.4
	config.DiceConfig.Memory.KeysLimit = 2000000

	if opt.Port != 0 {
		config.DiceConfig.AsyncServer.Port = opt.Port
	} else {
		config.DiceConfig.AsyncServer.Port = 8739
	}

	if opt.MaxClients != 0 {
		config.DiceConfig.Performance.MaxClients = opt.MaxClients
	}

	const totalRetries = 100
	var err error
	watchChan := make(chan dstore.QueryWatchEvent, config.DiceConfig.Memory.KeysLimit)
	gec := make(chan error)
	shardManager := shard.NewShardManager(1, watchChan, nil, gec)
	// Initialize the AsyncServer
	testServer := server.NewAsyncServer(shardManager, watchChan)

	// Try to bind to a port with a maximum of `totalRetries` retries.
	for i := 0; i < totalRetries; i++ {
		if err = testServer.FindPortAndBind(); err == nil {
			break
		}

		if err.Error() == "address already in use" {
			slog.Info("Port already in use, trying port",
				slog.Int("port", config.DiceConfig.AsyncServer.Port),
				slog.Int("new_port", config.DiceConfig.AsyncServer.Port+1),
			)
			config.DiceConfig.AsyncServer.Port++
		} else {
			slog.Error("Failed to bind port", slog.Any("error", err))
			return
		}
	}

	if err != nil {
		slog.Error("Failed to bind to a port after retries",
			slog.Any("error", err),
			slog.Int("retry_count", totalRetries),
		)
		os.Exit(1)
		return
	}

	// Inform the user that the server is starting
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
}

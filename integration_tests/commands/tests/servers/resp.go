package servers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"sync"

	"github.com/dicedb/dice/config"
	derrors "github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/iothread"
	"github.com/dicedb/dice/internal/server/resp"
	"github.com/dicedb/dice/internal/shard"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dice/internal/wal"
	"github.com/dicedb/dice/internal/watchmanager"
)

func GetRespConn() net.Conn {
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", config.DiceConfig.RespServer.Port))
	if err != nil {
		panic(err)
	}
	return conn
}

func RunRespServer(ctx context.Context, wg *sync.WaitGroup, opt TestServerOptions) {
	config.DiceConfig.Network.IOBufferLength = 16
	config.DiceConfig.Persistence.WriteAOFOnCleanup = false

	// #1261: Added here to prevent resp integration tests from failing on lower-spec machines
	config.DiceConfig.Memory.KeysLimit = 2000
	if opt.Port != 0 {
		config.DiceConfig.RespServer.Port = opt.Port
	} else {
		config.DiceConfig.RespServer.Port = 9739
	}

	queryWatchChan := make(chan dstore.QueryWatchEvent, config.DiceConfig.Performance.WatchChanBufSize)
	cmdWatchChan := make(chan dstore.CmdWatchEvent, config.DiceConfig.Performance.WatchChanBufSize)
	cmdWatchSubscriptionChan := make(chan watchmanager.WatchSubscription)
	gec := make(chan error)
	shardManager := shard.NewShardManager(1, queryWatchChan, cmdWatchChan, gec)
	ioThreadManager := iothread.NewManager(20000, shardManager)
	// Initialize the RESP Server
	wl, _ := wal.NewNullWAL()
	testServer := resp.NewServer(shardManager, ioThreadManager, cmdWatchSubscriptionChan, cmdWatchChan, gec, wl)

	fmt.Println("Starting the test RESP server on the port", config.DiceConfig.RespServer.Port)

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
				_ = ctx.Err()
			}
		}
	}()
}

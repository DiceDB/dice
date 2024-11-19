package worker_test

import(
	"testing"
	"runtime"
	"context"
	"sync"
	"log/slog"
	"fmt"
	"errors"
	"net/http"
	"time"

	dice "github.com/dicedb/dicedb-go"
	diceerrors "github.com/dicedb/dice/internal/errors"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dice/internal/watchmanager"
	"github.com/dicedb/dice/internal/wal"
	"github.com/dicedb/dice/internal/shard"
	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/internal/server/resp"
	"github.com/dicedb/dice/internal/server/abstractserver"
	"github.com/dicedb/dice/internal/worker"
)

func BenchmarkWorkerStart(b *testing.B) {
	numShards := runtime.NumCPU()
	runtime.GOMAXPROCS(numShards)

	var queryWatchChan chan dstore.QueryWatchEvent
	var cmdWatchChan   chan dstore.CmdWatchEvent
	var cmdWatchSubscriptionChan = make(chan watchmanager.WatchSubscription)
	var serverErrCh = make(chan error, 2)
	var wl wal.AbstractWAL

	shardManager := shard.NewShardManager(uint8(numShards), queryWatchChan, cmdWatchChan, serverErrCh)

	var ctx, cancel = context.WithCancel(context.Background())

	var wg = sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		shardManager.Run(ctx)
	}()
	
	var serverWg sync.WaitGroup

	workerManager := worker.NewWorkerManager(config.DiceConfig.Performance.MaxClients, shardManager)
	respServer := resp.NewServer(shardManager, workerManager, cmdWatchSubscriptionChan, cmdWatchChan, serverErrCh, wl)
	serverWg.Add(1)
	go runServer(ctx, &serverWg, respServer, serverErrCh)

	time.Sleep(10 * time.Second)

	commands := []string{"PING", "SET foo bar", "GET foo", "INCR counter", "DEL foo"}
	concurrencyLevels := []int{10, 50, 100, 200}

	for _, concurrency := range concurrencyLevels {
		b.Run("Concurrency="+fmt.Sprintf("%d",concurrency), func(b *testing.B) {
			client := dice.NewClient(&dice.Options{
				Addr: "localhost:7379",
				Password: "",
				DB: 0,
			})
			defer client.Close()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var cmdWg sync.WaitGroup
				cmdWg.Add(concurrency)
				for j := 0; j < concurrency; j++ {
					go func() {
						defer cmdWg.Done()
						for _, cmd := range commands {
							switch cmd {
							case "PING":
								client.Ping(ctx)
							case "SET foo bar":
								client.Set(ctx, "foo", "bar", 0)
							case "GET foo":
								client.Get(ctx, "foo")
							case "INCR counter":
								client.Incr(ctx, "counter")
							case "DEL foo":
								client.Del(ctx, "foo")
							}
						}
					}()
				}
				cmdWg.Wait()
			}
		})
	}
	cancel()
	wg.Wait()
	serverWg.Wait() 
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
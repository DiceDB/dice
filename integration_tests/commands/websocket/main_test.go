package websocket

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/dicedb/dice/internal/logger"
)

func TestMain(m *testing.M) {
	l := logger.New(logger.Opts{WithTimestamp: false})
	slog.SetDefault(l)
	var wg sync.WaitGroup

	// Run the test server
	// This is a synchronous method, because internally it
	// checks for available port and then forks a goroutine
	// to start the server
	opts := TestServerOptions{
		Port:   testPort1,
		Logger: l,
	}
	ctx, cancel := context.WithCancel(context.Background())
	RunWebsocketServer(ctx, &wg, opts)

	// Wait for the server to start
	time.Sleep(2 * time.Second)

	executor := NewWebsocketCommandExecutor()

	// Run the test suite
	exitCode := m.Run()

	// abort
	conn := executor.ConnectToServer()
	executor.FireCommand(conn, "abort")

	cancel()
	wg.Wait()
	os.Exit(exitCode)
}

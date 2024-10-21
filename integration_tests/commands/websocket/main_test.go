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

	opts := TestServerOptions{
		Port:   testPort1,
		Logger: l,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	RunWebsocketServer(ctx, &wg, opts)

	// Wait for the server to start
	time.Sleep(2 * time.Second)

	// Run the test suite
	exitCode := m.Run()

	// shut down gracefully
	executor := NewWebsocketCommandExecutor()
	conn := executor.ConnectToServer()
	executor.FireCommand(conn, "abort")

	wg.Wait()
	os.Exit(exitCode)
}

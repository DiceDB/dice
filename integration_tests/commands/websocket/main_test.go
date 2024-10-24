package websocket

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	var wg sync.WaitGroup

	opts := TestServerOptions{
		Port: testPort1,
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

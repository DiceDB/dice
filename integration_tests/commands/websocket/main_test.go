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

	// Run the test server
	// This is a synchronous method, because internally it
	// checks for available port and then forks a goroutine
	// to start the server
	opts := TestServerOptions{
		Port: testPort1,
	}
	ctx, cancel := context.WithCancel(context.Background())
	RunWebsocketServer(ctx, &wg, opts)

	// Wait for the server to start
	time.Sleep(2 * time.Second)

	executor := NewWebsocketCommandExecutor()

	// Run the test suite
	exitCode := m.Run()

	conn := executor.ConnectToServer()
	executor.FireCommand(conn, "abort")

	cancel()
	wg.Wait()
	os.Exit(exitCode)
}

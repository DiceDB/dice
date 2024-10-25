package http

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
		Port: 8083,
	}
	ctx, cancel := context.WithCancel(context.Background())
	RunHTTPServer(ctx, &wg, opts)

	// Wait for the server to start
	time.Sleep(2 * time.Second)

	executor := NewHTTPCommandExecutor()

	// Run the test suite
	exitCode := m.Run()

	executor.FireCommand(HTTPCommand{
		Command: "ABORT",
		Body:    map[string]interface{}{},
	})

	cancel()
	wg.Wait()
	if exitCode != 0 {
		os.Exit(exitCode)
	}
}

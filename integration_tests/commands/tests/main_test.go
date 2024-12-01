package tests

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/dicedb/dice/integration_tests/commands/tests/servers"
)

func TestMain(m *testing.M) {
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	opts := servers.TestServerOptions{
		Port: 9738,
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		servers.RunRespServer(ctx, &wg, opts)
	}()
	//TODO: run all three in paraller
	//RunWebSocketServer
	//RunHTTPServer

	// Wait for the server to start
	time.Sleep(2 * time.Second)

	// Run the test suite
	exitCode := m.Run()

	// Signal all servers to stop
	cancel()

	// Wait for all goroutines to finish
	wg.Wait()

	// Exit with the appropriate code
	os.Exit(exitCode)
}

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
	respOpts := servers.TestServerOptions{
		Port: 9738,
	}
	httpOpts := servers.TestServerOptions{
		Port: 8083,
	}
	wsOpts := servers.TestServerOptions{
		Port: 8380,
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		servers.RunRespServer(ctx, &wg, respOpts)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		servers.RunHTTPServer(ctx, &wg, httpOpts)
	}()

	//TODO: RunWebSocketServer
	wg.Add(1)
	go func() {
		defer wg.Done()
		servers.RunWebsocketServer(ctx, &wg, wsOpts)
	}()

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

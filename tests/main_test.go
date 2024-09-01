package tests

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(context.Background())
	// Run the test server
	// This is a synchronous method, because internally it
	// checks for available port and then forks a goroutine
	// to start the server
	runTestServer(ctx, &wg)

	// Wait for the server to start
	time.Sleep(1 * time.Second)

	conn := getLocalConnection()
	if conn == nil {
		panic("Failed to connect to the test server")
	}
	defer conn.Close()

	// Run the test suite
	exitCode := m.Run()

	cancel()

	wg.Wait()
	os.Exit(exitCode)
}

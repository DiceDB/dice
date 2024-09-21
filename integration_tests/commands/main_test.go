package commands

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
	logger := logger.New(logger.Opts{WithTimestamp: false})
	slog.SetDefault(logger)

	var wg sync.WaitGroup

	// Run the test server
	// This is a synchronous method, because internally it
	// checks for available port and then forks a goroutine
	// to start the server
	opts := TestServerOptions{
		Port:   8739,
		Logger: logger,
	}
	RunTestServer(context.Background(), &wg, opts)

	// Wait for the server to start
	time.Sleep(2 * time.Second)

	conn := getLocalConnection()
	if conn == nil {
		panic("Failed to connect to the test server")
	}
	defer conn.Close()

	// Run the test suite
	exitCode := m.Run()

	result := FireCommand(conn, "ABORT")
	if result != "OK" {
		panic("Failed to abort the server")
	}

	wg.Wait()
	os.Exit(exitCode)
}

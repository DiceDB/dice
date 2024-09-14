package commands

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"
	slogzerolog "github.com/samber/slog-zerolog/v2"
)

func TestMain(m *testing.M) {
	zerologLogger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr})
	logger := slog.New(slogzerolog.Option{Logger: &zerologLogger}.NewZerologHandler())
	slog.SetDefault(logger)

	var wg sync.WaitGroup

	// Run the test server
	// This is a synchronous method, because internally it
	// checks for available port and then forks a goroutine
	// to start the server
	opts := TestServerOptions{
		Port: 8739,
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

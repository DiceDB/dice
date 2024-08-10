package tests

import (
	"os"
	"sync"
	"testing"
	"time"

	"github.com/dicedb/dice/core"
)

func TestMain(m *testing.M) {
	var wg sync.WaitGroup

	// Run the test server
	// This is a synchronous method, because internally it
	// checks for available port and then forks a goroutine
	// to start the server
	runTestServer(&wg)

	// Wait for the server to start
	time.Sleep(1 * time.Second)

	conn := getLocalConnection()
	if conn == nil {
		panic("Failed to connect to the test server")
	}
	defer conn.Close()

	// Run the test suite
	exitCode := m.Run()

	core.ResetStore()

	// Fire the ABORT command to gracefully terminate the server
	result := fireCommand(conn, "ABORT")
	if result != "OK" {
		panic("Failed to abort the server")
	}

	wg.Wait()
	os.Exit(exitCode)
}

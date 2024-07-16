package tests

import (
	"os"
	"sync"
	"testing"
	"time"
)

type DTestCase struct {
	InCmds []string
	Out    []interface{}
}

func TestMain(m *testing.M) {
	var wg sync.WaitGroup

	// Run the test server
	// This is a synchronous method, because internally it
	// checks for available port and then forks a goroutine
	// to start the server
	runTestServer(&wg)

	time.Sleep(1 * time.Second)
	conn := getLocalConnection()

	// run the test suite
	exitCode := m.Run()

	// Fire the ABORT command to gracefully terminate the server
	fireCommand(conn, "ABORT")

	// wait for all the goroutines to finish
	wg.Wait()
	os.Exit(exitCode)
}

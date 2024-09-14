package server

import (
	"context"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/dicedb/dice/integration_tests/commands"
	"gotest.tools/v3/assert"

	"github.com/dicedb/dice/config"
)

func getConnection() (net.Conn, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", maxConnTestOptions.Port))
	if err != nil {
		return nil, err
	}
	return conn, nil
}

var maxConnTestOptions = commands.TestServerOptions{
	Port: 8741,
}

func TestMaxConnection(t *testing.T) {
	var wg sync.WaitGroup
	commands.RunTestServer(context.Background(), &wg, maxConnTestOptions)

	time.Sleep(2 * time.Second)

	var maxConnLimit = config.TestServerMaxClients + 2
	connections := make([]net.Conn, maxConnLimit)
	for i := 0; i < maxConnLimit; i++ {
		conn, err := getConnection()
		if err == nil {
			connections[i] = conn
		} else {
			t.Fatalf("unexpected error while getting connection %d: %v", i, err)
		}
		time.Sleep(500 * time.Millisecond)
	}
	assert.Equal(t, maxConnLimit, len(connections), "should have reached the max connection limit")

	_, err := getConnection()
	assert.ErrorContains(t, err, "connect: connection refused")

	result := commands.FireCommand(connections[0], "ABORT")
	if result != "OK" {
		t.Fatalf("Unexpected response to ABORT command: %v", result)
	}

	// Wait for the server to shut down
	time.Sleep(1 * time.Second)
	wg.Wait()
}

package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"testing"
	"time"

	"gotest.tools/v3/assert"

	"github.com/dicedb/dice/config"
	commands "github.com/dicedb/dice/integration_tests/commands/async"
)

func getConnection(port int) (net.Conn, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func TestMaxConnection(t *testing.T) {
	var wg sync.WaitGroup
	var maxConnTestOptions = commands.TestServerOptions{
		Port:       8741,
		MaxClients: 50,
		Logger:     slog.Default(),
	}
	commands.RunTestServer(context.Background(), &wg, maxConnTestOptions)

	time.Sleep(2 * time.Second)

	var maxConnLimit = config.TestServerMaxClients + 2
	connections := make([]net.Conn, maxConnLimit)
	for i := 0; i < maxConnLimit; i++ {
		conn, err := getConnection(maxConnTestOptions.Port)
		if err == nil {
			connections[i] = conn
		} else {
			t.Fatalf("unexpected error while getting connection %d: %v", i, err)
		}
	}
	assert.Equal(t, maxConnLimit, len(connections), "should have reached the max connection limit")

	_, err := getConnection(maxConnTestOptions.Port)
	assert.ErrorContains(t, err, "connect: connection refused")

	result := commands.FireCommand(connections[0], "ABORT")
	if result != "OK" {
		t.Fatalf("Unexpected response to ABORT command: %v", result)
	}
	wg.Wait()
}

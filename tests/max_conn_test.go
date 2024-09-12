package tests

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/dicedb/dice/config"
	"gotest.tools/v3/assert"
)

func getConnection() (net.Conn, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", config.Port))
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func TestMaxConnAccept(t *testing.T) {
	var maxConnLimit = config.Test_ServerMaxClients + 1
	connections := make([]net.Conn, maxConnLimit)
	for i := 0; i < maxConnLimit; i++ {
		conn, err := getConnection()
		if err == nil {
			connections[i] = conn
			defer connections[i].Close()
		} else {
			t.Fatalf("unexpected error  while getting connection: %v", err)
		}
		time.Sleep(100 * time.Millisecond)
	}
	assert.Equal(t, maxConnLimit, len(connections), "should have reached the max connection limit")

	_, err := getConnection()
	assert.Error(t, err, "dial tcp 127.0.0.1:8739: connect: connection refused")
}

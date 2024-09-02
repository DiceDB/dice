package tests

import (
	"net"
	"sync"
	"testing"
	"time"
)

func TestAbortCommand(t *testing.T) {
	// Ensure the port is released by checking if it's available
	if !CheckPortAvailibility("localhost:YOUR_PORT") {
		t.Fatalf("Server port was not released after shutdown")
	} else {
		t.Log("Port was successfully released after server shutdown")
	}
	
	var wg sync.WaitGroup
	// Restart the server to ensure it restarts successfully
	runTestServer(&wg)

	// Wait for the server to start again
	time.Sleep(1 * time.Second)

	restartedConn := getLocalConnection()
	if restartedConn == nil {
		t.Fatalf("Failed to restart the test server")
	} else {
		t.Log("Server successfully restarted")
	}

	defer restartedConn.Close()

	response := fireCommand(restartedConn, "PING")
	if response != "PONG" {
		t.Errorf("Unexpected server response after restart, expected 'PONG', got '%s'", response)
	} else {
		t.Log("Server responded correctly after restart")
	}
}

func CheckPortAvailibility(address string) bool {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return true // Port is available
	}
	conn.Close()
	return false // Port is in use
}

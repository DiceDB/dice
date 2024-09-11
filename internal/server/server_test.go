package server_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/charmbracelet/log"
	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/internal/clientio"
	"github.com/dicedb/dice/internal/server"
	"github.com/dicedb/dice/testutils"
)

func TestAbortCommand(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Cleanup(cancel)

	var wg sync.WaitGroup
	runTestServer(ctx, &wg)

	time.Sleep(2 * time.Second)

	// Test 1: Ensure the server is running
	t.Run("ServerIsRunning", func(t *testing.T) {
		conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", config.Port))
		if err != nil {
			t.Fatalf("Failed to connect to server: %v", err)
		}
		conn.Close()
	})

	//Test 2: Send ABORT command and check if the server shuts down
	t.Run("AbortCommandShutdown", func(t *testing.T) {
		conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", config.Port))
		if err != nil {
			t.Fatalf("Failed to connect to server: %v", err)
		}
		defer conn.Close()

		// Send ABORT command
		result := fireCommand(conn, "ABORT")
		if result != "OK" {
			t.Fatalf("Unexpected response to ABORT command: %v", result)
		}

		// Wait for the server to shut down
		time.Sleep(1 * time.Second)

		// Try to connect again, it should fail
		_, err = net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", config.Port))
		if err == nil {
			t.Fatal("Server did not shut down as expected")
		}
	})

	// Test 3: Ensure the server port is released
	t.Run("PortIsReleased", func(t *testing.T) {
		// Try to bind to the same port
		listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", config.Port))
		if err != nil {
			t.Fatalf("Port should be available after server shutdown: %v", err)
		}
		listener.Close()
	})

	wg.Wait()
}

func TestServerRestartAfterAbort(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	// start test server.
	var wg sync.WaitGroup
	runTestServer(ctx, &wg)

	time.Sleep(1 * time.Second)

	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", config.Port))
	if err != nil {
		t.Fatalf("Server should be running after restart: %v", err)
	}

	// Send ABORT command to shut down server
	result := fireCommand(conn, "ABORT")
	if result != "OK" {
		t.Fatalf("Unexpected response to ABORT command: %v", result)
	}
	conn.Close()

	// wait for the server to shutdown
	time.Sleep(2 * time.Second)

	// restart server
	runTestServer(ctx, &wg)

	// wait for the server to start up
	time.Sleep(2 * time.Second)

	// Check if the server is running
	conn, err = net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", config.Port))
	if err != nil {
		t.Fatalf("Server should be running after restart: %v", err)
	}

	// Clean up
	result = fireCommand(conn, "ABORT")
	if result != "OK" {
		t.Fatalf("Unexpected response to ABORT command: %v", result)
	}
	conn.Close()

	wg.Wait()
}

//nolint:unused
func runTestServer(ctx context.Context, wg *sync.WaitGroup) {
	config.IOBufferLength = 16
	config.Port = 8739
	config.WriteAOFOnCleanup = true

	const totalRetries = 10
	var err error

	// Initialize the AsyncServer
	testServer := server.NewAsyncServer()

	// Try to bind to a port with a maximum of totalRetries retries.
	for i := 0; i < totalRetries; i++ {
		if err = testServer.FindPortAndBind(); err == nil {
			break
		}

		if err.Error() == "address already in use" {
			log.Infof("Port %d already in use, trying port %d", config.Port, config.Port+1)
			config.Port++
		} else {
			log.Fatalf("Failed to bind port: %v", err)
			return
		}
	}

	if err != nil {
		log.Fatalf("Failed to bind to a port after %d retries: %v", totalRetries, err)
		return
	}

	// Inform the user that the server is starting
	fmt.Println("Starting the test server on port", config.Port)

	// Start the server in a goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := testServer.Run(ctx); err != nil {
			if errors.Is(err, server.ErrAborted) {
				return
			}
			log.Fatalf("Test server encountered an error: %v", err)
		}
	}()
}

//nolint:unused
func fireCommand(conn net.Conn, cmd string) interface{} {
	var err error
	args := testutils.ParseCommand(cmd)
	_, err = conn.Write(clientio.Encode(args, false))
	if err != nil {
		log.Fatalf("error %s while firing command: %s", err, cmd)
	}

	rp := clientio.NewRESPParser(conn)
	v, err := rp.DecodeOne()
	if err != nil {
		if err == io.EOF {
			return nil
		}
		log.Fatalf("error %s while firing command: %s", err, cmd)
	}
	return v
}

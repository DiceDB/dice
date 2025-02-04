// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/dicedb/dice/config"
)

func init() {
	config.Config.Port = 8379
	log.Print("Setting port to ", config.Config.Port)
}

func TestAbortCommand(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer ctx.Done()
	t.Cleanup(cancel)

	var wg sync.WaitGroup
	time.Sleep(2 * time.Second)

	// Test 1: Ensure the server is running
	t.Run("ServerIsRunning", func(t *testing.T) {
		conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", config.Config.Port))
		if err != nil {
			t.Fatalf("Failed to connect to server: %v", err)
		}
		conn.Close()
	})

	//Test 2: Send ABORT command and check if the server shuts down
	t.Run("AbortCommandShutdown", func(t *testing.T) {
		client := getLocalConnection()
		defer client.Close()

		// Send ABORT command
		resp := client.FireString("ABORT")
		if resp.GetVStr() != "OK" {
			t.Fatalf("Unexpected response to ABORT command: %v", resp)
		}

		// Wait for the server to shut down
		time.Sleep(1 * time.Second)

		client = getLocalConnection()
		if client != nil {
			t.Fatalf("Server did not shut down as expected")
		}
	})

	// Test 3: Ensure the server port is released
	t.Run("PortIsReleased", func(t *testing.T) {
		// Try to bind to the same port
		listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", config.Config.Port))
		if err != nil {
			t.Fatalf("Port should be available after server shutdown: %v", err)
		}
		listener.Close()
	})

	wg.Wait()
}

func TestServerRestartAfterAbort(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer ctx.Done()
	t.Cleanup(cancel)

	// start test server.
	var wg sync.WaitGroup
	resp.RunTestServer(&wg, testServerOptions)

	time.Sleep(1 * time.Second)

	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", config.Config.Port))
	if err != nil {
		t.Fatalf("Server should be running after restart: %v", err)
	}

	// Send ABORT command to shut down server
	result := resp.client.FireString("ABORT")
	if result != "OK" {
		t.Fatalf("Unexpected response to ABORT command: %v", result)
	}
	conn.Close()

	// wait for the server to shutdown
	time.Sleep(2 * time.Second)

	wg.Wait()

	// restart server
	ctx2, cancel2 := context.WithCancel(context.Background())
	defer ctx2.Done()
	t.Cleanup(cancel2)

	// start test server.
	// use different waitgroups and contexts to avoid race conditions.;
	var wg2 sync.WaitGroup
	resp.RunTestServer(&wg2, testServerOptions)

	// wait for the server to start up
	time.Sleep(2 * time.Second)

	// Check if the server is running
	conn2, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", config.Config.Port))
	if err != nil {
		t.Fatalf("Server should be running after restart: %v", err)
	}

	// Clean up
	result = resp.FireCommand(conn2, "ABORT")
	if result != "OK" {
		t.Fatalf("Unexpected response to ABORT command: %v", result)
	}
	conn2.Close()

	wg2.Wait()
}

package tests

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/dicedb/dice/internal/clientio"

	"github.com/dicedb/dice/internal/server"

	"github.com/charmbracelet/log"
	"github.com/dicedb/dice/config"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dice/testutils"
	redis "github.com/dicedb/go-dice"
)

//nolint:unused
func getLocalConnection() net.Conn {
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", config.Port))
	if err != nil {
		panic(err)
	}
	return conn
}

// deleteTestKeys is a utility to delete a list of keys before running a test
//
//nolint:unused
func deleteTestKeys(keysToDelete []string, store *dstore.Store) {
	for _, key := range keysToDelete {
		store.Del(key)
	}
}

//nolint:unused
func getLocalSdk() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf(":%d", config.Port),

		DialTimeout:           10 * time.Second,
		ReadTimeout:           30 * time.Second,
		WriteTimeout:          30 * time.Second,
		ContextTimeoutEnabled: true,

		MaxRetries: -1,

		PoolSize:        10,
		PoolTimeout:     30 * time.Second,
		ConnMaxIdleTime: time.Minute,
	})
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

//nolint:unused
func fireCommandAndGetRESPParser(conn net.Conn, cmd string) *clientio.RESPParser {
	args := testutils.ParseCommand(cmd)
	_, err := conn.Write(clientio.Encode(args, false))
	if err != nil {
		log.Fatalf("error %s while firing command: %s", err, cmd)
	}

	return clientio.NewRESPParser(conn)
}

//nolint:unused
func runTestServer(ctx context.Context, wg *sync.WaitGroup) {
	config.IOBufferLength = 16
	config.Port = 8739

	const totalRetries = 100
	var err error

	// Initialize the AsyncServer
	testServer := server.NewAsyncServer()

	// Try to bind to a port with a maximum of `totalRetries` retries.
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

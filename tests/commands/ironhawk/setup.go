// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"sync"
	"time"

	"github.com/dicedb/dice/internal/server/ironhawk"
	"github.com/dicedb/dice/internal/shardmanager"
	"github.com/dicedb/dice/internal/wal"

	"github.com/dicedb/dice/config"
	derrors "github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dicedb-go"
)

//nolint:unused
func getLocalConnection() *dicedb.Client {
	client, err := dicedb.NewClient("localhost", config.Config.Port)
	if err != nil {
		panic(err)
	}
	return client
}

func ClosePublisherSubscribers(publisher net.Conn, subscribers []net.Conn) error {
	if err := publisher.Close(); err != nil {
		return fmt.Errorf("error closing publisher connection: %v", err)
	}
	for _, sub := range subscribers {
		time.Sleep(100 * time.Millisecond) // [TODO] why is this needed?
		if err := sub.Close(); err != nil {
			return fmt.Errorf("error closing subscriber connection: %v", err)
		}
	}
	return nil
}

// //nolint:unused
// func unsubscribeFromWatchUpdates(t *testing.T, subscribers []net.Conn, cmd, fingerprint string) {
// 	t.Helper()
// 	for _, subscriber := range subscribers {
// 		rp := fireCommandAndGetRESPParser(subscriber, fmt.Sprintf("%s.UNWATCH %s", cmd, fingerprint))
// 		assert.NotNil(t, rp)
// 		v, err := rp.DecodeOne()
// 		assert.NoError(t, err)
// 		castedValue, ok := v.(string)
// 		if !ok {
// 			t.Errorf("Type assertion to string failed for value: %v", v)
// 		}
// 		assert.Equal(t, castedValue, "OK")
// 	}
// }

// //nolint:unused
// func unsubscribeFromWatchUpdatesSDK(t *testing.T, subscribers []WatchSubscriber, cmd, fingerprint string) {
// 	for _, subscriber := range subscribers {
// 		err := subscriber.watch.Unwatch(context.Background(), cmd, fingerprint)
// 		assert.Nil(t, err)
// 	}
// }

// // deleteTestKeys is a utility to delete a list of keys before running a test
// //
// //nolint:unused
// func deleteTestKeys(keysToDelete []string, store *dstore.Store) {
// 	for _, key := range keysToDelete {
// 		store.Del(key)
// 	}
// }

//nolint:unused
func getLocalSdk() *dicedb.Client {
	client, err := dicedb.NewClient("localhost", config.Config.Port)
	if err != nil {
		panic(err)
	}
	return client
}

// type WatchSubscriber struct {
// 	client *dicedb.Client
// 	watch  *dicedb.WatchConn
// }

// func ClosePublisherSubscribersSDK(publisher *dicedb.Client, subscribers []WatchSubscriber) error {
// 	if err := publisher.Close(); err != nil {
// 		return fmt.Errorf("error closing publisher connection: %v", err)
// 	}
// 	for _, sub := range subscribers {
// 		if err := sub.watch.Close(); err != nil {
// 			return fmt.Errorf("error closing subscriber watch connection: %v", err)
// 		}
// 		if err := sub.client.Close(); err != nil {
// 			return fmt.Errorf("error closing subscriber connection: %v", err)
// 		}
// 	}
// 	return nil
// }

func RunTestServer(wg *sync.WaitGroup) {
	// #1261: Added here to prevent resp integration tests from failing on lower-spec machines
	gec := make(chan error)
	shardManager := shardmanager.NewShardManager(1, gec)
	ioThreadManager := ironhawk.NewIOThreadManager()
	watchManager := &ironhawk.WatchManager{}
	wal.SetupWAL()

	testServer := ironhawk.NewServer(shardManager, ioThreadManager, watchManager)

	ctx, cancel := context.WithCancel(context.Background())
	fmt.Println("Starting the test server on port", config.Config.Port)

	shardManagerCtx, cancelShardManager := context.WithCancel(ctx)
	wg.Add(1)
	go func() {
		defer wg.Done()
		shardManager.Run(shardManagerCtx)
	}()

	// Start the server in a goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := testServer.Run(ctx); err != nil {
			if errors.Is(err, derrors.ErrAborted) {
				cancelShardManager()
				return
			}
			slog.Error("Test server encountered an error", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	go func() {
		for err := range gec {
			if err != nil && errors.Is(err, derrors.ErrAborted) {
				// if either the AsyncServer/RESPServer or the HTTPServer received an abort command,
				// cancel the context, helping gracefully exiting all servers
				cancel()
			}
		}
	}()
}

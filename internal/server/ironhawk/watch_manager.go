// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"context"
	"log/slog"
	"strconv"
	"strings"
	"sync"

	"github.com/dicedb/dice/internal/cmd"
	"github.com/dicedb/dice/internal/shardmanager"
)

type WatchManager struct {
	mu                   sync.RWMutex
	clientWatchThreadMap map[string]*IOThread

	keyFPMap    map[string]map[uint64]bool
	fpClientMap map[uint64]map[string]bool
	fpCmdMap    map[uint64]*cmd.Cmd
}

func NewWatchManager() *WatchManager {
	return &WatchManager{
		clientWatchThreadMap: map[string]*IOThread{},

		keyFPMap:    map[string]map[uint64]bool{},
		fpClientMap: map[uint64]map[string]bool{},
		fpCmdMap:    map[uint64]*cmd.Cmd{},
	}
}

func (w *WatchManager) RegisterThread(t *IOThread) {
	if t.Mode == "watch" {
		// Only acquire lock if we are in "watch" mode.
		w.mu.Lock()
		defer w.mu.Unlock()
		w.clientWatchThreadMap[t.ClientID] = t
	}
}

func (w *WatchManager) HandleWatch(c *cmd.Cmd, t *IOThread) {
	w.mu.Lock()
	defer w.mu.Unlock()

	fp, key := c.Fingerprint(), c.Key()
	slog.Debug("creating a new subscription",
		slog.String("key", key),
		slog.String("cmd", c.String()),
		slog.Any("fingerprint", fp),
		slog.String("client_id", t.ClientID))

	// For the key that will be watched through any .WATCH command
	// Create an entry in the map that holds, key <--> [command fingerprint] as map
	if _, ok := w.keyFPMap[key]; !ok {
		w.keyFPMap[key] = make(map[uint64]bool)
	}
	w.keyFPMap[key][fp] = true

	// For the fingerprint
	// Create an entry in the map that holds, fingerprint <--> [client id] as map
	// This tells us which clients are subscribed to a particular fingerprint
	if _, ok := w.fpClientMap[fp]; !ok {
		w.fpClientMap[fp] = make(map[string]bool)
	}
	w.fpClientMap[fp][t.ClientID] = true

	// Store the fingerprint <--> command mapping
	// so that we understand what should we execute when the data changes
	w.fpCmdMap[fp] = c

	w.clientWatchThreadMap[t.ClientID] = t
}

func (w *WatchManager) HandleUnwatch(c *cmd.Cmd, t *IOThread) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if len(c.C.Args) != 1 {
		return
	}

	// Parse the fingerprint from the command
	fp, err := strconv.ParseUint(c.C.Args[0], 10, 64)
	if err != nil {
		return
	}

	// Multiple clients can unsubscribe from the same fingerprint
	// So, we need to delete the one that is unsubscribing
	delete(w.fpClientMap[fp], t.ClientID)

	// If a fingerprint has no clients subscribed to it, delete the fingerprint from the map.
	if len(w.fpClientMap[fp]) == 0 {
		delete(w.fpClientMap, fp)

		// If we have deleted the fingerprint, delete the command from the map
		delete(w.fpCmdMap, fp)
	}

	// Delete the mapping where we have the key <--> [command fingerprint]
	// This seems to be an O(n) operation.
	// Downside of keeping this entry laying around is that
	// If key k changes, then we may be iterating through the fingerprint that does not have any active watcher
	// Hence we should do a lazy deletion.

	// TODO: If the key gets deleted from the database
	// delete the subscriptions against that key from all the places.
}

func (w *WatchManager) CleanupThreadWatchSubscriptions(t *IOThread) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Delete the mapping of Watch thread to client id
	delete(w.clientWatchThreadMap, t.ClientID)

	// Delete all the subscriptions of the client from the fingerprint maps
	// Note: this is an O(n) operation and hence if there are large number of clients, this might be expensive.
	// We can do a lazy deletion of the fingerprint map if this becomes a problem.
	for fp := range w.fpClientMap {
		delete(w.fpClientMap[fp], t.ClientID)
		if len(w.fpClientMap[fp]) == 0 {
			delete(w.fpClientMap, fp)
		}
	}
}

func (w *WatchManager) NotifyWatchers(c *cmd.Cmd, shardManager *shardmanager.ShardManager, t *IOThread) {
	// Use RLock instead as we are not really modifying any shared maps here.
	w.mu.RLock()
	defer w.mu.RUnlock()

	key := c.Key()
	for fp := range w.keyFPMap[key] {
		_c := w.fpCmdMap[fp]
		if _c == nil {
			// TODO: Not having a command for a fingerprint is a bug.
			continue
		}

		r, err := _c.Execute(shardManager)
		if err != nil {
			slog.Error("failed to execute command as part of watch notification",
				slog.Any("cmd", _c.String()),
				slog.Any("error", err))
			continue
		}

		for clientID := range w.fpClientMap[fp] {
			thread := w.clientWatchThreadMap[clientID]
			if thread == nil {
				// if there is no thread against the client, delete the client from the map
				delete(w.clientWatchThreadMap, clientID)
				continue
			}

			// If this is first time a client is connecting it'd be sending a .WATCH command
			// in that case we don't need to notify all other clients subscribed to the key
			if strings.HasSuffix(c.C.Cmd, ".WATCH") && t.ClientID != clientID {
				continue
			}

			err := thread.serverWire.Send(context.Background(), r.Rs)
			if err != nil {
				slog.Error("failed to write response to thread",
					slog.Any("client_id", thread.ClientID),
					slog.String("mode", thread.Mode),
					slog.Any("error", err))
			}
		}

		slog.Debug("notifying watchers for key", slog.String("key", key), slog.Int("watchers", len(w.fpClientMap[fp])))
	}
}

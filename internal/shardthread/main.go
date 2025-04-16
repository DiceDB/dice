// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package shardthread

import (
	"context"
	"time"

	"github.com/dicedb/dice/config"
	dstore "github.com/dicedb/dice/internal/store"
)

type ShardThread struct {
	id               int           // id is the unique identifier for the shard.
	store            *dstore.Store // store that the shard is responsible for.
	globalErrorChan  chan error    // globalErrorChan is the channel for sending system-level errors.
	lastCronExecTime time.Time     // lastCronExecTime is the last time the shard executed cron tasks.
	cronFrequency    time.Duration // cronFrequency is the frequency at which the shard executes cron tasks.
}

// NewShardThread creates a new ShardThread instance with the given shard id and error channel.
func NewShardThread(id int, gec chan error, evictionStrategy dstore.EvictionStrategy) *ShardThread {
	return &ShardThread{
		id:               id,
		store:            dstore.NewStore(nil, evictionStrategy, id),
		globalErrorChan:  gec,
		lastCronExecTime: time.Now(),
		cronFrequency:    config.ShardCronFrequency,
	}
}

// Start starts the shard thread, listening for incoming requests.
func (shard *ShardThread) Start(ctx context.Context) {
	ticker := time.NewTicker(shard.cronFrequency)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			shard.runCronTasks()
		case <-ctx.Done():
			shard.cleanup()
			return
		}
	}
}

// runCronTasks runs the cron tasks for the shard. This includes deleting expired keys.
func (shard *ShardThread) runCronTasks() {
	dstore.DeleteExpiredKeys(shard.store)
	shard.lastCronExecTime = time.Now()
}

// cleanup handles cleanup logic when the shard stops.
func (shard *ShardThread) cleanup() {
	if !config.Config.EnableWAL {
		return
	}
}

func (shard *ShardThread) Store() *dstore.Store {
	return shard.store
}

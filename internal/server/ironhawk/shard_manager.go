// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/cespare/xxhash/v2"
	"github.com/dicedb/dice/config"

	"github.com/dicedb/dice/internal/eval"
	"github.com/dicedb/dice/internal/ops"
	dstore "github.com/dicedb/dice/internal/store"
)

type Shard struct {
	ID     int
	Thread *ShardThread
}

type ShardManager struct {
	shards  []*Shard
	sigChan chan os.Signal // sigChan is the signal channel for the shard manager
}

// NewShardManager creates a new ShardManager instance with the given number of Shards and a parent context.
func NewShardManager(shardCount int, cmdWatchChan chan dstore.CmdWatchEvent, globalErrorChan chan error) *ShardManager {
	shards := make([]*Shard, shardCount)
	maxKeysPerShard := config.DefaultKeysLimit / shardCount
	for i := 0; i < shardCount; i++ {
		shards[i] = &Shard{
			ID:     i,
			Thread: NewShardThread(i, globalErrorChan, dstore.NewPrimitiveEvictionStrategy(maxKeysPerShard)),
		}
	}

	return &ShardManager{
		shards:  shards,
		sigChan: make(chan os.Signal, 1),
	}
}

// Execute executes a command on the appropriate shard.
func (manager *ShardManager) Execute(op *ops.StoreOp) *eval.EvalResponse {
	shard := manager.getShardForKey(op.Cmd.Key())
	return shard.Thread.processRequest(op)
}

// Run starts the ShardManager, manages its lifecycle, and listens for errors.
func (manager *ShardManager) Run(ctx context.Context) {
	signal.Notify(manager.sigChan, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup
	shardCtx, cancelShard := context.WithCancel(ctx)
	defer cancelShard()

	manager.start(shardCtx, &wg)

	select {
	case <-ctx.Done():
		// Parent context was canceled, trigger shutdown
	case <-manager.sigChan:
		// OS signal received, trigger shutdown
	}

	wg.Wait() // Wait for all shard goroutines to exit.
}

// start initializes and starts the shard threads.
func (manager *ShardManager) start(ctx context.Context, wg *sync.WaitGroup) {
	for _, shard := range manager.shards {
		shard := shard

		wg.Add(1)
		go func() {
			defer wg.Done()
			shard.Thread.Start(ctx)
		}()
	}
}

func (manager *ShardManager) getShardForKey(key string) *Shard {
	return manager.shards[xxhash.Sum64String(key)%uint64(manager.GetShardCount())]
}

// GetShardCount returns the number of shards managed by this ShardManager.
func (manager *ShardManager) GetShardCount() int8 {
	return int8(len(manager.shards))
}

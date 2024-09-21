package shard

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/dicedb/dice/internal/ops"
	dstore "github.com/dicedb/dice/internal/store"
)

type ShardManager struct {
	// shards is a constant slice of all Shards managed by this manager, indexed by ShardID. The shards slice is
	// instantiated during ShardManager creation, and never modified after wards. Therefore, it can be accessed
	// concurrently without synchronization.
	shards          []*ShardThread
	shardReqMap     map[ShardID]chan *ops.StoreOp // shardReqMap is a map of shard id to its respective request channel
	globalErrorChan chan *ShardError              // globalErrorChan is the common global error channel for all Shards
	sigChan         chan os.Signal                // sigChan is the signal channel for the shard manager
}

// NewShardManager creates a new ShardManager instance with the given number of Shards and a parent context.
func NewShardManager(shardCount int8, watchChan chan dstore.WatchEvent, logger *slog.Logger) *ShardManager {
	shards := make([]*ShardThread, shardCount)
	shardReqMap := make(map[ShardID]chan *ops.StoreOp)
	globalErrorChan := make(chan *ShardError)

	for i := int8(0); i < shardCount; i++ {
		// Shards are numbered from 0 to shardCount-1
		shard := NewShardThread(ShardID(i), globalErrorChan, watchChan, logger)
		shards[i] = shard
		shardReqMap[ShardID(i)] = shard.ReqChan
	}

	return &ShardManager{
		shards:          shards,
		shardReqMap:     shardReqMap,
		globalErrorChan: globalErrorChan,
		sigChan:         make(chan os.Signal, 1),
	}
}

// Run starts the ShardManager, manages its lifecycle, and listens for errors.
func (manager *ShardManager) Run(ctx context.Context) {
	signal.Notify(manager.sigChan, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup
	shardCtx, cancelShard := context.WithCancel(ctx)
	defer cancelShard()

	manager.start(shardCtx, &wg)

	wg.Add(1)
	go func() {
		defer wg.Done()
		manager.listenForErrors()
	}()

	select {
	case <-ctx.Done():
		// Parent context was canceled, trigger shutdown
	case <-manager.sigChan:
		// OS signal received, trigger shutdown
	}

	close(manager.globalErrorChan) // Close the error channel after all Shards stop
	wg.Wait()                      // Wait for all shard goroutines to exit.
}

// start initializes and starts the shard threads.
func (manager *ShardManager) start(ctx context.Context, wg *sync.WaitGroup) {
	for _, shard := range manager.shards {
		shard := shard

		wg.Add(1)
		go func() {
			defer wg.Done()
			shard.Start(ctx)
		}()
	}
}

// listenForErrors listens to the global error channel and logs the errors. It exits when the error channel is closed.
func (manager *ShardManager) listenForErrors() {
	for err := range manager.globalErrorChan {
		// Handle or log shard errors here
		log.Printf("Shard %d error: %v", err.shardID, err.err)
	}
}

// GetShardCount returns the number of shards managed by this ShardManager.
func (manager *ShardManager) GetShardCount() int {
	return len(manager.shards)
}

// GetShard returns the ShardThread for the given ShardID.
func (manager *ShardManager) GetShard(id ShardID) *ShardThread {
	if int(id) < len(manager.shards) {
		return manager.shards[id]
	}
	return nil
}

// RegisterWorker registers a worker with all Shards present in the ShardManager.
func (manager *ShardManager) RegisterWorker(workerID string, workerChan chan *ops.StoreResponse) {
	for _, shard := range manager.shards {
		shard.registerWorker(workerID, workerChan)
	}
}

func (manager *ShardManager) UnregisterWorker(workerID string) {
	for _, shard := range manager.shards {
		shard.unregisterWorker(workerID)
	}
}

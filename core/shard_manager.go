package core

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type ShardManager struct {
	Shards          []*ShardThread            // slice of all Shards managed by this manager.
	shardsMutex     sync.Mutex                // mutex to protect the Shards slice.
	shardReqMap     map[ShardID]chan *StoreOp // map of shard id to its respective request channel.
	globalErrorChan chan *ShardError          // common global error channel for all Shards.
	sigChan         chan os.Signal            // signal channel for the shard manager.
}

// NewShardManager creates a new ShardManager instance with the given number of Shards and a parent context.
func NewShardManager(shardCount int8) *ShardManager {
	manager := &ShardManager{
		Shards:          make([]*ShardThread, shardCount),
		shardReqMap:     make(map[ShardID]chan *StoreOp),
		globalErrorChan: make(chan *ShardError),
		sigChan:         make(chan os.Signal, 1),
	}

	manager.initializeShards(shardCount)
	return manager
}

// initializeShards creates and configures shard threads.
func (manager *ShardManager) initializeShards(shardCount int8) {
	manager.shardsMutex.Lock()
	defer manager.shardsMutex.Unlock()

	for i := int8(0); i < shardCount; i++ {
		// Shards are numbered from 0 to shardCount-1
		shard := NewShardThread(ShardID(i), manager.globalErrorChan)
		manager.Shards[i] = shard
		manager.shardReqMap[ShardID(i)] = shard.ReqChan
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
	manager.shardsMutex.Lock()
	defer manager.shardsMutex.Unlock()
	for _, shard := range manager.Shards {
		shard := shard

		wg.Add(1)
		go func() {
			defer wg.Done()
			shard.Start(ctx)
		}()
	}
}

// RegisterWorker registers a worker with all Shards present in the ShardManager.
func (manager *ShardManager) RegisterWorker(workerID string, workerChan chan *StoreResponse) {
	manager.shardsMutex.Lock()
	defer manager.shardsMutex.Unlock()
	for _, shard := range manager.Shards {
		shard.registerWorker(workerID, workerChan)
	}
}

// listenForErrors listens to the global error channel and logs the errors. It exits when the error channel is closed.
func (manager *ShardManager) listenForErrors() {
	for err := range manager.globalErrorChan {
		// Handle or log shard errors here
		log.Printf("Shard %d error: %v", err.shardID, err.err)
	}
}

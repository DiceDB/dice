package shard

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/dicedb/dice/config"

	"github.com/cespare/xxhash/v2"
	"github.com/dicedb/dice/internal/ops"
	dstore "github.com/dicedb/dice/internal/store"
)

type ShardManager struct {
	// shards is a constant slice of all Shards managed by this manager, indexed by ShardID. The shards slice is
	// instantiated during ShardManager creation, and never modified after wards. Therefore, it can be accessed
	// concurrently without synchronization.
	shards          []*ShardThread
	shardReqMap     map[ShardID]chan *ops.StoreOp // shardReqMap is a map of shard id to its respective request channel
	globalErrorChan chan error                    // globalErrorChan is the common global error channel for all Shards
	ShardErrorChan  chan *ShardError              // ShardErrorChan is the channel for sending shard-level errors
	sigChan         chan os.Signal                // sigChan is the signal channel for the shard manager
	shardCount      uint8                         // shardCount is the number of shards managed by this manager
}

// NewShardManager creates a new ShardManager instance with the given number of Shards and a parent context.
func NewShardManager(shardCount uint8, queryWatchChan chan dstore.QueryWatchEvent, cmdWatchChan chan dstore.CmdWatchEvent, globalErrorChan chan error) *ShardManager {
	shards := make([]*ShardThread, shardCount)
	shardReqMap := make(map[ShardID]chan *ops.StoreOp)
	shardErrorChan := make(chan *ShardError)

	maxKeysPerShard := config.DiceConfig.Memory.KeysLimit / int(shardCount)
	for i := uint8(0); i < shardCount; i++ {
		evictionStrategy := dstore.NewBatchEvictionLRU(maxKeysPerShard, config.DiceConfig.Memory.EvictionRatio)
		// Shards are numbered from 0 to shardCount-1
		shard := NewShardThread(i, globalErrorChan, shardErrorChan, queryWatchChan, cmdWatchChan, evictionStrategy)
		shards[i] = shard
		shardReqMap[i] = shard.ReqChan
	}

	return &ShardManager{
		shards:          shards,
		shardReqMap:     shardReqMap,
		globalErrorChan: globalErrorChan,
		ShardErrorChan:  shardErrorChan,
		sigChan:         make(chan os.Signal, 1),
		shardCount:      shardCount,
	}
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

	close(manager.ShardErrorChan) // Close the error channel after all Shards stop
	wg.Wait()                     // Wait for all shard goroutines to exit.
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

func (manager *ShardManager) GetShardInfo(key string) (id ShardID, c chan *ops.StoreOp) {
	hash := xxhash.Sum64String(key)
	id = ShardID(hash % uint64(manager.GetShardCount()))
	return id, manager.GetShard(id).ReqChan
}

// GetShardCount returns the number of shards managed by this ShardManager.
func (manager *ShardManager) GetShardCount() int8 {
	return int8(len(manager.shards))
}

// GetShard returns the ShardThread for the given ShardID.
func (manager *ShardManager) GetShard(id ShardID) *ShardThread {
	if int(id) < len(manager.shards) {
		return manager.shards[id]
	}
	return nil
}

// RegisterIOThread registers a io-thread with all Shards present in the ShardManager.
func (manager *ShardManager) RegisterIOThread(id string, request, processing chan *ops.StoreResponse) {
	for _, shard := range manager.shards {
		shard.registerIOThread(id, request, processing)
	}
}

func (manager *ShardManager) UnregisterIOThread(id string) {
	for _, shard := range manager.shards {
		shard.unregisterIOThread(id)
	}
}

package shard

/*
#define _GNU_SOURCE
#cgo CFLAGS: -pthread
#cgo LDFLAGS: -pthread
#include <pthread.h>
#include <sched.h>

	void SetThreadAffinity(int cpu) {
	    cpu_set_t cpuset;
	    CPU_ZERO(&cpuset);
	    CPU_SET(cpu, &cpuset);
	    pthread_setaffinity_np(pthread_self(), sizeof(cpu_set_t), &cpuset);
	}

	int GetThreadAffinity() {
    cpu_set_t cpuset;
    pthread_t current_thread = pthread_self();
    CPU_ZERO(&cpuset);
    int ret = pthread_getaffinity_np(current_thread, sizeof(cpu_set_t), &cpuset);

    if (ret != 0) {
        printf("Error getting CPU affinity: %d\n", ret);
        exit(EXIT_FAILURE);
    }

    for (int i = 0; i < CPU_SETSIZE; i++) {
        if (CPU_ISSET(i, &cpuset)) {
            return i;  // Return the first CPU where the thread is running
        }
    }
    return -1; // Should not reach here
}
*/
import "C"

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"

	"github.com/cespare/xxhash/v2"
	"github.com/dicedb/dice/internal/ops"
	dstore "github.com/dicedb/dice/internal/store"
)

func setCPUAffinity(cpu int) {
	C.SetThreadAffinity(C.int(cpu))
}

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
func NewShardManager(shardCount uint8, watchChan chan dstore.QueryWatchEvent, globalErrorChan chan error, logger *slog.Logger) *ShardManager {
	shards := make([]*ShardThread, shardCount)
	shardReqMap := make(map[ShardID]chan *ops.StoreOp)
	shardErrorChan := make(chan *ShardError)

	for i := uint8(0); i < shardCount; i++ {
		// Shards are numbered from 0 to shardCount-1
		shard := NewShardThread(i, globalErrorChan, shardErrorChan, watchChan, logger)
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
	for i, shard := range manager.shards {
		shard := shard

		wg.Add(1)
		go func() {
			runtime.LockOSThread()
			setCPUAffinity(i)
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

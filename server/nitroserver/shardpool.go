package nitroserver

import (
	"context"
	"hash/fnv"
	"runtime"
	"strconv"
	"sync"

	"github.com/charmbracelet/log"
	"github.com/dicedb/dice/core"
	"github.com/dicedb/dice/server"
)

var shardPool *ShardPool

type ShardThread struct {
	store    *core.Store
	srdreqch chan *Operation
}

type ShardPool struct {
	shardThreads []*ShardThread
}

func NewShardPool(ctx context.Context, wg *sync.WaitGroup, cores int) *ShardPool {
	log.Info("Initializing Shard pool with Cores = " + strconv.Itoa(cores))
	runtime.GOMAXPROCS(cores)
	shardPool = &ShardPool{}

	shardPool.shardThreads = make([]*ShardThread, cores)
	for i := 0; i < cores; i++ {
		store := core.NewStore()

		shardPool.shardThreads[i] = &ShardThread{
			store:    store,
			srdreqch: make(chan *Operation),
		}

		go shardPool.shardThreads[i].Run()

		// Add watcher for Store
		// Need to verify if this wirks as expected
		go server.WatchKeys(ctx, wg, store)
	}

	return shardPool
}

// Look up functional options for abstracting if we need to switch the toppology
func findOwnerShard(key string) int {
	hash := fnv.New32a()
	hash.Write([]byte(key))
	hashValue := int(hash.Sum32())
	bucket := hashValue % GetDiceCores()
	return bucket
}

func (shardPool *ShardPool) Submit(op *Operation) {
	// Non Key Operation.
	// We will fan out and execute Operation on all Shards
	if len(op.keys) == 0 {
		op.targetExecutionShards = GetDiceCores()

		for i := 0; i < GetDiceCores(); i++ {
			shardPool.shardThreads[i].srdreqch <- op
		}
	} else {
		// We have a Key operation and we can target
		// specific shardds to execute the operation
		op.targetExecutionShards = len(op.keys)
		for _, key := range op.keys {
			index := findOwnerShard(key)
			shardPool.shardThreads[index].srdreqch <- op
		}
	}
}

func (t *ShardThread) Run() {
	for op := range t.srdreqch {
		var byteResponse = core.ExecuteCommand(op.cmd, op.conn, t.store)
		op.ResultCH <- &IOResult{message: byteResponse, targetExecutionShards: op.targetExecutionShards, cmd: op.cmd}
	}
}

func (shardPool *ShardPool) SubmitClientRequest(op *Operation) {
	shardPool.Submit(op)
}

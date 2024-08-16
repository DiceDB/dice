package server

import (
	"hash/fnv"
	"runtime"

	"github.com/dicedb/dice/core"
	"github.com/charmbracelet/log"
)

var SHARDPOOL_SIZE int = max(1, runtime.NumCPU() - 1)
var CONCURRENT_CLIENTS = 1000

var ipool *IOThreadPool
var spool *ShardPool

// break out OP pool from shard pool
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}


func init() {
	log.Info("Initializinf thread and shard pool")
	ipool = NewIOThreadPool(CONCURRENT_CLIENTS) // handle client connections // ? do we need a io pool here? Create new thread ondemand as a new client requests?
	runtime.GOMAXPROCS(SHARDPOOL_SIZE)
	spool = NewShardPool(SHARDPOOL_SIZE) // handles data
}

// Shouod we use reddis command here?
type Operation struct {
	conn *core.Client
	// cmd *core.RedisCmd
	cmds *core.RedisCmds
	keys []string
	ResultCH chan *IOResult
}

type IOResult struct {
	message string
}

type IORequest struct {
	conn *core.Client
	// cmd *core.RedisCmd
	cmds *core.RedisCmds
	keys []string
}

type IOThread struct {
	ioreqch chan *IORequest
	ioresch chan *IOResult
}

// Look up functional options for abstracting if we need to switch the toppology
func findOwnerShard(key string) int {
	log.Info("Finding owner shard for key:", key)
	hash := fnv.New32a()
	hash.Write([]byte(key))
	hashValue := int(hash.Sum32())
	bucket := hashValue % SHARDPOOL_SIZE
	return bucket
}


// Should we pass redis command here?
func (p *ShardPool) Submit(op *Operation) {
	// Make this a coordinator
	log.Info("Operation submitted on Shard")
	// Non Key Operation.
	// We need to fan out and execute Operation on all Shards
	if len(op.keys) == 0 {
		log.Info("Fan out to all shards")
		for i := 0; i < SHARDPOOL_SIZE; i++ {
			p.shardThreads[i].srdreqch <- op // send response channel 
			// Listener. Listen for N/Shard data points returned, aggregate and send to IO Thread
		}

		
	}

	// We have a Key operation and we can target
	// specific shardds to execute the operation
	for _, key := range op.keys {
		// from the operation, find the owner shard
		index := findOwnerShard(key)
		log.Info("Sending operation to shard:", index)

		// put the operation in that shard
		// right now `ch` is unbuffered, but we can create a buffer,
		// enqueue it, and then batch process it, or
		// when we look at transactions in multi-threaded setup
		// we can re-order it and process it in the correct order
		p.shardThreads[index].srdreqch <- op // also send result channel 
		// here - listen on that channel for result // timeout 5sec
		// ID for command to return response to ? should commands be blocking for response. sequential execution?
	}

	

	// how do we handle failures / command exec failure & system failure
	// Do we need a error channel along with the response channel

}

func (t *IOThread) Run() {
	for req := range t.ioreqch {
		// read the request
		// create the operation
		spool.Submit(&Operation{
			conn: req.conn,
			// cmd: req.cmd,
			cmds: req.cmds,
			keys: req.keys,
			ResultCH: t.ioresch,
		})
	}
}

type IOThreadPool struct {
	pool chan *IOThread
}

func NewIOThreadPool(poolsize int) *IOThreadPool {
	p := IOThreadPool{}
	p.Init(poolsize)
	return &p
}

func (p *IOThreadPool) Init(poolsize int) {
	log.Info("Initializing Thread pool with pool size =", poolsize)
	p.pool = make(chan *IOThread, poolsize)
	iothread := &IOThread{
		ioreqch: make(chan *IORequest),
		ioresch: make(chan *IOResult),
	}
	go iothread.Run()
	for i := 0; i < poolsize; i++ {
		p.pool <- iothread
	}
}

func (p *IOThreadPool) Get() *IOThread {
	return <-p.pool
}

func (p *IOThreadPool) Put(t *IOThread) {
	p.pool <- t
}

type ShardThread struct {
	store *core.Store
	srdreqch chan *Operation
}

func (t *ShardThread) Run() {
	for op := range t.srdreqch {
		// execute the operation and create the result
		log.Info("Executing command:", op.cmds)
		EvalAndRespondOnChannel(*op.cmds, op.conn, t.store, op.ResultCH)

		// var resp = core.ExecuteCommandThreaded(*op.cmds, op.conn, t.store)

		// op.ResultCH <- &IOResult{"Returned hard coded resp"} // Missing a listener since its not buffered

		//fmt.Println("Command execution response:", resp)
		//for _, respStr := range resp {
		//	op.ResultCH <- &IOResult{respStr}
		//}
	}
	log.Info("Done executing on Shard")
}


func EvalAndRespondOnChannel(cmds core.RedisCmds, c *core.Client, store *core.Store, resultCh chan *IOResult) {
	for _, cmd := range cmds {
		log.Info("Got command: ", cmd.Cmd)
		resultCh<- &IOResult{message: string(core.ExecuteCommand(cmd, c, store))}
	}
}

type ShardPool struct {
	shardThreads []*ShardThread
}

func NewShardPool(poolsize int) *ShardPool {
	p := ShardPool{}
	p.Init(poolsize)
	return &p
}

func (p *ShardPool) Init(poolsize int) {
	log.Info("Initializing Shard pool with poolsize=", poolsize)
	p.shardThreads = make([]*ShardThread, poolsize)
	for i := 0; i < poolsize; i++ {
		p.shardThreads[i] = &ShardThread{
			store: core.NewStore(),
			srdreqch: make(chan *Operation),
		}
		go p.shardThreads[i].Run()
	}
}

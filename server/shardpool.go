package server

import (
	"fmt"
	"hash/fnv"
	"net"
	"runtime"

	"github.com/dicedb/dice/core"
)

var SHARDPOOL_SIZE int = runtime.NumCPU()
var CONCURRENT_CLIENTS = 1000

var ipool *IOThreadPool
var spool *ShardPool

func init() {
	ipool = NewIOThreadPool(CONCURRENT_CLIENTS) // handle client connections
	runtime.GOMAXPROCS(SHARDPOOL_SIZE)
	spool = NewShardPool(SHARDPOOL_SIZE) // handles data
}

// Shouod we use reddis command here?
type Operation struct {
	Key      string
	Value    string
	Op       string
	ResultCH chan<- *Result
}

type Result struct {
	message string
}

type Request struct {
	conn net.Conn
}

type IOThread struct {
	reqch chan *Request
	resch chan *Result
}

func (t *IOThread) Run() {
	for req := range t.reqch {
		fmt.Println("handling req", req)
		// read the request
		// create the operation
		spool.Submit(&Operation{
			ResultCH: t.resch,
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
	p.pool = make(chan *IOThread, poolsize)
	iothread := &IOThread{
		reqch: make(chan *Request),
		resch: make(chan *Result),
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
	reqch chan *Operation
}

func (t *ShardThread) Run() {
	for op := range t.reqch {
		fmt.Println("handling op", op)
		// execute the operation and create the result
		var msg = ""
		executeCommand(cmd*RedisCmd, c*Client, t.store)

		fmt.Println(msg)
		op.ResultCH <- &Result{msg}
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
	p.shardThreads = make([]*ShardThread, poolsize)
	for i := 0; i < poolsize; i++ {
		p.shardThreads[i] = &ShardThread{
			store: core.NewStore(),
			reqch: make(chan *Operation),
		}
		go p.shardThreads[i].Run()
	}
}

func findOwnerShard(key string) int {
	hash := fnv.New32a()
	hash.Write([]byte(key))
	hashValue := int(hash.Sum32())
	bucket := hashValue % SHARDPOOL_SIZE
	return bucket
}

func (p *ShardPool) Submit(op *Operation) {
	// from the operation, find the owner shard
	index := findOwnerShard(op.Key)

	// put the operation in that shard
	// right now `ch` is unbuffered, but we can create a buffer,
	// enqueue it, and then batch process it, or
	// when we look at transactions in multi-threaded setup
	// we can re-order it and process it in the correct order
	p.shardThreads[index].reqch <- op
}

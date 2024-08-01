package server

import (
	"fmt"
	"net"
	"runtime"

	"github.com/dicedb/dice/core"
)

var numCPU int = runtime.NumCPU()

var ipool *IOThreadPool
var spool *ShardPool

func init() {
	ipool = NewIOThreadPool(numCPU)
	spool = NewShardPool(numCPU)
}

type Operation struct {
	Key      string
	Value    string
	Op       string
	ResultCH chan<- *Result
}

type Result struct{}

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
		op.ResultCH <- &Result{}
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

func (p *ShardPool) Submit(op *Operation) {
	// from the operation, find the owner shard
	index := 0

	// put the operation in that shard
	// right now `ch` is unbuffered, but we can create a buffer,
	// enqueue it, and then batch process it, or
	// when we look at transactions in multi-threaded setup
	// we can re-order it and process it in the correct order
	p.shardThreads[index].reqch <- op
}

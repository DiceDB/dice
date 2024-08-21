package nitroserver

import (
	"github.com/dicedb/dice/core"
)

type Operation struct {
	conn                  *core.Client
	cmd                   *core.RedisCmd
	keys                  []string
	ResultCH              chan *IOResult
	targetExecutionShards int
}

type IOResult struct {
	message               []byte
	cmd                   *core.RedisCmd
	targetExecutionShards int
}

type IORequest struct {
	conn *core.Client
	cmd  *core.RedisCmd
	keys []string
}

type IOThread struct {
	ioreqch chan *IORequest
	ioresch chan *IOResult
}

func NewIOThread() *IOThread {
	iothread := &IOThread{
		ioreqch: make(chan *IORequest),
		ioresch: make(chan *IOResult),
	}
	return iothread
}

func (t *IOThread) Run(sPool *ShardPool) {
	for req := range t.ioreqch {
		sPool.SubmitClientRequest(
			&Operation{
				conn:     req.conn,
				cmd:      req.cmd,
				keys:     req.keys,
				ResultCH: t.ioresch,
			})
	}
}

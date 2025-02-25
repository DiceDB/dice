package shard

import "github.com/dicedb/dice/internal/shardthread"

type Shard struct {
	ID     int
	Thread *shardthread.ShardThread
}

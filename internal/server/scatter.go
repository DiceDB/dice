package server

import (
	"github.com/dicedb/dice/internal/cmd"
	"github.com/dicedb/dice/internal/comm"
	"github.com/dicedb/dice/internal/shard"
)

// Since PING doesn't require special logic for sending to all shards
// It just checks health on all shards
func scatterPING(mgr *shard.ShardManager, redisCmd *cmd.RedisCmd, c *comm.Client) (cm []cmd.RedisCmd) {
	for i := 0; i < mgr.GetShardCount(); i++ {
		cm = append(cm, *redisCmd)
	}
	return cm
}

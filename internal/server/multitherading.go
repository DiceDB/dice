package server

import (
	"bytes"

	"github.com/dicedb/dice/internal/cmd"
	"github.com/dicedb/dice/internal/comm"
	"github.com/dicedb/dice/internal/eval"
	"github.com/dicedb/dice/internal/ops"
	"github.com/dicedb/dice/internal/shard"
	"github.com/twmb/murmur3"
)

func getShard(key string, n uint32) uint32 {
	// Hash the key using MurmurHash3
	hash := murmur3.Sum32([]byte(key))
	// Apply modulo operation
	return hash % n
}

type ShardingMeta struct {
	Cmd     string
	Breakup func(mgr *shard.ShardManager, redisCmd *cmd.RedisCmd, c *comm.Client) []cmd.RedisCmd
	Gather  func(responses ...eval.EvalResponse) []byte
}

var (
	MultithreadingCmds = map[string]ShardingMeta{}

	pingCmdMeta = ShardingMeta{
		Cmd:     "PING",
		Breakup: scatterPING,
		Gather:  gatherPING,
	}
)

func init() {
	MultithreadingCmds["PING"] = pingCmdMeta
}

func (s *AsyncServer) cmdsBreakup(redisCmd *cmd.RedisCmd, c *comm.Client) []cmd.RedisCmd {
	val, ok := MultithreadingCmds[redisCmd.Cmd]
	if !ok {
		return []cmd.RedisCmd{*redisCmd}
	}

	return val.Breakup(s.shardManager, redisCmd, c)
}

func (s *AsyncServer) scatter(cmds []cmd.RedisCmd, c *comm.Client) {
	// single sharded command
	if len(cmds) == 1 {
		key := cmds[0].Args[0]
		id := getShard(key, uint32(s.shardManager.GetShardCount()))
		s.shardManager.GetShard(shard.ShardID(id)).ReqChan <- &ops.StoreOp{
			Cmd:      &cmds[0],
			WorkerID: "server",
			ShardID:  0,
			Client:   c,
		}
	} else {
		// multishard command
		// Condition for command that requires all shards
		if len(cmds) == s.shardManager.GetShardCount() {
			for i := 0; i < len(cmds); i++ {
				s.shardManager.GetShard(shard.ShardID(i)).ReqChan <- &ops.StoreOp{
					Cmd:      &cmds[i],
					WorkerID: "server",
					ShardID:  0,
					Client:   c,
				}
			}
		} else {
			for i := 0; i < len(cmds); i++ {
				key := cmds[i].Args[0]
				id := getShard(key, uint32(s.shardManager.GetShardCount()))
				s.shardManager.GetShard(shard.ShardID(id)).ReqChan <- &ops.StoreOp{
					Cmd:      &cmds[i],
					WorkerID: "server",
					ShardID:  0,
					Client:   c,
				}
			}
		}

	}
}

func (s *AsyncServer) gather(redisCmd *cmd.RedisCmd, buf *bytes.Buffer, numShards int) {
	// Loop to wait for messages from numberof shards
	var evalResp []eval.EvalResponse
	for i := 0; i < numShards; i++ {
		select {
		case resp := <-s.ioChan:
			evalResp = append(evalResp, *&resp.EvalResponse)
			// should another case for time.Sleep(n) to max wait for response
		}
	}

	val, ok := MultithreadingCmds[redisCmd.Cmd]
	if !ok {
		buf.Write(evalResp[0].Result.([]byte))
	} else {
		buf.Write(val.Gather(evalResp...))
	}

}

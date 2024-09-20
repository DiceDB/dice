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
	hash := murmur3.Sum32([]byte(key))
	return hash % n
}

func (s *AsyncServer) cmdsBreakup(redisCmd *cmd.RedisCmd, c *comm.Client) []cmd.RedisCmd {
	val, ok := WorkerCmdsMeta[redisCmd.Cmd]
	if !ok {
		return []cmd.RedisCmd{*redisCmd}
	}

	// if command supports multisharding then send the command
	// to the respective breakup function
	// Which can return array of broken down commands
	return val.Breakup(s.shardManager, redisCmd, c)
}

func (s *AsyncServer) scatter(cmds []cmd.RedisCmd, c *comm.Client) {
	// Otherwise check for the shard based on the key using hash
	// and send it to the particular shard
	for i := 0; i < len(cmds); i++ {
		var id uint32
		if len(cmds[i].Args) > 0 {
			key := cmds[i].Args[i]
			id = getShard(key, uint32(s.shardManager.GetShardCount()))
		}
		s.shardManager.GetShard(shard.ShardID(id)).ReqChan <- &ops.StoreOp{
			Cmd:      &cmds[i],
			WorkerID: "server",
			ShardID:  int(id),
			Client:   c,
		}
	}
}

func (s *AsyncServer) gather(redisCmd *cmd.RedisCmd, buf *bytes.Buffer, numShards int, c CmdType) {
	// Loop to wait for messages from numberof shards
	var evalResp []eval.EvalResponse
	for i := 0; i < numShards; i++ {
		resp, ok := <-s.ioChan
		if ok {
			evalResp = append(evalResp, resp.EvalResponse)
		}
	}

	// Check if command supports multisharding
	val, ok := WorkerCmdsMeta[redisCmd.Cmd]
	if !ok {
		buf.Write(evalResp[0].Result.([]byte))
		return
	}

	switch c {
	case SingleShard, Custom:
		if evalResp[0].Error != nil {
			buf.Write([]byte(evalResp[0].Error.Error()))
			return
		}
		buf.Write(evalResp[0].Result.([]byte))

	case Multishard:
		buf.Write(val.Gather(evalResp...))
	}
}

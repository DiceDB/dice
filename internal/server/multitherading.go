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

// getShard calculates the shard ID for a given key using Murmur3 hashing.
// It returns the shard ID by computing the hash modulo the number of shards (n).
func getShard(key string, n uint32) uint32 {
	hash := murmur3.Sum32([]byte(key))
	return hash % n
}

// cmdsBreakup breaks down a Redis command into smaller commands if multisharding is supported.
// It uses the metadata to check if the command supports multisharding and calls the respective breakup function.
// If multisharding is not supported, it returns the original command in a slice.
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

// scatter distributes the Redis commands to the respective shards based on the key.
// For each command, it calculates the shard ID and sends the command to the shard's request channel for processing.
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

// gather collects the responses from multiple shards and writes the results into the provided buffer.
// It first waits for responses from all the shards and then processes the result based on the command type (SingleShard, Custom, or Multishard).
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

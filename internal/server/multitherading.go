package server

import (
	"bytes"
	"fmt"

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
		Breakup: breakupPING,
		Gather:  gatherPING,
	}
)

func init() {
	MultithreadingCmds["PING"] = pingCmdMeta
}

func (s *AsyncServer) cmdsBreakup(redisCmd *cmd.RedisCmd, c *comm.Client) []cmd.RedisCmd {

	// check if command supports multisharding
	val, ok := MultithreadingCmds[redisCmd.Cmd]
	// If not then just return the command as it is
	// no need to do futher breakup
	if !ok {
		return []cmd.RedisCmd{*redisCmd}
	}

	// If yes then send the command to the respective breakup function
	// Which can return array of broken down commands
	return val.Breakup(s.shardManager, redisCmd, c)
}

func (s *AsyncServer) scatter(cmds []cmd.RedisCmd, c *comm.Client) {
	// single sharded command
	if len(cmds) == 1 {
		var id uint32
		if len(cmds[0].Args) > 0 {
			key := cmds[0].Args[0]
			id = getShard(key, uint32(s.shardManager.GetShardCount()))
		}
		fmt.Println("Sending to the shard: ", id)
		s.shardManager.GetShard(shard.ShardID(id)).ReqChan <- &ops.StoreOp{
			Cmd:      &cmds[0],
			WorkerID: "server",
			ShardID:  int(id),
			Client:   c,
		}
	} else {
		// multishard command
		// Condition for command that requires all shards such as PING
		if len(cmds) == s.shardManager.GetShardCount() {
			for i := 0; i < len(cmds); i++ {
				s.shardManager.GetShard(shard.ShardID(i)).ReqChan <- &ops.StoreOp{
					Cmd:      &cmds[i],
					WorkerID: "server",
					ShardID:  i,
					Client:   c,
				}
			}
		} else {
			// Otherwise check for the shard based on the key using hash
			// and send it to the particular shard
			for i := 0; i < len(cmds); i++ {
				var id uint32
				if len(cmds[0].Args) > 0 {
					key := cmds[0].Args[0]
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

	}
}

func (s *AsyncServer) gather(redisCmd *cmd.RedisCmd, buf *bytes.Buffer, numShards int) {
	// Loop to wait for messages from numberof shards
	var evalResp []eval.EvalResponse
	for i := 0; i < numShards; i++ {
		select {
		case resp := <-s.ioChan:
			fmt.Println("Getting response from Shard:", i)
			evalResp = append(evalResp, *&resp.EvalResponse)
			// should another case for time.Sleep(n) to max wait for response
		}
	}

	// Check if command supports multisharding
	val, ok := MultithreadingCmds[redisCmd.Cmd]

	// If single shard then jsut get the reponse, encode it before sending to client
	// Encoding function is not yet implemented
	if !ok {
		buf.Write(evalResp[0].Result.([]byte))
	} else {
		// If multishard command, then send it to the respective Gather function
		buf.Write(val.Gather(evalResp...))
	}
}

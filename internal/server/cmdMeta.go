package server

import (
	"github.com/dicedb/dice/internal/cmd"
	"github.com/dicedb/dice/internal/comm"
	"github.com/dicedb/dice/internal/eval"
	"github.com/dicedb/dice/internal/shard"
)

// Define a custom type for the enum
type CmdType int

// Declare enum values using iota
const (
	Global      CmdType = iota // Global commands don't need to touch shards
	SingleShard                // Single shard commands work with only one shard
	Multishard                 // Multishard commands work with multiple shards and needs scatter, gather logic
	Custom                     // Custom flows needs to directly connect to the client
)

type CmdsMeta struct {
	Cmd          string
	Breakup      func(mgr *shard.ShardManager, redisCmd *cmd.RedisCmd, c *comm.Client) []cmd.RedisCmd
	Gather       func(responses ...eval.EvalResponse) []byte
	RespNoShards func(args []string) []byte
	CmdType
}

var (
	WorkerCmdsMeta = map[string]CmdsMeta{}

	// Global commands which doesn't need to touch shards
	infoCmdMeta = CmdsMeta{
		Cmd:          "INFO",
		CmdType:      Global,
		RespNoShards: respINFO,
	}
	pingCmdMeta = CmdsMeta{
		Cmd:          "PING",
		CmdType:      Global,
		RespNoShards: respPING,
	}

	// Single shard commands which doesn't need breakup and gather logic
	setCmdMeta = CmdsMeta{
		Cmd:     "SET",
		CmdType: SingleShard,
	}
	getCmdMeta = CmdsMeta{
		Cmd:     "GET",
		CmdType: SingleShard,
	}
	getsetCmdMeta = CmdsMeta{
		Cmd:     "GETSET",
		CmdType: SingleShard,
	}

	// Multishard commands which needs breakup and gather logic functions

	// Custom commands like Qwatch needs custom logic
)

func init() {
	WorkerCmdsMeta["INFO"] = infoCmdMeta
	WorkerCmdsMeta["PING"] = pingCmdMeta

	WorkerCmdsMeta["SET"] = setCmdMeta
	WorkerCmdsMeta["GET"] = getCmdMeta
	WorkerCmdsMeta["GETSET"] = getsetCmdMeta
}

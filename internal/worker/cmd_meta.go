package worker

import (
	"fmt"

	"github.com/dicedb/dice/internal/cmd"
	"github.com/dicedb/dice/internal/eval"
	"github.com/dicedb/dice/internal/logger"
)

type CmdType int

const (
	Global CmdType = iota
	SingleShard
	MultiShard
	Custom
)

const (
	// Global commands
	CmdPing  = "PING"
	CmdAbort = "ABORT"
	CmdAuth  = "AUTH"

	// Single-shard commands.
	CmdSet    = "SET"
	CmdGet    = "GET"
	CmdGetSet = "GETSET"
)

type CommandsMeta struct {
	CmdType
	Cmd                  string
	isSimpleEnc          bool
	WorkerCommandHandler func([]string) []byte
	decomposeCommand     func(redisCmd *cmd.RedisCmd) []*cmd.RedisCmd
	composeResponse      func(responses ...eval.EvalResponse) interface{}
}

var WorkerCommandsMeta = map[string]CommandsMeta{
	// Global commands.
	CmdPing: {
		CmdType:              Global,
		WorkerCommandHandler: eval.RespPING,
	},
	CmdAbort: {
		CmdType:     Custom,
		isSimpleEnc: true,
	},
	CmdAuth: {
		CmdType:     Custom,
		isSimpleEnc: true,
	},

	// Single-shard commands.
	CmdSet: {
		CmdType:     SingleShard,
		isSimpleEnc: true,
	},
	CmdGet: {
		CmdType:     SingleShard,
		isSimpleEnc: true,
	},
	CmdGetSet: {
		CmdType:     SingleShard,
		isSimpleEnc: true,
	},
}

func init() {
	l := logger.New(logger.Opts{WithTimestamp: true})
	// Validate the metadata for each command
	for c, meta := range WorkerCommandsMeta {
		if err := validateCmdMeta(c, meta); err != nil {
			l.Error("error validating worker command metadata %s: %v", c, err)
		}
	}
}

// validateCmdMeta ensures that the metadata for each command is properly configured
func validateCmdMeta(c string, meta CommandsMeta) error {
	switch meta.CmdType {
	case Global:
		if meta.WorkerCommandHandler == nil {
			return fmt.Errorf("global command %s must have WorkerCommandHandler function", c)
		}
	case MultiShard:
		if meta.decomposeCommand == nil || meta.composeResponse == nil {
			return fmt.Errorf("multi-shard command %s must have both decomposeCommand and composeResponse implemented", c)
		}
	case SingleShard, Custom:
		// No specific validations for these types currently
	default:
		return fmt.Errorf("unknown command type for %s", c)
	}

	return nil
}

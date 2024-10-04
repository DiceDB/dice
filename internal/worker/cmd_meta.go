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

// Global commands
const (
	CmdPing  = "PING"
	CmdAbort = "ABORT"
	CmdAuth  = "AUTH"
)

// Single-shard commands.
const (
	CmdSet    = "SET"
	CmdGet    = "GET"
	CmdGetSet = "GETSET"
)

type CmdMeta struct {
	CmdType
	Cmd                  string
	WorkerCommandHandler func([]string) []byte

	// decomposeCommand is a function that takes a DiceDB command and breaks it down into smaller,
	// manageable DiceDB commands for each shard processing. It returns a slice of DiceDB commands.
	decomposeCommand func(DiceDBCmd *cmd.DiceDBCmd) []*cmd.DiceDBCmd

	// composeResponse is a function that combines multiple responses from the execution of commands
	// into a single response object. It accepts a variadic parameter of EvalResponse objects
	// and returns a unified response interface.
	composeResponse func(responses ...eval.EvalResponse) interface{}
}

var CommandsMeta = map[string]CmdMeta{
	// Global commands.
	CmdPing: {
		CmdType:              Global,
		WorkerCommandHandler: eval.RespPING,
	},
	CmdAbort: {
		CmdType: Custom,
	},
	CmdAuth: {
		CmdType: Custom,
	},

	// Single-shard commands.
	CmdSet: {
		CmdType: SingleShard,
	},
	CmdGet: {
		CmdType: SingleShard,
	},
	CmdGetSet: {
		CmdType: SingleShard,
	},
}

func init() {
	l := logger.New(logger.Opts{WithTimestamp: true})
	// Validate the metadata for each command
	for c, meta := range CommandsMeta {
		if err := validateCmdMeta(c, meta); err != nil {
			l.Error("error validating worker command metadata %s: %v", c, err)
		}
	}
}

// validateCmdMeta ensures that the metadata for each command is properly configured
func validateCmdMeta(c string, meta CmdMeta) error {
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

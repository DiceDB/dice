package worker

import (
	"fmt"

	"github.com/dicedb/dice/internal/cmd"
	"github.com/dicedb/dice/internal/eval"
	"github.com/dicedb/dice/internal/logger"
	"github.com/dicedb/dice/internal/querymanager"
)

type CmdType int

const (
	Global CmdType = iota
	SingleShard
	MultiShard
	Custom
	Watch
)

// Global commands
const (
	CmdPing  = "PING"
	CmdAbort = "ABORT"
	CmdAuth  = "AUTH"
)

// Single-shard commands.
const (
	CmdSet      = "SET"
	CmdGet      = "GET"
	CmdGetSet   = "GETSET"
	CmdGetWatch = "GET.WATCH"
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
	// and returns a unified response interface. It is used in the command type "MultiShard"
	composeResponse func(responses ...eval.EvalResponse) interface{}

	// watchResponse constructs a response object from a command execution.
	// It accepts the command name, key, and result as parameters and returns a slice of interfaces.
	// The returned slice contains the command, the key, and the result, which could be any type including an error.
	// response from this function is required when server sends to clients without the client explicitly requesting them
	watchResponse func(cmd, key string, result interface{}) []interface{}
}

var CommandsMeta = map[string]CmdMeta{
	// Global commands.
	CmdPing: {
		CmdType:              Global,
		WorkerCommandHandler: eval.RespPING,
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

	// Custom commands.
	CmdAbort: {
		CmdType: Custom,
	},
	CmdAuth: {
		CmdType: Custom,
	},

	// Watch commands
	CmdGetWatch: {
		CmdType:       Watch,
		watchResponse: querymanager.GenericWatchResponse,
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
	case SingleShard, Watch, Custom:
		// No specific validations for these types currently
	default:
		return fmt.Errorf("unknown command type for %s", c)
	}

	return nil
}

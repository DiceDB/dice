package worker

import (
	"fmt"

	"github.com/dicedb/dice/internal/cmd"
	"github.com/dicedb/dice/internal/eval"
	"github.com/dicedb/dice/internal/logger"
)

type CmdType int

const (
	// Global represents a command that applies globally across all shards or nodes.
	// This type of command doesn't target a specific shard but affects the entire system.
	Global CmdType = iota

	// SingleShard represents a command that operates on a single shard.
	// This command is scoped to execute on one specific shard, optimizing for shard-local operations.
	SingleShard

	// MultiShard represents a command that operates across multiple shards.
	// This type of command spans more than one shard and may involve coordination between shards.
	MultiShard

	// Custom represents a command that is user-defined or has custom logic.
	// This command type allows for flexibility in executing specific, non-standard operations.
	Custom

	// Watch represents a command that is used to monitor changes or events.
	// This type of command listens for changes on specific keys or resources and responds accordingly.
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
	CmdSet         = "SET"
	CmdGet         = "GET"
	CmdGetSet      = "GETSET"
	CmdGetWatch    = "GET.WATCH"
	CmdZRangeWatch = "ZRANGE.WATCH"
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
		CmdType: Watch,
	},
	CmdZRangeWatch: {
		CmdType: Watch,
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

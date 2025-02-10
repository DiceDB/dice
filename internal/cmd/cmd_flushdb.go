package cmd

import (
	"github.com/dicedb/dice/internal/store"
)

var cFLUSHDB = &DiceDBCommand{
	Name:      "FLUSHDB",
	HelpShort: "FLUSHDB deletes all keys.",
	Eval:      evalFLUSHDB,
}

func init() {
	commandRegistry.AddCommand(cFLUSHDB)
}

// TODO: FLUSHDB is a multi-shard command.
// It should be executed on all shards, hence we need to
// scatter and gather the results.

// FLUSHDB deletes all keys.
// The function expects no arguments
//
// Parameters:
//   - c *Cmd: The command context containing the arguments
//   - s *dstore.Store: The data store instance
//
// Returns:
//   - *CmdRes: OK or nil
//   - error: Error if wrong number of arguments
func evalFLUSHDB(c *Cmd, s *store.Store) (*CmdRes, error) {
	if len(c.C.Args) != 0 {
		return cmdResNil, errWrongArgumentCount("FLUSHDB")
	}

	store.Reset(s)
	return cmdResOK, nil
}

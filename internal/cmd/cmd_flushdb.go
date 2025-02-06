package cmd

import (
	"strings"

	"github.com/dicedb/dice/internal/store"
	dstore "github.com/dicedb/dice/internal/store"
)

const (
	SYNC  string = "sync"
	ASYNC string = "async"
)

var cFLUSHDB = &DiceDBCommand{
	Name:      "FLUSHDB",
	HelpShort: "FLUSHDB removes all keys from the currently selected database.",
	Eval:      evalFLUSHDB,
}

func init() {
	commandRegistry.AddCommand(cFLUSHDB)
}

// FLUSHDB is used to remove all keys from the currently selected database in a DiceDB instance.
//
// # The function expects no arguments
//
// Parameters:
//   - c *Cmd: The command context containing the arguments
//   - s *dstore.Store: The data store instance
//
// Returns:
//   - *CmdRes: OK or nil
//   - error: Error if wrong number of arguments
func evalFLUSHDB(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) > 1 {
		return cmdResNil, errWrongArgumentCount("FLUSHDB")
	}

	flushType := SYNC
	if len(c.C.Args) == 1 {
		flushType = strings.ToUpper(c.C.Args[0])
	}

	switch flushType {
	case SYNC, ASYNC:
		store.ResetStore(s)
	default:
		return cmdResNil, errInvalidSyntax("FLUSHDB")
	}

	return cmdResOK, nil
}

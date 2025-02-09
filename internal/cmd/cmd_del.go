package cmd

import (
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dicedb-go/wire"
)

var cDEL = &DiceDBCommand{
	Name:      "DEL",
	HelpShort: "DEL deletes all the specified keys in args list",
	Eval:      evalDEL,
}

func init() {
	commandRegistry.AddCommand(cDEL)
}

// TODO: DEL command is actually a multi-key command so this needs
// to be scattered and gathered one step before this.

// evalDEL deletes all the specified keys in args list.
//
// Parameters:
//   - c *Cmd: The command context containing the arguments
//   - s *dstore.Store: The data store instance
//
// Returns:
//   - *CmdRes: Response containing the count of total deleted keys
//   - error: Error if wrong number of arguments or wrong value type
func evalDEL(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) < 1 {
		return cmdResNil, errWrongArgumentCount("DEL")
	}

	var count int
	for _, key := range c.C.Args {
		if ok := s.Del(key); ok {
			count++
		}
	}

	return &CmdRes{R: &wire.Response{
		Value: &wire.Response_VInt{VInt: int64(count)},
	}}, nil
}

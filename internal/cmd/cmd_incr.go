package cmd

import (
	dstore "github.com/dicedb/dice/internal/store"
)

var cINCR = &DiceDBCommand{
	Name:      "INCR",
	HelpShort: "INCR increments the value of the specified key in args by 1",
	Eval:      evalDECR,
}

func init() {
	commandRegistry.AddCommand(cINCR)
}

// evalINCR increments an integer value stored at the specified key by 1.
//
// The function expects exactly one argument: the key to increment.
// If the key does not exist, it is initialized with value 1.
// If the key exists but does not contain an integer, an error is returned.
//
// Parameters:
//   - c *Cmd: The command context containing the arguments
//   - s *dstore.Store: The data store instance
//
// Returns:
//   - *CmdRes: Response containing the new integer value after increment
//   - error: Error if wrong number of arguments or wrong value type
func evalINCR(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) != 1 {
		return cmdResNil, errWrongArgumentCount("INCR")
	}

	delta := int64(1)

	return incrDecr(c,s,delta)
}

// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	dstore "github.com/dicedb/dice/internal/store"
)

var cDECR = &DiceDBCommand{
	Name:      "DECR",
	HelpShort: "DECR decrements the value of the specified key in args by 1",
	Eval:      evalDECR,
}

func init() {
	commandRegistry.AddCommand(cDECR)
}

// evalDECR decrements an integer value stored at the specified key by 1.
//
// The function expects exactly one argument: the key to decrement.
// If the key does not exist, it is initialized with value -1.
// If the key exists but does not contain an integer, an error is returned.
//
// Parameters:
//   - c *Cmd: The command context containing the arguments
//   - s *dstore.Store: The data store instance
//
// Returns:
//   - *CmdRes: Response containing the new integer value after decrement
//   - error: Error if wrong number of arguments or wrong value type
func evalDECR(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) != 1 {
		return cmdResNil, errWrongArgumentCount("DECR")
	}

	return doIncr(c, s, -1)
}

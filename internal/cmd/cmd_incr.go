// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
)

var cINCR = &CommandMeta{
	Name:      "INCR",
	HelpShort: "INCR increments the value of the specified key in args by 1",
	Eval:      evalINCR,
	Execute:   executeINCR,
}

func init() {
	CommandRegistry.AddCommand(cINCR)
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
		return cmdResNil, errors.ErrWrongArgumentCount("INCR")
	}
	return doIncr(c, s, 1)
}

func executeINCR(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) != 1 {
		return cmdResNil, errors.ErrWrongArgumentCount("INCR")
	}
	shard := sm.GetShardForKey(c.C.Args[0])
	return evalINCR(c, shard.Thread.Store())
}

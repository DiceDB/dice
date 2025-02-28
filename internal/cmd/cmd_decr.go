// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
)

var cDECR = &CommandMeta{
	Name:      "DECR",
	Syntax:    "DECR key",
	HelpShort: "DECR decrements the value of the specified key in args by 1",
	HelpLong: `
DECR command decrements the integer at 'key' by one. Creates 'key' as -1 if absent.
Errors on wrong type or non-integer string. Limited to 64-bit signed integers.

Returns the new value of 'key' on success.
	`,
	Examples: `
localhost:7379> SET k 43
OK OK
localhost:7379> DECR k
OK 42
	`,
	Eval:    evalDECR,
	Execute: executeDECR,
}

func init() {
	CommandRegistry.AddCommand(cDECR)
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
		return cmdResNil, errors.ErrWrongArgumentCount("DECR")
	}

	return doIncr(c, s, -1)
}

func executeDECR(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) != 1 {
		return cmdResNil, errors.ErrWrongArgumentCount("DECR")
	}

	shard := sm.GetShardForKey(c.C.Args[0])
	return evalDECR(c, shard.Thread.Store())
}

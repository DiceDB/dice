// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"strconv"

	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
)

var cDECRBY = &CommandMeta{
	Name:      "DECRBY",
	HelpShort: "DECRBY decrements the specified key by the specified delta",
	Syntax:    "DECRBY key delta",
	Documentation: `
DECRBY command decrements the integer at 'key' by the delta specified. Creates 'key' with value (-delta) if absent.
Errors on wrong type or non-integer string. Limited to 64-bit signed integers.

Returns the new value of 'key' on success.

## Examples

` + "```" + `
localhost:7379> SET k 43
OK OK
localhost:7379> DECRBY k 10
OK 33
` + "```" + `
	`,
	Eval:    evalDECRBY,
	Execute: executeDECRBY,
}

func init() {
	CommandRegistry.AddCommand(cDECRBY)
}

func evalDECRBY(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) != 2 {
		return cmdResNil, errors.ErrWrongArgumentCount("DECRBY")
	}

	delta, err := strconv.ParseInt(c.C.Args[1], 10, 64)
	if err != nil {
		return cmdResNil, errors.ErrIntegerOutOfRange
	}

	return doIncr(c, s, -delta)
}

func executeDECRBY(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) != 2 {
		return cmdResNil, errors.ErrWrongArgumentCount("DECRBY")
	}
	shard := sm.GetShardForKey(c.C.Args[0])
	return evalDECRBY(c, shard.Thread.Store())
}

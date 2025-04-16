// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"strconv"

	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dicedb-go/wire"
)

var cDECRBY = &CommandMeta{
	Name:      "DECRBY",
	Syntax:    "DECRBY key delta",
	HelpShort: "DECRBY decrements the specified key by the specified delta",
	HelpLong: `
DECRBY command decrements the integer at 'key' by the delta specified. Creates 'key' with value (-delta) if absent.
The command raises an error if the value is a non-integer.

Returns the new value of 'key' on success.
	`,
	Examples: `
localhost:7379> SET k 43
OK
localhost:7379> DECRBY k 10
OK 33
localhost:7379> DECRBY k2 50
OK -50
localhost:7379> GET k2
OK "-50"
	`,
	Eval:    evalDECRBY,
	Execute: executeDECRBY,
}

func init() {
	CommandRegistry.AddCommand(cDECRBY)
}

func newDECRBYRes(newValue int64) *CmdRes {
	return &CmdRes{
		Rs: &wire.Result{
			Message: "OK",
			Status:  wire.Status_OK,
			Response: &wire.Result_DECRBYRes{
				DECRBYRes: &wire.DECRBYRes{Value: newValue},
			},
		},
	}
}

var (
	DECRBYResNilRes = newDECRBYRes(0)
)

func evalDECRBY(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) != 2 {
		return DECRBYResNilRes, errors.ErrWrongArgumentCount("DECRBY")
	}

	delta, err := strconv.ParseInt(c.C.Args[1], 10, 64)
	if err != nil {
		return DECRBYResNilRes, errors.ErrIntegerOutOfRange
	}

	_, newValue, err := doIncr(c, s, -delta)
	if err != nil {
		return DECRBYResNilRes, err
	}

	return newDECRBYRes(newValue), nil
}

func executeDECRBY(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) != 2 {
		return DECRBYResNilRes, errors.ErrWrongArgumentCount("DECRBY")
	}
	shard := sm.GetShardForKey(c.C.Args[0])
	return evalDECRBY(c, shard.Thread.Store())
}

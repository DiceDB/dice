// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dicedb-go/wire"
)

var cDECR = &CommandMeta{
	Name:      "DECR",
	Syntax:    "DECR key",
	HelpShort: "DECR decrements the value of the specified key in args by 1",
	HelpLong: `
DECR command decrements the integer at 'key' by one. Creates 'key' as -1 if absent.
The command raises an error if the value is a non-integer.

Returns the new value of 'key' on success.
	`,
	Examples: `
localhost:7379> SET k 43
OK
localhost:7379> DECR k
OK 42
localhost:7379> DECR k1
OK -1
localhost:7379> SET k2 v
OK
localhost:7379> DECR k2
ERR wrongtype operation against a key holding the wrong kind of value
	`,
	Eval:    evalDECR,
	Execute: executeDECR,
}

func init() {
	CommandRegistry.AddCommand(cDECR)
}

func newDECRRes(newValue int64) *CmdRes {
	return &CmdRes{
		Rs: &wire.Result{
			Message: "OK",
			Status:  wire.Status_OK,
			Response: &wire.Result_DECRRes{
				DECRRes: &wire.DECRRes{Value: newValue},
			},
		},
	}
}

var (
	DECRResNilRes = newDECRRes(0)
)

func evalDECR(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) != 1 {
		return DECRResNilRes, errors.ErrWrongArgumentCount("DECR")
	}

	_, newValue, err := doIncr(c, s, -1)
	if err != nil {
		return DECRResNilRes, err
	}

	return newDECRRes(newValue), nil
}

func executeDECR(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) != 1 {
		return DECRResNilRes, errors.ErrWrongArgumentCount("DECR")
	}

	shard := sm.GetShardForKey(c.C.Args[0])
	return evalDECR(c, shard.Thread.Store())
}

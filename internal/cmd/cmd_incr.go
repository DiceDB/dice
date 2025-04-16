// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dicedb-go/wire"
)

var cINCR = &CommandMeta{
	Name:      "INCR",
	Syntax:    "INCR key",
	HelpShort: "INCR increments the value of the specified key in args by 1",
	HelpLong: `
INCR increments the integer at 'key' by one. Creates 'key' as 1 if absent.
The command raises an error if the value is a non-integer.

Returns the new value of 'key' on success.
	`,
	Examples: `
localhost:7379> SET k 43
OK
localhost:7379> INCR k
OK 44
localhost:7379> INCR k2
OK 1
localhost:7379> GET k2
OK "1"
	`,
	Eval:    evalINCR,
	Execute: executeINCR,
}

func init() {
	CommandRegistry.AddCommand(cINCR)
}

func newINCRRes(newValue int64) *CmdRes {
	return &CmdRes{
		Rs: &wire.Result{
			Message: "OK",
			Status:  wire.Status_OK,
			Response: &wire.Result_INCRRes{
				INCRRes: &wire.INCRRes{Value: newValue},
			},
		},
	}
}

var (
	INCRResNilRes = newINCRRes(0)
)

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
		return INCRResNilRes, errors.ErrWrongArgumentCount("INCR")
	}
	_, newValue, err := doIncr(c, s, 1)
	if err != nil {
		return INCRResNilRes, err
	}

	return newINCRRes(newValue), nil
}

func executeINCR(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) != 1 {
		return INCRResNilRes, errors.ErrWrongArgumentCount("INCR")
	}
	shard := sm.GetShardForKey(c.C.Args[0])
	return evalINCR(c, shard.Thread.Store())
}

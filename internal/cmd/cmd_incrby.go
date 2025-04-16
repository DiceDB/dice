// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"strconv"

	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/object"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dicedb-go/wire"
)

var cINCRBY = &CommandMeta{
	Name:      "INCRBY",
	Syntax:    "INCRBY key delta",
	HelpShort: "INCRBY increments the specified key by the specified delta",
	HelpLong: `
INCRBY command increments the integer at 'key' by the delta specified. Creates 'key' with value (delta) if absent.
The command raises an error if the value is a non-integer.

Returns the new value of 'key' on success.
	`,
	Examples: `
localhost:7379> SET k 43
OK
localhost:7379> INCRBY k 10
OK 53
localhost:7379> INCRBY k2 50
OK 50
localhost:7379> GET k2
OK "50"
	`,
	Eval:    evalINCRBY,
	Execute: executeINCRBY,
}

func init() {
	CommandRegistry.AddCommand(cINCRBY)
}

func newINCRBYRes(newValue int64) *CmdRes {
	return &CmdRes{
		Rs: &wire.Result{
			Message: "OK",
			Status:  wire.Status_OK,
			Response: &wire.Result_INCRBYRes{
				INCRBYRes: &wire.INCRBYRes{Value: newValue},
			},
		},
	}
}

var (
	INCRBYResNilRes = newINCRBYRes(0)
)

func evalINCRBY(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) != 2 {
		return INCRBYResNilRes, errors.ErrWrongArgumentCount("INCRBY")
	}

	delta, err := strconv.ParseInt(c.C.Args[1], 10, 64)
	if err != nil {
		return INCRBYResNilRes, errors.ErrIntegerOutOfRange
	}

	_, newValue, err := doIncr(c, s, delta)
	if err != nil {
		return INCRBYResNilRes, err
	}

	return newINCRBYRes(newValue), nil
}

func executeINCRBY(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) != 2 {
		return INCRBYResNilRes, errors.ErrWrongArgumentCount("INCRBY")
	}
	shard := sm.GetShardForKey(c.C.Args[0])
	return evalINCRBY(c, shard.Thread.Store())
}

//nolint:unparam
func doIncr(c *Cmd, s *dstore.Store, delta int64) (oldValue, newValue int64, err error) {
	key := c.C.Args[0]
	obj := s.Get(key)
	if obj == nil {
		obj = s.NewObj(delta, -1, object.ObjTypeInt)
		s.Put(key, obj)
		return 0, delta, nil
	}

	switch obj.Type {
	case object.ObjTypeInt:
		break
	default:
		return 0, 0, errors.ErrWrongTypeOperation
	}

	oldValue, _ = obj.Value.(int64)

	newValue = oldValue + delta
	obj.Value = newValue

	return oldValue, newValue, nil
}

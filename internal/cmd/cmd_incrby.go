// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"strconv"

	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/object"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
)

var cINCRBY = &CommandMeta{
	Name:      "INCRBY",
	Syntax:    "INCRBY key delta",
	HelpShort: "INCRBY increments the specified key by the specified delta",
	HelpLong: `
INCRBY command increments the integer at 'key' by the delta specified. Creates 'key' with value (delta) if absent.
Errors on wrong type or non-integer string. Limited to 64-bit signed integers.

Returns the new value of 'key' on success.
	`,
	Examples: `
localhost:7379> SET k 43
OK OK
localhost:7379> INCRBY k 10
OK 53
	`,
	Eval:    evalINCRBY,
	Execute: executeINCRBY,
}

func init() {
	CommandRegistry.AddCommand(cINCRBY)
}

func evalINCRBY(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) != 2 {
		return cmdResNil, errors.ErrWrongArgumentCount("INCRBY")
	}

	delta, err := strconv.ParseInt(c.C.Args[1], 10, 64)
	if err != nil {
		return cmdResNil, errors.ErrIntegerOutOfRange
	}

	return doIncr(c, s, delta)
}

func doIncr(c *Cmd, s *dstore.Store, delta int64) (*CmdRes, error) {
	key := c.C.Args[0]
	obj := s.Get(key)
	if obj == nil {
		obj = s.NewObj(delta, -1, object.ObjTypeInt)
		s.Put(key, obj)
		return cmdResInt(delta), nil
	}

	switch obj.Type {
	case object.ObjTypeInt:
		break
	default:
		return cmdResNil, errors.ErrWrongTypeOperation
	}

	value, _ := obj.Value.(int64)

	value += delta
	obj.Value = value

	return cmdResInt(value), nil
}

func executeINCRBY(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) != 2 {
		return cmdResNil, errors.ErrWrongArgumentCount("INCRBY")
	}
	shard := sm.GetShardForKey(c.C.Args[0])
	return evalINCRBY(c, shard.Thread.Store())
}

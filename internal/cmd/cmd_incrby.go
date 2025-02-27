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
	HelpShort: "INCRBY decrements the value of the specified key in args by the specified decrement",
	Eval:      evalINCRBY,
	Execute:   executeINCRBY,
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
		return &CmdRes{R: &wire.Response{
			Value: &wire.Response_VInt{VInt: delta},
		}}, nil
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

	return &CmdRes{R: &wire.Response{
		Value: &wire.Response_VInt{VInt: value},
	}}, nil
}

func executeINCRBY(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	shard := sm.GetShardForKey(c.C.Args[0])
	return evalINCRBY(c, shard.Thread.Store())
}

// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/object"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dicedb-go/wire"
	"google.golang.org/protobuf/types/known/structpb"
)

var cHGETALL = &CommandMeta{
	Name:      "HGETALL",
	Syntax:    "HGETALL key",
	HelpShort: "HGET returns all the field and value for the key",
	HelpLong: `
HGET returns all the field and value for the key.

The command returns (nil) if the key does not exist.
	`,
	Examples: `
localhost:7379> HSET k1 f1 v1
OK OK
localhost:7379> HGETALL k1
OK f1
v1
localhost:7379> HGETALL k2
(nil)
	`,
	Eval:    evalHGETALL,
	Execute: executeHGETALL,
}

func init() {
	CommandRegistry.AddCommand(cHGETALL)
}

func evalHGETALL(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	key := c.C.Args[0]
	obj := s.Get(key)

	var hashMap HashMap

	if obj != nil {
		if err := object.AssertType(obj.Type, object.ObjTypeHashMap); err != nil {
			return cmdResNil, errors.ErrWrongTypeOperation
		}
		hashMap = obj.Value.(HashMap)
	}

	var vlist []*structpb.Value
	for key, val := range hashMap {
		field, err1 := structpb.NewValue(key)
		value, err2 := structpb.NewValue(val)
		if err1 != nil || err2 != nil {
			return nil, errors.ErrUnknownObjectType
		}
		vlist = append(vlist, field, value)
	}

	return &CmdRes{R: &wire.Response{
		VList: vlist,
	}}, nil
}

func executeHGETALL(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) != 1 {
		return cmdResNil, errors.ErrWrongArgumentCount("HGETALL")
	}
	shard := sm.GetShardForKey(c.C.Args[0])
	return evalHGETALL(c, shard.Thread.Store())
}

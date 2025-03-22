// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/object"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dicedb-go/wire"
)

var cHGETALL = &CommandMeta{
	Name:      "HGETALL",
	Syntax:    "HGETALL key",
	HelpShort: "HGETALL returns all the field-value pairs for the key from the string-string map",
	HelpLong: `
HGETALL returns all the field-value pairs for the key from the string-string map.

The command returns (nil) if the key does not exist or the map is empty.
	`,
	Examples: `
localhost:7379> HSET k1 f1 v1 f2 v2 f3 v3
OK 3
localhost:7379> HGETALL k1
OK 
f1=v1
f2=v2
f3=v3
localhost:7379> HGETALL k2
OK (nil)
	`,
	Eval:    evalHGETALL,
	Execute: executeHGETALL,
}

func init() {
	CommandRegistry.AddCommand(cHGETALL)
}

func evalHGETALL(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	key := c.C.Args[0]
	var m SSMap

	obj := s.Get(key)
	if obj != nil {
		if err := object.AssertType(obj.Type, object.ObjTypeSSMap); err != nil {
			return cmdResNil, errors.ErrWrongTypeOperation
		}
		m = obj.Value.(SSMap)
	}

	if len(m) == 0 {
		return cmdResNil, nil
	}

	return &CmdRes{R: &wire.Response{
		VSsMap: m,
	}}, nil
}

func executeHGETALL(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) != 1 {
		return cmdResNil, errors.ErrWrongArgumentCount("HGETALL")
	}
	shard := sm.GetShardForKey(c.C.Args[0])
	return evalHGETALL(c, shard.Thread.Store())
}

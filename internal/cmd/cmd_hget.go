// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dicedb-go/wire"
)

var cHGET = &CommandMeta{
	Name:      "HGET",
	Syntax:    "HGET key field",
	HelpShort: "HGET returns the value of field for the key",
	HelpLong: `
HGET returns the field value for the hash key in args.

The command returns (nil) if the key or field does not exist.
	`,
	Examples: `
localhost:7379> HSET k1 f1 v1
OK OK
localhost:7379> HGET k1 f1
OK v1
localhost:7379> HGET k2
(nil)
localhost:7379> HGET k1 f2
(nil)
	`,
	Eval:    evalHGET,
	Execute: executeHGET,
}

func init() {
	CommandRegistry.AddCommand(cHGET)
}

func evalHGET(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	key := c.C.Args[0]
	field := c.C.Args[1]
	obj := s.Get(key)

	if obj == nil {
		return cmdResNil, nil
	}

	hashMap, ok := obj.Value.(HashMap)

	if !ok {
		return cmdResNil, errors.ErrWrongTypeOperation
	}

	val, present := hashMap.Get(field)
	if !present {
		return cmdResNil, nil
	}

	return &CmdRes{R: &wire.Response{
		Value: &wire.Response_VStr{VStr: *val},
	}}, nil
}

func executeHGET(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) != 2 {
		return cmdResNil, errors.ErrWrongArgumentCount("HGET")
	}
	shard := sm.GetShardForKey(c.C.Args[0])
	return evalHGET(c, shard.Thread.Store())
}

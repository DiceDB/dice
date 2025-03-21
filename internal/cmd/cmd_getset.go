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

var cGETSET = &CommandMeta{
	Name:      "GETSET",
	Syntax:    "GETSET key value",
	HelpShort: "GETSET sets the value for the key and returns the old value",
	HelpLong: `
GETSET sets the value for the key and returns the old value.

The command returns (nil) if the key does not exist.
	`,
	Examples: `
localhost:7379> SET k1 v1
OK OK
localhost:7379> GETSET k1 v2
OK v1
localhost:7379> GET k1
OK v2
	`,
	Eval:    evalGETSET,
	Execute: executeGETSET,
}

func init() {
	CommandRegistry.AddCommand(cGETSET)
}

func evalGETSET(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) != 2 {
		return cmdResNil, errors.ErrWrongArgumentCount("GETSET")
	}
	key := c.C.Args[0]
	value := c.C.Args[1]
	obj := s.Get(key)
	intValue, err := strconv.ParseInt(value, 10, 64)

	if err == nil {
		s.Put(key, s.NewObj(intValue, -1, object.ObjTypeInt))
	} else {
		floatValue, err := strconv.ParseFloat(value, 64)
		if err == nil {
			s.Put(key, s.NewObj(floatValue, -1, object.ObjTypeFloat))
		} else {
			s.Put(key, s.NewObj(value, -1, object.ObjTypeString))
		}
	}
	if obj == nil {
		return cmdResNil, nil
	}
	return cmdResFromObject(obj)
}

func executeGETSET(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) != 2 {
		return cmdResNil, errors.ErrWrongArgumentCount("GETSET")
	}
	shard := sm.GetShardForKey(c.C.Args[0])
	return evalGETSET(c, shard.Thread.Store())
}

// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"

	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/object"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dicedb-go/wire"
)

var cGET = &CommandMeta{
	Name:      "GET",
	Syntax:    "GET key",
	HelpShort: "GET returns the value as a string for the key in args",
	HelpLong: `
GET returns the value as a string for the key in args.

The command returns an empty string if the key does not exist.
	`,
	Examples: `
localhost:7379> SET k1 v1
OK
localhost:7379> GET k1
OK "v1"
localhost:7379> GET k2
OK ""
	`,
	Eval:    evalGET,
	Execute: executeGET,
}

func init() {
	CommandRegistry.AddCommand(cGET)
}

func newGETRes(obj *object.Obj) *CmdRes {
	return &CmdRes{
		Rs: &wire.Result{
			Response: &wire.Result_GETRes{
				GETRes: &wire.GETRes{
					Value: getWireValueFromObj(obj),
				},
			},
		},
	}
}

var (
	GETResNilRes = newGETRes(nil)
)

func evalGET(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) != 1 {
		return GETResNilRes, errors.ErrWrongArgumentCount("GET")
	}

	key := c.C.Args[0]
	obj := s.Get(key)

	return newGETRes(obj), nil
}

func executeGET(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) != 1 {
		return GETResNilRes, errors.ErrWrongArgumentCount("GET")
	}
	shard := sm.GetShardForKey(c.C.Args[0])
	return evalGET(c, shard.Thread.Store())
}

func getWireValueFromObj(obj *object.Obj) string {
	if obj == nil {
		return ""
	}

	switch obj.Type {
	case object.ObjTypeInt:
		return fmt.Sprintf("%d", obj.Value.(int64))
	case object.ObjTypeString:
		return obj.Value.(string)
	case object.ObjTypeByteArray, object.ObjTypeHLL:
		return string(obj.Value.([]byte))
	default:
		panic("unknown object type " + obj.Type.String())
	}
}

// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"log/slog"

	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/object"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dicedb-go/wire"
)

var cGET = &CommandMeta{
	Name:      "GET",
	Syntax:    "GET key",
	HelpShort: "GET returns the value for the key",
	HelpLong: `
GET returns the value for the key in args.

The command returns (nil) if the key does not exist.
	`,
	Examples: `
localhost:7379> SET k1 v1
OK OK
localhost:7379> GET k1
OK v1
localhost:7379> GET k2
(nil)
	`,
	Eval:    evalGET,
	Execute: executeGET,
}

func init() {
	CommandRegistry.AddCommand(cGET)
}

var (
	GETResNilRes = &CmdRes{
		Rs: &wire.Result{
			Response: &wire.Result_GETRes{
				GETRes: &wire.GETRes{
					Value: getWireValueFromObj(nil),
				},
			},
		},
	}
)

func evalGET(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) != 1 {
		return GETResNilRes, errors.ErrWrongArgumentCount("GET")
	}

	key := c.C.Args[0]
	obj := s.Get(key)

	return createGETResFromObj(obj), nil
}

func executeGET(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) != 1 {
		return GETResNilRes, errors.ErrWrongArgumentCount("GET")
	}
	shard := sm.GetShardForKey(c.C.Args[0])
	return evalGET(c, shard.Thread.Store())
}

func createGETResFromObj(obj *object.Obj) *CmdRes {
	return &CmdRes{Rs: &wire.Result{
		Response: &wire.Result_GETRes{
			GETRes: &wire.GETRes{Value: getWireValueFromObj(obj)},
		},
	}}
}

func getWireValueFromObj(obj *object.Obj) *wire.Value {
	if obj == nil {
		return &wire.Value{Value: &wire.Value_Nil{Nil: true}}
	}

	switch obj.Type {
	case object.ObjTypeInt:
		return &wire.Value{Value: &wire.Value_I64{I64: obj.Value.(int64)}}
	case object.ObjTypeString:
		return &wire.Value{Value: &wire.Value_Str{Str: obj.Value.(string)}}
	case object.ObjTypeByteArray, object.ObjTypeHLL:
		return &wire.Value{Value: &wire.Value_Bytes{Bytes: obj.Value.([]byte)}}
	default:
		// This should never happen
		panic("unknown object type " + obj.Type.String())
	}
}

func cmdResFromObject(obj *object.Obj) (*CmdRes, error) {
	if obj == nil {
		return GETResNilRes, nil
	}

	switch obj.Type {
	case object.ObjTypeInt:
		return &CmdRes{R: &wire.Response{
			Value: &wire.Response_VInt{VInt: obj.Value.(int64)},
		}}, nil
	case object.ObjTypeString:
		return &CmdRes{R: &wire.Response{
			Value: &wire.Response_VStr{VStr: obj.Value.(string)},
		}}, nil
	case object.ObjTypeByteArray, object.ObjTypeHLL:
		return &CmdRes{R: &wire.Response{
			Value: &wire.Response_VBytes{VBytes: obj.Value.([]byte)},
		}}, nil
	default:
		slog.Error("unknown object type", "type", obj.Type)
		return GETResNilRes, errors.ErrUnknownObjectType
	}
}

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
	HelpShort: "HGETALL returns all the field-value pairs from the string-string map stored at key",
	HelpLong: `
HGETALL returns all the field-value pairs (we call it HElements) from the string-string map stored at key.

The command returns empty list if the key does not exist or the map is empty. Note that the order of the elements is not guaranteed.
	`,
	Examples: `
localhost:7379> HSET k1 f1 v1 f2 v2 f3 v3
OK 3
localhost:7379> HGETALL k1
OK
0) f1="v1"
1) f2="v2"
2) f3="v3"
localhost:7379> HGETALL k2
OK
	`,
	Eval:        evalHGETALL,
	Execute:     executeHGETALL,
	IsWatchable: true,
}

func init() {
	CommandRegistry.AddCommand(cHGETALL)
}

func newHGETALLRes(elements []*wire.HElement) *CmdRes {
	return &CmdRes{
		Rs: &wire.Result{
			Message: "OK",
			Status:  wire.Status_OK,
			Response: &wire.Result_HGETALLRes{
				HGETALLRes: &wire.HGETALLRes{
					Elements: elements,
				},
			},
		},
	}
}

var (
	HGETALLResNilRes = newHGETALLRes([]*wire.HElement{})
)

func evalHGETALL(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	key := c.C.Args[0]
	var m SSMap

	obj := s.Get(key)
	if obj != nil {
		if err := object.AssertType(obj.Type, object.ObjTypeSSMap); err != nil {
			return HGETALLResNilRes, errors.ErrWrongTypeOperation
		}
		m = obj.Value.(SSMap)
	}

	if len(m) == 0 {
		return HGETALLResNilRes, nil
	}

	elements := make([]*wire.HElement, 0, len(m))
	for k, v := range m {
		elements = append(elements, &wire.HElement{Key: k, Value: v})
	}

	return newHGETALLRes(elements), nil
}

func executeHGETALL(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) != 1 {
		return HGETALLResNilRes, errors.ErrWrongArgumentCount("HGETALL")
	}
	shard := sm.GetShardForKey(c.C.Args[0])
	return evalHGETALL(c, shard.Thread.Store())
}

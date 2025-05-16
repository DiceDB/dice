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
	HelpShort: "HGET returns the value of field present in the string-string map held at key.",
	HelpLong: `
HGET returns the value of field present in the string-string map held at key.

The command returns empty string "" if the key or field does not exist.
	`,
	Examples: `
localhost:7379> HSET k1 f1 v1
OK 1
localhost:7379> HGET k1 f1
OK "v1"
localhost:7379> HGET k2 f1
OK ""
localhost:7379> HGET k1 f2
OK ""
	`,
	Eval:        evalHGET,
	Execute:     executeHGET,
	IsWatchable: true,
}

func init() {
	CommandRegistry.AddCommand(cHGET)
}

func newHGETRes(value string) *CmdRes {
	return &CmdRes{
		Rs: &wire.Result{
			Message: "OK",
			Status:  wire.Status_OK,
			Response: &wire.Result_HGETRes{
				HGETRes: &wire.HGETRes{
					Value: value,
				},
			},
		},
	}
}

var (
	HGETResNilRes = newHGETRes("")
)

func evalHGET(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	key, field := c.C.Args[0], c.C.Args[1]

	obj := s.Get(key)
	if obj == nil {
		return HGETResNilRes, nil
	}

	m, ok := obj.Value.(SSMap)
	if !ok {
		return HGETResNilRes, errors.ErrWrongTypeOperation
	}

	val, ok := m.Get(field)
	if !ok {
		return HGETResNilRes, nil
	}

	return newHGETRes(val), nil
}

func executeHGET(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) != 2 {
		return HGETResNilRes, errors.ErrWrongArgumentCount("HGET")
	}
	shard := sm.GetShardForKey(c.C.Args[0])
	return evalHGET(c, shard.Thread.Store())
}

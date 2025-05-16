// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/object"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dice/internal/types"
	"github.com/dicedb/dicedb-go/wire"
)

var cZCARD *CommandMeta = &CommandMeta{
	Name:      "ZCARD",
	Syntax:    "ZCARD key",
	HelpShort: `ZCARD returns the cardinality of the sorted set stored at key.`,
	HelpLong: `
ZCARD returns the cardinality of the sorted set stored at key.

The command returns 0 if key does not exist.
	`,
	Examples: `
locahost:7379> ZADD users 1 alice 2 bob 3 charlie
OK 3
locahost:7379> ZCARD users
OK 3
localhost:7379> ZCARD nonexistent_key
OK 0
	`,
	Eval:        evalZCARD,
	Execute:     executeZCARD,
	IsWatchable: true,
}

func init() {
	CommandRegistry.AddCommand(cZCARD)
}

func newZCARDRes(count int64) *CmdRes {
	return &CmdRes{
		Rs: &wire.Result{
			Message: "OK",
			Status:  wire.Status_OK,
			Response: &wire.Result_ZCARDRes{
				ZCARDRes: &wire.ZCARDRes{
					Count: count,
				},
			},
		},
	}
}

var (
	ZCARDResNilRes = newZCARDRes(0)
)

func evalZCARD(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) != 1 {
		return ZCARDResNilRes, errors.ErrWrongArgumentCount("ZCARD")
	}

	key := c.C.Args[0]
	var ss *types.SortedSet

	obj := s.Get(key)
	if obj == nil {
		return ZCARDResNilRes, nil
	}

	if obj.Type != object.ObjTypeSortedSet {
		return ZCARDResNilRes, errors.ErrWrongTypeOperation
	}

	ss = obj.Value.(*types.SortedSet)
	return newZCARDRes(int64(ss.GetCount())), nil
}

func executeZCARD(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) != 1 {
		return ZCARDResNilRes, errors.ErrWrongArgumentCount("ZCARD")
	}
	shard := sm.GetShardForKey(c.C.Args[0])
	return evalZCARD(c, shard.Thread.Store())
}

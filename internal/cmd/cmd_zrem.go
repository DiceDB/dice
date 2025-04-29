// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/object"
	"github.com/dicedb/dice/internal/shardmanager"
	dsstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dice/internal/types"
	"github.com/dicedb/dicedb-go/wire"
)

var cZREM = &CommandMeta{
	Name:      "ZREM",
	Syntax:    "ZREM key member [member ...]",
	HelpShort: "Removes the specified members by key from the sorted set.",
	HelpLong: `
Removes the specified members by key from the sorted set. Non existing members are ignored.

Returns the number of members removed from the sorted set.
	`,
	Examples: `
localhost:7379> ZADD users 10 alice 20 bob 30 charlie
OK 3
localhost:7379> ZRANGE users 0 60 BYSCORE
OK
1) 10, alice
2) 20, bob
3) 30, charlie
localhost:7379> ZREM users alice bob
OK 2
localhost:7379> ZRANGE users 0 60 BYSCORE
OK
1) 30, charlie
`,
	Eval:    evalZREM,
	Execute: executeZREM,
}

func init() {
	CommandRegistry.AddCommand(cZREM)
}

func newZREMRes(count int64) *CmdRes {
	return &CmdRes{
		Rs: &wire.Result{
			Message:  "OK",
			Status:   wire.Status_OK,
			Response: &wire.Result_ZREMRes{ZREMRes: &wire.ZREMRes{Count: count}},
		},
	}
}

var (
	ZREMResNilRes = newZREMRes(0)
	ZREMRes0      = newZREMRes(0)
)

func executeZREM(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) < 2 {
		return ZREMResNilRes, errors.ErrWrongArgumentCount("ZREM")
	}

	shard := sm.GetShardForKey(c.C.Args[0])
	return evalZREM(c, shard.Thread.Store())
}

func evalZREM(c *Cmd, s *dsstore.Store) (*CmdRes, error) {
	if len(c.C.Args) < 2 {
		return ZREMResNilRes, errors.ErrWrongArgumentCount("ZREM")
	}
	key := c.C.Args[0]

	var ss *types.SortedSet

	obj := s.Get(key)
	if obj == nil {
		return ZREMResNilRes, nil
	}

	if obj.Type != object.ObjTypeSortedSet {
		return ZREMResNilRes, errors.ErrWrongTypeOperation
	}

	ss = obj.Value.(*types.SortedSet)

	countRem := int64(0)
	for i := 1; i < len(c.C.Args); i++ {
		n := ss.Remove(c.C.Args[i])
		if n != nil {
			countRem++
		}
	}

	return newZREMRes(countRem), nil
}

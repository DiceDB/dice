// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"math"
	"strconv"

	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/object"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dice/internal/types"
	"github.com/dicedb/dicedb-go/wire"
)

var cZCOUNT = &CommandMeta{
	Name:      "ZCOUNT",
	Syntax:    "ZCOUNT key min max",
	HelpShort: "ZCOUNT counts the number of members in a sorted set between min and max (both inclusive)",
	HelpLong: `
ZCOUNT counts the number of members in a sorted set between min and max (both inclusive)

If you want to use unbounded ranges, use -inf and +inf for min and max respectively.
The command returns the count of members in a sorted set between min and max (both inclusive). Returns 0 if the key does not exist.
`,
	Examples: `
localhost:7379> ZADD k 10 k1
OK 1
localhost:7379> ZADD k 20 k2
OK 1
localhost:7379> ZADD k 30 k3
OK 1
localhost:7379> ZCOUNT k 10 20
OK 2
localhost:7379> ZCOUNT k 10 30
OK 3
localhost:7379> ZCOUNT k 10 40
OK 3
localhost:7379> ZCOUNT k 1 2
OK 0
localhost:7379> ZCOUNT k -inf +inf
OK 3
localhost:7379> ZCOUNT k 10 10
OK 1
	`,
	Eval:        evalZCOUNT,
	Execute:     executeZCOUNT,
	IsWatchable: true,
}

func init() {
	CommandRegistry.AddCommand(cZCOUNT)
}

func newZCOUNTRes(count int64) *CmdRes {
	return &CmdRes{
		Rs: &wire.Result{
			Message: "OK",
			Status:  wire.Status_OK,
			Response: &wire.Result_ZCOUNTRes{
				ZCOUNTRes: &wire.ZCOUNTRes{
					Count: count,
				},
			},
		},
	}
}

var (
	ZCOUNTResNilRes = newZCOUNTRes(0)
	ZCOUNTRes0      = newZCOUNTRes(0)
)

func evalZCOUNT(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	var err error
	var minVal, maxVal int64

	if len(c.C.Args) != 3 {
		return ZCOUNTResNilRes, errors.ErrWrongArgumentCount("ZCOUNT")
	}

	key, minArg, maxArg := c.C.Args[0], c.C.Args[1], c.C.Args[2]

	if minArg == "-inf" {
		minVal = math.MinInt64
	} else if minArg == "+inf" {
		minVal = math.MaxInt64
	} else {
		minVal, err = strconv.ParseInt(minArg, 10, 64)
		if err != nil {
			return ZCOUNTResNilRes, errors.ErrInvalidNumberFormat
		}
	}

	if maxArg == "-inf" {
		maxVal = math.MinInt64
	} else if maxArg == "+inf" {
		maxVal = math.MaxInt64
	} else {
		maxVal, err = strconv.ParseInt(maxArg, 10, 64)
		if err != nil {
			return ZCOUNTResNilRes, errors.ErrInvalidNumberFormat
		}
	}

	if minVal > maxVal {
		return ZCOUNTRes0, nil
	}

	var ss *types.SortedSet

	obj := s.Get(key)
	if obj == nil {
		return ZCOUNTRes0, nil
	}

	if obj.Type != object.ObjTypeSortedSet {
		return ZCOUNTResNilRes, errors.ErrWrongTypeOperation
	}

	ss = obj.Value.(*types.SortedSet)

	count := ss.ZCOUNT(minVal, maxVal)
	return newZCOUNTRes(count), nil
}

func executeZCOUNT(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) != 3 {
		return ZCOUNTResNilRes, errors.ErrWrongArgumentCount("ZCOUNT")
	}

	shard := sm.GetShardForKey(c.C.Args[0])
	return evalZCOUNT(c, shard.Thread.Store())
}

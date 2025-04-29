// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"strconv"

	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/object"
	"github.com/dicedb/dice/internal/shardmanager"
	dsstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dice/internal/types"
	"github.com/dicedb/dicedb-go/wire"
)

var cZRANGE = &CommandMeta{
	Name:      "ZRANGE",
	Syntax:    "ZRANGE key start stop",
	HelpShort: "ZRANGE returns the range of elements from the sorted set stored at key.",
	HelpLong: `
ZRANGE returns the range of elements from the sorted set stored at key.

The elements are considered to be ordered from the lowest to the highest score. Both start and
stop are 0-based indexes, where 0 is the first element, 1 is the next element and so on.`,
	Examples: `
localhost:7379> ZADD s 1 a 2 b 3 c 4 d 5 e
OK 5
localhost:7379> ZRANGE s 1 3
OK
0) 1, a
1) 2, b
2) 3, c
`,
	Eval:    evalZRANGE,
	Execute: executeZRANGE,
}

func init() {
	CommandRegistry.AddCommand(cZRANGE)
}

func newZRANGERes(elements []*wire.ZElement) *CmdRes {
	return &CmdRes{
		Rs: &wire.Result{
			Message: "OK",
			Status:  wire.Status_OK,
			Response: &wire.Result_ZRANGERes{
				ZRANGERes: &wire.ZRANGERes{Elements: elements},
			},
		},
	}
}

var (
	ZRANGEResNilRes = newZRANGERes([]*wire.ZElement{})
)

func evalZRANGE(c *Cmd, s *dsstore.Store) (*CmdRes, error) {
	if len(c.C.Args) != 3 {
		return ZRANGEResNilRes, errors.ErrWrongArgumentCount("ZRANGE")
	}
	key := c.C.Args[0]
	startStr := c.C.Args[1]
	stopStr := c.C.Args[2]

	start, err := strconv.Atoi(startStr)
	if err != nil {
		return ZRANGEResNilRes, errors.ErrInvalidNumberFormat
	}

	stop, err := strconv.Atoi(stopStr)
	if err != nil {
		return ZRANGEResNilRes, errors.ErrInvalidNumberFormat
	}

	obj := s.Get(key)
	if obj == nil {
		return ZRANGEResNilRes, nil
	}

	if obj.Type != object.ObjTypeSortedSet {
		return ZRANGEResNilRes, errors.ErrWrongTypeOperation
	}

	ss := obj.Value.(*types.SortedSet)
	elements := ss.ZRANGE(start, stop)

	return newZRANGERes(elements), nil
}

func executeZRANGE(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) != 3 {
		return ZRANGEResNilRes, errors.ErrWrongArgumentCount("ZRANGE")
	}

	shard := sm.GetShardForKey(c.C.Args[0])
	return evalZRANGE(c, shard.Thread.Store())
}

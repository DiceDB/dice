// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"strconv"
	"strings"

	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/object"
	"github.com/dicedb/dice/internal/shardmanager"
	dsstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dice/internal/types"
	"github.com/dicedb/dicedb-go/wire"
)

var cZRANGE = &CommandMeta{
	Name:      "ZRANGE",
	Syntax:    "ZRANGE key start stop [BYSCORE | BYRANK]",
	HelpShort: "ZRANGE returns the range of elements from the sorted set stored at key.",
	HelpLong: `
ZRANGE returns the range of elements from the sorted set stored at key.

The default range is by rank "BYRANK" and this can be changed to "BYSCORE" if you want to range by score spanning the start and stop values.
The rank is 1-based, which means that the first element is at rank 1 and not rank 0.
The 1), 2), 3), ... is the rank of the element in the sorted set.

Both the start and stop values are inclusive and hence the elements having either of the values will be included. The
elements are considered to be ordered from the lowest to the highest. If you want reverse order, consider
storing score with flipped sign.`,
	Examples: `
localhost:7379> ZADD s 10 a 20 b 30 c 40 d 50 e
OK 5
localhost:7379> ZRANGE s 1 3
OK
1) 10, a
2) 20, b
3) 30, c
localhost:7379> ZRANGE s 1 4 BYRANK
OK
1) 10, a
2) 20, b
3) 30, c
4) 40, d
localhost:7379> ZRANGE s 1 3 BYSCORE
OK
localhost:7379> ZRANGE s 30 100 BYSCORE
OK
3) 30, c
4) 40, d
5) 50, e
`,
	Eval:        evalZRANGE,
	Execute:     executeZRANGE,
	IsWatchable: true,
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
	if len(c.C.Args) < 3 || len(c.C.Args) > 4 {
		return ZRANGEResNilRes, errors.ErrWrongArgumentCount("ZRANGE")
	}

	key := c.C.Args[0]
	startStr := c.C.Args[1]
	stopStr := c.C.Args[2]

	var byScore, byRank = false, true
	if len(c.C.Args) >= 4 {
		byScore = strings.EqualFold(c.C.Args[3], "BYSCORE")
		byRank = strings.EqualFold(c.C.Args[3], "BYRANK")
	}

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
	elements := ss.ZRANGE(start, stop, byScore, byRank)

	return newZRANGERes(elements), nil
}

func executeZRANGE(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) < 3 || len(c.C.Args) > 4 {
		return ZRANGEResNilRes, errors.ErrWrongArgumentCount("ZRANGE")
	}

	shard := sm.GetShardForKey(c.C.Args[0])
	return evalZRANGE(c, shard.Thread.Store())
}

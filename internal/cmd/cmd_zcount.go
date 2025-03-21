// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"strconv"

	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/eval/sortedset"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dicedb-go/wire"
)

var cZCOUNT = &CommandMeta{
	Name:      "ZCOUNT",
	Syntax:    "ZCOUNT key min max",
	HelpShort: "Counts the number of members in a sorted set between min and max (inclusive)",
	HelpLong: `
		Counts the number of members in a sorted set between min and max (inclusive)
		Use -inf and +inf for unbounded ranges
		Returns the count of members in a sorted set between min and max
		Returns 0 if the key does not exist
	`,
	Examples: `
		localhost:7379> ZCOUNT myzset 11 15
		OK 0

		localhost:7379> ZCOUNT myzset 5 10
		OK 3

		localhost:7379> ZCOUNT myzset 11
		ERR wrong number of arguments for 'ZCOUNT' command
	`,
	Eval:     evalZCOUNT,
	Execute: executeZCOUNT,
}

func init() {
	CommandRegistry.AddCommand(cZCOUNT)
}

func evalZCOUNT(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	// check number of arguments
	if len(c.C.Args) != 3 {
		return cmdResNil, errors.ErrWrongArgumentCount("ZCOUNT")
	}

	key := c.C.Args[0]
	minArg := c.C.Args[1]
	maxArg := c.C.Args[2]

	// parse minVal and maxVal scores
	minVal, errMin := strconv.ParseFloat(minArg, 64)
	maxVal, errMax := strconv.ParseFloat(maxArg, 64)
	if errMin != nil || errMax != nil {
		return cmdResNil, errors.ErrInvalidNumberFormat
	}

	// retrieve object from store
	obj := s.Get(key)
	if obj == nil {
		return cmdResInt0, nil
	}

	// ensure object is a valid sorted set
	var sortedSet *sortedset.Set
	sortedSet, err := sortedset.FromObject(obj)
	if err != nil {
		return cmdResNil, errors.ErrWrongTypeOperation
	}

	// get count of members within range from sorted set
	count := sortedSet.CountInRange(minVal, maxVal)

	return &CmdRes{R: &wire.Response{
		Value: &wire.Response_VInt{VInt: int64(count)},
	}}, nil
}

func executeZCOUNT(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) != 3 {
		return cmdResNil, errors.ErrWrongArgumentCount("ZCOUNT")
	}

	shard := sm.GetShardForKey(c.C.Args[0])
	return evalZCOUNT(c, shard.Thread.Store())
}


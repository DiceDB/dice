// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"strconv"

	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/object"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dice/internal/types"
	"github.com/dicedb/dicedb-go/wire"
)

var cZPOPMAX = &CommandMeta{
	Name:      "ZPOPMAX",
	Syntax:    "ZPOPMAX key [count]",
	HelpShort: "ZPOPMAX removes and returns the member with the highest score from the sorted set at the specified key.",
	HelpLong: `
ZPOPMAX removes and returns the member with the highest score from the sorted set at the specified key.

If the key does not exist, the command returns empty list. An optional "count" argument can be provided
to remove and return multiple members (up to the number specified).

When popped, the elements are returned in descending order of score and you get the rank of the element in the sorted set.
The rank is 1-based, which means that the first element is at rank 1 and not rank 0.
The 1), 2), 3), ... is the rank of the element in the sorted set.
	`,
	Examples: `
localhost:7379> ZADD users 10 alice 20 bob 30 charlie
OK 3
localhost:7379> ZPOPMAX users
OK
3) 30, charlie
localhost:7379> ZPOPMAX users 10
OK
2) 20, bob
1) 10, alice
	`,
	Eval:    evalZPOPMAX,
	Execute: executeZPOPMAX,
}

func init() {
	CommandRegistry.AddCommand(cZPOPMAX)
}

func newZPOPMAXRes(elements []*wire.ZElement) *CmdRes {
	return &CmdRes{
		Rs: &wire.Result{
			Message: "OK",
			Status:  wire.Status_OK,
			Response: &wire.Result_ZPOPMAXRes{
				ZPOPMAXRes: &wire.ZPOPMAXRes{
					Elements: elements,
				},
			},
		},
	}
}

var (
	ZPOPMAXResNilRes = newZPOPMAXRes([]*wire.ZElement{})
)

// evalZPOPMAX validates the arguments and executes the ZPOPMAX command logic.
// It returns the highest scoring members removed from the sorted set.
func evalZPOPMAX(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	// Validate that at least one argument (the key) is provided.
	if len(c.C.Args) < 1 {
		return ZPOPMAXResNilRes, errors.ErrWrongArgumentCount("ZPOPMAX")
	}
	key := c.C.Args[0]
	count := 1

	// If count is provided, convert it to an integer.
	if len(c.C.Args) > 1 {
		ops, err := strconv.Atoi(c.C.Args[1])
		if err != nil || ops <= 0 {
			return ZPOPMAXResNilRes, errors.ErrIntegerOutOfRange
		}
		count = ops
	}

	var ss *types.SortedSet

	obj := s.Get(key)
	if obj == nil {
		return ZPOPMAXResNilRes, nil
	}

	if obj.Type != object.ObjTypeSortedSet {
		return ZPOPMAXResNilRes, errors.ErrWrongTypeOperation
	}

	ss = obj.Value.(*types.SortedSet)
	elements := make([]*wire.ZElement, 0, count)
	totalElements := ss.SortedSet.GetCount()

	for i := 0; i < count; i++ {
		n := ss.PopMax()
		if n == nil {
			break
		}
		elements = append(elements, &wire.ZElement{
			Member: n.Key(),
			Score:  int64(n.Score()),
			Rank:   int64(totalElements) - int64(i),
		})
	}
	return newZPOPMAXRes(elements), nil
}

// executeZPOPMAX retrieves the appropriate shard for the key and evaluates the ZPOPMAX command.
// It returns the result of removing and returning the highest-scored elements.
func executeZPOPMAX(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	// Validate the existence of at least one argument (the key).
	if len(c.C.Args) < 1 {
		return ZPOPMAXResNilRes, errors.ErrWrongArgumentCount("ZPOPMAX")
	}
	// Determine the appropriate shard based on the key.
	shard := sm.GetShardForKey(c.C.Args[0])
	return evalZPOPMAX(c, shard.Thread.Store())
}

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

var cZPOPMIN = &CommandMeta{
	Name:      "ZPOPMIN",
	Syntax:    "ZPOPMIN key [count]",
	HelpShort: "ZPOPMIN removes and returns the member with the lowest score from the sorted set at the specified key.",
	HelpLong: `
ZPOPMIN removes and returns the member with the lowest score from the sorted set at the specified key.

If the key does not exist, the command returns empty list. An optional "count" argument can be provided
to remove and return multiple members (up to the number specified).

When popped, the elements are returned in ascending order of score and you get the rank of the element in the sorted set.
The rank is 1-based, which means that the first element is at rank 1 and not rank 0.
The 1), 2), 3), ... is the rank of the element in the sorted set.
	`,
	Examples: `
localhost:7379> ZADD users 10 alice 20 bob 30 charlie
OK 3
localhost:7379> ZPOPMIN users
OK
1) 10, alice
localhost:7379> ZPOPMIN users 10
OK
1) 20, bob
2) 30, charlie
	`,
	Eval:    evalZPOPMIN,
	Execute: executeZPOPMIN,
}

func init() {
	CommandRegistry.AddCommand(cZPOPMIN)
}

func newZPOPMINRes(elements []*wire.ZElement) *CmdRes {
	return &CmdRes{
		Rs: &wire.Result{
			Message: "OK",
			Status:  wire.Status_OK,
			Response: &wire.Result_ZPOPMINRes{
				ZPOPMINRes: &wire.ZPOPMINRes{
					Elements: elements,
				},
			},
		},
	}
}

var (
	ZPOPMINResNilRes = newZPOPMINRes([]*wire.ZElement{})
)

// evalZPOPMIN validates the arguments and executes the ZPOPMIN logic.
// It returns the lowest scoring member removed from the sorted set.
func evalZPOPMIN(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	// Validate that at least one argument is provided.
	if len(c.C.Args) < 1 {
		return ZPOPMINResNilRes, errors.ErrWrongArgumentCount("ZPOPMIN")
	}
	key := c.C.Args[0]
	count := 1

	if len(c.C.Args) > 1 {
		ops, err := strconv.Atoi(c.C.Args[1])
		if err != nil || ops <= 0 {
			return ZPOPMINResNilRes, errors.ErrIntegerOutOfRange
		}
		count = ops
	}

	var ss *types.SortedSet

	obj := s.Get(key)
	if obj == nil {
		return ZPOPMINResNilRes, nil
	}

	if obj.Type != object.ObjTypeSortedSet {
		return ZPOPMINResNilRes, errors.ErrWrongTypeOperation
	}

	ss = obj.Value.(*types.SortedSet)
	elements := make([]*wire.ZElement, 0, count)

	for i := 0; i < count; i++ {
		n := ss.PopMin()
		if n == nil {
			break
		}
		elements = append(elements, &wire.ZElement{
			Member: n.Key(),
			Score:  int64(n.Score()),
			Rank:   int64(i + 1),
		})
	}
	return newZPOPMINRes(elements), nil
}

func executeZPOPMIN(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	// Validate the existence atleast one argument.
	if len(c.C.Args) < 1 {
		return ZPOPMINResNilRes, errors.ErrWrongArgumentCount("ZPOPMIN")
	}
	// Determine the shard for the key.
	shard := sm.GetShardForKey(c.C.Args[0])
	return evalZPOPMIN(c, shard.Thread.Store())
}

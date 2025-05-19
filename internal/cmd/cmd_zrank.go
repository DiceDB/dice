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

var cZRANK = &CommandMeta{
	Name:      "ZRANK",
	Syntax:    "ZRANK key member",
	HelpShort: "ZRANK returns the rank of a member in a sorted set, ordered from low to high scores.",
	HelpLong: `
ZRANK returns the rank of a member in a sorted set, ordered from low to high scores.

The rank is 1-based which means that the member with the lowest score has rank 1, the next highest has rank 2, and so on.
The command returns the element - rank, score, and member.

Thus, 1), 2), 3) are the rank of the element in the sorted set, followed by the score and the member.

The the member passed as the second argument is not a member of the sorted set, the command returns a
valid response with a rank of 0 and score of 0. If the key does not exist, the command returns a
valid response with a rank of 0, score of 0, and the member as "".
	`,
	Examples: `
localhost:7379> ZADD users 10 alice 20 bob 30 charlie
OK 3
localhost:7379> ZRANK users bob
OK 2) 20, bob
localhost:7379> ZRANK users charlie
OK 3) 30, charlie
localhost:7379> ZRANK users daniel
OK 0) 0, daniel
	`,
	Eval:        evalZRANK,
	Execute:     executeZRANK,
	IsWatchable: true,
}

func init() {
	CommandRegistry.AddCommand(cZRANK)
}

func newZRANKRes(element *wire.ZElement) *CmdRes {
	return &CmdRes{
		Rs: &wire.Result{
			Message: "OK",
			Status:  wire.Status_OK,
			Response: &wire.Result_ZRANKRes{
				ZRANKRes: &wire.ZRANKRes{
					Element: element,
				},
			},
		},
	}
}

var (
	ZRANKResNilRes = newZRANKRes(nil)
)

// evalZRANK returns the rank of the member in the sorted set stored at key.
// The rank (or index) is 0-based, which means that the member with the lowest score has rank 0.
// If the 'WITHSCORE' option is specified, it returns both the rank and the score of the member.
// Returns nil if the key does not exist or the member is not a member of the sorted set.
func evalZRANK(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	key := c.C.Args[0]
	member := c.C.Args[1]

	obj := s.Get(key)
	if obj == nil {
		return ZRANKResNilRes, nil
	}

	if obj.Type != object.ObjTypeSortedSet {
		return ZRANGEResNilRes, errors.ErrWrongTypeOperation
	}

	ss := obj.Value.(*types.SortedSet)

	node := ss.GetByKey(member)
	rank := ss.FindRank(member)
	if node == nil || rank == 0 {
		return newZRANKRes(&wire.ZElement{
			Rank:   0,
			Score:  0,
			Member: member,
		}), nil
	}

	return newZRANKRes(&wire.ZElement{
		Rank:   int64(rank),
		Score:  int64(node.Score()),
		Member: node.Key(),
	}), nil
}

func executeZRANK(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) != 2 {
		return ZRANKResNilRes, errors.ErrWrongArgumentCount("ZRANK")
	}
	shard := sm.GetShardForKey(c.C.Args[0])
	return evalZRANK(c, shard.Thread.Store())
}

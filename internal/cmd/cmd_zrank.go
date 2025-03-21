// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"strconv"
	"strings"

	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/eval/sortedset"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dicedb-go/wire"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	WithScore string = "WITHSCORE"
)

var cZRANK = &CommandMeta{
	Name:      "ZRANK",
	Syntax:    "ZRANK key member [WITHSCORE]",
	HelpShort: "Retrieve the rank of `member1` in the sorted set `myzset`:",
	HelpLong:  ``, // TODO: Will do after code approval
	Examples:  ``,
	Eval:      evalZRANK,
	Execute:   executeZRANK,
}

func init() {
	CommandRegistry.AddCommand(cZRANK)
}

func evalZRANK(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	args := c.C.Args
	if len(args) < 2 || len(args) > 3 {
		return cmdResNil, errors.ErrWrongArgumentCount("ZRANK")
	}

	key := args[0]
	member := args[1]
	withScore := false

	if len(args) == 3 {
		if !strings.EqualFold(args[2], WithScore) {
			return cmdResNil, errors.ErrInvalidSyntax("ZRANK")
		}
		withScore = true
	}

	obj := s.Get(key)
	if obj == nil {
		return cmdResNil, nil
	}

	sortedSet, err := sortedset.FromObject(obj)
	if err != nil {
		return cmdResNil, errors.ErrWrongTypeOperation
	}

	rank, score := sortedSet.RankWithScore(member, false)
	if rank == -1 {
		return cmdResNil, nil
	}

	if withScore {
		scoreStr := strconv.FormatFloat(score, 'f', -1, 64)
		return &CmdRes{
			R: &wire.Response{VList: []*structpb.Value{
				structpb.NewNumberValue(float64(rank)),
				structpb.NewStringValue(scoreStr),
			}},
		}, nil
	}

	return &CmdRes{R: &wire.Response{
		Value: &wire.Response_VInt{VInt: rank},
	}}, nil
}

func executeZRANK(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) <= 1 {
		return cmdResNil, errors.ErrWrongArgumentCount("ZRANK")
	}
	shard := sm.GetShardForKey(c.C.Args[0])
	return evalZRANK(c, shard.Thread.Store())
}

// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/server/utils"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dicedb-go/wire"
)

var cTTL = &CommandMeta{
	Name:      "TTL",
	Syntax:    "TTL key",
	HelpShort: "TTL returns the remaining time to live in seconds",
	HelpLong: `
TTL returns the remaining time to live (in seconds) of a key that has an expiration set.

- Returns -1 if the key has no expiration.
- Returns -2 if the key does not exist.
	`,
	Examples: `
localhost:7379> SET k 43
OK OK
localhost:7379> TTL k
OK -1
localhost:7379> SET k 43 EX 10
OK OK
localhost:7379> TTL k
OK 9
localhost:7379> TTL kn
OK -2
	`,
	Eval:    evalTTL,
	Execute: executeTTL,
}

func init() {
	CommandRegistry.AddCommand(cTTL)
}

func evalTTL(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) != 1 {
		return cmdResNil, errors.ErrWrongArgumentCount("TTL")
	}

	var key = c.C.Args[0]

	obj := s.Get(key)
	if obj == nil {
		return &CmdRes{R: &wire.Response{
			Value: &wire.Response_VInt{VInt: -2},
		}}, nil
	}

	exp, isExpirySet := dstore.GetExpiry(obj, s)

	if !isExpirySet {
		return &CmdRes{R: &wire.Response{
			Value: &wire.Response_VInt{VInt: -1},
		}}, nil
	}

	durationMs := exp - uint64(utils.GetCurrentTime().UnixMilli())

	return &CmdRes{R: &wire.Response{
		Value: &wire.Response_VInt{VInt: int64(durationMs / 1000)},
	}}, nil
}

func executeTTL(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) != 1 {
		return cmdResNil, errors.ErrWrongArgumentCount("TTL")
	}
	shard := sm.GetShardForKey(c.C.Args[0])
	return evalTTL(c, shard.Thread.Store())
}

// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/shard"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dicedb-go/wire"
)

var cEXISTS = &CommandMeta{
	Name:      "EXISTS",
	Syntax:    "EXISTS key [key ...]",
	HelpShort: "EXISTS returns the count of keys that exist among the given",
	HelpLong:  `EXISTS command returns the count of keys that exist among the given arguments without modifying them.`,
	Examples: `
localhost:7379> SET k1 v1
OK
localhost:7379> SET k2 v2
OK
localhost:7379> EXISTS k1 k2 k3
OK 2
	`,
	Eval:    evalEXISTS,
	Execute: executeEXISTS,
}

func init() {
	CommandRegistry.AddCommand(cEXISTS)
}

// TODO: EXISTS command is actually a multi-key command so this needs
// to be scattered and gathered one step before this.

func evalEXISTS(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) < 1 {
		return cmdResNil, errors.ErrWrongArgumentCount("EXISTS")
	}

	var count int64
	for _, key := range c.C.Args {
		// GetNoTouch is used to check if a key exists in the store
		// without updating its last access time.
		if s.GetNoTouch(key) != nil {
			count++
		}
	}

	// Return the count as a response
	return &CmdRes{R: &wire.Response{
		Value: &wire.Response_VInt{VInt: count},
	}}, nil
}

func executeEXISTS(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) < 1 {
		return cmdResNil, errors.ErrWrongArgumentCount("EXISTS")
	}
	var count int64
	var shardMap = make(map[*shard.Shard][]string)
	for _, key := range c.C.Args {
		shard := sm.GetShardForKey(key)
		shardMap[shard] = append(shardMap[shard], key)
	}

	for shard, keys := range shardMap {
		c.C.Args = keys
		r, err := evalEXISTS(c, shard.Thread.Store())
		if err != nil {
			return nil, err
		}
		count += r.R.GetVInt()
	}
	return &CmdRes{R: &wire.Response{
		Value: &wire.Response_VInt{VInt: count},
	}}, nil
}

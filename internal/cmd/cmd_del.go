// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dicedb-go/wire"
)

var cDEL = &CommandMeta{
	Name:      "DEL",
	Syntax:    "DEL key [key ...]",
	HelpShort: "DEL deletes all the specified keys and returns the number of keys deleted on success.",
	HelpLong: `DEL deletes all the specified keys and returns the number of keys deleted on success.

If the key does not exist, it is ignored. The command returns the number of keys successfully deleted.`,
	Examples: `
localhost:7379> SET k1 v1
OK
localhost:7379> SET k2 v2
OK
localhost:7379> DEL k1 k2 k3
OK 2`,
	Eval:    evalDEL,
	Execute: executeDEL,
}

func init() {
	CommandRegistry.AddCommand(cDEL)
}

func newDELRes(count int64) *CmdRes {
	return &CmdRes{
		Rs: &wire.Result{
			Message: "OK",
			Status:  wire.Status_OK,
			Response: &wire.Result_DELRes{
				DELRes: &wire.DELRes{
					Count: count,
				},
			},
		},
	}
}

var (
	DELResNilRes = newDELRes(0)
)

func evalDEL(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) < 1 {
		return DELResNilRes, errors.ErrWrongArgumentCount("DEL")
	}

	var count int64
	for _, key := range c.C.Args {
		if ok := s.Del(key); ok {
			count++
		}
	}

	return newDELRes(count), nil
}

func executeDEL(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) < 1 {
		return DELResNilRes, errors.ErrWrongArgumentCount("DEL")
	}

	var count int64
	for _, key := range c.C.Args {
		shard := sm.GetShardForKey(key)
		r, err := evalDEL(c, shard.Thread.Store())
		if err != nil {
			return nil, err
		}
		count += r.Rs.GetDELRes().Count
	}
	return newDELRes(count), nil
}

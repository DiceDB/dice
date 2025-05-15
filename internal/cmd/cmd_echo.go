// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dicedb-go/wire"
)

var cECHO = &CommandMeta{
	Name:      "ECHO",
	Syntax:    "ECHO message",
	HelpShort: "ECHO returns the message passed to it",
	HelpLong:  `ECHO returns the message passed to it.`,
	Examples: `
localhost:7379> ECHO dicedb
OK dicedb`,
	Eval:    evalECHO,
	Execute: executeECHO,
}

func init() {
	CommandRegistry.AddCommand(cECHO)
}

func newECHORes(message string) *CmdRes {
	return &CmdRes{
		Rs: &wire.Result{
			Message: "OK",
			Status:  wire.Status_OK,
			Response: &wire.Result_ECHORes{
				ECHORes: &wire.ECHORes{
					Message: message,
				},
			},
		},
	}
}

func evalECHO(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) != 1 {
		return newECHORes(""), errors.ErrWrongArgumentCount("ECHO")
	}

	return newECHORes(c.C.Args[0]), nil
}

func executeECHO(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	shard := sm.GetShardForKey("-")
	return evalECHO(c, shard.Thread.Store())
}

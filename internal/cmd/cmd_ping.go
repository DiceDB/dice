// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dicedb-go/wire"
)

var cPING = &CommandMeta{
	Name:      "PING",
	Syntax:    "PING",
	HelpShort: "PING returns PONG if no argument is provided, otherwise it returns PONG with the message.",
	HelpLong: `
PING returns PONG if no argument is provided, otherwise it returns PONG with the message argument.
	`,
	Examples: `
localhost:7379> PING
OK "PONG"
localhost:7379> PING dicedb
OK "PONG dicedb"
	`,
	Eval:    evalPING,
	Execute: executePING,
}

func init() {
	CommandRegistry.AddCommand(cPING)
}

func newPINGRes(v string) *CmdRes {
	return &CmdRes{
		Rs: &wire.Result{
			Message: "OK",
			Status:  wire.Status_OK,
			Response: &wire.Result_PINGRes{
				PINGRes: &wire.PINGRes{
					Message: v,
				},
			},
		},
	}
}

var PINGResNilRes = newPINGRes("")

func evalPING(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) >= 2 {
		return PINGResNilRes, errors.ErrWrongArgumentCount("PING")
	}
	if len(c.C.Args) == 0 {
		return newPINGRes("PONG"), nil
	}
	return newPINGRes("PONG " + c.C.Args[0]), nil
}

func executePING(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	shard := sm.GetShardForKey("-")
	return evalPING(c, shard.Thread.Store())
}

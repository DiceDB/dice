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
	HelpShort: "PING returns PONG if no argument is provided, otherwise it returns the argument.",
	HelpLong: `
PING returns PONG if no argument is provided, otherwise it returns the argument.
	`,
	Examples: `
localhost:7379> PING
PONG
localhost:7379> PING Hello
Hello
	`,
	Eval:    evalPING,
	Execute: executePING,
}

func init() {
	CommandRegistry.AddCommand(cPING)
}

func evalPING(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) >= 2 {
		return cmdResNil, errors.ErrWrongArgumentCount("PING")
	}
	if len(c.C.Args) == 0 {
		return &CmdRes{R: &wire.Response{
			Value: &wire.Response_VStr{VStr: "PONG"},
		}}, nil
	}
	return &CmdRes{R: &wire.Response{
		Value: &wire.Response_VStr{VStr: c.C.Args[0]},
	}}, nil
}

func executePING(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	shard := sm.GetShardForKey("-")
	return evalPING(c, shard.Thread.Store())
}

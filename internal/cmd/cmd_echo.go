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
	HelpLong:  `ECHO command returns the message passed to it.`,
	Examples: `
	localhost:7379> ECHO hello!
OK hello!`,
	Eval:    evalECHO,
	Execute: executeECHO,
}

func init() {
	CommandRegistry.AddCommand(cECHO)
}

func evalECHO(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	message := strings.Join(c.C.Args, " ")
	return &CmdRes{R: &wire.Response{
		Value: &wire.Response_VStr{VStr: message},
	}}, nil
}

func executeECHO(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	shard := sm.GetShardForKey("-")
	return evalECHO(c, shard.Thread.Store())
}

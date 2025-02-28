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
	HelpShort: "ECHO returns the message passed to it",
	Eval:      evalECHO,
	Execute:   executeECHO,
}

func init() {
	CommandRegistry.AddCommand(cECHO)
}

func evalECHO(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	checkString := checkIsFullString(c.C.Args)
	if len(c.C.Args) != 1 && !checkString{
		return cmdResNil, errors.ErrWrongArgumentCount("ECHO")
	}

	return &CmdRes{R: &wire.Response{
		Value: &wire.Response_VStr{VStr: stripString(c.C.Args)},
	}}, nil
}

func executeECHO(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	shard := sm.GetShardForKey("-")
	return evalECHO(c, shard.Thread.Store())
}

// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
)

var cUNWATCH = &CommandMeta{
	Name:      "UNWATCH",
	HelpShort: "UNWATCH removes the previously created query subscription",
	Eval:      evalUNWATCH,
	Execute:   executeUNWATCH,
}

func init() {
	CommandRegistry.AddCommand(cUNWATCH)
}

func evalUNWATCH(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) != 1 {
		return cmdResNil, errWrongArgumentCount("UNWATCH")
	}

	return cmdResOK, nil
}

func executeUNWATCH(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	shard := sm.GetShardForKey("-")
	return evalUNWATCH(c, shard.Thread.Store())
}

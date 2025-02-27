// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
)

var cHANDSHAKE = &CommandMeta{
	Name:      "HANDSHAKE",
	HelpShort: "HANDSHAKE is used to handshake with the database; sends client_id and execution mode",
	Eval:      evalHANDSHAKE,
	Execute:   executeHANDSHAKE,
}

func init() {
	CommandRegistry.AddCommand(cHANDSHAKE)
}

func evalHANDSHAKE(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) != 2 {
		return cmdResNil, errors.ErrWrongArgumentCount("HANDSHAKE")
	}
	c.ClientID = c.C.Args[0]
	c.Mode = c.C.Args[1]
	return cmdResOK, nil
}

func executeHANDSHAKE(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	shard := sm.GetShardForKey("-")
	return evalHANDSHAKE(c, shard.Thread.Store())
}

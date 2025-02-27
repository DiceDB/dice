// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"strconv"

	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
)

var cEXPIREAT = &CommandMeta{
	Name:      "EXPIREAT",
	HelpShort: "EXPIREAT sets the expiration time of a key as an absolute Unix timestamp (in seconds)",
	Eval:      evalEXPIREAT,
	Execute:   executeEXPIREAT,
}

func init() {
	CommandRegistry.AddCommand(cEXPIREAT)
}

func evalEXPIREAT(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	// We need at least 2 arguments.
	if len(c.C.Args) < 2 {
		return cmdResNil, errors.ErrWrongArgumentCount("EXPIREAT")
	}

	var key = c.C.Args[0]
	exUnixTimeSec, err := strconv.ParseInt(c.C.Args[1], 10, 64)
	if err != nil || exUnixTimeSec < 0 {
		// TODO: Check for the upper bound of the input.
		return cmdResNil, errors.ErrInvalidExpireTime("EXPIREAT")
	}

	isExpirySet, err := dstore.EvaluateAndSetExpiry(c.C.Args[2:], exUnixTimeSec, key, s)
	if err != nil {
		return cmdResNil, err
	}

	if isExpirySet {
		return cmdResInt1, nil
	}

	return cmdResInt0, nil
}

func executeEXPIREAT(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) < 2 {
		return cmdResNil, errors.ErrWrongArgumentCount("EXPIREAT")
	}

	shard := sm.GetShardForKey(c.C.Args[0])
	return evalEXPIREAT(c, shard.Thread.Store())
}

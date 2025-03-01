// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"strconv"

	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/server/utils"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
)

var cEXPIRE = &CommandMeta{
	Name:      "EXPIRE",
	Syntax:    "EXPIRE key seconds [NX | XX]",
	HelpShort: "EXPIRE sets an expiry (in seconds) on a specified key",
	HelpLong: `
EXPIRE sets an expiry (in seconds) on a specified key. After the expiry time has elapsed, the key will be automatically deleted.

> If you want to delete the expirtation time on the key, you can use the PERSIST command.

The command returns 1 if the expiry was set, and 0 if the key already had an expiry set. The command supports the following options:

- NX: Set the expiration only if the key does not already have an expiration time.
- XX: Set the expiration only if the key already has an expiration time.
	`,
	Examples: `
locahost:7379> SET k1 v1
OK OK
locahost:7379> EXPIRE k1 10
OK 1
locahost:7379> SET k2 v2
OK OK
locahost:7379> EXPIRE k2 10 NX
OK 1
locahost:7379> EXPIRE k2 20 XX
OK 1
locahost:7379> EXPIRE k2 20 NX
OK 0
	`,
	Eval:    evalEXPIRE,
	Execute: executeEXPIRE,
}

func init() {
	CommandRegistry.AddCommand(cEXPIRE)
}

func evalEXPIRE(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) <= 1 {
		return cmdResNil, errors.ErrWrongArgumentCount("EXPIRE")
	}

	var key = c.C.Args[0]

	exDurationSec, err := strconv.ParseInt(c.C.Args[1], 10, 64)
	if err != nil || exDurationSec < 0 {
		return cmdResNil, errors.ErrInvalidExpireTime("EXPIRE")
	}

	obj := s.Get(key)
	if obj == nil {
		return cmdResInt0, nil
	}

	isExpirySet, err := dstore.EvaluateAndSetExpiry(c.C.Args[2:], utils.AddSecondsToUnixEpoch(exDurationSec), key, s)
	if err != nil {
		return cmdResNil, err
	}

	if isExpirySet {
		return cmdResInt1, nil
	}

	return cmdResInt0, nil
}

func executeEXPIRE(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) <= 1 {
		return cmdResNil, errors.ErrWrongArgumentCount("EXPIRE")
	}
	shard := sm.GetShardForKey(c.C.Args[0])
	return evalEXPIRE(c, shard.Thread.Store())
}

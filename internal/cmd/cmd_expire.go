// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"strconv"
	"time"

	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dicedb-go/wire"
)

var cEXPIRE = &CommandMeta{
	Name:      "EXPIRE",
	Syntax:    "EXPIRE key seconds [NX | XX]",
	HelpShort: "EXPIRE sets an expiry (in seconds) on a specified key",
	HelpLong: `
EXPIRE sets an expiry (in seconds) on a specified key. After the expiry time has elapsed, the key will be automatically deleted.

> If you want to delete the expiration time on the key, you can use the PERSIST command.

The command returns true if the expiry was set (changed), and false if the expiry could not be set (changed) due to key
not being present or due to the provided sub-command conditions not being met. The command
supports the following options:

- NX: Set the expiration only if the key does not already have an expiration time.
- XX: Set the expiration only if the key already has an expiration time.
	`,
	Examples: `
localhost:7379> SET k1 v1
OK
localhost:7379> EXPIRE k1 10
OK true
localhost:7379> SET k2 v2
OK
localhost:7379> EXPIRE k2 10 NX
OK true
localhost:7379> EXPIRE k2 20 XX
OK true
localhost:7379> EXPIRE k2 20 NX
OK false
	`,
	Eval:    evalEXPIRE,
	Execute: executeEXPIRE,
}

func init() {
	CommandRegistry.AddCommand(cEXPIRE)
}

func newEXPIRERes(isChanged bool) *CmdRes {
	return &CmdRes{
		Rs: &wire.Result{
			Message: "OK",
			Status:  wire.Status_OK,
			Response: &wire.Result_EXPIRERes{
				EXPIRERes: &wire.EXPIRERes{
					IsChanged: isChanged,
				},
			},
		},
	}
}

var (
	EXPIREResNilRes    = newEXPIRERes(false)
	EXPIREResSetRes    = newEXPIRERes(true)
	EXPIREResNotSetRes = newEXPIRERes(false)
)

func evalEXPIRE(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) <= 1 {
		return EXPIREResNilRes, errors.ErrWrongArgumentCount("EXPIRE")
	}

	var key = c.C.Args[0]

	exDurationSec, err := strconv.ParseInt(c.C.Args[1], 10, 64)
	if err != nil || exDurationSec < 0 {
		return EXPIREResNilRes, errors.ErrInvalidExpireTime("EXPIRE")
	}

	obj := s.Get(key)
	if obj == nil {
		return EXPIREResNotSetRes, nil
	}

	isExpirySet, err := dstore.EvaluateAndSetExpiry(c.C.Args[2:], time.Now().UnixMilli()+exDurationSec*1000, key, s)
	if err != nil {
		return EXPIREResNilRes, err
	}

	if isExpirySet {
		return EXPIREResSetRes, nil
	}

	return EXPIREResNotSetRes, nil
}

func executeEXPIRE(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) <= 1 {
		return EXPIREResNilRes, errors.ErrWrongArgumentCount("EXPIRE")
	}
	shard := sm.GetShardForKey(c.C.Args[0])
	return evalEXPIRE(c, shard.Thread.Store())
}

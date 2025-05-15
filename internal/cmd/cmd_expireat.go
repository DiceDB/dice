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

var cEXPIREAT = &CommandMeta{
	Name:      "EXPIREAT",
	Syntax:    "EXPIREAT key timestamp [NX | XX | GT | LT]",
	HelpShort: "EXPIREAT sets the expiration time of a key as an absolute Unix timestamp (in seconds)",
	HelpLong: `
EXPIREAT sets the expiration time of a key as an absolute Unix timestamp (in seconds).
After the expiry time has elapsed, the key will be automatically deleted.

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
localhost:7379> EXPIREAT k1 1740829942
OK true
localhost:7379> EXPIREAT k1 1740829942 NX
OK false
localhost:7379> EXPIREAT k1 1740829942 XX
OK false
	`,
	Eval:    evalEXPIREAT,
	Execute: executeEXPIREAT,
}

func init() {
	CommandRegistry.AddCommand(cEXPIREAT)
}

func newEXPIREATRes(isChanged bool) *CmdRes {
	return &CmdRes{
		Rs: &wire.Result{
			Message: "OK",
			Status:  wire.Status_OK,
			Response: &wire.Result_EXPIREATRes{
				EXPIREATRes: &wire.EXPIREATRes{
					IsChanged: isChanged,
				},
			},
		},
	}
}

var (
	EXPIREATResNilRes       = newEXPIREATRes(false)
	EXPIREATResChangedRes   = newEXPIREATRes(true)
	EXPIREATResUnchangedRes = newEXPIREATRes(false)
)

// EXPIREATMaxAbsTimestamp is the maximum allowed timestamp
// 10 years from the time server started
var EXPIREATMaxAbsTimestamp = time.Now().AddDate(10, 0, 0).Unix()

func evalEXPIREAT(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) < 2 {
		return EXPIREATResNilRes, errors.ErrWrongArgumentCount("EXPIREAT")
	}

	var key = c.C.Args[0]
	exUnixTimeSec, err := strconv.ParseInt(c.C.Args[1], 10, 64)
	if err != nil || exUnixTimeSec < 0 || exUnixTimeSec > EXPIREATMaxAbsTimestamp {
		return EXPIREATResNilRes, errors.ErrInvalidExpireTime("EXPIREAT")
	}

	isExpirySet, err := dstore.EvaluateAndSetExpiry(c.C.Args[2:], exUnixTimeSec*1000, key, s)
	if err != nil {
		return EXPIREATResNilRes, err
	}

	if isExpirySet {
		return EXPIREATResChangedRes, nil
	}

	return EXPIREATResUnchangedRes, nil
}

func executeEXPIREAT(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) < 2 {
		return EXPIREATResNilRes, errors.ErrWrongArgumentCount("EXPIREAT")
	}

	shard := sm.GetShardForKey(c.C.Args[0])
	return evalEXPIREAT(c, shard.Thread.Store())
}

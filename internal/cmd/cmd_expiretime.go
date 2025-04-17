// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dicedb-go/wire"
)

var cExpireTime *CommandMeta = &CommandMeta{
	Name:      "EXPIRETIME",
	Syntax:    "EXPIRETIME key",
	HelpShort: `EXPIRETIME returns the absolute Unix timestamp in seconds at which the given key will expire`,
	HelpLong: `
EXPIRETIME returns the absolute timestamp in Unix seconds at which the given key will expire.

The command returns -1 if the key exists but has no associated expiration time.
The command returns -2 if the key does not exist.
	`,
	Examples: `
localhost:7379> SET k1 v1
OK
localhost:7379> EXPIRE k1 10
OK true
localhost:7379> EXPIRETIME k1
OK 1744451192
	`,
	Eval:    evalEXPIRETIME,
	Execute: executeEXPIRETIME,
}

func init() {
	CommandRegistry.AddCommand(cExpireTime)
}

func newEXPIRETIMERes(expiryAt int64) *CmdRes {
	return &CmdRes{
		Rs: &wire.Result{
			Message: "OK",
			Status:  wire.Status_OK,
			Response: &wire.Result_EXPIRETIMERes{
				EXPIRETIMERes: &wire.EXPIRETIMERes{
					UnixSec: expiryAt,
				},
			},
		},
	}
}

var (
	EXPIRETIMEResNilRes = newEXPIRETIMERes(0)
	EXPIRETIMEResNegOne = newEXPIRETIMERes(-1)
	EXPIRETIMEResNegTwo = newEXPIRETIMERes(-2)
)

func evalEXPIRETIME(c *Cmd, dst *dstore.Store) (*CmdRes, error) {
	key := c.C.Args[0]
	obj := dst.Get(key)

	// returns -2 as the object is unavailable
	if obj == nil {
		return EXPIRETIMEResNegTwo, nil
	}

	expiry, ok := dstore.GetExpiry(obj, dst)
	if !ok {
		// returns -1 as the key doesn't have an expiration time set
		return EXPIRETIMEResNegOne, nil
	}

	// returns the absolute Unix timestamp (since January 1, 1970) in seconds at which the given key will expire
	return newEXPIRETIMERes(expiry / 1000), nil
}

func executeEXPIRETIME(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) != 1 {
		return EXPIRETIMEResNilRes, errors.ErrWrongArgumentCount("EXPIRETIME")
	}
	shard := sm.GetShardForKey(c.C.Args[0])
	return evalEXPIRETIME(c, shard.Thread.Store())
}

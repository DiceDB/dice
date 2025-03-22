// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/eval/sortedset"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dicedb-go/wire"
)

var cZCARD *CommandMeta = &CommandMeta{
	Name:      "ZCARD",
	Syntax:    "EXPIRETIME key",
	HelpShort: `EXPIRETIME returns the absolute Unix timestamp in seconds at which the given key will expire`,
	HelpLong: `
EXPIRETIME returns the absolute Unix timestamp in seconds at which the given key will expire.

The command returns -1 if the key exists but has no associated expiration time.
The command returns -2 if the key does not exist.
	`,
	Examples: `
locahost:7379> SET k1 v1
OK OK
locahost:7379> EXPIRE k1 10
OK 1
locahost:7379> EXPIRETIME k1
OK 1740829178
	`,
	Eval:    evalEXPIRETIME,
	Execute: executeEXPIRETIME,
}

func init() {
	CommandRegistry.AddCommand(cExpireTime)
}

// evalZCARD returns the cardinality (number of elements) of the sorted set stored at key.
// args should contain only 1 value, the key
// Returns expiration Unix timestamp in seconds.
// Returns -1 if the key exists but has no associated expiration time.
// Returns -2 if the key does not exist.
func evalZCARD(c *Cmd, dst *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) != 1{
		return cmdResNil,errors.ErrWrongArgumentCount("ZCARD")
	}

	key := c.C.Args[0]
	obj := dst.Get(key)
	if obj == nil {
		return cmdResInt0, nil
	}

	sortedSet,err := sortedset.FromObject(obj)

	if err != nil{
		return cmdResNil,errors.ErrWrongTypeOperation
	}

	return &CmdRes{R: &wire.Response{
		Value: &wire.Response_VInt{
			VInt: int64(sortedSet.Len()),
		},
	}}, nil

}

func executeZCARD(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) != 1 {
		return cmdResNil, errors.ErrWrongArgumentCount("EXPIRETIME")
	}
	shard := sm.GetShardForKey(c.C.Args[0])
	return evalZCARD(c, shard.Thread.Store())
}

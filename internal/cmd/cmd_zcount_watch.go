// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dicedb-go/wire"
)

var cZCOUNTWATCH = &CommandMeta{
	Name:      "ZCOUNT.WATCH",
	Syntax:    "ZCOUNT.WATCH key min max",
	HelpShort: "ZCOUNT.WATCH creates a query subscription over the ZCOUNT command",
	HelpLong: `
ZCOUNT.WATCH creates a query subscription over the ZCOUNT command. The client invoking the command
will receive the output of the ZCOUNT command (not just the notification) whenever the value against
the key is updated.

You can update the key in any other client. The ZCOUNT.WATCH client will receive the updated value.
	`,
	Examples: `
client1:7379> ZADD users 1 alice 2 bob 3 charlie
OK 3
client1:7379> ZCOUNT.WATCH users 1 5
entered the watch mode for ZCOUNT.WATCH users


client2:7379> ZADD users 4 daniel
OK 1


client1:7379> ...
entered the watch mode for ZCOUNT.WATCH users
OK [fingerprint=7042915837159566899] 4
	`,
	Eval:    evalZCOUNTWATCH,
	Execute: executeZCOUNTWATCH,
}

func init() {
	CommandRegistry.AddCommand(cZCOUNTWATCH)
}

func newZCOUNTWATCHRes() *CmdRes {
	return &CmdRes{
		Rs: &wire.Result{
			Message:  "OK",
			Status:   wire.Status_OK,
			Response: &wire.Result_ZCOUNTWATCHRes{},
		},
	}
}

var (
	ZCOUNTWATCHResNilRes = newZCOUNTWATCHRes()
)

func evalZCOUNTWATCH(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	r, err := evalZCOUNT(c, s)
	if err != nil {
		return nil, err
	}

	r.Rs.Fingerprint64 = c.Fingerprint()
	return r, nil
}

func executeZCOUNTWATCH(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) == 0 {
		return ZCOUNTWATCHResNilRes, errors.ErrWrongArgumentCount("ZCOUNT.WATCH")
	}
	shard := sm.GetShardForKey(c.C.Args[0])
	return evalZCOUNTWATCH(c, shard.Thread.Store())
}

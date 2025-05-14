// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dicedb-go/wire"
)

var cZCARDWATCH = &CommandMeta{
	Name:      "ZCARD.WATCH",
	Syntax:    "ZCARD.WATCH key",
	HelpShort: "ZCARD.WATCH creates a query subscription over the ZCARD command",
	HelpLong: `
ZCARD.WATCH creates a query subscription over the ZCARD command. The client invoking the command
will receive the output of the ZCARD command (not just the notification) whenever the value against
the key is updated.

You can update the key in any other client. The ZCARD.WATCH client will receive the updated value.
	`,
	Examples: `
client1:7379> ZADD users 1 alice 2 bob 3 charlie
OK 3
client1:7379> ZCARD.WATCH users
entered the watch mode for ZCARD.WATCH users


client2:7379> ZADD users 4 daniel
OK 1


client1:7379> ...
entered the watch mode for ZCARD.WATCH users
OK [fingerprint=8372868704969517043] 4
	`,
	Eval:    evalZCARDWATCH,
	Execute: executeZCARDWATCH,
}

func init() {
	CommandRegistry.AddCommand(cZCARDWATCH)
}

func newZCARDWATCHRes() *CmdRes {
	return &CmdRes{
		Rs: &wire.Result{
			Message:  "OK",
			Status:   wire.Status_OK,
			Response: &wire.Result_ZCARDWATCHRes{},
		},
	}
}

var (
	ZCARDWATCHResNilRes = newZCARDWATCHRes()
)

func evalZCARDWATCH(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	r, err := evalZCARD(c, s)
	if err != nil {
		return nil, err
	}

	r.Rs.Fingerprint64 = c.Fingerprint()
	return r, nil
}

func executeZCARDWATCH(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) == 0 {
		return ZCARDWATCHResNilRes, errors.ErrWrongArgumentCount("ZCARD.WATCH")
	}
	shard := sm.GetShardForKey(c.C.Args[0])
	return evalZCARDWATCH(c, shard.Thread.Store())
}

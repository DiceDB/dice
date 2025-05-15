// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dicedb-go/wire"
)

var cZRANGEWATCH = &CommandMeta{
	Name:      "ZRANGE.WATCH",
	Syntax:    "ZRANGE.WATCH key start stop [BYSCORE | BYRANK]",
	HelpShort: "ZRANGE.WATCH creates a query subscription over the ZRANGE command",
	HelpLong: `
ZRANGE.WATCH creates a query subscription over the ZRANGE command. The client invoking the command
will receive the output of the ZRANGE command (not just the notification) whenever the value against
the key is updated.

You can update the key in any other client. The ZRANGE.WATCH client will receive the updated value.
	`,
	Examples: `
client1:7379> ZADD users 10 alice 20 bob 30 charlie
OK 3
client1:7379> ZRANGE.WATCH users 1 5
entered the watch mode for ZRANGE.WATCH users


client2:7379> ZADD users 40 daniel
OK 1


client1:7379> ...
entered the watch mode for ZRANGE.WATCH users
OK [fingerprint=1007898011883907067]
1) 10, alice
2) 20, bob
3) 30, charlie
4) 40, daniel
	`,
	Eval:    evalZRANGEWATCH,
	Execute: executeZRANGEWATCH,
}

func init() {
	CommandRegistry.AddCommand(cZRANGEWATCH)
}

func newZRANGEWATCHRes() *CmdRes {
	return &CmdRes{
		Rs: &wire.Result{
			Message:  "OK",
			Status:   wire.Status_OK,
			Response: &wire.Result_ZRANGEWATCHRes{},
		},
	}
}

var (
	ZRANGEWATCHResNilRes = newZRANGEWATCHRes()
)

func evalZRANGEWATCH(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	r, err := evalZRANGE(c, s)
	if err != nil {
		return nil, err
	}

	r.Rs.Fingerprint64 = c.Fingerprint()
	return r, nil
}

func executeZRANGEWATCH(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) == 0 {
		return ZRANGEWATCHResNilRes, errors.ErrWrongArgumentCount("ZRANGE.WATCH")
	}
	shard := sm.GetShardForKey(c.C.Args[0])
	return evalZRANGEWATCH(c, shard.Thread.Store())
}

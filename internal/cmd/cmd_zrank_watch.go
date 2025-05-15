// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dicedb-go/wire"
)

var cZRANKWATCH = &CommandMeta{
	Name:      "ZRANK.WATCH",
	Syntax:    "ZRANK.WATCH key",
	HelpShort: "ZRANK.WATCH creates a query subscription over the ZRANK command",
	HelpLong: `
ZRANK.WATCH creates a query subscription over the ZRANK command. The client invoking the command
will receive the output of the ZRANK command (not just the notification) whenever the value against
the key is updated.

You can update the key in any other client. The ZRANK.WATCH client will receive the updated value.
	`,
	Examples: `
client1:7379> ZADD users 10 alice 20 bob 30 charlie
OK 3
client1:7379> ZRANK.WATCH users bob
entered the watch mode for ZRANK.WATCH users


client2:7379> ZADD users 10 bob
OK 0
client2:7379> ZADD users 100 alice
OK 0


client1:7379> ...
entered the watch mode for ZRANK.WATCH users
OK [fingerprint=3262833422269415227] 2) 10, bob
OK [fingerprint=3262833422269415227] 1) 10, bob
	`,
	Eval:    evalZRANKWATCH,
	Execute: executeZRANKWATCH,
}

func init() {
	CommandRegistry.AddCommand(cZRANKWATCH)
}

func newZRANKWATCHRes() *CmdRes {
	return &CmdRes{
		Rs: &wire.Result{
			Message:  "OK",
			Status:   wire.Status_OK,
			Response: &wire.Result_ZRANKWATCHRes{},
		},
	}
}

var (
	ZRANKWATCHResNilRes = newZRANKWATCHRes()
)

func evalZRANKWATCH(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	r, err := evalZRANK(c, s)
	if err != nil {
		return nil, err
	}

	r.Rs.Fingerprint64 = c.Fingerprint()
	return r, nil
}

func executeZRANKWATCH(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) == 0 {
		return ZRANKWATCHResNilRes, errors.ErrWrongArgumentCount("ZRANK.WATCH")
	}
	shard := sm.GetShardForKey(c.C.Args[0])
	return evalZRANKWATCH(c, shard.Thread.Store())
}

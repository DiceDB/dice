// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dicedb-go/wire"
)

var cHGETALLWATCH = &CommandMeta{
	Name:      "HGETALL.WATCH",
	Syntax:    "HGETALL.WATCH key",
	HelpShort: "HGETALL.WATCH creates a query subscription over the HGETALL command",
	HelpLong: `
HGETALL.WATCH creates a query subscription over the HGETALL command. The client invoking the command
will receive the output of the HGETALL command (not just the notification) whenever the value against
the key is updated.

You can update the key in any other client. The HGETALL.WATCH client will receive the updated value.
	`,
	Examples: `
client1:7379> HSET k f1 v1
OK 1
client1:7379> HGETALL.WATCH k
entered the watch mode for HGETALL.WATCH k


client2:7379> HSET k f2 v2
OK 1


client1:7379> ...
entered the watch mode for HGETALL.WATCH k
OK [fingerprint=4237011426]
0) f1="v1"
1) f2="v2"
	`,
	Eval:    evalHGETALLWATCH,
	Execute: executeHGETALLWATCH,
}

func init() {
	CommandRegistry.AddCommand(cHGETALLWATCH)
}

func newHGETALLWATCHRes() *CmdRes {
	return &CmdRes{
		Rs: &wire.Result{
			Message:  "OK",
			Status:   wire.Status_OK,
			Response: &wire.Result_HGETALLWATCHRes{},
		},
	}
}

var (
	HGETALLWATCHResNilRes = newHGETALLWATCHRes()
)

func evalHGETALLWATCH(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	r, err := evalHGETALL(c, s)
	if err != nil {
		return nil, err
	}

	r.Rs.Fingerprint64 = c.Fingerprint()
	return r, nil
}

func executeHGETALLWATCH(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) == 0 {
		return HGETALLWATCHResNilRes, errors.ErrWrongArgumentCount("HGETALL.WATCH")
	}
	shard := sm.GetShardForKey(c.C.Args[0])
	return evalHGETALLWATCH(c, shard.Thread.Store())
}

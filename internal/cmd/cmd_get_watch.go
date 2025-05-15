// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dicedb-go/wire"
)

var cGETWATCH = &CommandMeta{
	Name:      "GET.WATCH",
	Syntax:    "GET.WATCH key",
	HelpShort: "GET.WATCH creates a query subscription over the GET command",
	HelpLong: `
GET.WATCH creates a query subscription over the GET command. The client invoking the command
will receive the output of the GET command (not just the notification) whenever the value against
the key is updated.

You can update the key in any other client. The GET.WATCH client will receive the updated value.
	`,
	Examples: `
client1:7379> SET k1 v1
OK
client1:7379> GET.WATCH k1
entered the watch mode for GET.WATCH k1


client2:7379> SET k1 v2
OK


client1:7379> ...
entered the watch mode for GET.WATCH k1
OK [fingerprint=2356444921] "v2"
	`,
	Eval:    evalGETWATCH,
	Execute: executeGETWATCH,
}

func init() {
	CommandRegistry.AddCommand(cGETWATCH)
}

func newGETWATCHRes() *CmdRes {
	return &CmdRes{
		Rs: &wire.Result{
			Message: "OK",
			Status:  wire.Status_OK,
			Response: &wire.Result_GETWATCHRes{
				GETWATCHRes: &wire.GETWATCHRes{},
			},
		},
	}
}

var (
	GETWATCHResNilRes = newGETWATCHRes()
)

func evalGETWATCH(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	r, err := evalGET(c, s)
	if err != nil {
		return GETWATCHResNilRes, err
	}

	r.Rs.Fingerprint64 = c.Fingerprint()
	return r, nil
}

func executeGETWATCH(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) == 0 {
		return GETWATCHResNilRes, errors.ErrWrongArgumentCount("GET.WATCH")
	}
	shard := sm.GetShardForKey(c.C.Args[0])
	return evalGETWATCH(c, shard.Thread.Store())
}

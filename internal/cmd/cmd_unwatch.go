// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dicedb-go/wire"
)

var cUNWATCH = &CommandMeta{
	Name:      "UNWATCH",
	Syntax:    "UNWATCH <fingerprint>",
	HelpShort: "UNWATCH removes the previously created query subscription",
	HelpLong: `
WATCH command creates a subscription to the key and returns a fingerprint. Use this fingerprint to UNWATCH the subscription.
After running the UNWATCH command, the subscription will be removed and the data changes will no longer be sent to the client.

Note: If you are using the DiceDB CLI, then you need not run this command given the REPL will implicitly run this command when
you exit the watch mode.
	`,
	Examples: `
localhost:7379> UNWATCH 2356444921
OK
	`,
	Eval:    evalUNWATCH,
	Execute: executeUNWATCH,
}

func init() {
	CommandRegistry.AddCommand(cUNWATCH)
}

func newUNWATCHRes() *CmdRes {
	return &CmdRes{
		Rs: &wire.Result{
			Message:  "OK",
			Status:   wire.Status_OK,
			Response: &wire.Result_UNWATCHRes{},
		},
	}
}

var (
	UNWATCHResNilRes = newUNWATCHRes()
	UNWATCHResOKRes  = newUNWATCHRes()
)

// Note: We do not do anything here, because the UNWATCH command
// is handled by the iothread.
func evalUNWATCH(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) != 1 {
		return UNWATCHResNilRes, errors.ErrWrongArgumentCount("UNWATCH")
	}

	return UNWATCHResOKRes, nil
}

func executeUNWATCH(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) != 1 {
		return UNWATCHResNilRes, errors.ErrWrongArgumentCount("UNWATCH")
	}
	shard := sm.GetShardForKey("-")
	return evalUNWATCH(c, shard.Thread.Store())
}

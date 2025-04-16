// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/shardmanager"
	"github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dicedb-go/wire"
)

var cFLUSHDB = &CommandMeta{
	Name:      "FLUSHDB",
	Syntax:    "FLUSHDB",
	HelpShort: "FLUSHDB deletes all keys.",
	HelpLong: `
FLUSHDB deletes all keys present in the database.
	`,
	Examples: `
localhost:7379> SET k1 v1
OK
localhost:7379> SET k2 v2
OK
localhost:7379> FLUSHDB
OK
localhost:7379> GET k1
OK ""
localhost:7379> GET k2
OK ""
	`,
	Eval:    evalFLUSHDB,
	Execute: executeFLUSHDB,
}

func init() {
	CommandRegistry.AddCommand(cFLUSHDB)
}

func newFLUSHDBRes() *CmdRes {
	return &CmdRes{
		Rs: &wire.Result{
			Message: "OK",
			Status:  wire.Status_OK,
			Response: &wire.Result_FLUSHDBRes{
				FLUSHDBRes: &wire.FLUSHDBRes{},
			},
		},
	}
}

var (
	FLUSHDBResOKRes  = newFLUSHDBRes()
	FLUSHDBResNilRes = newFLUSHDBRes()
)

func evalFLUSHDB(c *Cmd, s *store.Store) (*CmdRes, error) {
	if len(c.C.Args) != 0 {
		return FLUSHDBResNilRes, errors.ErrWrongArgumentCount("FLUSHDB")
	}

	store.Reset(s)
	return FLUSHDBResOKRes, nil
}

func executeFLUSHDB(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	for _, shard := range sm.Shards() {
		_, err := evalFLUSHDB(c, shard.Thread.Store())
		if err != nil {
			return nil, err
		}
	}
	return FLUSHDBResOKRes, nil
}

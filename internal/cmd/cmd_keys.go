// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dicedb-go/wire"
)

var cKEYS = &CommandMeta{
	Name:      "KEYS",
	Syntax:    "KEYS pattern",
	HelpShort: "KEYS returns all keys matching the pattern",
	HelpLong: `
KEYS returns all keys matching the pattern.

The pattern can contain the following special characters to match multiple keys.
Supports glob-style patterns:
- *: matches any sequence of characters
- ?: matches any single character`,
	Examples: `
localhost:7379> SET k1 v1
OK
localhost:7379> SET k2 v2
OK
localhost:7379> SET k33 v33
OK
localhost:7379> KEYS k?
OK
0) k1
1) k2
localhost:7379> KEYS k*
OK
0) k1
1) k2
2) k33
localhost:7379> KEYS *
OK
0) k1
1) k2
2) k33
	`,
	Eval:    evalKEYS,
	Execute: executeKEYS,
}

func init() {
	CommandRegistry.AddCommand(cKEYS)
}

func newKEYSRes(keys []string) *CmdRes {
	return &CmdRes{
		Rs: &wire.Result{
			Message: "OK",
			Status:  wire.Status_OK,
			Response: &wire.Result_KEYSRes{
				KEYSRes: &wire.KEYSRes{
					Keys: keys,
				},
			},
		},
	}
}

var (
	KEYSResNilRes = newKEYSRes([]string{})
)

func evalKEYS(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) != 1 {
		return KEYSResNilRes, errors.ErrWrongArgumentCount("KEYS")
	}
	pattern := c.C.Args[0]
	keys, err := s.Keys(pattern)
	if err != nil {
		return KEYSResNilRes, err
	}
	return newKEYSRes(keys), nil
}

func executeKEYS(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) != 1 {
		return KEYSResNilRes, errors.ErrWrongArgumentCount("KEYS")
	}
	var keys []string
	for _, shard := range sm.Shards() {
		res, err := evalKEYS(c, shard.Thread.Store())
		if err != nil {
			return KEYSResNilRes, err
		}
		keys = append(keys, res.Rs.GetKEYSRes().Keys...)
	}
	return newKEYSRes(keys), nil
}

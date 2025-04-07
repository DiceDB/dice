// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dicedb-go/wire"
	"google.golang.org/protobuf/types/known/structpb"
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
OK OK
localhost:7379> SET k2 v2
OK OK
localhost:7379> SET k33 v33
OK OK
localhost:7379> KEYS k?
OK
1) "k1"
2) "k2"
localhost:7379> KEYS k*
OK
1) "k1"
2) "k2"
3) "k33"
localhost:7379> KEYS *
OK
1) "k1"
2) "k2"
3) "k33"
	`,
	Eval:    evalKEYS,
	Execute: executeKEYS,
}

func init() {
	CommandRegistry.AddCommand(cKEYS)
}

func evalKEYS(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) != 1 {
		return cmdResNil, errors.ErrWrongArgumentCount("KEYS")
	}
	pattern := c.C.Args[0]
	keys, err := s.Keys(pattern)
	if err != nil {
		return nil, err
	}
	return createResponseFromArray(keys), nil
}

func executeKEYS(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) != 1 {
		return cmdResNil, errors.ErrWrongArgumentCount("KEYS")
	}
	var keys []string
	for _, shard := range sm.Shards() {
		res, err := evalKEYS(c, shard.Thread.Store())
		if err != nil {
			return nil, err
		}
		for _, v := range res.R.GetVList() {
			keys = append(keys, v.GetStringValue())
		}
	}
	finalRes := createResponseFromArray(keys)
	return finalRes, nil
}

func createResponseFromArray(arr []string) *CmdRes {
	if len(arr) == 0 {
		return cmdResNil
	}
	var res []*structpb.Value
	for _, v := range arr {
		val := structpb.NewStringValue(v)
		res = append(res, val)
	}
	return &CmdRes{R: &wire.Response{
		VList: res}}
}

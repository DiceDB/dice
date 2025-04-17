// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dicedb-go/wire"
)

var cType = &CommandMeta{
	Name:      "TYPE",
	Syntax:    "TYPE key",
	HelpShort: "TYPE returns the type of the value stored at a specified key",
	HelpLong: `
TYPE returns the type of the value stored at a specified key. The type can be one of the following:

 - string
 - int
 
 Returns "none" if the key does not exist.
	`,
	Examples: `
localhost:7379> SET k 43
OK
localhost:7379> TYPE k
OK int
localhost:7379> TYPE kn
OK none
	`,
	Eval:    evalTYPE,
	Execute: executeTYPE,
}

func init() {
	CommandRegistry.AddCommand(cType)
}

func newTYPERes(t string) *CmdRes {
	return &CmdRes{
		Rs: &wire.Result{
			Message: "OK",
			Status:  wire.Status_OK,
			Response: &wire.Result_TYPERes{
				TYPERes: &wire.TYPERes{
					Type: t,
				},
			},
		},
	}
}

var (
	TYPEResNilRes = newTYPERes("none")
)

func evalTYPE(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) != 1 {
		return TYPEResNilRes, errors.ErrWrongArgumentCount("TYPE")
	}

	key := c.C.Args[0]

	obj := s.Get(key)
	if obj == nil {
		return TYPEResNilRes, nil
	}

	return newTYPERes(obj.Type.String()), nil
}

func executeTYPE(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) == 0 {
		return TYPEResNilRes, errors.ErrWrongArgumentCount("TYPE")
	}

	shard := sm.GetShardForKey(c.C.Args[0])
	return evalTYPE(c, shard.Thread.Store())
}

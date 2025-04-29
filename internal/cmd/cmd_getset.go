// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/object"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dicedb-go/wire"
)

var cGETSET = &CommandMeta{
	Name:      "GETSET",
	Syntax:    "GETSET key value",
	HelpShort: "GETSET sets the value for the key and returns the old value",
	HelpLong: `
GETSET sets the value for the key and returns the old value.

The command returns "" if the key does not exist.
	`,
	Examples: `
localhost:7379> SET k1 v1
OK
localhost:7379> GETSET k1 v2
OK "v1"
localhost:7379> GET k1
OK "v2"
	`,
	Eval:    evalGETSET,
	Execute: executeGETSET,
}

func init() {
	CommandRegistry.AddCommand(cGETSET)
}

func newGETSETRes(obj *object.Obj) *CmdRes {
	value, err := getWireValueFromObj(obj)
	if err != nil {
		return &CmdRes{
			Rs: &wire.Result{
				Message: err.Error(),
				Status:  wire.Status_ERR,
			},
		}
	}
	return &CmdRes{
		Rs: &wire.Result{
			Message: "OK",
			Status:  wire.Status_OK,
			Response: &wire.Result_GETSETRes{
				GETSETRes: &wire.GETSETRes{
					Value: value,
				},
			},
		},
	}
}

var (
	GETSETResNilRes = newGETSETRes(nil)
)

func evalGETSET(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) != 2 {
		return GETSETResNilRes, errors.ErrWrongArgumentCount("GETSET")
	}
	key, value := c.C.Args[0], c.C.Args[1]
	obj := s.Get(key)

	// Put the new value in the store
	s.Put(key, CreateObjectFromValue(s, value, -1))

	// Return the old value, if the key does not exist, return nil
	if obj == nil {
		return GETSETResNilRes, nil
	}

	// Return the old value
	return newGETSETRes(obj), nil
}

func executeGETSET(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) != 2 {
		return GETSETResNilRes, errors.ErrWrongArgumentCount("GETSET")
	}
	shard := sm.GetShardForKey(c.C.Args[0])
	return evalGETSET(c, shard.Thread.Store())
}

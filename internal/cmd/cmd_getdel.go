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

var cGETDEL = &CommandMeta{
	Name:      "GETDEL",
	Syntax:    "GETDEL key",
	HelpShort: "GETDEL returns the value of the key and then deletes the key.",
	HelpLong: `
GETDEL returns the value of the key and then deletes the key.

The command returns (nil) if the key does not exist.
	`,
	Examples: `
localhost:7379> SET k v
OK 
localhost:7379> GETDEL k
OK "v"
localhost:7379> GET k
OK ""
	`,
	Eval:    evalGETDEL,
	Execute: executeGETDEL,
}

func init() {
	CommandRegistry.AddCommand(cGETDEL)
}

func newGETDELRes(obj *object.Obj) *CmdRes {
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
			Response: &wire.Result_GETDELRes{
				GETDELRes: &wire.GETDELRes{
					Value: value,
				},
			},
		},
	}
}

var (
	GETDELResNilRes = newGETDELRes(nil)
)

func evalGETDEL(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) != 1 {
		return GETDELResNilRes, errors.ErrWrongArgumentCount("GETDEL")
	}

	key := c.C.Args[0]

	// Getting the key based on previous touch value
	obj := s.GetNoTouch(key)
	if obj == nil {
		return GETDELResNilRes, nil
	}

	// Get the key from the hash table
	// TODO: Evaluate the need for having GetDel
	// implemented in the store. It might be better if we can
	// keep the business logic untangled from the store.
	objVal := s.GetDel(key)
	return newGETDELRes(objVal), nil
}

func executeGETDEL(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) != 1 {
		return GETDELResNilRes, errors.ErrWrongArgumentCount("GETDEL")
	}

	shard := sm.GetShardForKey(c.C.Args[0])
	return evalGETDEL(c, shard.Thread.Store())
}

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
OK OK
localhost:7379> GETDEL k
OK v
localhost:7379> GET k
(nil)
	`,
	Eval:    evalGETDEL,
	Execute: executeGETDEL,
}

func init() {
	CommandRegistry.AddCommand(cGETDEL)
}

// GETDEL returns the value of the key and then deletes the key.
//
// The function expects exactly one argument: the key to get.
// If the key exists, it will be deleted before its value is returned.
// evalGETDEL returns cmdResNil if key is expired or it does not exist
//
// Parameters:
//   - c *Cmd: The command context containing the arguments
//   - s *dstore.Store: The data store instance
//
// Returns:
//   - *CmdRes: Response containing the value for the queried key.
//   - error: Error if wrong number of arguments or wrong value type.
func evalGETDEL(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) != 1 {
		return cmdResNil, errors.ErrWrongArgumentCount("GETDEL")
	}

	key := c.C.Args[0]

	// Getting the key based on previous touch value
	obj := s.GetNoTouch(key)
	if obj == nil {
		return cmdResNil, nil
	}

	// Get the key from the hash table
	// TODO: Evaluate the need for having GetDel
	// implemented in the store. It might be better if we can
	// keep the business logic untangled from the store.
	objVal := s.GetDel(key)

	// Decode and return the value based on its encoding
	switch oType := objVal.Type; oType {
	case object.ObjTypeInt:
		// Value is stored as an int64, so use type assertion
		return &CmdRes{R: &wire.Response{
			Value: &wire.Response_VInt{VInt: objVal.Value.(int64)},
		}}, nil
	case object.ObjTypeString:
		// Value is stored as a string, use type assertion
		return &CmdRes{R: &wire.Response{
			Value: &wire.Response_VStr{VStr: objVal.Value.(string)},
		}}, nil
	default:
		return cmdResNil, errors.ErrWrongTypeOperation
	}
}

func executeGETDEL(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) != 1 {
		return cmdResNil, errors.ErrWrongArgumentCount("GETDEL")
	}

	shard := sm.GetShardForKey(c.C.Args[0])
	return evalGETDEL(c, shard.Thread.Store())
}

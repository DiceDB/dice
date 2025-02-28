// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dicedb-go/wire"
)

var cDEL = &CommandMeta{
	Name:      "DEL",
	HelpShort: "DEL deletes all the specified keys",
	Documentation: `
Deletes all the specified keys and returns the number of keys deleted on success.

## Syntax

` + "```" + `
DEL key [key ...]
` + "```" + `

## Examples

` + "```" + `
localhost:7379> SET k1 v1
OK
localhost:7379> SET k2 v2
OK
localhost:7379> DEL k1 k2 k3
(integer) 2
` + "```" + `
	`,
	Eval:    evalDEL,
	Execute: executeDEL,
}

func init() {
	CommandRegistry.AddCommand(cDEL)
}

// TODO: DEL command is actually a multi-key command so this needs
// to be scattered and gathered one step before this.

// evalDEL deletes all the specified keys in args list.
//
// Parameters:
//   - c *Cmd: The command context containing the arguments
//   - s *dstore.Store: The data store instance
//
// Returns:
//   - *CmdRes: Response containing the count of total deleted keys
//   - error: Error if wrong number of arguments or wrong value type
func evalDEL(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) < 1 {
		return cmdResNil, errors.ErrWrongArgumentCount("DEL")
	}

	var count int
	for _, key := range c.C.Args {
		if ok := s.Del(key); ok {
			count++
		}
	}

	return &CmdRes{R: &wire.Response{
		Value: &wire.Response_VInt{VInt: int64(count)},
	}}, nil
}

func executeDEL(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) < 1 {
		return cmdResNil, errors.ErrWrongArgumentCount("DEL")
	}

	var count int64
	for _, key := range c.C.Args {
		shard := sm.GetShardForKey(key)
		r, err := evalDEL(c, shard.Thread.Store())
		if err != nil {
			return nil, err
		}
		count += r.R.GetVInt()
	}
	return &CmdRes{R: &wire.Response{
		Value: &wire.Response_VInt{VInt: count},
	}}, nil
}

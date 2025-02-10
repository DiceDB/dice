// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dicedb-go/wire"
)

var cEXISTS = &DiceDBCommand{
	Name:      "EXISTS",
	HelpShort: "Returns the count of keys that exist among the given arguments without modifying them",
	Eval:      evalEXISTS,
}

func init() {
	commandRegistry.AddCommand(cEXISTS)
}

// TODO: EXISTS command is actually a multi-key command so this needs
// to be scattered and gathered one step before this.

func evalEXISTS(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) < 1 {
		return cmdResNil, errWrongArgumentCount("EXISTS")
	}

	var count int64
	for _, key := range c.C.Args {
		// GetNoTouch is used to check if a key exists in the store
		// without updating its last access time.
		if s.GetNoTouch(key) != nil {
			count++
		}
	}

	// Return the count as a response
	return &CmdRes{R: &wire.Response{
		Value: &wire.Response_VInt{VInt: count},
	}}, nil
}

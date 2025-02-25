// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dicedb-go/wire"
)

var cType = &CommandMeta{
	Name:      "TYPE",
	HelpShort: "returns the type of the value stored at a specified key",
	Eval:      evalTYPE,
}

func init() {
	CommandRegistry.AddCommand(cType)
}

func evalTYPE(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) != 1 {
		return cmdResNil, errWrongArgumentCount("TYPE")
	}

	key := c.C.Args[0]
	obj := s.Get(key)

	if obj == nil {
		return &CmdRes{R: &wire.Response{
			Value: &wire.Response_VStr{
				VStr: "none",
			},
		}}, nil
	}

	typeStr := obj.Type.String()
	return &CmdRes{R: &wire.Response{
		Value: &wire.Response_VStr{
			VStr: typeStr,
		},
	}}, nil
}

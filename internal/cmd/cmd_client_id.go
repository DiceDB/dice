// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dicedb-go/wire"
)

var cCLIENTID = &DiceDBCommand{
	Name:      "CLIENT.ID",
	HelpShort: "CLIENT.ID gets and sets the client ID",
	Eval:      evalCLIENTID,
}

func init() {
	commandRegistry.AddCommand(cCLIENTID)
}

func evalCLIENTID(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) == 0 {
		return &CmdRes{R: &wire.Response{
			Value: &wire.Response_VStr{VStr: c.ClientID},
		}}, nil
	}
	c.ClientID = c.C.Args[0]
	return cmdResOK, nil
}

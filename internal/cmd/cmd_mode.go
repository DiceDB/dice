// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	dstore "github.com/dicedb/dice/internal/store"
)

var cMODE = &DiceDBCommand{
	Name:      "MODE",
	HelpShort: "MODE sets the mode of the client",
	Eval:      evalMODE,
}

func init() {
	commandRegistry.AddCommand(cMODE)
}

func evalMODE(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	return cmdResOK, nil
}

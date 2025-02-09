// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	dstore "github.com/dicedb/dice/internal/store"
)

var cHANDSHAKE = &DiceDBCommand{
	Name:      "HANDSHAKE",
	HelpShort: "HANDSHAKE is used to handshake with the database; sends client_id and execution mode",
	Eval:      evalHANDSHAKE,
}

func init() {
	commandRegistry.AddCommand(cHANDSHAKE)
}

func evalHANDSHAKE(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) != 2 {
		return cmdResNil, errWrongArgumentCount("HANDSHAKE")
	}
	c.ClientID = c.C.Args[0]
	c.Mode = c.C.Args[1]
	return cmdResOK, nil
}

package cmd

import (
	dstore "github.com/dicedb/dice/internal/store"
)

var cSHUTDOWN = &DiceDBCommand{
	Name:      "SHUTDOWN",
	HelpShort: "SHUTDOWN",
	Eval:      evalSHUTDOWN,
}

func init() {
	commandRegistry.AddCommand(cSHUTDOWN)
}

func evalSHUTDOWN(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) != 0 {
		return cmdResNil, errWrongArgumentCount("SHUTDOWN")
	}

	return cmdResOK, nil
}

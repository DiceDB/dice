package cmd

import (
	dstore "github.com/dicedb/dice/internal/store"
)

var cUNWATCH = &DiceDBCommand{
	Name:      "UNWATCH",
	HelpShort: "UNWATCH removes the previously created query subscription",
	Eval:      evalUNWATCH,
}

func init() {
	commandRegistry.AddCommand(cUNWATCH)
}

func evalUNWATCH(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) != 1 {
		return cmdResNil, errWrongArgumentCount("UNWATCH")
	}

	return cmdResOK, nil
}

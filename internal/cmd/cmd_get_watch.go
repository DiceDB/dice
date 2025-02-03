package cmd

import (
	dstore "github.com/dicedb/dice/internal/store"
)

var cGETWATCH = &DiceDBCommand{
	Name:      "GET.WATCH",
	HelpShort: "GET.WATCH creates a query subscription over the GET command",
	Eval:      evalGETWATCH,
}

func init() {
	commandRegistry.AddCommand(cGETWATCH)
}

func evalGETWATCH(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) != 1 {
		return cmdResNil, errWrongArgumentCount("GET.WATCH")
	}

	r, err := evalGET(c, s)
	if err != nil {
		return nil, err
	}

	return r, nil
}

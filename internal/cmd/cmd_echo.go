package cmd

import (
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dice/wire"
)

var cECHO = &DiceDBCommand{
	Name:      "ECHO",
	HelpShort: "ECHO returns the message passed to it",
	Eval:      evalECHO,
}

func init() {
	commandRegistry.AddCommand(cECHO)
}

func evalECHO(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) != 1 {
		return NewCmdResNil(), errWrongArgumentCount("ECHO")
	}

	return &CmdRes{R: &wire.Response{
		Value: &wire.Response_VStr{VStr: c.C.Args[0]},
	}}, nil
}

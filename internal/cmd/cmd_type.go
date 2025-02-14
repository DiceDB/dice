package cmd

import (
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dicedb-go/wire"
)

var cType = &DiceDBCommand{
	Name:      "TYPE",
	HelpShort: "returns the type of the value stored at a specified key",
	Eval:      evalTYPE,
}

func init() {
	commandRegistry.AddCommand(cType)
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

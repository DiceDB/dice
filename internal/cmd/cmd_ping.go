package cmd

import (
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dice/wire"
)

var cPING = &DiceDBCommand{
	Name:      "PING",
	HelpShort: "PING returns with an encoded \"PONG\" if no message is added with the ping command, the message will be returned.",
	Eval:      evalPING,
}

func init() {
	commandRegistry.AddCommand(cPING)
}

func evalPING(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) >= 2 {
		return NewCmdResNil(), errWrongArgumentCount("PING")
	}
	if len(c.C.Args) == 0 {
		return &CmdRes{R: &wire.Response{
			Value: &wire.Response_VStr{VStr: "PONG"},
		}}, nil
	}
	return &CmdRes{R: &wire.Response{
		Value: &wire.Response_VStr{VStr: c.C.Args[0]},
	}}, nil
}

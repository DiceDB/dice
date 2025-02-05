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

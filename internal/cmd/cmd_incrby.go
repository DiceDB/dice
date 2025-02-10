package cmd

import (
	"strconv"

	dstore "github.com/dicedb/dice/internal/store"
)

var cINCRBY = &DiceDBCommand{
	Name:      "INCRBY",
	HelpShort: "INCRBY decrements the value of the specified key in args by the specified decrement",
	Eval:      evalINCRBY,
}


func init() {
	commandRegistry.AddCommand(cDECRBY)
}

func evalINCRBY(c *Cmd, s *dstore.Store)  (*CmdRes, error) {
	if len(c.C.Args) != 2 {
		return cmdResNil, errWrongArgumentCount("INCRBY")
	}

	delta , err := strconv.ParseInt(c.C.Args[1], 10, 64)
	if err != nil {
		return cmdResNil, errIntegerOutOfRange
	}

	return incrDecr(c,s,delta)
}




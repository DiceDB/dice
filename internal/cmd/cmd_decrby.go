package cmd

import (
	"strconv"

	dstore "github.com/dicedb/dice/internal/store"
)

var cDECRBY = &DiceDBCommand{
	Name:      "DECRBY",
	HelpShort: "DECRBY decrements the value of the specified key in args by the specified decrement",
	Eval:      evalDECRBY,
}

func init() {
	commandRegistry.AddCommand(cDECRBY)
}

func evalDECRBY(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) != 2 {
		return cmdResNil, errWrongArgumentCount("DECRBY")
	}

	delta, err := strconv.ParseInt(c.C.Args[1], 10, 64)
	if err != nil {
		return cmdResNil, errIntegerOutOfRange
	}

	return doIncr(c, s, -delta)
}

package cmd

import (
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dice/wire"
)

var cEXISTS = &DiceDBCommand{
	Name: "EXISTS",
	HelpShort: "Returns the number of keys existing in the db",
	Eval: evalEXISTS,
}

func init() {
	commandRegistry.AddCommand(cEXISTS)
}

func evalEXISTS(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) < 1 {
		return cmdResNil, errWrongArgumentCount("EXISTS")
	}

	var count int64
	for _, key := range c.C.Args {
		// Check if the key exists in the store
		if s.GetNoTouch(key) != nil {
			count++
		}
	}

	// Return the count as a response
	return &CmdRes{R: &wire.Response{
		Value: &wire.Response_VInt{VInt: count},
	}}, nil
}
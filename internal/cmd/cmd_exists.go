package cmd

import (
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dicedb-go/wire"
)

var cEXISTS = &DiceDBCommand{
	Name: "EXISTS",
	HelpShort: "Returns the count of keys that exist among the given arguments without modifying them",
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
		// GetNoTouch is used to check if a key exists in the store 
		// without updating its last access time.
		if s.GetNoTouch(key) != nil {
			count++
		}
	}

	// Return the count as a response
	return &CmdRes{R: &wire.Response{
		Value: &wire.Response_VInt{VInt: count},
	}}, nil
}
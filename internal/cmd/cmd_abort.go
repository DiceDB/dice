package cmd

import (
	"context"

	dstore "github.com/dicedb/dice/internal/store"
)

var cSHUTDOWN = &DiceDBCommand{
	Name:      "ABORT",
	HelpShort: "ABORT",
	Eval:      evalSHUTDOWN,
}

func init() {
	commandRegistry.AddCommand(cSHUTDOWN)
}

func evalSHUTDOWN(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) != 0 {
		return cmdResNil, errWrongArgumentCount("ABORT")
	}
	if c.Ctx == nil {
		return cmdResNil, errContextNotFound
	}

	cancel := c.Ctx.Value("cancel").(context.CancelFunc)
	cancel()

	return cmdResOK, nil
}

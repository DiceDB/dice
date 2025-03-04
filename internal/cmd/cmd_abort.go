package cmd

import (
	"os"

	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
)

var cSHUTDOWN = &CommandMeta{
	Name:      "SHUTDOWN",
	HelpShort: "SHUTDOWN",
	Eval:      evalSHUTDOWN,
	Execute:   executeSHUTDOWN,
}

func init() {
	CommandRegistry.AddCommand(cSHUTDOWN)
}

func evalSHUTDOWN(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) > 1 {
		return cmdResNil, errors.ErrWrongArgumentCount("SHUTDOWN")
	}

	shutdownType := ""
	if len(c.C.Args) == 1 {
		shutdownType = c.C.Args[0]
	}

	p, _ := os.FindProcess(os.Getpid())
	var err error

	if shutdownType == "FORCE" {
		err = p.Signal(os.Kill)
	} else {
		err = p.Signal(os.Interrupt)
	}

	if err != nil {
		return cmdResNil, errors.NewErr(err.Error())
	}
	return cmdResOK, nil
}

func executeSHUTDOWN(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) > 1 {
		return cmdResNil, errors.ErrWrongArgumentCount("SHUTDOWN")
	}
	shard := sm.GetShardForKey("-")
	return evalSHUTDOWN(c, shard.Thread.Store())
}

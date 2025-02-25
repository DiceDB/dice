package cmd

import (
	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
)

var cABORT = &CommandMeta{
	Name:      "ABORT",
	HelpShort: "ABORT",
	Eval:      evalABORT,
	Execute:   executeABORT,
}

func init() {
	CommandRegistry.AddCommand(cABORT)
}

func evalABORT(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) != 0 {
		return cmdResNil, errors.ErrWrongArgumentCount("ABORT")
	}
	return cmdResOK, nil
}

func executeABORT(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	shard := sm.GetShardForKey("-")
	return evalABORT(c, shard.Thread.Store())
}

package cmd

import (
	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/eval/sortedset"
	"github.com/dicedb/dice/internal/shardmanager"
	dsstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dicedb-go/wire"
)

var cZREM = &CommandMeta{
	Name:      "ZREM",
	Syntax:    "ZREM key member [member ...]",
	HelpShort: "Removes the specified members from the sorted set stored at key. Non existing members are ignored.",
	HelpLong: `
Removes the specified members from the sorted set stored at key. Non existing members are ignored.
	`,
	Examples: `
localhost:7379> ZREM mySortedSet "key1" "key2"
`,
	Eval:    evalZREM,
	Execute: executeZREM,
}

func init() {
	CommandRegistry.AddCommand(cZREM)
}

func executeZREM(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) < 2 {
		return cmdResNil, errors.ErrWrongArgumentCount("ZREM")
	}

	shard := sm.GetShardForKey(c.C.Args[0])
	return evalZREM(c, shard.Thread.Store())
}

func evalZREM(c *Cmd, s *dsstore.Store) (*CmdRes, error) {
	if len(c.C.Args) < 2 {
		return cmdResNil, errors.ErrWrongArgumentCount("ZREM")
	}
	key := c.C.Args[0]
	obj := s.Get(key)
	if obj == nil {
		return cmdResInt0, nil
	}
	sortedSet, err := sortedset.FromObject(obj)
	if err != nil {
		return cmdResNil, errors.ErrWrongTypeOperation
	}

	countRem := 0
	for i := 1; i < len(c.C.Args); i++ {
		if sortedSet.Remove(c.C.Args[i]) {
			countRem++
		}
	}

	return &CmdRes{R: &wire.Response{
		Value: &wire.Response_VInt{VInt: int64(countRem)},
	}}, nil
}

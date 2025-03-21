package cmd

import (
	"strconv"
	"strings"

	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/eval/sortedset"
	"github.com/dicedb/dice/internal/shardmanager"
	dsstore "github.com/dicedb/dice/internal/store"
)

var cZRANGE = &CommandMeta{
	Name:      "ZRANGE",
	Syntax:    "ZRANGE key start stop [REV] [WITHSCORES]",
	HelpShort: "Returns the specified range of elements in the sorted set stored at <key>.",
	HelpLong: `
Returns the specified range of elements in the sorted set stored at key.
The elements are considered to be ordered from the lowest to the highest score.
Both start and stop are 0-based indexes, where 0 is the first element, 1 is the next element and so on.
These indexes can also be negative numbers indicating offsets from the end of the sorted set, with -1 being the last element of the sorted set, -2 the penultimate element and so on.
Returns the specified range of elements in the sorted set.
	`,
	Examples: `
localhost:7379> ZRANGE mySortedSet 1 3
`,
	Eval:    evalZRANGE,
	Execute: executeZRANGE,
}

func init() {
	CommandRegistry.AddCommand(cZRANGE)
}

func executeZRANGE(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) < 3 {
		return cmdResNil, errors.ErrWrongArgumentCount("ZRANGE")
	}

	shard := sm.GetShardForKey(c.C.Args[0])
	return evalZRANGE(c, shard.Thread.Store())
}

func evalZRANGE(c *Cmd, s *dsstore.Store) (*CmdRes, error) {
	if len(c.C.Args) < 3 {
		return cmdResNil, errors.ErrWrongArgumentCount("ZRANGE")
	}
	key := c.C.Args[0]
	startStr := c.C.Args[1]
	stopStr := c.C.Args[2]

	withScores := false
	reverse := false

	for i := 3; i < len(c.C.Args); i++ {
		arg := strings.ToUpper(c.C.Args[i])
		if arg == "WITHSCORES" {
			withScores = true
		} else if arg == "REV" {
			reverse = true
		} else {
			return cmdResNil, errors.ErrInvalidSyntax("ZRANGE")
		}
	}

	start, err := strconv.Atoi(startStr)
	if err != nil {
		return cmdResNil, errors.ErrInvalidNumberFormat
	}

	stop, err := strconv.Atoi(stopStr)
	if err != nil {
		return cmdResNil, errors.ErrInvalidNumberFormat
	}

	obj := s.Get(key)
	if obj == nil {
		return cmdResNil, nil
	}

	sortedSet, errMsg := sortedset.FromObject(obj)

	if errMsg != nil {
		return cmdResNil, errors.ErrWrongTypeOperation
	}

	result := sortedSet.GetRange(start, stop, withScores, reverse)

	return GetCmdResStringArray(result)
}

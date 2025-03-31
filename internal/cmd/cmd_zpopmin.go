package cmd

import (
	"strconv"

	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/eval/sortedset"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
)

var cZPOPMIN = &CommandMeta{
	Name:      "ZPOPMIN",
	Syntax:    "ZPOPMIN key [count]",
	HelpShort: "ZPOPMIN returns the lowest-scoring members from a sorted set after removing them. Deletes the sorted set if the last member was popped.",
	HelpLong: `
ZPOPMIN command is used to remove and return the member(s) with the lowest score(s) from a sorted set.
The command supports the following options:
- count: The number of members to remove and return.
	`,
	Examples: `
localhost:7379> ZADD myzset 1 "one"
(integer) 1
localhost:7379> ZADD myzset 2 "two"
(integer) 1
localhost:7379> ZADD myzset 3 "three"
(integer) 1
localhost:7379> ZADD myzset 4 "four"
(integer) 1
localhost:7379> ZPOPMIN myzset 2
1) "one"
2) "1"
3) "two"
4) "2"
	`,
	Eval:    evalZPOPMIN,
	Execute: executeZPOPMIN,
}

func init() {
	CommandRegistry.AddCommand(cZPOPMIN)
}

func evalZPOPMIN(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	key, count, err := parseZPOPMINArgs(c)
	if err != nil {
		return cmdResNil, err
	}

	obj := s.Get(key)
	if obj == nil {
		return cmdResEmptySlice, nil
	}

	sortedSet, errInfo := sortedset.FromObject(obj)
	if errInfo != nil {
		return cmdResNil, errors.ErrWrongTypeOperation
	}

	res := sortedSet.PopMin(count)
	return cmdResSlice(res), nil
}

func parseZPOPMINArgs(c *Cmd) (key string, count int, err error) {
	if len(c.C.Args) == 0 || len(c.C.Args) > 2 {
		return "", 0, errors.ErrWrongArgumentCount("ZPOPMIN")
	}
	key = c.C.Args[0]
	count = 1
	if len(c.C.Args) > 1 {
		count, err = strconv.Atoi(c.C.Args[1])
		if err != nil || count <= 0 {
			return "", 0, errors.ErrIntegerOutOfRange
		}
	}
	return key, count, nil
}

func executeZPOPMIN(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) < 1 {
		return cmdResNil, errors.ErrWrongArgumentCount("ZPOPMIN")
	}
	shard := sm.GetShardForKey(c.C.Args[0])
	return evalZPOPMIN(c, shard.Thread.Store())
}

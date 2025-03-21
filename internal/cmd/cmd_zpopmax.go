package cmd

import (
	"strconv"

	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/eval/sortedset"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
)

var cZPOPMAX = &CommandMeta{
	Name:      "ZPOPMAX",
	Syntax:    "ZPOPMAX key [count]",
	HelpShort: "ZPOPMAX removes and returns the member with the highest score from the sorted set at the specified key.",
	HelpLong: `
Removes and returns the member with the highest score from the sorted set stored at the specified key.

If the key does not exist, the command returns (nil). An optional "count" argument can be provided 
to remove and return multiple members (up to the number specified).

Usage Notes:
- count: The number of members to remove and return.
	`,
	Examples: `
localhost:7379> ZADD myzset 1 "one"
OK
localhost:7379> ZADD myzset 2 "two"
OK
localhost:7379> ZADD myzset 3 "three"
OK
localhost:7379> ZPOPMAX myzset
1) "three"
2) "3"
localhost:7379> ZPOPMAX myzset 2
1) "two"
2) "2"
3) "one"
4) "1"
	`,
	Eval:    evalZPOPMAX,
	Execute: executeZPOPMAX,
}

func init() {
	CommandRegistry.AddCommand(cZPOPMAX)
}

// evalZPOPMAX validates the arguments and executes the ZPOPMAX command logic.
// It returns the highest scoring members removed from the sorted set.
func evalZPOPMAX(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	// Validate that at least one argument (the key) is provided.
	if len(c.C.Args) < 1 {
		return cmdResNil, errors.ErrWrongArgumentCount("ZPOPMAX")
	}
	key := c.C.Args[0]
	count := 1

	// If count is provided, convert it to an integer.
	if len(c.C.Args) > 1 {
		ops, err := strconv.Atoi(c.C.Args[1])
		if err != nil {
			return cmdResNil, errors.ErrInvalidSyntax("ZPOPMAX: count must be an integer")
		}
		if ops <= 0 {
			return cmdResNil, errors.ErrIntegerOutOfRange
		}
		count = ops
	}

	// Retrieve the object from the data store.
	obj := s.Get(key)
	if obj == nil {
		return cmdResNil, nil
	}

	// Attempt to cast the object to a sorted set.
	sortedSet, errMsg := sortedset.FromObject(obj)
	if errMsg != nil {
		return cmdResNil, errors.ErrWrongTypeOperation
	}

	// Remove and return the maximum elements from the sorted set.
	res := sortedSet.PopMax(count)
	response, err := createResponseWithList(res)
	if err != nil {
		return cmdResNil, err
	}
	return &CmdRes{R: response}, nil
}

// executeZPOPMAX retrieves the appropriate shard for the key and evaluates the ZPOPMAX command.
// It returns the result of removing and returning the highest-scored elements.
func executeZPOPMAX(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	// Validate the existence of at least one argument (the key).
	if len(c.C.Args) < 1 {
		return cmdResNil, errors.ErrWrongArgumentCount("ZPOPMAX")
	}
	// Determine the appropriate shard based on the key.
	shard := sm.GetShardForKey(c.C.Args[0])
	return evalZPOPMAX(c, shard.Thread.Store())
}

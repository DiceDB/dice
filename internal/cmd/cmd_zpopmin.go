package cmd

import (
	"strconv"

	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/eval/sortedset"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dicedb-go/wire"
	"google.golang.org/protobuf/types/known/structpb"
)

var cZPOPMIN = &CommandMeta{
	Name:      "ZPOPMIN",
	Syntax:    "ZPOPMIN key [count]",
	HelpShort: "ZPOPMIN removes and returns the member with the lowest score from the sorted set at the specified key.",
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
localhost:7379> ZPOPMIN myzset
1) "one"
2) "1"
localhost:7379> ZPOPMIN myzset 2
1) "two"
2) "2"
3) "three"
4) "3"
	`,
	Eval:    evalZPOPMIN,
	Execute: executeZPOPMIN,
}

func init() {
	CommandRegistry.AddCommand(cZPOPMIN)
}

// evalZPOPMIN validates the arguments and executes the ZPOPMIN logic.
// It returns the lowest scoring member removed from the sorted set.
func evalZPOPMIN(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	// Validate that at least one argument is provided.

	if len(c.C.Args) < 1 {
		return cmdResNil, errors.ErrWrongArgumentCount("ZPOPMIN")
	}
	key := c.C.Args[0]
	count := 1

	if len(c.C.Args) > 1 {
		ops, err := strconv.Atoi(c.C.Args[1])
		if err != nil {
			return cmdResNil, errors.ErrInvalidSyntax("ZPOPMIN: count must be an integer")
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

	res := sortedSet.PopMin(count)
	response, err := createResponseWithList(res)
	if err != nil {
		return cmdResNil, err
	}
	return &CmdRes{R: response}, nil
}

// executeZPOPMIN retrieves the appropriate shard for the key and evaluates the ZPOPMIN command.
// It returns the result of removing and returning the highest-scored elements.
func executeZPOPMIN(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	//Validate the existence atleast one argument.
	if len(c.C.Args) < 1 {
		return cmdResNil, errors.ErrWrongArgumentCount("ZPOPMIN")
	}
	//Determine the shard for the key.
	shard := sm.GetShardForKey(c.C.Args[0])
	return evalZPOPMIN(c, shard.Thread.Store())
}

func createResponseWithList(strings []string) (*wire.Response, error) {
	var values []*structpb.Value

	// Convert each string to structpb.Value
	for _, str := range strings {
		val, err := structpb.NewValue(str)
		if err != nil {
			return nil, err
		}
		values = append(values, val)
	}

	return &wire.Response{
		VList: values,
	}, nil
}

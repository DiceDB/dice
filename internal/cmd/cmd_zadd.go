package cmd

import (
	"math"
	"strconv"
	"strings"

	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/eval/sortedset"
	"github.com/dicedb/dice/internal/object"
	"github.com/dicedb/dice/internal/shardmanager"
	dsstore "github.com/dicedb/dice/internal/store"
)

var cZADD = &CommandMeta{
	Name:      "ZADD",
	Syntax:    "ZADD key [NX | XX] [GT | LT] [CH] [INCR] score member [score member...]",
	HelpShort: "Adds all the specified members with the specified scores to the sorted set stored at key",
	HelpLong: `
Adds all the specified members with the specified scores to the sorted set stored at key
Options: NX, XX, CH, INCR, GT, LT, CH
- NX: Only add new elements and do not update existing elements
- XX: Only update existing elements and do not add new elements
- CH: Modify the return value from the number of new elements added, to the total number of elements changed
- INCR: When this option is specified, the elements are treated as increments to the score of the existing elements
- GT: Only add new elements if the score is greater than the existing score
- LT: Only add new elements if the score is less than the existing score
Returns the number of elements added to the sorted set, not including elements already existing for which the score was updated.
	`,
	Examples: `
localhost:7379> ZADD mySortedSet 1 foo 2 bar
OK 2
`,
	Eval:    evalZADD,
	Execute: executeZADD,
}

func init() {
	CommandRegistry.AddCommand(cZADD)
}

func evalZADD(c *Cmd, s *dsstore.Store) (*CmdRes, error) {
	if len(c.C.Args) < 3 {
		return cmdResNil, errors.ErrWrongArgumentCount("ZADD")
	}

	key := c.C.Args[0]
	params := map[float64]string{}
	flags := map[string]bool{}

	for i := 1; i < len(c.C.Args); i++ {
		arg := strings.ToUpper(c.C.Args[i])
		switch arg {
		// Only valid options are allowed
		case "XX", "NX", "LT", "GT", "CH", "INCR":
			flags[arg] = true
		default:
			// This should be a float repesneting the score. If its not, its a problem.
			score, err := strconv.ParseFloat(arg, 64)
			if err != nil || math.IsNaN(score) {
				return cmdResNil, errors.ErrInvalidNumberFormat
			}
			if i+1 >= len(c.C.Args) {
				return cmdResNil, errors.ErrWrongArgumentCount("ZADD")
			}
			// Its a score-member pair
			params[score] = c.C.Args[i+1]
			i++
		}
	}

	if flags["NX"] && flags["XX"] {
		return cmdResNil, errors.ErrGeneral("XX and NX options at the same time are not compatible")
	}
	if (flags["GT"] && flags["NX"]) || (flags["LT"] && flags["NX"]) || (flags["GT"] && flags["LT"]) {
		return cmdResNil, errors.ErrGeneral("GT, LT, and/or NX options at the same time are not compatible")
	}
	if flags["INCR"] && len(params) > 1 {
		return cmdResNil, errors.ErrGeneral("INCR option supports a single increment-element pair")
	}
	sortedSet, err := getOrCreateSortedSet(s, key)
	if err != nil {
		return cmdResNil, err
	}
	// all processing takes place here
	return processMembersWithFlags(params, sortedSet, s, key, flags)
}

func executeZADD(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) < 3 {
		return cmdResNil, errors.ErrWrongArgumentCount("ZADD")
	}

	shard := sm.GetShardForKey(c.C.Args[0])
	return evalZADD(c, shard.Thread.Store())
}

func getOrCreateSortedSet(store *dsstore.Store, key string) (*sortedset.Set, error) {
	obj := store.Get(key)
	if obj != nil {
		sortedSet, err := sortedset.FromObject(obj)
		if err != nil {
			return nil, errors.ErrWrongTypeOperation
		}
		return sortedSet, nil
	}
	return sortedset.New(), nil
}

// processMembersWithFlags processes the members and scores while handling flags.
func processMembersWithFlags(params map[float64]string, sortedSet *sortedset.Set, store *dsstore.Store, key string, flags map[string]bool) (*CmdRes, error) {
	added, updated := 0, 0

	for score, member := range params {
		currentScore, exists := sortedSet.Get(member)

		// If INCR is used, increment the score first
		if flags["INCR"] {
			if exists {
				score += currentScore
			} else {
				score = 0.0 + score
			}

			// Now check GT and LT conditions based on the incremented score and return accordingly
			if (flags["GT"] && exists && score <= currentScore) ||
				(flags["LT"] && exists && score >= currentScore) {
				return nil, nil
			}
		}

		// Check if the member should be skipped based on NX or XX flags
		if shouldSkipMember(score, currentScore, exists, flags) {
			continue
		}

		// Insert or update the member in the sorted set
		wasInserted := sortedSet.Upsert(score, member)

		if wasInserted && !exists {
			added++
		} else if exists && score != currentScore {
			updated++
		}

		// If INCR is used, exit after processing one score-member pair
		if flags["INCR"] {
			return cmdResFloat(score), nil
		}
	}

	// Store the updated sorted set in the store
	storeUpdatedSet(store, key, sortedSet)

	if flags["CH"] {
		return cmdResInt(int64(added + updated)), nil
	}

	// Return only the count of added members
	return cmdResInt(int64(added)), nil
}

// shouldSkipMember determines if a member should be skipped based on flags.
func shouldSkipMember(score, currentScore float64, exists bool, flags map[string]bool) bool {
	if flags["NX"] && exists || flags["XX"] && !exists {
		return true
	}
	if exists {
		if flags["LT"] && score >= currentScore {
			return true
		}
		if flags["GT"] && score <= currentScore {
			return true
		}
	}
	return false
}

// storeUpdatedSet stores the updated sorted set in the store.
func storeUpdatedSet(store *dsstore.Store, key string, sortedSet *sortedset.Set) {
	store.Put(key, store.NewObj(sortedSet, -1, object.ObjTypeSortedSet), dsstore.WithPutCmd(dsstore.ZAdd))
}

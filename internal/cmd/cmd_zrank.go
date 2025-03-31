package cmd

import (
	"strings"

	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/eval/sortedset"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
	"google.golang.org/protobuf/types/known/structpb"
)

var cZRANK = &CommandMeta{
	Name:      "ZRANK",
	Syntax:    "ZRANK key member [WITHSCORE]",
	HelpShort: "ZRANK returns the rank of a member in a sorted set, ordered from low to high scores.",
	HelpLong: `
ZRANK returns the 0-based rank (position) of a member in a sorted set, 
with scores ordered from low to high. If the optional WITHSCORE flag is provided, 
the score will also be returned.
- returns nil if the member does not exist in the sorted set
- rank starts at 0 for the member with the lowest score
- WITHSCORE returns both rank and score
	`,
	Examples: `
localhost:7379> ZADD myzset 1 "one"
(integer) 1
localhost:7379> ZADD myzset 2 "two"
(integer) 1
localhost:7379> ZADD myzset 3 "three"
(integer) 1
localhost:7379> ZRANK myzset "two"
(integer) 1
localhost:7379> ZRANK myzset "five"
(nil)
localhost:7379> ZRANK myzset "three" WITHSCORE
1) (integer) 2
2) "3"
	`,
	Eval:    evalZRANK,
	Execute: executeZRANK,
}

func init() {
	CommandRegistry.AddCommand(cZRANK)
}

// evalZRANK returns the rank of the member in the sorted set stored at key.
// The rank (or index) is 0-based, which means that the member with the lowest score has rank 0.
// If the 'WITHSCORE' option is specified, it returns both the rank and the score of the member.
// Returns nil if the key does not exist or the member is not a member of the sorted set.
func evalZRANK(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	key, member, withScore, err := parseZRANKArgs(c)
	if err != nil {
		return cmdResNil, err
	}
	obj := s.Get(key)
	if obj == nil {
		return cmdResNil, errors.ErrWrongTypeOperation
	}

	sortedSet, errInfo := sortedset.FromObject(obj)
	if errInfo != nil {
		return cmdResNil, errors.ErrWrongTypeOperation
	}
	rank, score := sortedSet.RankWithScore(member, false)
	if rank == -1 {
		return cmdResNil, errors.ErrWrongTypeOperation
	}

	result := cmdResIntSlice([]int64{rank})

	if withScore {
		result.R.VList = append(result.R.VList, structpb.NewNumberValue(float64(score)))
	}

	return result, nil
}

func parseZRANKArgs(c *Cmd) (key string, member string, withScore bool, err error) {
	key = c.C.Args[0]
	member = c.C.Args[1]
	withScore = false
	if len(c.C.Args) > 2 {
		if strings.EqualFold(c.C.Args[2], "WITHSCORE") {
			withScore = true
		} else {
			err = errors.ErrIntegerOutOfRange
			return
		}
	}
	return
}

func executeZRANK(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) < 2 {
		return cmdResNil, errors.ErrWrongArgumentCount("ZRANK")
	}
	shard := sm.GetShardForKey(c.C.Args[0])
	return evalZRANK(c, shard.Thread.Store())
}

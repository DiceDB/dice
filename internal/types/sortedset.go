// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package types

import (
	"errors"

	"github.com/dicedb/dicedb-go/wire"
	"github.com/wangjia184/sortedset"
)

type SortedSet struct {
	*sortedset.SortedSet
}

func NewSortedSet() *SortedSet {
	return &SortedSet{
		SortedSet: sortedset.New(),
	}
}

func (s *SortedSet) ZADD(scores []int64, members []string, params map[Param]string) (int64, error) {
	addedCount, updatedCount := 0, 0

	if params[XX] != "" && params[NX] != "" {
		return 0, errors.New("XX and NX options at the same time are not compatible")
	}
	if (params[GT] != "" && params[NX] != "") ||
		(params[LT] != "" && params[NX] != "") ||
		(params[GT] != "" && params[LT] != "") {
		return 0, errors.New("GT, LT, and/or NX options at the same time are not compatible")
	}
	if params[INCR] != "" && len(members) > 1 {
		return 0, errors.New("INCR option supports a single increment-element pair")
	}

	for i := range scores {
		score, member := scores[i], members[i]
		n := s.GetByKey(member)
		exists := n != nil
		currentScore := sortedset.SCORE(0)
		if exists {
			currentScore = n.Score()
		}

		// Handle INCR option
		if params[INCR] != "" {
			if exists {
				score = int64(currentScore) + score
			}
			s.AddOrUpdate(member, sortedset.SCORE(score), nil)
			return score, nil
		}

		// Skip based on NX/XX flags
		if (params[NX] != "" && exists) || (params[XX] != "" && !exists) {
			continue
		}

		// Skip based on GT/LT conditions
		if exists {
			if params[GT] != "" && sortedset.SCORE(score) <= currentScore {
				continue
			}
			if params[LT] != "" && sortedset.SCORE(score) >= currentScore {
				continue
			}
		}

		// Add or update the member
		wasInserted := s.AddOrUpdate(member, sortedset.SCORE(score), nil)
		if wasInserted && !exists {
			addedCount++
		} else if exists && sortedset.SCORE(score) != currentScore {
			updatedCount++
		}
	}

	// Return appropriate count based on CH flag
	if params[CH] != "" {
		return int64(addedCount + updatedCount), nil
	}
	return int64(addedCount), nil
}

func (s *SortedSet) ZCOUNT(minScore, maxScore int64) int64 {
	return int64(len(s.SortedSet.GetByScoreRange(sortedset.SCORE(minScore), sortedset.SCORE(maxScore), nil)))
}

func (s *SortedSet) ZRANGE(start, stop int, byScore, byRank bool) []*wire.ZElement {
	var nodes []*sortedset.SortedSetNode
	if byScore {
		nodes = s.SortedSet.GetByScoreRange(
			sortedset.SCORE(start),
			sortedset.SCORE(stop),
			&sortedset.GetByScoreRangeOptions{})
	} else {
		nodes = s.SortedSet.GetByRankRange(start, stop, false)
	}

	var rank int64 = 0
	if len(nodes) > 0 {
		rank = int64(s.SortedSet.FindRank(nodes[0].Key()))
	}

	result := make([]*wire.ZElement, len(nodes))
	for i, node := range nodes {
		result[i] = &wire.ZElement{
			Member: node.Key(),
			Score:  int64(node.Score()),
			Rank:   rank + int64(i),
		}
	}
	return result
}

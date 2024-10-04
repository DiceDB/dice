package sortedset

import (
	"strconv"
	"strings"

	diceerrors "github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/object"
	"github.com/google/btree"
)

// Item represents a member of a sorted set. It includes a score and a member.
type Item struct {
	btree.Item
	Score  float64
	Member string
}

// Less compares two Items. Required by the btree.Item interface.
func (a *Item) Less(b btree.Item) bool {
	other := b.(*Item)
	if a.Score != other.Score {
		return a.Score < other.Score
	}
	return a.Member < other.Member
}

// is a sorted set data structure that stores members with associated scores.
type Set struct {
	// tree is a btree that stores Items.
	tree *btree.BTree
	// memberMap is a map that stores members and their scores.
	memberMap map[string]float64
}

// New creates a new .
func New() *Set {
	return &Set{
		tree:      btree.New(2),
		memberMap: make(map[string]float64),
	}
}

func FromObject(obj *object.Obj) (value *Set, err []byte) {
	if err := object.AssertTypeAndEncoding(obj.TypeEncoding, object.ObjTypeSortedSet, object.ObjEncodingBTree); err != nil {
		return nil, err
	}
	value, ok := obj.Value.(*Set)
	if !ok {
		return nil, diceerrors.NewErrWithMessage("Invalid sorted set object")
	}
	return value, nil
}

// Add adds a member with a score to the  and returns true if the member was added, false if it already existed.
func (ss *Set) Upsert(score float64, member string) bool {
	existingScore, exists := ss.memberMap[member]

	if exists {
		oldItem := &Item{Score: existingScore, Member: member}
		ss.tree.Delete(oldItem)
	}

	item := &Item{Score: score, Member: member}
	ss.tree.ReplaceOrInsert(item)
	ss.memberMap[member] = score

	return !exists
}

// Remove removes a member from the  and returns true if the member was removed, false if it did not exist.
func (ss *Set) Remove(member string) bool {
	score, exists := ss.memberMap[member]
	if !exists {
		return false
	}

	item := &Item{Score: score, Member: member}
	ss.tree.Delete(item)
	delete(ss.memberMap, member)

	return true
}

// GetRange returns a slice of members with scores between min and max, inclusive.
// it returns the members in ascending order if reverse is false, and descending order if reverse is true.
// If withScores is true, the members will be returned with their scores.
func (ss *Set) GetRange(
	start, stop int,
	withScores bool,
	reverse bool,
) []string {
	length := ss.tree.Len()
	if start < 0 {
		start += length
	}
	if stop < 0 {
		stop += length
	}

	if start < 0 {
		start = 0
	}
	if stop >= length {
		stop = length - 1
	}

	if start > stop || start >= length {
		return []string{}
	}

	var result []string

	index := 0

	// iterFunc is the function that will be called for each item in the B-tree. It will append the item to the result if it is within the specified range.
	// It will return false if the specified range has been reached.
	iterFunc := func(item btree.Item) bool {
		if index > stop {
			return false
		}

		if index >= start {
			ssi := item.(*Item)
			result = append(result, ssi.Member)
			if withScores {
				// Use 'g' format to match Redis's float formatting
				scoreStr := strings.ToLower(strconv.FormatFloat(ssi.Score, 'g', -1, 64))
				result = append(result, scoreStr)
			}
		}
		index++
		return true
	}

	if reverse {
		ss.tree.Descend(iterFunc)
	} else {
		ss.tree.Ascend(iterFunc)
	}

	return result
}

func (ss *Set) Get(member string) (float64, bool) {
	score, exists := ss.memberMap[member]
	return score, exists
}

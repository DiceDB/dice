package eval

import "github.com/google/btree"

// SortedSetItem represents a member of a sorted set. It includes a score and a member.
type SortedSetItem struct {
	btree.Item
	Score  float64
	Member string
}

// Less compares two SortedSetItems. Required by the btree.Item interface.
func (a *SortedSetItem) Less(b btree.Item) bool {
	other := b.(*SortedSetItem)
	if a.Score != other.Score {
		return a.Score < other.Score
	}
	return a.Member < other.Member
}

/*
 * Test Suite for Sorted Set implementation for Dice.
 * The sorted set is backed by a SkipList and a dictionary.
 *
 * The implementation is inspired by the following open-source projects:
 * - Redis: https://github.com/redis/redis/blob/unstable/src/t_zset.c
 * - https://github.com/wangjia184/sortedset/
 * - https://github.com/liyiheng/zset/
 */

package core

import (
	"testing"
)

func checkOrder(t *testing.T, nodes []*zskiplistNode, expectedOrder []int64) {
	if len(expectedOrder) != len(nodes) {
		t.Errorf("nodes does not contain %d elements", len(expectedOrder))
	}
	for i := 0; i < len(expectedOrder); i++ {
		if nodes[i].Key() != expectedOrder[i] {
			t.Errorf("nodes[%d] is %q, but the expected key is %q", i, nodes[i].Key(), expectedOrder[i])
		}

	}
}
func checkIterByRankRange(t *testing.T, sortedset *ZSet, start int, end int, expectedOrder []int64) {
	var keys []int64

	// check nil callback should do nothing
	sortedset.IterFuncByRankRange(start, end, nil)

	sortedset.IterFuncByRankRange(start, end, func(key int64, _ interface{}) bool {
		keys = append(keys, key)
		return true
	})
	if len(expectedOrder) != len(keys) {
		t.Errorf("keys does not contain %d elements", len(expectedOrder))
	}
	for i := 0; i < len(expectedOrder); i++ {
		if keys[i] != expectedOrder[i] {
			t.Errorf("keys[%d] is %q, but the expected key is %q", i, keys[i], expectedOrder[i])
		}
	}

	// check return early
	if len(expectedOrder) < 1 {
		return
	}
	// reset data
	keys = []int64{}
	var i int
	sortedset.IterFuncByRankRange(start, end, func(key int64, _ interface{}) bool {
		keys = append(keys, key)
		i++
		// return early
		return i < len(expectedOrder)-1
	})
	if len(expectedOrder)-1 != len(keys) {
		t.Errorf("keys does not contain %d elements", len(expectedOrder)-1)
	}
	for i := 0; i < len(expectedOrder)-1; i++ {
		if keys[i] != expectedOrder[i] {
			t.Errorf("keys[%d] is %q, but the expected key is %q", i, keys[i], expectedOrder[i])
		}
	}

}

func checkRankRangeIterAndOrder(t *testing.T, sortedset *ZSet, start int, end int, remove bool, expectedOrder []int64) {
	checkIterByRankRange(t, sortedset, start, end, expectedOrder)
	nodes := sortedset.GetByRankRange(start, end, remove)
	checkOrder(t, nodes, expectedOrder)
}

func TestCase1(t *testing.T) {
	sortedset := NewZSet()

	sortedset.AddOrUpdate(1, 89, "Kelly")
	sortedset.AddOrUpdate(2, 100, "Staley")
	sortedset.AddOrUpdate(3, 100, "Jordon")
	sortedset.AddOrUpdate(4, -321, "Park")
	sortedset.AddOrUpdate(5, 101, "Albert")
	sortedset.AddOrUpdate(6, 99, "Lyman")
	sortedset.AddOrUpdate(7, 99, "Singleton")
	sortedset.AddOrUpdate(8, 70, "Audrey")

	sortedset.AddOrUpdate(5, 99, "ntrnrt")

	sortedset.Remove(2)

	node := sortedset.GetByRank(3, false)
	if node == nil || node.Key() != 1 {
		t.Error("GetByRank() does not return expected value `1`")
	}

	node = sortedset.GetByRank(-3, false)
	if node == nil || node.Key() != 6 {
		t.Error("GetByRank() does not return expected value `6`")
	}

	// get all nodes since the first one to last one
	checkRankRangeIterAndOrder(t, sortedset, 1, -1, false, []int64{4, 8, 1, 5, 6, 7, 3})

	// get & remove the 2nd/3rd nodes in reserve order
	checkRankRangeIterAndOrder(t, sortedset, -2, -3, true, []int64{7, 6})

	// get all nodes since the last one to first one
	checkRankRangeIterAndOrder(t, sortedset, -1, 1, false, []int64{3, 5, 1, 8, 4})

}

func TestCase2(t *testing.T) {

	// create a new set
	sortedset := NewZSet()

	// fill in new node
	sortedset.AddOrUpdate(1, 89, "Kelly")
	sortedset.AddOrUpdate(2, 100, "Staley")
	sortedset.AddOrUpdate(3, 100, "Jordon")
	sortedset.AddOrUpdate(4, -321, "Park")
	sortedset.AddOrUpdate(5, 101, "Albert")
	sortedset.AddOrUpdate(6, 99, "Lyman")
	sortedset.AddOrUpdate(7, 99, "Singleton")
	sortedset.AddOrUpdate(8, 70, "Audrey")

	// update an existing node
	sortedset.AddOrUpdate(5, 99, "ntrnrt")

	// remove node
	sortedset.Remove(2)

	nodes := sortedset.GetByScoreRange(-500, 500, nil)
	checkOrder(t, nodes, []int64{4, 8, 1, 5, 6, 7, 3})

	nodes = sortedset.GetByScoreRange(500, -500, nil)
	//t.Logf("%v", nodes)
	checkOrder(t, nodes, []int64{3, 7, 6, 5, 1, 8, 4})

	nodes = sortedset.GetByScoreRange(600, 500, nil)
	checkOrder(t, nodes, []int64{})

	nodes = sortedset.GetByScoreRange(500, 600, nil)
	checkOrder(t, nodes, []int64{})

	rank := sortedset.FindRank(6)
	if rank != 5 {
		t.Error("FindRank() does not return expected value `5`")
	}

	rank = sortedset.FindRank(4)
	if rank != 1 {
		t.Error("FindRank() does not return expected value `1`")
	}

	nodes = sortedset.GetByScoreRange(99, 100, nil)
	checkOrder(t, nodes, []int64{5, 6, 7, 3})

	nodes = sortedset.GetByScoreRange(90, 50, nil)
	checkOrder(t, nodes, []int64{1, 8})

	nodes = sortedset.GetByScoreRange(99, 100, &GetByScoreRangeOptions{
		ExcludeStart: true,
	})
	checkOrder(t, nodes, []int64{3})

	nodes = sortedset.GetByScoreRange(100, 99, &GetByScoreRangeOptions{
		ExcludeStart: true,
	})
	checkOrder(t, nodes, []int64{7, 6, 5})

	nodes = sortedset.GetByScoreRange(99, 100, &GetByScoreRangeOptions{
		ExcludeEnd: true,
	})
	checkOrder(t, nodes, []int64{5, 6, 7})

	nodes = sortedset.GetByScoreRange(100, 99, &GetByScoreRangeOptions{
		ExcludeEnd: true,
	})
	checkOrder(t, nodes, []int64{3})

	nodes = sortedset.GetByScoreRange(50, 100, &GetByScoreRangeOptions{
		Limit: 2,
	})
	checkOrder(t, nodes, []int64{8, 1})

	nodes = sortedset.GetByScoreRange(100, 50, &GetByScoreRangeOptions{
		Limit: 2,
	})
	checkOrder(t, nodes, []int64{3, 7})

	minNode := sortedset.PeekMin()
	if minNode == nil || minNode.Key() != 4 {
		t.Error("PeekMin() does not return expected value `4`")
	}

	minNode = sortedset.PopMin()
	if minNode == nil || minNode.Key() != 4 {
		t.Error("PopMin() does not return expected value `4`")
	}

	nodes = sortedset.GetByScoreRange(-500, 500, nil)
	checkOrder(t, nodes, []int64{8, 1, 5, 6, 7, 3})

	maxNode := sortedset.PeekMax()
	if maxNode == nil || maxNode.Key() != 3 {
		t.Error("PeekMax() does not return expected value `3`")
	}

	maxNode = sortedset.PopMax()
	if maxNode == nil || maxNode.Key() != 3 {
		t.Error("PopMax() does not return expected value `3`")
	}

	nodes = sortedset.GetByScoreRange(500, -500, nil)
	checkOrder(t, nodes, []int64{7, 6, 5, 1, 8})
}

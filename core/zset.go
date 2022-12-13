/*
 * ZSet is the sorted set implementation for Dice.
 * The sorted set is backed by a SkipList and a dictionary.
 *
 * The implementation is inspired by the following open-source projects:
 * - Redis: https://github.com/redis/redis/blob/unstable/src/t_zset.c
 * - https://github.com/wangjia184/sortedset/
 * - https://github.com/liyiheng/zset/
 */

package core

import "math/rand"

type SCORE float64

const zskiplist_max_level = 32 /* Should be enough for 2^32 elements */
const zskiplist_p = 0.25       /* Skiplist P = 1/4 */

type (
	zskiplistLevel struct {
		// Forward is the first node in the level
		forward *zskiplistNode
		// span at a particular node stores the number of nodes between the current
		// node and node->forward at the current level.
		// span is used to calculate the 1-based rank of element in the skip list.
		span int64
	}

	/*
	 * Represents a node in the SkipList.
	 * We store only integer values since we can use the pointer to the actual
	 * objects as the keys in the dictionary.
	 * The pointers are all 8 bytes, so we can use int64 to store them.
	 */
	zskiplistNode struct {
		key      int64       // Integer key of this object
		value    interface{} // Value of this object
		score    SCORE       // Score of this node
		backward *zskiplistNode
		level    []zskiplistLevel
	}

	/*
	 * Represents the SkipList.
	 * Each skiplist keeps track of its header, tail, length, and total levels.
	 */
	zskiplist struct {
		header *zskiplistNode
		tail   *zskiplistNode
		length int64
		level  int16
	}

	/*
	 * Dice Sorted Set.
	 * Provides a data structure optimized for storing and retrieving elements
	 * based on score, in addition to rank and range queries.
	 */
	ZSet struct {
		dict map[int64]*zskiplistNode // Dictionary of all the elements for O(1) access
		zsl  *zskiplist               // SkipList of all elements in sorted order of Score.
	}
)

// Get the Key of the node.
func (zn *zskiplistNode) Key() int64 {
	return zn.key
}

// Get the score of the node.
func (zn *zskiplistNode) Score() SCORE {
	return zn.score
}

// Creates a new zskiplistNode.
func zslCreateNode(level int16, score SCORE, key int64, value interface{}) *zskiplistNode {
	// Initialize the node
	node := &zskiplistNode{
		key:   key,
		value: value,
		score: score,
		level: make([]zskiplistLevel, level),
	}
	return node
}

// Creates a new zskiplist.
func zslCreate() *zskiplist {
	return &zskiplist{
		level:  1,
		header: zslCreateNode(zskiplist_max_level, 0, 0, 0),
	}
}

/* Returns a random level for the new skiplist node we are going to create.
 * The return value of this function is between 1 and SKIPLIST_P
 * (both inclusive), with a powerlaw-alike distribution where higher
 * levels are less likely to be returned.
 */
func randomLevel() int16 {
	// default level
	level := int16(1)

	// Increase the level while the "coin flips as heads".
	// Higher levels are less likely to be returned.
	for float32(rand.Int31()&0xFFFF) < (zskiplist_p * 0xFFFF) {
		level++
	}
	if level < zskiplist_max_level {
		return level
	}
	return zskiplist_max_level
}

/*
 * zslInsert a new node in the skiplist. Assumes the element does not already
 * exist (up to the caller to enforce that).
 */
func (zsl *zskiplist) zslInsert(score SCORE, key int64, value interface{}) *zskiplistNode {
	// The update array is used to keep track of the nodes that are updated
	// during the insertion process. This is used to update the backward
	// pointers of the nodes that are inserted.
	var update [zskiplist_max_level]*zskiplistNode
	var rank [zskiplist_max_level]int64

	x := zsl.header

	// Start traversing the skiplist from the highest level.
	for currentLevel := zsl.level - 1; currentLevel >= 0; currentLevel-- {
		// Store the rank that is crossed to reach the insert position
		if currentLevel == zsl.level-1 {
			// Rank at the top level is always 0.
			rank[currentLevel] = 0
		} else {
			rank[currentLevel] = rank[currentLevel+1]
		}

		// Keep traversing the skiplist while the score is less than the current
		// node.
		// If the score is equal, then we check the key.
		for x.level[currentLevel].forward != nil &&
			(x.level[currentLevel].forward.score < score ||
				// TODO: This might not be a good way to handle the same score.
				// We probably want the actual value stored at the pointer represented
				// by this key.
				(x.level[currentLevel].forward.score == score &&
					x.level[currentLevel].forward.key < key)) {
			// Update the rank by adding the span of the current node to the
			// rank. This is essentially the position of the current node in
			// the overall list.
			rank[currentLevel] += x.level[currentLevel].span
			// Move forward
			x = x.level[currentLevel].forward

		}
		update[currentLevel] = x
	}

	// We assume the element does not exist, since we allow duplicated scores,
	// reinserting the same element should never happen since the caller of
	// zslInsert() should test in the hash table if the element is already
	// inside or not.
	level := randomLevel()

	// If the new node is going to have a higher level than the current
	// skiplist level, we need to update our `rank` and `update` arrays.
	if level > zsl.level {
		for currentLevel := zsl.level; currentLevel < level; currentLevel++ {
			rank[currentLevel] = 0
			update[currentLevel] = zsl.header
			update[currentLevel].level[currentLevel].span = zsl.length
		}
		zsl.level = level
	}

	x = zslCreateNode(level, score, key, value)

	// Update the level pointers of the new node we are inserting.
	for currentLevel := int16(0); currentLevel < level; currentLevel++ {
		prevNode := update[currentLevel]
		// Update the forward pointer of the new node we are inserting at
		// currentLevel to the forward pointer of the update node at this level.
		x.level[currentLevel].forward = prevNode.level[currentLevel].forward
		// Update the forward pointer of the update node to point to the new
		// node we are inserting.
		prevNode.level[currentLevel].forward = x

		// Set the span of the new node = (span of prevNode) - (difference in
		// position of new node at 0th level (actual position) and current level)
		x.level[currentLevel].span = prevNode.level[currentLevel].span - (rank[0] - rank[currentLevel])
		// Update span covered by update[i] as x is inserted here.
		prevNode.level[currentLevel].span = (rank[0] - rank[currentLevel]) + 1
	}

	// increment span for all the untouched levels
	for currentLevel := level; currentLevel < zsl.level; currentLevel++ {
		update[currentLevel].level[currentLevel].span++
	}

	// Update the backward pointer of the new node we are inserting.
	if update[0] == zsl.header {
		x.backward = nil
	} else {
		x.backward = update[0]
	}

	// Update the backward pointer of the node that is forward of the new node
	if x.level[0].forward != nil {
		x.level[0].forward.backward = x
	} else {
		zsl.tail = x
	}

	zsl.length++
	return x
}

// Internal function used for delete options.
//
// `x` is the node to delete.
//
// `update` is the array of nodes that are updated during the deletion process.
func (zsl *zskiplist) zslDeleteNode(x *zskiplistNode, update [zskiplist_max_level]*zskiplistNode) {
	// Update the node before the node we are deleting to point to the node
	// after the node we are deleting. Also update the span covered by this
	// node.
	for currentLevel := int16(0); currentLevel < zsl.level; currentLevel++ {
		prevNode := update[currentLevel]
		if prevNode.level[currentLevel].forward == x {
			prevNode.level[currentLevel].span += x.level[currentLevel].span - 1
			prevNode.level[currentLevel].forward = x.level[currentLevel].forward
		} else {
			prevNode.level[currentLevel].span--
		}
	}

	// Update the backward pointer of the node after the node we are deleting.
	if x.level[0].forward != nil {
		x.level[0].forward.backward = x.backward
	} else {
		zsl.tail = x.backward
	}

	// Update the level of the skiplist if the forward node of the header at the
	// highest level is nil.
	for zsl.level > 1 && zsl.header.level[zsl.level-1].forward == nil {
		zsl.level--
	}

	zsl.length-- // Decrement the length of the skiplist.
}

// Delete an element with matching score/key from the skiplist.
func (zsl *zskiplist) zslDelete(score SCORE, key int64) int {
	var update [zskiplist_max_level]*zskiplistNode

	x := zsl.header

	// Find the node that we want to delete.
	// Also store the nodes that will be updated during the deletion process.
	for currentLevel := zsl.level - 1; currentLevel >= 0; currentLevel-- {
		for x.level[currentLevel].forward != nil &&
			// We may have multiple elements with the same score, what we need is to find
			// the element with both the right score and object.
			(x.level[currentLevel].forward.score < score ||
				(x.level[currentLevel].forward.score == score &&
					x.level[currentLevel].forward.key < key)) {
			x = x.level[currentLevel].forward
		}
		update[currentLevel] = x
	}

	x = x.level[0].forward
	if x != nil && score == x.score && key == x.key {
		zsl.zslDeleteNode(x, update)
		return 1 // deleted
	}

	return 0 // not found
}

/*-------------------------------------------------------
* Common Sorted Set API
-------------------------------------------------------*/

// NewZSet creates a new SortedSet and returns its pointer.
func NewZSet() *ZSet {
	return &ZSet{
		dict: make(map[int64]*zskiplistNode),
		zsl:  zslCreate(),
	}
}

// Length returns the total number of elements in the SortedSet.
func (z *ZSet) Length() int64 {
	return z.zsl.length
}

// Add an element into the sorted set with specific key / value / score.
// If the element is added, this method returns true; otherwise false means
// the existing value was updated.
//
// Time complexity: O(log(N))
func (z *ZSet) AddOrUpdate(key int64, score SCORE, value interface{}) bool {
	var newNode *zskiplistNode = nil

	// Try to find the value in the dictionary.
	found := z.dict[key]

	// Case 1: The element is present in the sorted set.
	if found != nil {
		// score does not change, only update value
		if found.score == score {
			found.value = value
		} else {
			// score has changed, delete and re-insert
			z.zsl.zslDelete(found.score, key)
			newNode = z.zsl.zslInsert(score, key, value)
		}
	} else {
		// Case 2: The element is not in the sorted set, insert it.
		newNode = z.zsl.zslInsert(score, key, value)
	}

	if newNode != nil {
		z.dict[key] = newNode
	}

	return found == nil
}

// Delete the element at a specific key
//
// Time complexity: O(log(N))
func (z *ZSet) Remove(key int64) (ok bool) {
	found := z.dict[key]
	if found != nil {
		z.zsl.zslDelete(found.score, key)
		delete(z.dict, key)
		return true
	}
	return false
}

// sanitizeIndexes return start, end, and reverse flag.
// Internal method for standardizing the start and end indexes.
func (z *ZSet) sanitizeIndexes(start int, end int) (int, int, bool) {
	if start < 0 {
		start = int(z.Length()) + start + 1
	}
	if end < 0 {
		end = int(z.Length()) + end + 1
	}
	if start <= 0 {
		start = 1
	}
	if end <= 0 {
		end = 1
	}

	reverse := start > end
	if reverse { // swap start and end
		start, end = end, start
	}
	return start, end, reverse
}

func (z *ZSet) findNodeByRank(start int, remove bool) (traversed int,
	x *zskiplistNode, update [zskiplist_max_level]*zskiplistNode) {
	x = z.zsl.header

	for currentLevel := z.zsl.level - 1; currentLevel >= 0; currentLevel-- {
		for x.level[currentLevel].forward != nil &&
			traversed+int(x.level[currentLevel].span) < start {
			traversed += int(x.level[currentLevel].span)
			x = x.level[currentLevel].forward
		}
		if remove {
			update[currentLevel] = x
		} else {
			// Check if next node is the target.
			if traversed+1 == start {
				break
			}
		}
		// loop down onto the lower level.
	}
	return
}

// Get nodes within specific rank range [start, end]
// Note that the rank is 1-based integer. Rank 1 means the first node;
// Rank -1 means the last node;
//
// If start is greater than end, the returned array is in reserved order
// If remove is true, the returned nodes are removed
//
// Time complexity of this method is : O(log(N))
func (z *ZSet) GetByRankRange(start int, end int, remove bool) []*zskiplistNode {
	start, end, reverse := z.sanitizeIndexes(start, end)

	var nodes []*zskiplistNode

	traversed, x, update := z.findNodeByRank(start, remove)

	// traversed now keeps track of the rank we are currently at.
	traversed++
	x = x.level[0].forward
	for x != nil && traversed <= end {
		next := x.level[0].forward

		nodes = append(nodes, x)

		if remove {
			z.zsl.zslDeleteNode(x, update)
		}

		traversed++
		x = next
	}

	if reverse {
		for i, j := 0, len(nodes)-1; i < j; i, j = i+1, j-1 {
			nodes[i], nodes[j] = nodes[j], nodes[i]
		}
	}
	return nodes
}

// Get node by rank.
// Note that the rank is 1-based integer. Rank 1 means the first node;
// Rank -1 means the last node;
//
// If remove is true, the returned nodes are removed
// If node is not found at specific rank, nil is returned
//
// Time complexity of this method is : O(log(N))
func (z *ZSet) GetByRank(rank int, remove bool) *zskiplistNode {
	nodes := z.GetByRankRange(rank, rank, remove)

	if len(nodes) == 1 {
		return nodes[0]
	}

	return nil
}

// Get node by key
//
// If node is not found, nil is returned
// Time complexity : O(1)
func (z *ZSet) GetByKey(key int64) *zskiplistNode {
	return z.dict[key]
}

// Find the rank of the node specified by key
// Note that the rank is 1-based integer. Rank 1 means the first node
//
// If the node is not found, 0 is returned. Otherwise rank(> 0) is returned
//
// Time complexity of this method is : O(log(N))
func (z *ZSet) FindRank(key int64) int {
	var rank int = 0
	node := z.dict[key]
	if node != nil {
		x := z.zsl.header
		for i := z.zsl.level - 1; i >= 0; i-- {
			for x.level[i].forward != nil &&
				(x.level[i].forward.score < node.score ||
					(x.level[i].forward.score == node.score &&
						x.level[i].forward.key <= node.key)) {
				rank += int(x.level[i].span)
				x = x.level[i].forward
			}

			if x.key == key {
				return rank
			}
		}
	}
	return 0
}

// IterFuncByRankRange apply fn to node within specific rank range [start, end]
// or until fn return false
//
// Note that the rank is 1-based integer. Rank 1 means the first node; Rank -1 means the last node;
// If start is greater than end, apply fn in reserved order
// If fn is nil, this function return without doing anything
func (z *ZSet) IterFuncByRankRange(start int, end int, fn func(key int64, value interface{}) bool) {
	if fn == nil {
		return
	}

	start, end, reverse := z.sanitizeIndexes(start, end)
	traversed, x, _ := z.findNodeByRank(start, false)
	var nodes []*zskiplistNode

	x = x.level[0].forward
	for x != nil && traversed < end {
		next := x.level[0].forward

		if reverse {
			nodes = append(nodes, x)
		} else if !fn(x.key, x.value) {
			return
		}

		traversed++
		x = next
	}

	if reverse {
		for i := len(nodes) - 1; i >= 0; i-- {
			if !fn(nodes[i].key, nodes[i].value) {
				return
			}
		}
	}
}

// Get the element with the minimum score, nil if the set is empty.
//
// Time complexity : O(1)
func (z *ZSet) PeekMin() *zskiplistNode {
	return z.zsl.header.level[0].forward
}

// Get and remove the element with the minimum score, nil if the set is empty.
//
// Time complexity : O(log(N))
func (z *ZSet) PopMin() *zskiplistNode {
	x := z.zsl.header.level[0].forward
	if x != nil {
		z.Remove(x.key)
	}
	return x
}

// Get the element with maximum score, nil if the set is empty.
//
// Time complexity : O(1)
func (z *ZSet) PeekMax() *zskiplistNode {
	return z.zsl.tail
}

// Get and remove the element with the maximum score, nil if the set is empty.
//
// Time complexity : O(log(N))
func (z *ZSet) PopMax() *zskiplistNode {
	x := z.zsl.tail
	if x != nil {
		z.Remove(x.key)
	}
	return x
}

type GetByScoreRangeOptions struct {
	Limit        int  // limit the max nodes to return
	ExcludeStart bool // exclude start value, so it search in interval (start, end] or (start, end)
	ExcludeEnd   bool // exclude end value, so it search in interval [start, end) or (start, end)
}

// Get the nodes whose score within the specific range
//
// If options is nil, it searchs in interval [start, end] without any limit
// by default.
//
// Time complexity : O(log(N) + M) with M being the number of elements in the
// specified range.
func (z *ZSet) GetByScoreRange(start SCORE, end SCORE,
	options *GetByScoreRangeOptions) []*zskiplistNode {
	// Prepare the parameters
	var limit int = int((^uint(0)) >> 1)
	if options != nil && options.Limit > 0 {
		limit = options.Limit
	}

	excludeStart := options != nil && options.ExcludeStart
	excludeEnd := options != nil && options.ExcludeEnd
	reverse := start > end

	if reverse {
		start, end = end, start
		excludeStart, excludeEnd = excludeEnd, excludeStart
	}

	// Result set.
	var nodes []*zskiplistNode

	// Check if the set is empty and return early.
	if z.Length() == 0 {
		return nodes
	}

	if reverse { // search from end to start
		x := z.zsl.header

		// Find the ending node to start with.
		if excludeEnd { // ..start, end)
			for currentLevel := z.zsl.level - 1; currentLevel >= 0; currentLevel-- {
				for x.level[currentLevel].forward != nil &&
					// reach the node with score 'just' less than the end.
					x.level[currentLevel].forward.score < end {
					x = x.level[currentLevel].forward
				}
			}
		} else { // ..start, end]
			for currentLevel := z.zsl.level - 1; currentLevel >= 0; currentLevel-- {
				for x.level[currentLevel].forward != nil &&
					// reach the node with the score equal to or just smaller than the end.
					x.level[currentLevel].forward.score <= end {
					x = x.level[currentLevel].forward
				}
			}
		}

		// Start collecting nodes from the end towards the front until either
		// you run out of nodes or satisfy the limit.
		for x != nil && limit > 0 {
			if excludeStart { // (start, end..
				if x.score <= start {
					break
				}
			} else { // [start, end..
				if x.score < start {
					break
				}
			}

			next := x.backward

			nodes = append(nodes, x)
			limit--

			x = next
		}
	} else { // search from start to end
		x := z.zsl.header

		// Finding the starting node.
		if excludeStart { // (start, end..
			for currentLevel := z.zsl.level - 1; currentLevel >= 0; currentLevel-- {
				for x.level[currentLevel].forward != nil &&
					// We want to stop just before the target start node (to avoid
					//overshooting initial elements that match this condition.)
					x.level[currentLevel].forward.score <= start {
					x = x.level[currentLevel].forward
				}
			}
		} else { // [start, end..
			for currentLevel := z.zsl.level - 1; currentLevel >= 0; currentLevel-- {
				for x.level[currentLevel].forward != nil &&
					// We want to stop just before the target start node (to avoid
					// overshooting initial elements that match this condition.)
					x.level[currentLevel].forward.score < start {
					x = x.level[currentLevel].forward
				}
			}
		}

		// Current node is the last with score < or <= start.
		x = x.level[0].forward

		for x != nil && limit > 0 {
			if excludeEnd { // ..start, end)
				if x.score >= end {
					break
				}
			} else { // ..start, end]
				if x.score > end {
					break
				}
			}

			next := x.level[0].forward

			nodes = append(nodes, x)
			limit--

			x = next
		}
	}

	return nodes
}

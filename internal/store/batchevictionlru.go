// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package store

import (
	"container/heap"
	"math"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/internal/object"
)

// evictionItemHeap is a max-heap of evictionItems based on lastAccessed.
type evictionItemHeap []evictionItem

func (h *evictionItemHeap) Len() int { return len(*h) }

func (h *evictionItemHeap) Less(i, j int) bool {
	// For a max-heap, we want higher lastAccessed at the top.
	return (*h)[i].lastAccessed > (*h)[j].lastAccessed
}

func (h *evictionItemHeap) Swap(i, j int) { (*h)[i], (*h)[j] = (*h)[j], (*h)[i] }

func (h *evictionItemHeap) Push(x interface{}) {
	*h = append(*h, x.(evictionItem))
}

func (h *evictionItemHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

// Encapsulate heap operations to avoid interface{} in main code.
func (h *evictionItemHeap) push(item evictionItem) {
	heap.Push(h, item)
}

func (h *evictionItemHeap) pop() evictionItem {
	return heap.Pop(h).(evictionItem)
}

// PrimitiveEvictionStrategy implements batch eviction of least recently used keys
type PrimitiveEvictionStrategy struct {
	BaseEvictionStrategy
	maxKeys int
}

func NewPrimitiveEvictionStrategy(maxKeys int) *PrimitiveEvictionStrategy {
	return &PrimitiveEvictionStrategy{
		maxKeys: maxKeys,
	}
}

func (e *PrimitiveEvictionStrategy) ShouldEvict(store *Store) int {
	currentKeyCount := store.GetKeyCount()

	// Check if eviction is necessary only till the number of keys remains less than maxKeys
	if currentKeyCount < e.maxKeys {
		return 0 // No eviction needed
	}

	// Calculate target key count after eviction
	targetKeyCount := int(math.Ceil(float64(e.maxKeys) * (1 - config.EvictionRatio)))

	// Calculate the number of keys to evict to reach the target key count
	toEvict := currentKeyCount - targetKeyCount
	if toEvict < 1 {
		toEvict = 1 // Ensure at least one key is evicted if eviction is triggered
	}

	return toEvict
}

// EvictVictims deletes keys with the lowest LastAccessedAt values from the store.
func (e *PrimitiveEvictionStrategy) EvictVictims(store *Store, toEvict int) {
	if toEvict <= 0 {
		return
	}

	h := make(evictionItemHeap, 0, toEvict)
	heap.Init(&h)

	store.GetStore().All(func(k string, obj *object.Obj) bool {
		item := evictionItem{
			key:          k,
			lastAccessed: obj.LastAccessedAt,
		}
		if h.Len() < toEvict {
			h.push(item)
			return true
		}

		if item.lastAccessed < h[0].lastAccessed {
			h.pop()
			h.push(item)
		}
		return true
	})

	for h.Len() > 0 {
		item := h.pop()
		store.Del(item.key, WithDelCmd(Evict))
	}

	e.stats.recordEviction(int64(toEvict))
}

func (e *PrimitiveEvictionStrategy) OnAccess(key string, obj *object.Obj, accessType AccessType) {
	// Nothing to do for LRU batch eviction
}

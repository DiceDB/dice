// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package store

import (
	"time"

	"github.com/dicedb/dice/internal/object"
)

// EvictionStats tracks common statistics for all eviction strategies
type EvictionStats struct {
	totalEvictions     uint64
	totalKeysEvicted   uint64
	lastEvictionCount  int64
	lastEvictionTimeMs int64
}

func (s *EvictionStats) recordEviction(count int64) {
	s.totalEvictions++
	s.totalKeysEvicted += uint64(count)
	s.lastEvictionCount = count
	s.lastEvictionTimeMs = time.Now().UnixMilli()
}

// EvictionResult represents the outcome of an eviction operation
type EvictionResult struct {
	Victims map[string]*object.Obj // Keys and objects that were selected for eviction
	Count   int64                  // Number of items selected for eviction
}

// AccessType represents different types of access to a key
type AccessType int

const (
	AccessGet AccessType = iota
	AccessSet
	AccessDel
)

// evictionItem stores essential data needed for eviction decision
type evictionItem struct {
	key          string
	lastAccessed int64
}

// EvictionStrategy defines the interface for different eviction strategies
type EvictionStrategy interface {
	// ShouldEvict checks if eviction should be triggered based on the current store state
	// Returns the number of items that should be evicted, or 0 if no eviction is needed
	ShouldEvict(store *Store) int

	// EvictVictims evicts items from the store based on the eviction strategy
	EvictVictims(store *Store, toEvict int)

	// AfterEviction is called after victims have been evicted from the store
	// This allows strategies to update their internal state if needed
	// AfterEviction(result EvictionResult)

	// OnAccess is called when an item is accessed (get/set)
	// This allows strategies to update access patterns/statistics
	OnAccess(key string, obj *object.Obj, accessType AccessType)
}

// BaseEvictionStrategy provides common functionality for all eviction strategies
type BaseEvictionStrategy struct {
	stats EvictionStats
}

func (b *BaseEvictionStrategy) AfterEviction(result EvictionResult) {
	b.stats.recordEviction(result.Count)
}

func (b *BaseEvictionStrategy) GetStats() EvictionStats {
	return b.stats
}

func GetIdleTime(lastAccessedAt int64) int64 {
	return time.Now().UnixMilli() - lastAccessedAt
}

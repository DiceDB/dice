// This file is part of DiceDB.
// Copyright (C) 2024 DiceDB (dicedb.io).
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package store

import (
	"time"

	"github.com/dicedb/dice/internal/object"
	"github.com/dicedb/dice/internal/server/utils"
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
	lastAccessed uint32
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

func getCurrentClock() uint32 {
	return uint32(utils.GetCurrentTime().Unix()) & 0x00FFFFFF
}

func GetIdleTime(lastAccessedAt uint32) uint32 {
	c := getCurrentClock()
	lastAccessedAt &= 0x00FFFFFF
	if c >= lastAccessedAt {
		return c - lastAccessedAt
	}
	return (0x00FFFFFF - lastAccessedAt) + c
}

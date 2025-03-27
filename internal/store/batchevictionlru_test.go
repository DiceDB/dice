// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package store

import (
	"strconv"
	"testing"
	"time"

	"github.com/dicedb/dice/internal/object"
	"github.com/dicedb/dice/internal/server/utils"
	"github.com/stretchr/testify/assert"
)

func TestEvictVictims_BelowMaxKeys(t *testing.T) {
	eviction := NewPrimitiveEvictionStrategy(5)
	s := NewStore(nil, eviction, 0)

	// Add 3 keys (below maxKeys of 5)
	for i := 1; i <= 3; i++ {
		key := "key" + strconv.Itoa(i)
		obj := &object.Obj{
			LastAccessedAt: getCurrentClock() + uint32(i),
		}
		s.Put(key, obj)
	}

	initialKeyCount := s.GetKeyCount()
	toEvict := eviction.ShouldEvict(s)
	assert.Equal(t, 0, toEvict, "Should not evict any keys when below maxKeys")

	eviction.EvictVictims(s, toEvict)
	assert.Equal(t, initialKeyCount, s.GetKeyCount(), "No keys should be evicted when below maxKeys")
}

func TestEvictVictims_ExceedsMaxKeys(t *testing.T) {
	maxKeys := 5
	eviction := NewPrimitiveEvictionStrategy(maxKeys)
	s := NewStore(nil, eviction, 0)

	// Add 10 keys, exceeding maxKeys of 5
	for i := 1; i <= 10; i++ {
		key := "key" + strconv.Itoa(i)
		obj := &object.Obj{
			LastAccessedAt: getCurrentClock() + uint32(i),
		}

		s.Put(key, obj)
	}

	// Ensure number of keys are equal to or below maxKeys after eviction
	keyCount := s.GetKeyCount()
	assert.True(t, keyCount <= maxKeys, "Should have max or lesser number of keys remaining after eviction")
}

func TestEvictVictims_EvictsLRU(t *testing.T) {
	mockTime := &utils.MockClock{CurrTime: time.Now()}
	utils.CurrentTime = mockTime

	eviction := NewPrimitiveEvictionStrategy(10)
	s := NewStore(nil, eviction, 0)

	// Add keys with varying LastAccessedAt
	keyIDs := []int{0, 7, 1, 9, 4, 6, 5, 2, 8, 3, 10}
	for _, id := range keyIDs {
		// Ensure LastAccessedAt is unique
		key := "key" + strconv.Itoa(id)
		obj := &object.Obj{}
		mockTime.SetTime(mockTime.GetTime().Add(5 * time.Second))
		s.Put(key, obj)
	}

	// Expected to evict 4 keys with lowest LastAccessedAt, i.e. the first 4 keys added to the store
	evictedKeys := []string{"key0", "key7", "key1", "key9"} // Indices correspond to accessTimes

	remainingKeyCount := s.GetKeyCount()
	assert.Equal(t, 7, remainingKeyCount, "Should have seven, 6(Post Eviction) + 1(new key added post eviction) keys remaining after eviction")

	// Verify that evicted keys are no longer in the store
	for _, key := range evictedKeys {
		obj := s.GetNoTouch(key)
		assert.Nil(t, obj, "Key %s should have been evicted", key)
	}
}

func TestEvictVictims_IdenticalLastAccessedAt(t *testing.T) {
	currentTime := time.Now()
	mockTime := &utils.MockClock{CurrTime: currentTime}
	utils.CurrentTime = mockTime
	eviction := NewPrimitiveEvictionStrategy(10)
	s := NewStore(nil, eviction, 0)

	// Add 10 keys with identical LastAccessedAt
	for i := 0; i <= 10; i++ {
		key := "key" + strconv.Itoa(i)
		obj := &object.Obj{}
		mockTime.SetTime(currentTime) // Not needed, added explicitly for better clarity
		s.Put(key, obj)
	}

	expectedRemainingKeys := 6 // 5(Post Eviction) + 1 (key added after eviction)
	assert.Equal(t, expectedRemainingKeys, s.GetKeyCount(), "Should have evicted 5 keys")
}

func TestEvictVictims_EvictsAtLeastOne(t *testing.T) {
	eviction := NewPrimitiveEvictionStrategy(10) // 0% eviction rate
	s := NewStore(nil, eviction, 0)

	// Add 10 keys (equals maxKeys)
	for i := 0; i < 10; i++ {
		key := "key" + strconv.Itoa(i)
		obj := &object.Obj{}
		s.Put(key, obj)
	}

	toEvict := eviction.ShouldEvict(s)
	assert.Equal(t, 1, toEvict, "Should evict at least one key")
}

func TestEvictVictims_EmptyStore(t *testing.T) { // Handles Empty Store Gracefully
	eviction := NewPrimitiveEvictionStrategy(5)
	s := NewStore(nil, eviction, 0)

	toEvict := eviction.ShouldEvict(s)
	assert.Equal(t, 0, toEvict, "Should not evict any keys when store is empty")

	eviction.EvictVictims(s, toEvict)
	assert.Equal(t, 0, s.GetKeyCount(), "Store should remain empty after eviction")
}

func TestEvictVictims_LastAccessedAtUpdated(t *testing.T) {
	currentTime := time.Now()
	mockTime := &utils.MockClock{CurrTime: currentTime}
	utils.CurrentTime = mockTime
	eviction := NewPrimitiveEvictionStrategy(10)
	s := NewStore(nil, eviction, 0)

	// Add keys with initial LastAccessedAt
	for i := 1; i <= 10; i++ {
		key := "key" + strconv.Itoa(i)
		obj := &object.Obj{}
		s.Put(key, obj)
	}

	// Simulate access to some keys, updating LastAccessedAt
	accessedKeys := []string{"key2", "key3", "key4", "key5", "key6", "key7", "key8", "key10"}
	for _, key := range accessedKeys {
		mockTime.SetTime(mockTime.GetTime().Add(5 * time.Second))
		s.Get(key) // This should update LastAccessedAt
	}

	// The keys that were not accessed should be more likely to be evicted
	// Verify that verify accessed keys are still in the store
	for _, key := range accessedKeys {
		obj := s.GetNoTouch(key)
		assert.NotNil(t, obj, "Key %s should remain after eviction", key)
	}

	s.Put("key11", &object.Obj{}) // Trigger eviction

	// Verify that unaccessed keys were evicted
	unaccessedKeys := []string{"key1", "key9"}
	for _, key := range unaccessedKeys {
		obj := s.GetNoTouch(key)
		assert.Nil(t, obj, "Key %s should remain after eviction", key)
	}

	// Verify that some of the accessed keys were evicted
	numRemovedKeys := 0
	for _, key := range accessedKeys {
		obj := s.GetNoTouch(key)
		if obj == nil {
			numRemovedKeys++
		}
	}

	assert.True(t, numRemovedKeys == 2, "2 accessed key should have been evicted")
}

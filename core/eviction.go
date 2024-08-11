package core

import (
	"time"

	"github.com/dicedb/dice/config"
)

// Evicts the first key it found while iterating the map
// TODO: Make it efficient by doing thorough sampling
func evictFirst() {
	withLocks(DefaultLockIdentifier, func() {
		for keyPtr := range store {
			delByPtr(keyPtr)
			return
		}
	}, WithStoreLock())
}

// Randomly removes keys to make space for the new data added.
// The number of keys removed will be sufficient to free up at least 10% space
func evictAllkeysRandom() {
	evictCount := int64(config.EvictionRatio * float64(config.KeysLimit))
	withLocks(DefaultLockIdentifier, func() {
		// Iteration of Golang dictionary can be considered as a random
		// because it depends on the hash of the inserted key
		for keyPtr := range store {
			delByPtr(keyPtr)
			evictCount--
			if evictCount <= 0 {
				break
			}
		}
	}, WithStoreLock())
}

/*
 *  The approximated LRU algorithm
 */
func getCurrentClock() uint32 {
	return uint32(time.Now().Unix()) & 0x00FFFFFF
}

func getIdleTime(lastAccessedAt uint32) uint32 {
	c := getCurrentClock()
	if c >= lastAccessedAt {
		return c - lastAccessedAt
	}
	return (0x00FFFFFF - lastAccessedAt) + c
}

func populateEvictionPool() {
	sampleSize := 5

	withLocks(DefaultLockIdentifier, func() {
		for k := range store {
			ePool.Push(k, store[k].LastAccessedAt)
			sampleSize--
			if sampleSize == 0 {
				break
			}
		}
	}, WithStoreRLock())
}

// TODO: no need to populate everytime. should populate
// only when the number of keys to evict is less than what we have in the pool
func evictAllkeysLRU() {
	populateEvictionPool()
	evictCount := int16(config.EvictionRatio * float64(config.KeysLimit))
	for i := 0; i < int(evictCount) && len(ePool.pool) > 0; i++ {
		item := ePool.Pop()
		if item == nil {
			return
		}
		DelByPtr(item.keyPtr)
	}
}

// TODO: implement LFU
func evict() {
	switch config.EvictionStrategy {
	case "simple-first":
		evictFirst()
	case "allkeys-random":
		evictAllkeysRandom()
	case "allkeys-lru":
		evictAllkeysLRU()
	}
}

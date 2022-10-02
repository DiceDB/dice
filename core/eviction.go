package core

import (
	"time"

	"github.com/dicedb/dice/config"
)

// Evicts the first key it found while iterating the map
// TODO: Make it efficient by doing thorough sampling
func evictFirst() {
	for k := range store {
		Del(k)
		return
	}
}

// Randomly removes keys to make space for the new data added.
// The number of keys removed will be sufficient to free up least 10% space
func evictAllkeysRandom() {
	evictCount := int64(config.EvictionRatio * float64(config.KeysLimit))
	// Iteration of Golang dictionary can be considered as a random
	// because it depends on the hash of the inserted key
	for k := range store {
		Del(k)
		evictCount--
		if evictCount <= 0 {
			break
		}
	}
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
	for k := range store {
		ePool.Push(k, store[k].LastAccessedAt)
		sampleSize--
		if sampleSize == 0 {
			break
		}
	}
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
		Del(item.key)
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

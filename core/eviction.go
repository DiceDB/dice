package core

import (
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

// TODO: Make the eviction strategy configuration driven
// TODO: Support multiple eviction strategies
func evict() {
	switch config.EvictionStrategy {
	case "simple-first":
		evictFirst()
	case "allkeys-random":
		evictAllkeysRandom()
	}
}

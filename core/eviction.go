package core

import "github.com/dicedb/dice/config"

// Evicts the first key it found while iterating the map
// TODO: Make it efficient by doing thorough sampling
func evictFirst() {
	for k := range store {
		delete(store, k)
		return
	}
}

// TODO: Make the eviction strategy configuration driven
// TODO: Support multiple eviction strategies
func evict() {
	switch config.EvictionStrategy {
	case "simple-first":
		evictFirst()
	}
}

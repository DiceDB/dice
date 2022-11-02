package eviction

import (
	dbEngine "github.com/dicedb/dice/IStorageEngines"
	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/object"
	"github.com/dicedb/dice/utils"
)

// Evicts the first key it found while iterating the map
// TODO: Make it efficient by doing thorough sampling
func evictFirst(dh dbEngine.IKVStorage) {
	dh.GetStorage().Range(func (k, v interface{}) bool {
		key := k.(string)
		dh.Del(key)
		return false
	})
}

// Randomly removes keys to make space for the new data added.
// The number of keys removed will be sufficient to free up least 10% space
func evictAllkeysRandom(dh dbEngine.IKVStorage) {
	evictCount := int64(config.EvictionRatio * float64(config.KeysLimit))
	// Iteration of Golang dictionary can be considered as a random
	// because it depends on the hash of the inserted key
	dh.GetStorage().Range(func(k, v interface{}) bool {
		key := k.(string)
		dh.Del(key)
		evictCount--
		return evictCount > 0
	})
}

/*
 *  The approximated LRU algorithm
 */
func getIdleTime(lastAccessedAt uint32) uint32 {
	c := utils.GetCurrentClock()
	if c >= lastAccessedAt {
		return c - lastAccessedAt
	}
	return (0x00FFFFFF - lastAccessedAt) + c
}

func populateEvictionPool(dh dbEngine.IKVStorage) {
	sampleSize := 5

	dh.GetStorage().Range(func(k, v interface{}) bool {
		key := k.(string)
		value := v.(*object.Obj)
		ePool.Push(key, value.LastAccessedAt)
		return sampleSize != 0
	})
}

// TODO: no need to populate everytime. should populate
// only when the number of keys to evict is less than what we have in the pool
func EvictAllkeysLRU(dh dbEngine.IKVStorage) {
	populateEvictionPool(dh)
	evictCount := int16(config.EvictionRatio * float64(config.KeysLimit))
	for i := 0; i < int(evictCount) && len(ePool.pool) > 0; i++ {
		item := ePool.Pop()
		if item == nil {
			return
		}
		dh.Del(item.key)
	}
}

// TODO: implement LFU
func Evict(dh dbEngine.IKVStorage) {
	switch config.EvictionStrategy {
	case "simple-first":
		evictFirst(dh)
	case "allkeys-random":
		evictAllkeysRandom(dh)
	case "allkeys-lru":
		EvictAllkeysLRU(dh)
	}
}

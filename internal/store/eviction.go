package store

import (
	"math/rand"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/internal/object"
	"github.com/dicedb/dice/internal/server/utils"
)

// Evicts the first key it found while iterating the map
// TODO: Make it efficient by doing thorough sampling
func evictFirst(store *Store) {
	store.store.All(func(k string, obj *object.Obj) bool {
		store.delByPtr(k, WithDelCmd(Del))
		// stop after iterating over the first element
		return false
	})
}

// Randomly removes keys to make space for the new data added.
// The number of keys removed will be sufficient to free up at least 10% space
func evictAllkeysRandom(store *Store) {
	evictCount := int64(config.DiceConfig.Memory.EvictionRatio * float64(config.DiceConfig.Memory.KeysLimit))
	// Iteration of Golang dictionary can be considered as a random
	// because it depends on the hash of the inserted key
	store.store.All(func(k string, obj *object.Obj) bool {
		store.delByPtr(k, WithDelCmd(Del))
		evictCount--
		// continue if evictCount > 0
		return evictCount > 0
	})
}

/*
 *  The approximated LRU algorithm
 */
func getCurrentClock() uint32 {
	return uint32(utils.GetCurrentTime().Unix()) & 0x00FFFFFF
}

func GetLFULogCounter(lastAccessedAt uint32) uint8 {
	return uint8((lastAccessedAt & 0xFF000000) >> 24)
}

func UpdateLFULastAccessedAt(lastAccessedAt uint32) uint32 {
	currentUnixTime := getCurrentClock()
	counter := GetLFULogCounter(lastAccessedAt)

	counter = incrLogCounter(counter)
	return (uint32(counter) << 24) | currentUnixTime
}

func GetLastAccessedAt(lastAccessedAt uint32) uint32 {
	return lastAccessedAt & 0x00FFFFFF
}

func UpdateLastAccessedAt(lastAccessedAt uint32) uint32 {
	if config.DiceConfig.Memory.EvictionPolicy == config.EvictAllKeysLFU {
		return UpdateLFULastAccessedAt(lastAccessedAt)
	}
	return getCurrentClock()
}

/*
  - Similar to redis implementation of increasing access counter for a key
  - The larger the counter value, the lesser is probability of its increment in counter value
  - This counter is 8-bit number that will represent an approximate access counter of a key and will
    piggyback first 8 bits of `LastAccessedAt` field of Dice Object
*/
func incrLogCounter(counter uint8) uint8 {
	if counter == 255 {
		return 255
	}
	randomFactor := rand.Float32() //nolint:gosec
	approxFactor := 1.0 / float32(counter*uint8(config.DiceConfig.Memory.LFULogFactor)+1)
	if approxFactor > randomFactor {
		counter++
	}
	return counter
}

func GetIdleTime(lastAccessedAt uint32) uint32 {
	c := getCurrentClock()
	lastAccessedAt &= 0x00FFFFFF
	if c >= lastAccessedAt {
		return c - lastAccessedAt
	}
	return (0x00FFFFFF - lastAccessedAt) + c
}

func PopulateEvictionPool(store *Store) {
	sampleSize := 5
	// TODO: if we already have obj, why do we need to
	// look up in store.store again?
	store.store.All(func(k string, obj *object.Obj) bool {
		v, ok := store.store.Get(k)
		if ok {
			EPool.Push(k, v.LastAccessedAt)
			sampleSize--
		}
		// continue if sample size > 0
		// stop as soon as it hits 0
		return sampleSize > 0
	})
}

// EvictAllkeysLRUOrLFU evicts keys based on LRU or LFU policy.
// TODO: no need to populate every time. should populate only when the number of keys to evict is less than what we have in the pool
func EvictAllkeysLRUOrLFU(store *Store) {
	PopulateEvictionPool(store)
	evictCount := int16(config.DiceConfig.Memory.EvictionRatio * float64(config.DiceConfig.Memory.KeysLimit))

	for i := 0; i < int(evictCount) && len(EPool.pool) > 0; i++ {
		item := EPool.Pop()
		if item == nil {
			return
		}
		store.DelByPtr(item.keyPtr, WithDelCmd(Del))
	}
}

func (store *Store) evict() {
	switch config.DiceConfig.Memory.EvictionPolicy {
	case config.EvictSimpleFirst:
		evictFirst(store)
	case config.EvictAllKeysRandom:
		evictAllkeysRandom(store)
	case config.EvictAllKeysLRU:
		EvictAllkeysLRUOrLFU(store)
	case config.EvictAllKeysLFU:
		EvictAllkeysLRUOrLFU(store)
	}
}

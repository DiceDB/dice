package store

import (
	"math/rand"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/internal/server/utils"
)

// Evicts the first key it found while iterating the map
// TODO: Make it efficient by doing thorough sampling
func evictFirst(store *Store) {
	withLocks(func() {
		store.store.All(func(k string, obj *Obj) bool {
			store.delByPtr(k)
			// stop after iterating over the first element
			return false
		})
	}, store, WithStoreLock())
}

// Randomly removes keys to make space for the new data added.
// The number of keys removed will be sufficient to free up at least 10% space
func evictAllkeysRandom(store *Store) {
	evictCount := int64(config.EvictionRatio * float64(config.KeysLimit))
	withLocks(func() {
		// Iteration of Golang dictionary can be considered as a random
		// because it depends on the hash of the inserted key
		store.store.All(func(k string, obj *Obj) bool {
			store.delByPtr(k)
			evictCount--
			// continue if evictCount > 0
			return evictCount > 0
		})
	}, store, WithStoreLock())
}

/*
 *  The approximated LRU algorithm
 */
func getCurrentClock() uint32 {
	return uint32(utils.GetCurrentTime().Unix()) & 0x00FFFFFF
}

func getLFULogCounter(lastAccessedAt uint32) uint8 {
	return uint8(lastAccessedAt & 0xFF000000)
}

func updateLFULastAccessedAt(lastAccessedAt uint32) uint32 {
	currentUnixTime := getCurrentClock()
	counter := getLFULogCounter(lastAccessedAt)

	counter = incrLogCounter(counter)
	return (uint32(counter) & 0xFF000000) | currentUnixTime
}

func getLastAccessedAt(lastAccessedAt uint32) uint32 {
	if config.EvictionStrategy == config.AllKeysLFU {
		return updateLFULastAccessedAt(lastAccessedAt)
	}
	return getCurrentClock()
}

/*
  - Similar to redis implementation of increasing access counter for a key
  - The larger the counter value, the lesser is probability of its increment in counter value
  - This counter is 8 bit number that will represent an approximate access counter of a key and will
    piggyback first 8 bits of `LastAccessedAt` field of Dice Object
*/
func incrLogCounter(counter uint8) uint8 {
	if counter == 255 {
		return 255
	}
	randomFactor := rand.Float32() //nolint:gosec
	approxFactor := 1.0 / float32(counter*config.LFULogFactor+1)
	if approxFactor > randomFactor {
		counter++
	}
	return counter
}

func GetIdleTime(lastAccessedAt uint32) uint32 {
	c := getCurrentClock()
	if c >= lastAccessedAt {
		return c - lastAccessedAt
	}
	return (0x00FFFFFF - lastAccessedAt) + c
}

func populateEvictionPool(store *Store) {
	sampleSize := 5

	// TODO: if we already have obj, why do we need to
	// look up in store.store again?
	withLocks(func() {
		store.store.All(func(k string, obj *Obj) bool {
			v, ok := store.store.Get(k)
			if ok {
				ePool.Push(k, v.LastAccessedAt)
				sampleSize--
			}
			// continue if sample size > 0
			// stop as soon as it hits 0
			return sampleSize > 0
		})
	}, store, WithStoreRLock())
}

// TODO: no need to populate everytime. should populate
// only when the number of keys to evict is less than what we have in the pool
func EvictAllkeysLRUOrLFU(store *Store) {
	populateEvictionPool(store)
	evictCount := int16(config.EvictionRatio * float64(config.KeysLimit))
	for i := 0; i < int(evictCount) && len(ePool.pool) > 0; i++ {
		item := ePool.Pop()
		if item == nil {
			return
		}
		store.DelByPtr(item.keyPtr)
	}
}

func (store *Store) evict() {
	switch config.EvictionStrategy {
	case config.SimpleFirst:
		evictFirst(store)
	case config.AllKeysRandom:
		evictAllkeysRandom(store)
	case config.AllKeysLRU:
		EvictAllkeysLRUOrLFU(store)
	case config.AllKeysLFU:
		EvictAllkeysLRUOrLFU(store)
	}
}

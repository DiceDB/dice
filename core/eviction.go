package core

import (
	"time"
  "math/rand"
  "fmt"

	"github.com/dicedb/dice/config"
)

const (
  LFU_LOG_FACTOR = 10
)

// Evicts the first key it found while iterating the map
// TODO: Make it efficient by doing thorough sampling
func evictFirst() {
	withLocks(func() {
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
	withLocks(func() {
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

func getLFULogCounter(lastAccessedAt uint32) uint8 {
  return uint8(lastAccessedAt & 0xFF000000)
}

func updateLFULastAccessedAt(lastAccessedAt uint32) uint32 {
  currentUnixTime := getCurrentClock()
  counter := getLFULogCounter(lastAccessedAt)

  counter = incrLogCounter(counter)
  return (uint32(counter) & 0xFF000000) | currentUnixTime
}

/*
  - Similar to redis implementation of increasing access counter for a key
  - The larger the counter value, the lesser is probability of its increment in counter value
  - This counter is 8 bit number that will represent an approximate access counter of a key and will 
    piggyback first 8 bits of `LastAccessedAt` field of Dice Object
*/
func incrLogCounter(counter uint8) uint8 {
  if (counter == 255) {
    return 255
  }
  randomFactor := rand.Float64()
  approxFactor := 1.0 / float64(counter * LFU_LOG_FACTOR + 1) 
  if (approxFactor > randomFactor) {
    counter++
  }
  return counter
}

func getIdleTime(lastAccessedAt uint32) uint32 {
	c := getCurrentClock()
	if c >= lastAccessedAt {
		return c - lastAccessedAt
	}
	return (0x00FFFFFF - lastAccessedAt) + c
}

func populateEvictionPool() {
  fmt.Println("populateEvictionPool method", len(store))
	sampleSize := 2

	withLocks(func() {
    fmt.Println("INSIDE LOCK")
		for k := range store {
      fmt.Println("KEY: ", k)
			ePool.Push(k, store[k].LastAccessedAt)
      fmt.Println("sample size", sampleSize)
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

func evictAllkeysLFU() {
  fmt.Println("evictAllKeysLFU called")
  populateEvictionPool()
  fmt.Println("populated eviction pool", len(ePool.pool))
	evictCount := int16(config.EvictionRatio * float64(config.KeysLimit))
  fmt.Println("evict count", evictCount)
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
	case config.SIMPLE_FIRST:
		evictFirst()
  case config.ALL_KEYS_RANDOM:
		evictAllkeysRandom()
  case config.ALL_KEYS_LRU:
		evictAllkeysLRU()
  case config.ALL_KEYS_LFU:
    evictAllkeysLFU()
	}
}

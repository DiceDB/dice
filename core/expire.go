package core

import (
	"time"
	"unsafe"
)

func hasExpired(obj *Obj, store *Store) bool {
	exp, ok := store.expires[obj]
	if !ok {
		return false
	}
	return exp <= uint64(time.Now().UnixMilli())
}

func getExpiry(obj *Obj, store *Store) (uint64, bool) {
	exp, ok := store.expires[obj]
	return exp, ok
}

// TODO: Optimize
//   - Sampling
//   - Unnecessary iteration
func expireSample(store *Store) float32 {
	var limit int = 20
	var expiredCount int = 0
	var keysToDelete []unsafe.Pointer

	store.storeMutex.RLock()
	// Collect keys to be deleted
	for keyPtr, obj := range store.store {
		// once we iterated to 20 keys that have some expiration set
		// we break the loop
		if limit == 0 {
			break
		}
		limit--
		if hasExpired(obj, store) {
			keysToDelete = append(keysToDelete, keyPtr)
			expiredCount++
		}
	}
	store.storeMutex.RUnlock()

	// Delete the keys outside the read lock
	for _, keyPtr := range keysToDelete {
		store.DelByPtr(keyPtr)
	}

	return float32(expiredCount) / float32(20.0)
}

// Deletes all the expired keys - the active way
// Sampling approach: https://redis.io/commands/expire/
func DeleteExpiredKeys(store *Store) {
	for {
		frac := expireSample(store)
		// if the sample had less than 25% keys expired
		// we break the loop.
		if frac < 0.25 {
			break
		}
	}
}

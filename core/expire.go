package core

import (
	"github.com/dicedb/dice/server/utils"
)

func hasExpired(obj *Obj, store *Store) bool {
	exp, ok := store.expires.Get(obj)
	if !ok {
		return false
	}
	return exp <= uint64(utils.GetCurrentTime().UnixMilli())
}

func getExpiry(obj *Obj, store *Store) (uint64, bool) {
	exp, ok := store.expires.Get(obj)
	return exp, ok
}

func delExpiry(obj *Obj, store *Store) {
	store.expires.Delete(obj)
}

// TODO: Optimize
//   - Sampling
//   - Unnecessary iteration
func expireSample(store *Store) float32 {
	var limit int = 20
	var expiredCount int = 0
	var keysToDelete []string

	withLocks(func() {
		// Collect keys to be deleted
		store.store.All(func(keyPtr string, obj *Obj) bool {
			limit--
			if hasExpired(obj, store) {
				keysToDelete = append(keysToDelete, keyPtr)
				expiredCount++
			}
			// once we iterated to 20 keys that have some expiration set
			// we break the loop
			return limit >= 0
		})
	}, store, WithStoreRLock())

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

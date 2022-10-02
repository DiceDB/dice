package core

import (
	"time"
)

func hasExpired(obj *Obj) bool {
	exp, ok := expires[obj]
	if !ok {
		return false
	}
	return exp <= uint64(time.Now().UnixMilli())
}

func getExpiry(obj *Obj) (uint64, bool) {
	exp, ok := expires[obj]
	return exp, ok
}

// TODO: Optimize
//   - Sampling
//   - Unnecessary iteration
func expireSample() float32 {
	var limit int = 20
	var expiredCount int = 0

	// assuming iteration of golang hash table in randomized
	for key, obj := range store {
		limit--
		if hasExpired(obj) {
			Del(key)
			expiredCount++
		}

		// once we iterated to 20 keys that have some expiration set
		// we break the loop
		if limit == 0 {
			break
		}
	}

	return float32(expiredCount) / float32(20.0)
}

// Deletes all the expired keys - the active way
// Sampling approach: https://redis.io/commands/expire/
func DeleteExpiredKeys() {
	for {
		frac := expireSample()
		// if the sample had less than 25% keys expired
		// we break the loop.
		if frac < 0.25 {
			break
		}
	}
}

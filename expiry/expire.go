package expiry

import (
	dbEngine "github.com/dicedb/dice/IStorageEngines"
	"github.com/dicedb/dice/object"
)

func expireSample(dh dbEngine.IKVStorage) float32 {
	var limit int = 20
	var expiredCount int = 0
	dh.GetStorage().Range(func(k, v interface{}) bool {
		key := k.(string)
		value := v.(*object.Obj)
		// fmt.Printf("The key is : %v and the value is %v\n", key, *value)
		limit--
		if object.GetDiceExpiryStore().HasExpired(value) {
			dh.Del(key)
			expiredCount++
		}
		if limit == 0 {
			return true
		}
		return true
	})
	return float32(expiredCount) / float32(limit)
}

// Deletes all the expired keys - the active way
// Sampling approach: https://redis.io/commands/expire/
func DeleteExpiredKeys(dh dbEngine.IKVStorage) {
	for {
		frac := expireSample(dh)
		// if the sample had less than 25% keys expired
		// we break the loop.
		if frac < 0.25 {
			break
		}
	}
}

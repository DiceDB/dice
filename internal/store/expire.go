package store

import (
	"strings"

	diceerrors "github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/server/utils"

	ds "github.com/dicedb/dice/internal/datastructures"
)

const (
	NX string = "NX"
	XX string = "XX"
	GT string = "GT"
	LT string = "LT"
)

func hasExpired[T ds.DSInterface](obj *T, store *Store[T]) bool {
	exp, ok := store.expires.Get(obj)
	if !ok {
		return false
	}
	return exp <= uint64(utils.GetCurrentTime().UnixMilli())
}

func GetExpiry[T ds.DSInterface](obj *T, store *Store[T]) (uint64, bool) {
	exp, ok := store.expires.Get(obj)
	return exp, ok
}

func DelExpiry[T ds.DSInterface](obj *T, store *Store[T]) {
	store.expires.Delete(obj)
}

// TODO: Optimize
//   - Sampling
//   - Unnecessary iteration
func expireSample[T ds.DSInterface](store *Store[T]) float32 {
	var limit = 20
	var expiredCount = 0
	var keysToDelete []string

	// Collect keys to be deleted
	store.store.All(func(keyPtr string, obj *T) bool {
		limit--
		if hasExpired(&obj, store) {
			keysToDelete = append(keysToDelete, keyPtr)
			expiredCount++
		}
		// once we iterated to 20 keys that have some expiration set
		// we break the loop
		return limit >= 0
	})

	// Delete the keys outside the read lock
	for _, keyPtr := range keysToDelete {
		store.DelByPtr(keyPtr, WithDelCmd(Del))
	}

	return float32(expiredCount) / float32(20.0)
}

// DeleteExpiredKeys deletes all the expired keys - the active way
// Sampling approach: https://redis.io/commands/expire/
func DeleteExpiredKeys[T ds.DSInterface](store *Store[T]) {
	for {
		frac := expireSample(store)
		// if the sample had less than 25% keys expired
		// we break the loop.
		if frac < 0.25 {
			break
		}
	}
}

// NX: Set the expiration only if the key does not already have an expiration time.
// XX: Set the expiration only if the key already has an expiration time.
// GT: Set the expiration only if the new expiration time is greater than the current one.
// LT: Set the expiration only if the new expiration time is less than the current one.
// Returns Boolean True and error nil if expiry was set on the key successfully.
// Returns Boolean False and error nil if conditions didn't met.
// Returns Boolean False and error not-nil if invalid combination of subCommands or if subCommand is invalid
func EvaluateAndSetExpiry[T ds.DSInterface](subCommands []string, newExpiry int64, key string,
	store *Store[T]) (shouldSetExpiry bool, err error) {
	var newExpInMilli = newExpiry * 1000
	var prevExpiry *uint64 = nil
	var nxCmd, xxCmd, gtCmd, ltCmd bool

	obj := store.Get(key)
	//  key doesn't exist
	if obj == nil {
		return false, nil
	}
	shouldSetExpiry = true
	// if no condition exists
	if len(subCommands) == 0 {
		store.SetUnixTimeExpiry(obj, newExpiry)
		return shouldSetExpiry, nil
	}

	expireTime, ok := GetExpiry(obj, store)
	if ok {
		prevExpiry = &expireTime
	}

	for i := range subCommands {
		subCommand := strings.ToUpper(subCommands[i])

		switch subCommand {
		case NX:
			nxCmd = true
			if prevExpiry != nil {
				shouldSetExpiry = false
			}
		case XX:
			xxCmd = true
			if prevExpiry == nil {
				shouldSetExpiry = false
			}
		case GT:
			gtCmd = true
			if prevExpiry == nil || *prevExpiry > uint64(newExpInMilli) {
				shouldSetExpiry = false
			}
		case LT:
			ltCmd = true
			if prevExpiry != nil && *prevExpiry < uint64(newExpInMilli) {
				shouldSetExpiry = false
			}
		default:
			return false, diceerrors.ErrGeneral("Unsupported option " + subCommands[i])
		}
	}

	if !nxCmd && gtCmd && ltCmd {
		return false, diceerrors.ErrGeneral("GT and LT options at the same time are not compatible")
	}

	if nxCmd && (xxCmd || gtCmd || ltCmd) {
		return false, diceerrors.ErrGeneral("NX and XX," +
			" GT or LT options at the same time are not compatible")
	}

	if shouldSetExpiry {
		store.SetUnixTimeExpiry(obj, newExpiry)
	}
	return shouldSetExpiry, nil
}

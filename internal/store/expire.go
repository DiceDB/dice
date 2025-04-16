// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package store

import (
	"strings"
	"time"

	diceerrors "github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/object"
)

const (
	NX string = "NX"
	XX string = "XX"
	GT string = "GT"
	LT string = "LT"
)

func hasExpired(obj *object.Obj, store *Store) bool {
	exp, ok := store.expires.Get(obj)
	if !ok {
		return false
	}
	return exp <= time.Now().UnixMilli()
}

func GetExpiry(obj *object.Obj, store *Store) (int64, bool) {
	exp, ok := store.expires.Get(obj)
	return exp, ok
}

func DelExpiry(obj *object.Obj, store *Store) {
	store.expires.Delete(obj)
}

// TODO: Optimize
func deleteAllExpiredKeys(store *Store) {
	store.store.All(func(keyPtr string, obj *object.Obj) bool {
		if hasExpired(obj, store) {
			store.DelByPtr(keyPtr, WithDelCmd(Del))
		}
		return true
	})
}

// DeleteExpiredKeys deletes all the expired keys - the active way
func DeleteExpiredKeys(store *Store) {
	deleteAllExpiredKeys(store)
}

// NX: Set the expiration only if the key does not already have an expiration time.
// XX: Set the expiration only if the key already has an expiration time.
// GT: Set the expiration only if the new expiration time is greater than the current one.
// LT: Set the expiration only if the new expiration time is less than the current one.
// Returns Boolean True and error nil if expiry was set on the key successfully.
// Returns Boolean False and error nil if conditions didn't met.
// Returns Boolean False and error not-nil if invalid combination of subCommands or if subCommand is invalid
func EvaluateAndSetExpiry(subCommands []string, newExpiryAbsMillis int64, key string,
	store *Store) (shouldSetExpiry bool, err error) {
	var newExpInMilli = newExpiryAbsMillis
	var prevExpiry *int64 = nil
	var nxCmd, xxCmd, gtCmd, ltCmd bool

	obj := store.Get(key)
	//  key doesn't exist
	if obj == nil {
		return false, nil
	}
	shouldSetExpiry = false

	// If no sub-command is provided, set the expiry
	if len(subCommands) == 0 {
		store.SetUnixTimeExpiry(obj, newExpiryAbsMillis)
		return true, nil
	}

	// Get the previous expiry time
	expireTime, ok := GetExpiry(obj, store)
	if ok {
		prevExpiry = &expireTime
	}

	for i := range subCommands {
		subCommand := strings.ToUpper(subCommands[i])

		switch subCommand {
		case NX:
			nxCmd = true

			// Set the expiration only if the key does not already have an expiration time.
			if prevExpiry == nil {
				shouldSetExpiry = true
			}
		case XX:
			xxCmd = true

			// Set the expiration only if the key already has an expiration time.
			if prevExpiry != nil {
				shouldSetExpiry = true
			}
		case GT:
			gtCmd = true

			// Set the expiration only if the new expiration time is greater than the current one.
			if prevExpiry != nil && newExpInMilli > *prevExpiry {
				shouldSetExpiry = true
			}
		case LT:
			ltCmd = true

			// Set the expiration only if the new expiration time is less than the current one.
			if prevExpiry != nil && newExpInMilli < *prevExpiry {
				shouldSetExpiry = true
			}
		default:
			return false, diceerrors.ErrGeneral("unsupported option " + subCommands[i])
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
		store.SetUnixTimeExpiry(obj, newExpiryAbsMillis)
	}
	return shouldSetExpiry, nil
}

// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package eval

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"strconv"

	diceerrors "github.com/dicedb/dice/internal/errors"
	dstore "github.com/dicedb/dice/internal/store"
)

type HashMap map[string]string

func (h HashMap) Get(k string) (*string, bool) {
	value, ok := h[k]
	if !ok {
		return nil, false
	}
	return &value, true
}

func (h HashMap) Set(k, v string) (*string, bool) {
	value, ok := h[k]
	if ok {
		oldValue := value
		h[k] = v
		return &oldValue, true
	}

	h[k] = v
	return nil, false
}

func hashMapBuilder(keyValuePairs []string, currentHashMap HashMap) (HashMap, int64, error) {
	var hmap HashMap
	var numKeysNewlySet int64

	if currentHashMap == nil {
		hmap = make(HashMap)
	} else {
		hmap = currentHashMap
	}

	iter := 0
	argLength := len(keyValuePairs)

	for iter <= argLength-1 {
		if iter >= argLength-1 || iter+1 > argLength-1 {
			return hmap, -1, diceerrors.ErrWrongArgumentCount("HSET")
		}

		k := keyValuePairs[iter]
		v := keyValuePairs[iter+1]

		_, present := hmap.Set(k, v)
		if !present {
			numKeysNewlySet++
		}
		iter += 2
	}

	return hmap, numKeysNewlySet, nil
}

func getValueFromHashMap(key, field string, store *dstore.Store) *EvalResponse {
	obj := store.Get(key)
	if obj == nil {
		return &EvalResponse{
			Result: NIL,
			Error:  nil,
		}
	}

	hashMap, ok := obj.Value.(HashMap)
	if !ok {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	val, present := hashMap.Get(field)
	if !present {
		return &EvalResponse{
			Result: NIL,
			Error:  nil,
		}
	}

	return &EvalResponse{
		Result: *val,
		Error:  nil,
	}
}

func (h HashMap) incrementValue(field string, increment int64) (int64, error) {
	val, ok := h[field]
	if !ok {
		h[field] = fmt.Sprintf("%v", increment)
		return increment, nil
	}

	i, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return -1, diceerrors.NewErr(diceerrors.HashValueNotIntegerErr)
	}

	if (i > 0 && increment > 0 && i > math.MaxInt64-increment) || (i < 0 && increment < 0 && i < math.MinInt64-increment) {
		return -1, diceerrors.NewErr(diceerrors.IncrDecrOverflowErr)
	}

	total := i + increment
	h[field] = fmt.Sprintf("%v", total)

	return total, nil
}

func (h HashMap) incrementFloatValue(field string, incr float64) (string, error) {
	val, ok := h[field]
	if !ok {
		h[field] = fmt.Sprintf("%v", incr)
		strValue := formatFloat(incr, false)
		return strValue, nil
	}

	i, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return "-1", diceerrors.NewErr(diceerrors.IntOrFloatErr)
	}

	if math.IsInf(i+incr, 1) || math.IsInf(i+incr, -1) {
		return "-1", diceerrors.NewErr(diceerrors.IncrDecrOverflowErr)
	}

	total := i + incr
	strValue := formatFloat(total, false)
	h[field] = fmt.Sprintf("%v", total)

	return strValue, nil
}

// selectRandomFields returns random fields from a hashmap.
func selectRandomFields(hashMap HashMap, count int, withValues bool) *EvalResponse {
	keys := make([]string, 0, len(hashMap))
	for k := range hashMap {
		keys = append(keys, k)
	}

	var results []string
	resultSet := make(map[string]struct{})

	abs := func(x int) int {
		if x < 0 {
			return -x
		}
		return x
	}

	for i := 0; i < abs(count); i++ {
		if count > 0 && len(resultSet) == len(keys) {
			break
		}

		randomIndex, _ := rand.Int(rand.Reader, big.NewInt(int64(len(keys))))
		randomField := keys[randomIndex.Int64()]

		if count > 0 {
			if _, exists := resultSet[randomField]; exists {
				i--
				continue
			}
			resultSet[randomField] = struct{}{}
		}

		results = append(results, randomField)
		if withValues {
			results = append(results, hashMap[randomField])
		}
	}

	return &EvalResponse{
		Result: results,
		Error:  nil,
	}
}

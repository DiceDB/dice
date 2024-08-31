package core

import (
	"fmt"

	"github.com/cockroachdb/swiss"
	"github.com/dicedb/dice/core/diceerrors"
)

type HashMap struct {
	Map *swiss.Map[string, string]
}

func (h *HashMap) Get(k string) (string, bool) {
	value, ok := h.Map.Get(k)
	if !ok {
		return "", false
	}
	return value, true
}

func (h *HashMap) Set(k, v string) (string, bool) {
	value, present := h.Get(k)
	var returnStr string = ""

	if present {
		returnStr = value
	}

	h.Map.Put(k, v)

	return returnStr, present
}

func hashMapBuilder(keyValuePairs []string, currentHashMap *HashMap) (*HashMap, int64, error) {
	var hmap *HashMap
	var numKeysNewlySet int64

	if currentHashMap == nil {
		hmap = &HashMap{
			Map: swiss.New[string, string](0),
		}
	} else {
		hmap = currentHashMap
	}

	iter := 0
	argLength := len(keyValuePairs)

	for iter <= argLength-1 {
		if iter >= argLength-1 || iter+1 > argLength-1 {
			return hmap, -1, diceerrors.NewErr(fmt.Sprintf(diceerrors.ArityErr, "HSET"))
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

//nolint:unused
func getValueFromHashMap(key, field string, store *Store) ([]byte, error) {
	var value string

	obj := store.Get(key)

	if obj == nil {
		return RespNIL, nil
	}

	switch currentVal := obj.Value.(type) {
	case *HashMap:
		val, present := currentVal.Get(field)
		if !present {
			return RespNIL, nil
		}
		value = val
	default:
		return nil, diceerrors.NewErr(diceerrors.WrongTypeErr)
	}

	return []byte(value), nil
}

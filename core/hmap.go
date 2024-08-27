package core

import (
	"errors"
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
			return hmap, -1, errors.New("mismatch in number of fields and values provided to HSET")
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

func getValueFromHashMap(key, field string, store *Store) ([]byte, error) {
	var value string

	obj := store.Get(key)

	if obj == nil {
		return RespNIL, nil
	}

	switch currentVal := obj.Value.(type) {
	case HashMap:
		val, present := currentVal.Get(field)
		if !present {
			return RespNIL, nil
		}
		value = *val
	default:
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	// return Encode(value, false), nil
	return []byte(value), nil
}

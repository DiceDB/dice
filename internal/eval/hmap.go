package eval

import (
	"fmt"

	"github.com/dicedb/dice/internal/clientio"
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
func getValueFromHashMap(key, field string, store *dstore.Store) ([]byte, error) {
	var value string

	obj := store.Get(key)

	if obj == nil {
		return clientio.RespNIL, nil
	}

	switch currentVal := obj.Value.(type) {
	case HashMap:
		val, present := currentVal.Get(field)
		if !present {
			return clientio.RespNIL, nil
		}
		value = *val
	default:
		return nil, diceerrors.NewErr(diceerrors.WrongTypeErr)
	}

	return []byte(value), nil
}

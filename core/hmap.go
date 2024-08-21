package core

import (
	"errors"

	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type HMap = *orderedmap.OrderedMap[string, string]

func hashMapBuilder(args []string, currentHashMap HMap) (HMap, int64, error) {
	var hmap HMap
	var numKeys int64

	if currentHashMap == nil {
		hmap = orderedmap.New[string, string]()
	} else {
		hmap = currentHashMap
	}

	iter := 1 // NOTE: to move past args[0] (aka key) within args

	for iter <= len(args)-1 {
		if iter < len(args)-1 && iter+1 <= len(args)-1 {
			k := args[iter]
			v := args[iter+1]

			_, present := hmap.Set(k, v)
			if !present {
				numKeys++
			}
			iter += 2
		} else {
			return hmap, -1, errors.New("ERR mismatch in number of fields and values provided to HSET")
		}
	}

	return hmap, numKeys, nil
}

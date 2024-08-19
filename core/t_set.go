package core

import (
	"errors"
	"sort"
)

func evalSADD(args []string) []byte {
	if len(args) < 2 {
		return Encode(errors.New("ERR wrong number of arguments for 'SADD' command"), false)
	}
	key := args[0]

	// Get the set object from the store.
	obj := Get(key)

	var count int = 0
	if obj == nil {
		var exDurationMs int64 = -1
		var keepttl bool = false
		// If the object does not exist, create a new set object.
		value := make(map[string]bool)

		// Add the value to the set.
		for i := 1; i < len(args); i++ {
			value[args[i]] = true
			count++
		}

		// Create a new object.
		Put(key, NewObj(value, exDurationMs, ObjTypeSet, ObjEncodingHT), WithKeepTTL(keepttl))
	} else {
		// If the object exists, check if it is a set object.
		if obj.TypeEncoding != ObjTypeSet|ObjEncodingHT {
			return Encode(errors.New("WRONGTYPE Operation against a key holding the wrong kind of value"), false)
		}

		// Get the set object.
		set := obj.Value.(map[string]bool)

		// Add the value to the set.
		for i := 1; i < len(args); i++ {
			if _, ok := set[args[i]]; !ok {
				set[args[i]] = true
				count++
			}
		}
	}
	return Encode(count, false)
}

func evalSMEMBERS(args []string) []byte {
	if len(args) != 1 {
		return Encode(errors.New("ERR wrong number of arguments for 'SMEMBERS' command"), false)
	}
	key := args[0]

	// Get the set object from the store.
	obj := Get(key)

	if obj == nil {
		return Encode([]string{}, false)
	}

	// If the object exists, check if it is a set object.
	if obj.TypeEncoding != ObjTypeSet|ObjEncodingHT {
		return Encode(errors.New("WRONGTYPE Operation against a key holding the wrong kind of value"), false)
	}

	// Get the set object.
	set := obj.Value.(map[string]bool)

	// Get the members of the set.
	var members []string
	for k, flag := range set {
		if flag {
			members = append(members, k)
		}
	}

	return Encode(members, false)
}

func evalSREM(args []string) []byte {
	if len(args) < 2 {
		return Encode(errors.New("ERR wrong number of arguments for 'SREM' command"), false)
	}
	key := args[0]

	// Get the set object from the store.
	obj := Get(key)

	var count int = 0
	if obj == nil {
		return Encode(count, false)
	}
	if obj.TypeEncoding != ObjTypeSet|ObjEncodingHT {
		return Encode(errors.New("WRONGTYPE Operation against a key holding the wrong kind of value"), false)
	}

	// Get the set object.
	set := obj.Value.(map[string]bool)
	for i := 1; i < len(args); i++ {
		if set[args[i]] {
			delete(set, args[i])
			count++
		}
	}
	return Encode(count, false)
}

func evalSCARD(args []string) []byte {
	if len(args) != 1 {
		return Encode(errors.New("ERR wrong number of arguments for 'SCARD' command"), false)
	}

	key := args[0]

	// Get the set object from the store.
	obj := Get(key)

	if obj == nil {
		return Encode(0, false)
	}

	// If the object exists, check if it is a set object.
	if obj.TypeEncoding != ObjTypeSet|ObjEncodingHT {
		return Encode(errors.New("WRONGTYPE Operation against a key holding the wrong kind of value"), false)
	}

	// Get the set object.
	set := obj.Value.(map[string]bool)
	var count int = 0
	for k, flag := range set {
		if !flag {
			delete(set, k)
		} else {
			count++
		}
	}
	return Encode(count, false)
}

func evalSDIFF(args []string) []byte {
	if len(args) < 1 {
		return Encode(errors.New("ERR wrong number of arguments for 'SDIFF' command"), false)
	}

	srcKey := args[0]
	obj := Get(srcKey)

	srcSet := make(map[string]bool)

	// Get the set object from the store.
	// store the count as the number of elements in the first set
	// we decrement the count as we find the elements in the other sets
	// if the count is 0, we skip further sets but still get them from
	// the store to check if they are set objects and update their last accessed time

	var count int = 0
	if obj != nil {
		if obj.TypeEncoding != ObjTypeSet|ObjEncodingHT {
			return Encode(errors.New("WRONGTYPE Operation against a key holding the wrong kind of value"), false)
		}
		// create a deep copy of the set object
		srcSet = make(map[string]bool)
		for k, flag := range obj.Value.(map[string]bool) {
			if flag {
				srcSet[k] = flag
				count++
			}
		}
	}

	for i := 1; i < len(args); i++ {
		// Get the set object from the store.
		obj := Get(args[i])

		if obj == nil {
			continue
		}

		// If the object exists, check if it is a set object.
		if obj.TypeEncoding != ObjTypeSet|ObjEncodingHT {
			return Encode(errors.New("WRONGTYPE Operation against a key holding the wrong kind of value"), false)
		}

		// only if the count is greater than 0, we need to check the other sets
		if count > 0 {
			// Get the set object.
			set := obj.Value.(map[string]bool)
			for k, flag := range set {
				if flag && srcSet[k] {
					srcSet[k] = false
					count--
				}
			}
		}
	}

	if count == 0 {
		return Encode([]string{}, false)
	}

	// Get the members of the set.
	var members []string
	for k, flag := range srcSet {
		if flag {
			members = append(members, k)
		}
	}

	return Encode(members, false)
}

func evalSINTER(args []string) []byte {
	if len(args) < 2 {
		return Encode(errors.New("ERR wrong number of arguments for 'SINTER' command"), false)
	}

	sets := make([]map[string]bool, len(args))

	var empty int = 0

	for i := 0; i < len(args); i++ {
		// Get the set object from the store.
		obj := Get(args[i])

		if obj == nil {
			empty++
			continue
		}

		// If the object exists, check if it is a set object.
		if obj.TypeEncoding != ObjTypeSet|ObjEncodingHT {
			return Encode(errors.New("WRONGTYPE Operation against a key holding the wrong kind of value"), false)
		}

		// Get the set object.
		set := obj.Value.(map[string]bool)
		sets[i] = set
	}

	if empty > 0 {
		return Encode([]string{}, false)
	}

	// sort the sets by the number of elements in the set
	// we will iterate over the smallest set
	// and check if the element is present in all the other sets
	sort.Slice(sets, func(i, j int) bool {
		return len(sets[i]) < len(sets[j])
	})

	count := 0
	resultSet := make(map[string]bool)
	var members []string

	// init the result set with the first set
	// store the number of elements in the first set in count
	// we will decrement the count if we do not find the elements in the other sets
	for k, flag := range sets[0] {
		if !flag {
			continue
		}
		count++
		resultSet[k] = true
	}

	for i := 1; i < len(sets); i++ {
		if count == 0 {
			break
		}
		for k, flag := range resultSet {
			if !flag {
				continue
			}
			if !sets[i][k] {
				delete(resultSet, k)
				count--
			}
		}
	}
	for k, flag := range resultSet {
		if flag {
			members = append(members, k)
		}
	}
	return Encode(members, false)
}

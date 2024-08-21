package core

import (
	"errors"
	"sort"
)

func evalSADD(args []string, store *Store) []byte {
	if len(args) < 2 {
		return Encode(errors.New("ERR wrong number of arguments for 'SADD' command"), false)
	}
	key := args[0]

	// Get the set object from the store.
	obj := store.Get(key)

	var count int = 0
	if obj == nil {
		var exDurationMs int64 = -1
		var keepttl bool = false
		// If the object does not exist, create a new set object.
		value := make(map[string]bool)
		// Create a new object.
		obj = store.NewObj(value, exDurationMs, ObjTypeSet, ObjEncodingHT)
		store.Put(key, obj, WithKeepTTL(keepttl))
	}

	if assertType(obj.TypeEncoding, ObjTypeSet) != nil {
		return Encode(errors.New("WRONGTYPE Operation against a key holding the wrong kind of value"), false)
	}
	if assertEncoding(obj.TypeEncoding, ObjEncodingHT) != nil {
		return Encode(errors.New("WRONGTYPE Operation against a key holding the wrong kind of value"), false)
	}

	// Get the set object.
	set := obj.Value.(map[string]bool)

	for _, arg := range args[1:] {
		if _, ok := set[arg]; !ok {
			set[arg] = true
			count++
		}
	}

	return Encode(count, false)
}

func evalSMEMBERS(args []string, store *Store) []byte {
	if len(args) != 1 {
		return Encode(errors.New("ERR wrong number of arguments for 'SMEMBERS' command"), false)
	}
	key := args[0]

	// Get the set object from the store.
	obj := store.Get(key)

	if obj == nil {
		return Encode([]string{}, false)
	}

	// If the object exists, check if it is a set object.
	if assertEncoding(obj.TypeEncoding, ObjEncodingHT) != nil {
		return Encode(errors.New("WRONGTYPE Operation against a key holding the wrong kind of value"), false)
	}

	if assertType(obj.TypeEncoding, ObjTypeSet) != nil {
		return Encode(errors.New("WRONGTYPE Operation against a key holding the wrong kind of value"), false)
	}

	// Get the set object.
	set := obj.Value.(map[string]bool)
	// Get the members of the set.
	var members = make([]string, 0, len(set))
	for k, flag := range set {
		if flag {
			members = append(members, k)
		}
	}

	return Encode(members, false)
}

func evalSREM(args []string, store *Store) []byte {
	if len(args) < 2 {
		return Encode(errors.New("ERR wrong number of arguments for 'SREM' command"), false)
	}
	key := args[0]

	// Get the set object from the store.
	obj := store.Get(key)

	var count int = 0
	if obj == nil {
		return Encode(count, false)
	}

	// If the object exists, check if it is a set object.
	if assertEncoding(obj.TypeEncoding, ObjEncodingHT) != nil {
		return Encode(errors.New("WRONGTYPE Operation against a key holding the wrong kind of value"), false)
	}

	if assertType(obj.TypeEncoding, ObjTypeSet) != nil {
		return Encode(errors.New("WRONGTYPE Operation against a key holding the wrong kind of value"), false)
	}

	// Get the set object.
	set := obj.Value.(map[string]bool)
	for _, arg := range args[1:] {
		if set[arg] {
			delete(set, arg)
			count++
		}
	}

	return Encode(count, false)
}

func evalSCARD(args []string, store *Store) []byte {
	if len(args) != 1 {
		return Encode(errors.New("ERR wrong number of arguments for 'SCARD' command"), false)
	}

	key := args[0]

	// Get the set object from the store.
	obj := store.Get(key)

	if obj == nil {
		return Encode(0, false)
	}

	// If the object exists, check if it is a set object.
	if assertEncoding(obj.TypeEncoding, ObjEncodingHT) != nil {
		return Encode(errors.New("WRONGTYPE Operation against a key holding the wrong kind of value"), false)
	}

	if assertType(obj.TypeEncoding, ObjTypeSet) != nil {
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

func evalSDIFF(args []string, store *Store) []byte {
	if len(args) < 1 {
		return Encode(errors.New("ERR wrong number of arguments for 'SDIFF' command"), false)
	}

	srcKey := args[0]
	obj := store.Get(srcKey)

	srcSet := make(map[string]bool)

	// Get the set object from the store.
	// store the count as the number of elements in the first set
	// we decrement the count as we find the elements in the other sets
	// if the count is 0, we skip further sets but still get them from
	// the store to check if they are set objects and update their last accessed time

	var count int = 0
	if obj != nil {
		if assertEncoding(obj.TypeEncoding, ObjEncodingHT) != nil {
			return Encode(errors.New("WRONGTYPE Operation against a key holding the wrong kind of value"), false)
		}

		if assertType(obj.TypeEncoding, ObjTypeSet) != nil {
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

	for _, arg := range args[1:] {
		// Get the set object from the store.
		obj := store.Get(arg)

		if obj == nil {
			continue
		}

		// If the object exists, check if it is a set object.
		if assertEncoding(obj.TypeEncoding, ObjEncodingHT) != nil {
			return Encode(errors.New("WRONGTYPE Operation against a key holding the wrong kind of value"), false)
		}

		if assertType(obj.TypeEncoding, ObjTypeSet) != nil {
			return Encode(errors.New("WRONGTYPE Operation against a key holding the wrong kind of value"), false)
		}

		// only if the count is greater than 0, we need to check the other sets
		if count > 0 {
			// Get the set object.
			set := obj.Value.(map[string]bool)
			for k, flag := range set {
				if flag && srcSet[k] {
					delete(srcSet, k)
					count--
				}
			}
		}
	}

	if count == 0 {
		return Encode([]string{}, false)
	}

	// Get the members of the set.
	var members = make([]string, 0, len(srcSet))
	for k, flag := range srcSet {
		if flag {
			members = append(members, k)
		}
	}

	return Encode(members, false)
}

func evalSINTER(args []string, store *Store) []byte {
	if len(args) < 2 {
		return Encode(errors.New("ERR wrong number of arguments for 'SINTER' command"), false)
	}

	sets := make([]map[string]bool, 0, len(args))

	var empty int = 0

	for _, arg := range args {
		// Get the set object from the store.
		obj := store.Get(arg)

		if obj == nil {
			empty++
			continue
		}

		// If the object exists, check if it is a set object.
		if assertEncoding(obj.TypeEncoding, ObjEncodingHT) != nil {
			return Encode(errors.New("WRONGTYPE Operation against a key holding the wrong kind of value"), false)
		}

		if assertType(obj.TypeEncoding, ObjTypeSet) != nil {
			return Encode(errors.New("WRONGTYPE Operation against a key holding the wrong kind of value"), false)
		}

		// Get the set object.
		set := obj.Value.(map[string]bool)
		sets = append(sets, set)
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

	// init the result set with the first set
	// store the number of elements in the first set in count
	// we will decrement the count if we do not find the elements in the other sets
	for k := range sets[0] {
		count++
		resultSet[k] = true
	}

	for i := 1; i < len(sets); i++ {
		if count == 0 {
			break
		}
		for k := range resultSet {
			if !sets[i][k] {
				delete(resultSet, k)
				count--
			}
		}
	}

	if count == 0 {
		return Encode([]string{}, false)
	}

	var members = make([]string, 0, len(resultSet))
	for k := range resultSet {
		members = append(members, k)
	}
	return Encode(members, false)
}

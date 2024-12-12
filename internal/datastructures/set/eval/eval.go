package eval

import (
	"strconv"

	ds "github.com/dicedb/dice/internal/datastructures"
	set "github.com/dicedb/dice/internal/datastructures/set"
	"github.com/dicedb/dice/internal/eval" // Ensure this package contains the definition for Object
	dstore "github.com/dicedb/dice/internal/store"
)

// deduce new encoding from the set of encodings
// creates a new set with the new encoding
func createNewSetAndAdd(obj *ds.DSInterface, oldEncoding int, item string) bool {
	encs := make(map[int]struct{})
	encs[oldEncoding] = struct{}{}
	itemEncoding := ds.GetElementType(item)
	encs[itemEncoding] = struct{}{}
	newEncoding := set.DeduceEncodingFromItems(encs)
	newSet := set.NewTypedSetFromEncoding(newEncoding)
	var items []string
	switch oldEncoding {
	case ds.EncodingInt8:
		//convert all the items from int8 to newEncoding
		set, ok := set.GetIfTypeTypedSet[int8](*obj)
		if !ok {
			return false
		}
		for _, i := range set.All() {
			items = append(items, ds.ToString(i))
		}

	case ds.EncodingInt16:
		//convert all the items from int16 to newEncoding
		set, ok := set.GetIfTypeTypedSet[int16](*obj)
		if !ok {
			return false
		}
		for _, i := range set.All() {
			items = append(items, ds.ToString(i))
		}
	case ds.EncodingInt32:
		//convert all the items from int32 to newEncoding
		set, ok := set.GetIfTypeTypedSet[int32](*obj)
		if !ok {
			return false
		}
		for _, i := range set.All() {
			items = append(items, ds.ToString(i))
		}
	case ds.EncodingInt64:
		//convert all the items from int64 to newEncoding
		set, ok := set.GetIfTypeTypedSet[int64](*obj)
		if !ok {
			return false
		}
		for _, i := range set.All() {
			items = append(items, ds.ToString(i))
		}
	case ds.EncodingFloat32:
		//convert all the items from float32 to newEncoding
		set, ok := set.GetIfTypeTypedSet[float32](*obj)
		if !ok {
			return false
		}
		for _, i := range set.All() {
			items = append(items, ds.ToString(i))
		}
	case ds.EncodingFloat64:
		//convert all the items from float64 to newEncoding
		set, ok := set.GetIfTypeTypedSet[float64](*obj)
		if !ok {
			return false
		}
		for _, i := range set.All() {
			items = append(items, ds.ToString(i))
		}
	default:
		//convert all the items from string to newEncoding
		set, ok := set.GetIfTypeTypedSet[string](*obj)
		if !ok {
			return false
		}
		for _, i := range set.All() {
			items = append(items, i)
		}
	}
	items = append(items, item)
	newSet, _ = set.NewTypedSetFromEncodingAndItems(items, newEncoding)
	obj = newSet
	return true
}

// tryAndAddToSet tries to add an item to a set object.
// If the item is of a different type than the set object,
// it creates a new set object with the correct type and adds the item to it.
func tryAndAddToSet[T ds.DSInterface](obj *T, item string) bool {
	internalEncoding := (*obj).GetEncoding()
	itemEncoding := ds.GetElementType(item)
	if internalEncoding == itemEncoding {
		return addToSetObj(obj, item)
	}
	return createNewSetAndAdd(obj.(ds.DSInterface), internalEncoding, item)
}

// addToSetObj adds an item to a set object.
// addToSetObj should only be called if the item is of the same type as the set object.
// If the item is not of the same type, it WILL return false.
func addToSetObj[T ds.DSInterface](obj *T, item string) bool {

	switch (*obj).GetEncoding() {
	case ds.EncodingInt8:
		// Convert the item to an int8.
		intItem, err := strconv.ParseInt(item, 10, 8)
		if err != nil {
			// create a new set and add the item
			return false
		}
		set, ok := set.GetIfTypeTypedSet[int8](*obj)
		if !ok {
			return false
		}
		return set.Add(int8(intItem))
	case ds.EncodingInt16:
		// Convert the item to an int16.
		intItem, err := strconv.ParseInt(item, 10, 16)
		if err != nil {
			// create a new set and add the item
			return false
		}
		set, ok := set.GetIfTypeTypedSet[int16](*obj)
		if !ok {
			return false
		}
		return set.Add(int16(intItem))
	case ds.EncodingInt32:
		// Convert the item to an int32.
		intItem, err := strconv.ParseInt(item, 10, 32)
		if err != nil {
			// create a new set and add the item
			return false
		}
		set, ok := set.GetIfTypeTypedSet[int32](*obj)
		if !ok {
			return false
		}
		return set.Add(int32(intItem))
	case ds.EncodingInt64:
		// Convert the item to an int64.
		intItem, err := strconv.ParseInt(item, 10, 64)
		if err != nil {
			// create a new set and add the item
			return false
		}
		set, ok := set.GetIfTypeTypedSet[int64](*obj)
		if !ok {
			return false
		}
		return set.Add(int64(intItem))
	case ds.EncodingFloat32:
		// Convert the item to a float32.
		floatItem, err := strconv.ParseFloat(item, 32)
		if err != nil {
			// create a new set and add the item
			return false
		}
		set, ok := set.GetIfTypeTypedSet[float32](*obj)
		if !ok {
			return false
		}
		return set.Add(float32(floatItem))
	case ds.EncodingFloat64:
		// Convert the item to a float64.
		floatItem, err := strconv.ParseFloat(item, 64)
		if err != nil {
			// create a new set and add the item
			return false
		}
		set, ok := set.GetIfTypeTypedSet[float64](*obj)
		if !ok {
			return false
		}
		return set.Add(float64(floatItem))
	default:
		set, ok := set.GetIfTypeTypedSet[string](*obj)
		if !ok {
			return false
		}
		return set.Add(item)
	}
}

func evalSADD[T ds.DSInterface](args []string, store *dstore.Store[T]) *eval.EvalResponse {
	if len(args) < 2 {
		return eval.MakeEvalError(ErrWrongArgumentCount("sadd"))
	}
	key := args[0]

	// Get the set object from the store.
	obj := store.Get(key)

	var count = 0
	if obj == nil {
		var exDurationMs int64 = -1
		var keepttl = false
		// If the object does not exist, create a new set object.
		value, _ := set.NewTypedSetFromItems(args[1:])
		// Create a new object.
		obj = store.NewObj(*value, exDurationMs)
		store.Put(key, obj, dstore.WithKeepTTL(keepttl))
		return eval.MakeEvalResult(len(args) - 1)
	}

	if (*obj).GetType() != ds.ObjTypeSet {
		return eval.MakeEvalError(ErrWrongTypeOperation)
	}

	for _, arg := range args[1:] {
		if addToSetObj(obj, arg) {
			count++
		}
	}

	return eval.MakeEvalResult(count)
}

func evalSCARD[T ds.DSInterface](args []string, store *dstore.Store[T]) *eval.EvalResponse {
	if len(args) != 1 {
		return eval.MakeEvalError(ErrWrongArgumentCount("scard"))
	}
	key := args[0]

	// Get the set object from the store.
	obj := store.Get(key)

	if obj == nil {
		return eval.MakeEvalResult(0)
	}

	if (*obj).GetType() != ds.ObjTypeSet {
		return eval.MakeEvalError(ErrWrongTypeOperation)
	}

	switch (*obj).GetEncoding() {
	case ds.EncodingInt8:
		set, ok := set.GetIfTypeTypedSet[int8](*obj)
		if !ok {
			return eval.MakeEvalError(ErrWrongTypeOperation)
		}
		return eval.MakeEvalResult(set.Size())
	case ds.EncodingInt16:
		set, ok := set.GetIfTypeTypedSet[int16](*obj)
		if !ok {
			return eval.MakeEvalError(ErrWrongTypeOperation)
		}
		return eval.MakeEvalResult(set.Size())
	case ds.EncodingInt32:
		set, ok := set.GetIfTypeTypedSet[int32](*obj)
		if !ok {
			return eval.MakeEvalError(ErrWrongTypeOperation)
		}
		return eval.MakeEvalResult(set.Size())
	case ds.EncodingInt64:
		set, ok := set.GetIfTypeTypedSet[int64](*obj)
		if !ok {
			return eval.MakeEvalError(ErrWrongTypeOperation)
		}
		return eval.MakeEvalResult(set.Size())
	case ds.EncodingFloat32:
		set, ok := set.GetIfTypeTypedSet[float32](*obj)
		if !ok {
			return eval.MakeEvalError(ErrWrongTypeOperation)
		}
		return eval.MakeEvalResult(set.Size())
	case ds.EncodingFloat64:
		set, ok := set.GetIfTypeTypedSet[float64](*obj)
		if !ok {
			return eval.MakeEvalError(ErrWrongTypeOperation)
		}
		return eval.MakeEvalResult(set.Size())
	default:
		set, ok := set.GetIfTypeTypedSet[string](*obj)
		if !ok {
			return eval.MakeEvalError(ErrWrongTypeOperation)
		}
		return eval.MakeEvalResult(set.Size())
	}
}

func evalSMEMBERS[T ds.DSInterface](args []string, store *dstore.Store[T]) *eval.EvalResponse {
	if len(args) != 1 {
		return eval.MakeEvalError(ErrWrongArgumentCount("smembers"))
	}
	key := args[0]

	// Get the set object from the store.
	obj := store.Get(key)

	if obj == nil {
		return eval.MakeEvalResult([]string{})
	}

	if (*obj).GetType() != ds.ObjTypeSet {
		return eval.MakeEvalError(ErrWrongTypeOperation)
	}

	encoding := (*obj).GetEncoding()
	switch encoding {
	case ds.EncodingInt8:
		set, ok := set.GetIfTypeTypedSet[int8](*obj)
		if !ok {
			return eval.MakeEvalError(ErrWrongTypeOperation)
		}
		return eval.MakeEvalResult(set.All())
	case ds.EncodingInt16:
		set, ok := set.GetIfTypeTypedSet[int16](*obj)
		if !ok {
			return eval.MakeEvalError(ErrWrongTypeOperation)
		}
		return eval.MakeEvalResult(set.All())
	case ds.EncodingInt32:
		set, ok := set.GetIfTypeTypedSet[int32](*obj)
		if !ok {
			return eval.MakeEvalError(ErrWrongTypeOperation)
		}
		return eval.MakeEvalResult(set.All())
	case ds.EncodingInt64:
		set, ok := set.GetIfTypeTypedSet[int64](*obj)
		if !ok {
			return eval.MakeEvalError(ErrWrongTypeOperation)
		}
		return eval.MakeEvalResult(set.All())
	case ds.EncodingFloat32:
		set, ok := set.GetIfTypeTypedSet[float32](*obj)
		if !ok {
			return eval.MakeEvalError(ErrWrongTypeOperation)
		}
		return eval.MakeEvalResult(set.All())
	case ds.EncodingFloat64:
		set, ok := set.GetIfTypeTypedSet[float64](*obj)
		if !ok {
			return eval.MakeEvalError(ErrWrongTypeOperation)
		}
		return eval.MakeEvalResult(set.All())
	default:
		set, ok := set.GetIfTypeTypedSet[string](*obj)
		if !ok {
			return eval.MakeEvalError(ErrWrongTypeOperation)
		}
		return eval.MakeEvalResult(set.All())
	}
}

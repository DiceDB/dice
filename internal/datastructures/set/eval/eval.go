package eval

import (
	"strconv"

	ds "github.com/dicedb/dice/internal/datastructures"
	set "github.com/dicedb/dice/internal/datastructures/set"
	"github.com/dicedb/dice/internal/eval"
	object "github.com/dicedb/dice/internal/object" // Ensure this package contains the definition for Object
	"github.com/dicedb/dice/internal/shard"
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
	case set.EncodingInt8:
		//convert all the items from int8 to newEncoding
		for _, i := range obj.Value.(*TypedSet[int8]).All() {
			items = append(items, ds.ToString(i))
		}
	case set.EncodingInt16:
		//convert all the items from int16 to newEncoding
		for _, i := range obj.Value.(*TypedSet[int16]).All() {
			items = append(items, ds.ToString(i))
		}
	case set.EncodingInt32:
		//convert all the items from int32 to newEncoding
		for _, i := range obj.Value.(*TypedSet[int32]).All() {
			items = append(items, ds.ToString(i))
		}
	case set.EncodingInt64:
		//convert all the items from int64 to newEncoding
		for _, i := range obj.Value.(*TypedSet[int64]).All() {
			items = append(items, ds.ToString(i))
		}
	case set.EncodingFloat32:
		//convert all the items from float32 to newEncoding
		for _, i := range obj.Value.(*TypedSet[float32]).All() {
			items = append(items, ds.ToString(i))
		}
	case set.EncodingFloat64:
		//convert all the items from float64 to newEncoding
		for _, i := range obj.Value.(*TypedSet[float64]).All() {
			items = append(items, ds.ToString(i))
		}
	default:
		//convert all the items from string to newEncoding
		for _, i := range obj.Value.(*TypedSet[string]).All() {
			items = append(items, i)
		}
	}
	items = append(items, item)
	newSet, _ = set.NewTypedSetFromEncodingAndItems(items, newEncoding)
	obj.Value = newSet
	return true
}

// tryAndAddToSet tries to add an item to a set object.
// If the item is of a different type than the set object,
// it creates a new set object with the correct type and adds the item to it.
func tryAndAddToSet(obj *object.Obj, item string) bool {
	internalEncoding := obj.Value.(ds.BaseDataStructure[ds.DSInterface]).Encoding
	itemEncoding := GetElementType(item)
	if internalEncoding == itemEncoding {
		return addToSetObj(obj, item)
	}
	return createNewSetAndAdd(obj, internalEncoding, item)
}

// addToSetObj adds an item to a set object.
// addToSetObj should only be called if the item is of the same type as the set object.
// If the item is not of the same type, it WILL return false.
func addToSetObj(obj *object.Obj, item string) bool {
	internalEncoding := obj.Value.(*ds.BaseDataStructure[ds.DSInterface]).GetEncoding()
	switch internalEncoding {
	case EncodingInt8:
		// Convert the item to an int8.
		intItem, err := strconv.ParseInt(item, 10, 8)
		if err != nil {
			// create a new set and add the item
			return false
		}
		return obj.Value.(*TypedSet[int64]).Add(intItem)
	case EncodingInt16:
		// Convert the item to an int16.
		intItem, err := strconv.ParseInt(item, 10, 16)
		if err != nil {
			// create a new set and add the item
			return false
		}
		return obj.Value.(*TypedSet[int64]).Add(intItem)
	case EncodingInt32:
		// Convert the item to an int32.
		intItem, err := strconv.ParseInt(item, 10, 32)
		if err != nil {
			// create a new set and add the item
			return false
		}
		return obj.Value.(*TypedSet[int64]).Add(intItem)
	case EncodingInt64:
		// Convert the item to an int64.
		intItem, err := strconv.ParseInt(item, 10, 64)
		if err != nil {
			// create a new set and add the item
			return false
		}
		return obj.Value.(*TypedSet[int64]).Add(intItem)
	case EncodingFloat32:
		// Convert the item to a float32.
		floatItem, err := strconv.ParseFloat(item, 32)
		if err != nil {
			// create a new set and add the item
			return false
		}
		return obj.Value.(*TypedSet[float32]).Add(float32(floatItem))
	case EncodingFloat64:
		// Convert the item to a float64.
		floatItem, err := strconv.ParseFloat(item, 64)
		if err != nil {
			// create a new set and add the item
			return false
		}
		return obj.Value.(*TypedSet[float64]).Add(floatItem)
	default:
		return obj.Value.(*TypedSet[string]).Add(item)
	}
}

func evalSADD(args []string, store *dstore.Store) *shard.EvalResponse {
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
		value, _ := NewTypedSetFromItems(args[1:])
		// Create a new object.
		obj = store.NewObj(value, exDurationMs, object.ObjTypeSet)
		store.Put(key, obj, dstore.WithKeepTTL(keepttl))
		return eval.MakeEvalResult(len(args) - 1)
	}

	if obj.Type != object.ObjTypeSet {
		return eval.MakeEvalError(ErrWrongTypeOperation)
	}

	for _, arg := range args[1:] {
		if addToSetObj(obj, arg) {
			count++
		}
	}

	return eval.MakeEvalResult(count)
}

func evalSCARD(args []string, store *dstore.Store) *shard.EvalResponse {
	if len(args) != 1 {
		return eval.MakeEvalError(ErrWrongArgumentCount("scard"))
	}
	key := args[0]

	// Get the set object from the store.
	obj := store.Get(key)

	if obj == nil {
		return eval.MakeEvalResult(0)
	}

	if obj.Type != object.ObjTypeSet {
		return eval.MakeEvalError(ErrWrongTypeOperation)
	}

	internalEncoding := obj.Value.(*ds.BaseDataStructure[ds.DSInterface]).GetEncoding()
	switch internalEncoding {
	case EncodingInt8:
		set := obj.Value.(*TypedSet[int8])
		return eval.MakeEvalResult(set.Size())
	case EncodingInt16:
		set := obj.Value.(*TypedSet[int16])
		return eval.MakeEvalResult(set.Size())
	case EncodingInt32:
		set := obj.Value.(*TypedSet[int32])
		return eval.MakeEvalResult(set.Size())
	case EncodingInt64:
		set := obj.Value.(*TypedSet[int64])
		return eval.MakeEvalResult(set.Size())
	case EncodingFloat32:
		set := obj.Value.(*TypedSet[float32])
		return eval.MakeEvalResult(set.Size())
	case EncodingFloat64:
		set := obj.Value.(*TypedSet[float64])
		return eval.MakeEvalResult(set.Size())
	default:
		set := obj.Value.(*TypedSet[string])
		return eval.MakeEvalResult(set.Size())
	}
}

func evalSMEMBERS(args []string, store *dstore.Store) *shard.EvalResponse {
	if len(args) != 1 {
		return eval.MakeEvalError(ErrWrongArgumentCount("smembers"))
	}
	key := args[0]

	// Get the set object from the store.
	obj := store.Get(key)

	if obj == nil {
		return eval.MakeEvalResult([]string{})
	}

	if obj.Type != object.ObjTypeSet {
		return eval.MakeEvalError(ErrWrongTypeOperation)
	}

	internalEncoding := obj.Value.(*dstore.BaseDataStructure[dstore.DSInterface]).GetEncoding()
	switch internalEncoding {
	case EncodingInt8:
		set := obj.Value.(*TypedSet[int8])
		return eval.MakeEvalResult(set.All())
	case EncodingInt16:
		set := obj.Value.(*TypedSet[int16])
		return eval.MakeEvalResult(set.All())
	case EncodingInt32:
		set := obj.Value.(*TypedSet[int32])
		return eval.MakeEvalResult(set.All())
	case EncodingInt64:
		set := obj.Value.(*TypedSet[int64])
		return eval.MakeEvalResult(set.All())
	case EncodingFloat32:
		set := obj.Value.(*TypedSet[float32])
		return eval.MakeEvalResult(set.All())
	case EncodingFloat64:
		set := obj.Value.(*TypedSet[float64])
		return eval.MakeEvalResult(set.All())
	default:
		set := obj.Value.(*TypedSet[string])
		return eval.MakeEvalResult(set.All())
	}
}

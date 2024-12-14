package eval

import (
	"strconv"

	"github.com/dicedb/dice/internal/cmd"
	"github.com/dicedb/dice/internal/comm"
	ds "github.com/dicedb/dice/internal/datastructures"
	set "github.com/dicedb/dice/internal/datastructures/set"
	"github.com/dicedb/dice/internal/store"
	dstore "github.com/dicedb/dice/internal/store"
)

type EvalResponse struct {
	Error  error
	Result interface{}
}

func MakeEvalError(err error) *EvalResponse {
	return &EvalResponse{
		Result: nil,
		Error:  err,
	}
}

func MakeEvalResult(result interface{}) *EvalResponse {
	return &EvalResponse{
		Result: result,
		Error:  nil,
	}
}

type Eval struct {
	store                 *store.Store
	cmd                   *cmd.DiceDBCmd
	client                *comm.Client
	isHTTPOperation       bool
	isWebSocketOperation  bool
	isPreprocessOperation bool
	_                     [5]byte
}

// deduce new encoding from the set of encodings
// creates a new set with the new encoding
func createNewSetAndAdd(obj ds.DSInterface, oldEncoding int, item string) bool {
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
		set, ok := set.GetIfTypeTypedSet[int8](obj)
		if !ok {
			return false
		}
		for _, i := range set.All() {
			items = append(items, ds.ToString(i))
		}

	case ds.EncodingInt16:
		//convert all the items from int16 to newEncoding
		set, ok := set.GetIfTypeTypedSet[int16](obj)
		if !ok {
			return false
		}
		for _, i := range set.All() {
			items = append(items, ds.ToString(i))
		}
	case ds.EncodingInt32:
		//convert all the items from int32 to newEncoding
		set, ok := set.GetIfTypeTypedSet[int32](obj)
		if !ok {
			return false
		}
		for _, i := range set.All() {
			items = append(items, ds.ToString(i))
		}
	case ds.EncodingInt64:
		//convert all the items from int64 to newEncoding
		set, ok := set.GetIfTypeTypedSet[int64](obj)
		if !ok {
			return false
		}
		for _, i := range set.All() {
			items = append(items, ds.ToString(i))
		}
	case ds.EncodingFloat32:
		//convert all the items from float32 to newEncoding
		set, ok := set.GetIfTypeTypedSet[float32](obj)
		if !ok {
			return false
		}
		for _, i := range set.All() {
			items = append(items, ds.ToString(i))
		}
	case ds.EncodingFloat64:
		//convert all the items from float64 to newEncoding
		set, ok := set.GetIfTypeTypedSet[float64](obj)
		if !ok {
			return false
		}
		for _, i := range set.All() {
			items = append(items, ds.ToString(i))
		}
	default:
		//convert all the items from string to newEncoding
		set, ok := set.GetIfTypeTypedSet[string](obj)
		if !ok {
			return false
		}
		for _, i := range set.All() {
			items = append(items, i)
		}
	}
	items = append(items, item)
	newSet = set.NewTypedSetFromEncodingAndItems(items, newEncoding)
	obj = newSet
	return true
}

// tryAndAddToSet tries to add an item to a set object.
// If the item is of a different type than the set object,
// it creates a new set object with the correct type and adds the item to it.
func tryAndAddToSet(obj ds.DSInterface, item string) bool {
	isSet, ok := set.GetIfTypeSet(obj)
	if !ok {
		return false
	}
	setEncoding := isSet.GetEncoding()
	itemEncoding := ds.GetElementType(item)
	if setEncoding == itemEncoding {
		return addToSetObj(obj, setEncoding, item)
	}
	return createNewSetAndAdd(obj, setEncoding, item)
}

// addToSetObj adds an item to a set object.
// addToSetObj should only be called if the item is of the same type as the set object.
// If the item is not of the same type, it WILL return false.
func addToSetObj(obj ds.DSInterface, encoding int, item string) bool {

	switch encoding {
	case ds.EncodingInt8:
		// Convert the item to an int8.
		intItem, err := strconv.ParseInt(item, 10, 8)
		if err != nil {
			// create a new set and add the item
			return false
		}
		set, ok := set.GetIfTypeTypedSet[int8](obj)
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
		set, ok := set.GetIfTypeTypedSet[int16](obj)
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
		set, ok := set.GetIfTypeTypedSet[int32](obj)
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
		set, ok := set.GetIfTypeTypedSet[int64](obj)
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
		set, ok := set.GetIfTypeTypedSet[float32](obj)
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
		set, ok := set.GetIfTypeTypedSet[float64](obj)
		if !ok {
			return false
		}
		return set.Add(float64(floatItem))
	default:
		set, ok := set.GetIfTypeTypedSet[string](obj)
		if !ok {
			return false
		}
		return set.Add(item)
	}
}

func tryAndAddToSet2(obj ds.DSInterface, item string) bool {
	switch any(obj).(type) {
	case *set.TypedSet[int8]:
		i, ok := strconv.ParseInt(item, 10, 8)
		if ok != nil {
			return createNewSetAndAdd(obj, ds.EncodingInt8, item)
		}
		return obj.(*set.TypedSet[int8]).Add(int8(i))
	case *set.TypedSet[int16]:
		i, ok := strconv.ParseInt(item, 10, 16)
		if ok != nil {
			return createNewSetAndAdd(obj, ds.EncodingInt16, item)
		}
		return obj.(*set.TypedSet[int16]).Add(int16(i))
	case *set.TypedSet[int32]:
		i, ok := strconv.ParseInt(item, 10, 32)
		if ok != nil {
			return createNewSetAndAdd(obj, ds.EncodingInt32, item)
		}
		return obj.(*set.TypedSet[int32]).Add(int32(i))
	case *set.TypedSet[int64]:
		i, ok := strconv.ParseInt(item, 10, 64)
		if ok != nil {
			return createNewSetAndAdd(obj, ds.EncodingInt64, item)
		}
		return obj.(*set.TypedSet[int64]).Add(int64(i))
	case *set.TypedSet[float32]:
		i, ok := strconv.ParseFloat(item, 32)
		if ok != nil {
			return createNewSetAndAdd(obj, ds.EncodingFloat32, item)
		}
		return obj.(*set.TypedSet[float32]).Add(float32(i))
	case *set.TypedSet[float64]:
		i, ok := strconv.ParseFloat(item, 64)
		if ok != nil {
			return createNewSetAndAdd(obj, ds.EncodingFloat64, item)
		}
		return obj.(*set.TypedSet[float64]).Add(float64(i))
	default:
		return obj.(*set.TypedSet[string]).Add(item)
	}
}
func evalSADD(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 2 {
		return MakeEvalError(ErrWrongArgumentCount("sadd"))
	}
	key := args[0]

	// Get the set object from the store.
	obj := store.Get(key)

	var count = 0
	if obj == nil {
		var exDurationMs int64 = -1
		var keepttl = false
		// If the object does not exist, create a new set object.
		value := set.NewTypedSetFromItems(args[1:])
		// Create a new object.
		obj = store.NewObj(value, exDurationMs)
		store.Put(key, obj, dstore.WithKeepTTL(keepttl))
		return MakeEvalResult(len(args) - 1)
	}

	_, ok := set.GetIfTypeSet(obj)
	if !ok {
		return MakeEvalError(ErrWrongTypeOperation)
	}

	for _, arg := range args[1:] {
		if tryAndAddToSet(obj, arg) {
			count++
		}
	}

	return MakeEvalResult(count)
}

func evalSCARD(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 1 {
		return MakeEvalError(ErrWrongArgumentCount("scard"))
	}
	key := args[0]

	// Get the set object from the store.
	obj := store.Get(key)

	if obj == nil {
		return MakeEvalResult(0)
	}

	setObj, ok := set.GetIfTypeSet(obj)
	if !ok {
		return MakeEvalError(ErrWrongTypeOperation)
	}

	switch setObj.GetEncoding() {
	case ds.EncodingInt8:
		set, ok := set.GetIfTypeTypedSet[int8](obj)
		if !ok {
			return MakeEvalError(ErrWrongTypeOperation)
		}
		return MakeEvalResult(set.Size())
	case ds.EncodingInt16:
		set, ok := set.GetIfTypeTypedSet[int16](obj)
		if !ok {
			return MakeEvalError(ErrWrongTypeOperation)
		}
		return MakeEvalResult(set.Size())
	case ds.EncodingInt32:
		set, ok := set.GetIfTypeTypedSet[int32](obj)
		if !ok {
			return MakeEvalError(ErrWrongTypeOperation)
		}
		return MakeEvalResult(set.Size())
	case ds.EncodingInt64:
		set, ok := set.GetIfTypeTypedSet[int64](obj)
		if !ok {
			return MakeEvalError(ErrWrongTypeOperation)
		}
		return MakeEvalResult(set.Size())
	case ds.EncodingFloat32:
		set, ok := set.GetIfTypeTypedSet[float32](obj)
		if !ok {
			return MakeEvalError(ErrWrongTypeOperation)
		}
		return MakeEvalResult(set.Size())
	case ds.EncodingFloat64:
		set, ok := set.GetIfTypeTypedSet[float64](obj)
		if !ok {
			return MakeEvalError(ErrWrongTypeOperation)
		}
		return MakeEvalResult(set.Size())
	default:
		set, ok := set.GetIfTypeTypedSet[string](obj)
		if !ok {
			return MakeEvalError(ErrWrongTypeOperation)
		}
		return MakeEvalResult(set.Size())
	}
}

func evalSMEMBERS(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 1 {
		return MakeEvalError(ErrWrongArgumentCount("smembers"))
	}
	key := args[0]

	// Get the set object from the store.
	obj := store.Get(key)

	if obj == nil {
		return MakeEvalResult([]string{})
	}

	setObj, ok := set.GetIfTypeSet(obj)
	if !ok {
		return MakeEvalError(ErrWrongTypeOperation)
	}

	switch setObj.GetEncoding() {
	case ds.EncodingInt8:
		set, ok := set.GetIfTypeTypedSet[int8](obj)
		if !ok {
			return MakeEvalError(ErrWrongTypeOperation)
		}
		return MakeEvalResult(set.All())
	case ds.EncodingInt16:
		set, ok := set.GetIfTypeTypedSet[int16](obj)
		if !ok {
			return MakeEvalError(ErrWrongTypeOperation)
		}
		return MakeEvalResult(set.All())
	case ds.EncodingInt32:
		set, ok := set.GetIfTypeTypedSet[int32](obj)
		if !ok {
			return MakeEvalError(ErrWrongTypeOperation)
		}
		return MakeEvalResult(set.All())
	case ds.EncodingInt64:
		set, ok := set.GetIfTypeTypedSet[int64](obj)
		if !ok {
			return MakeEvalError(ErrWrongTypeOperation)
		}
		return MakeEvalResult(set.All())
	case ds.EncodingFloat32:
		set, ok := set.GetIfTypeTypedSet[float32](obj)
		if !ok {
			return MakeEvalError(ErrWrongTypeOperation)
		}
		return MakeEvalResult(set.All())
	case ds.EncodingFloat64:
		set, ok := set.GetIfTypeTypedSet[float64](obj)
		if !ok {
			return MakeEvalError(ErrWrongTypeOperation)
		}
		return MakeEvalResult(set.All())
	default:
		set, ok := set.GetIfTypeTypedSet[string](obj)
		if !ok {
			return MakeEvalError(ErrWrongTypeOperation)
		}
		return MakeEvalResult(set.All())
	}
}

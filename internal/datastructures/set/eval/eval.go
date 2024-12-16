package eval

import (
	"cmp"
	"strconv"

	"github.com/dicedb/dice/internal/cmd"
	"github.com/dicedb/dice/internal/comm"
	ds "github.com/dicedb/dice/internal/datastructures"
	"github.com/dicedb/dice/internal/datastructures/set"
	"github.com/dicedb/dice/internal/store"
	dstore "github.com/dicedb/dice/internal/store"
)

type EvalResponse struct {
	Error  error
	Result interface{}
}

type constraint interface {
	cmp.Ordered
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

func getAllItemsFromSet(obj ds.DSInterface) []string {
	switch any(obj).(type) {
	case *set.TypedSet[int8]:
		set, ok := obj.(*set.TypedSet[int8])
		if !ok {
			return []string{}
		}
		items := make([]string, 0, len(set.Value))
		for i := range set.Value {
			items = append(items, strconv.Itoa(int(i)))
		}
		return items
	case *set.TypedSet[int16]:
		set, ok := obj.(*set.TypedSet[int16])
		if !ok {
			return []string{}
		}
		items := make([]string, 0, len(set.Value))
		for i := range set.Value {
			items = append(items, strconv.Itoa(int(i)))
		}
		return items
	case *set.TypedSet[int32]:
		set, ok := obj.(*set.TypedSet[int32])
		if !ok {
			return []string{}
		}
		items := make([]string, 0, len(set.Value))
		for i := range set.Value {
			items = append(items, strconv.Itoa(int(i)))
		}
		return items
	case *set.TypedSet[int64]:
		set, ok := obj.(*set.TypedSet[int64])
		if !ok {
			return []string{}
		}
		items := make([]string, 0, len(set.Value))
		for i := range set.Value {
			items = append(items, strconv.Itoa(int(i)))
		}
		return items
	case *set.TypedSet[float32]:
		set, ok := obj.(*set.TypedSet[float32])
		if !ok {
			return []string{}
		}
		items := make([]string, 0, len(set.Value))
		for i := range set.Value {
			items = append(items, strconv.FormatFloat(float64(i), 'f', -1, 32))
		}
		return items
	case *set.TypedSet[float64]:
		set, ok := obj.(*set.TypedSet[float64])
		if !ok {
			return []string{}
		}
		items := make([]string, 0, len(set.Value))
		for i := range set.Value {
			items = append(items, strconv.FormatFloat(i, 'f', -1, 64))
		}
		return items
	default:
		set, ok := obj.(*set.TypedSet[string])
		if !ok {
			return []string{}
		}
		items := make([]string, 0, len(set.Value))
		for i := range set.Value {
			items = append(items, i)
		}
		return items
	}
}

func lenOfSet(obj ds.DSInterface) int {
	switch any(obj).(type) {
	case *set.TypedSet[int8]:
		return len(obj.(*set.TypedSet[int8]).Value)
	case *set.TypedSet[int16]:
		return len(obj.(*set.TypedSet[int16]).Value)
	case *set.TypedSet[int32]:
		return len(obj.(*set.TypedSet[int32]).Value)
	case *set.TypedSet[int64]:
		return len(obj.(*set.TypedSet[int64]).Value)
	case *set.TypedSet[float32]:
		return len(obj.(*set.TypedSet[float32]).Value)
	case *set.TypedSet[float64]:
		return len(obj.(*set.TypedSet[float64]).Value)
	default:
		return len(obj.(*set.TypedSet[string]).Value)
	}
}

// deduce new encoding from the set of encodings
// creates a new set with the new encoding
func createNewSetAndAdd(obj ds.DSInterface, oldEncoding int, item string) bool {
	encs := make(map[int]struct{})
	encs[oldEncoding] = struct{}{}
	itemEncoding := ds.GetElementType(item)
	encs[itemEncoding] = struct{}{}
	newEncoding := set.DeduceEncodingFromItems(encs)
	var newSet ds.DSInterface
	switch newEncoding {
	case ds.EncodingInt8, ds.EncodingInt16, ds.EncodingInt32, ds.EncodingInt64:
		items := make([]int, 0, lenOfSet(obj))
		switch any(obj).(type) {
		case *set.TypedSet[int8]:
			for i := range obj.(*set.TypedSet[int8]).Value {
				items = append(items, int(i))
			}
		case *set.TypedSet[int16]:
			for i := range obj.(*set.TypedSet[int16]).Value {
				items = append(items, int(i))
			}
		case *set.TypedSet[int32]:
			for i := range obj.(*set.TypedSet[int32]).Value {
				items = append(items, int(i))
			}
		case *set.TypedSet[int64]:
			for i := range obj.(*set.TypedSet[int64]).Value {
				items = append(items, int(i))
			}
		}
		newSet = set.NewIntegerTypedSetFromEncodingAndItems(items, newEncoding)
	case ds.EncodingFloat32, ds.EncodingFloat64:
		items := make([]float64, 0, lenOfSet(obj))
		switch any(obj).(type) {
		case *set.TypedSet[float32]:
			for i := range obj.(*set.TypedSet[float32]).Value {
				items = append(items, float64(i))
			}
		case *set.TypedSet[float64]:
			for i := range obj.(*set.TypedSet[float64]).Value {
				items = append(items, i)
			}
		}
		newSet = set.NewFloatTypedSetFromEncodingAndItems(items, newEncoding)
	default:
		items := make([]string, 0, lenOfSet(obj))
		items = append(items, getAllItemsFromSet(obj)...) // optimise this
		newSet = set.NewTypedSetFromEncodingAndItems(items, newEncoding)
	}
	obj = newSet
	return true
}

func insertIntoTypedSet[T constraint](obj ds.DSInterface, item T) bool {
	typedSet, ok := obj.(*set.TypedSet[T])
	if !ok {
		return false
	}

	if _, exists := typedSet.Value[item]; !exists {
		typedSet.Value[item] = struct{}{}
		return true
	}
	return false
}
func tryAndAddToSet(obj ds.DSInterface, item string) bool {
	switch any(obj).(type) {
	case *set.TypedSet[int8]:
		i, ok := strconv.ParseInt(item, 10, 8)
		if ok != nil {
			return createNewSetAndAdd(obj, ds.EncodingInt8, item)
		}

		ii := int8(i)
		return insertIntoTypedSet(obj, ii)
	case *set.TypedSet[int16]:
		i, ok := strconv.ParseInt(item, 10, 16)
		if ok != nil {
			return createNewSetAndAdd(obj, ds.EncodingInt16, item)
		}
		ii := int16(i)
		return insertIntoTypedSet(obj, ii)
	case *set.TypedSet[int32]:
		i, ok := strconv.ParseInt(item, 10, 32)
		if ok != nil {
			return createNewSetAndAdd(obj, ds.EncodingInt32, item)
		}
		ii := int32(i)
		return insertIntoTypedSet(obj, ii)
	case *set.TypedSet[int64]:
		i, ok := strconv.ParseInt(item, 10, 64)
		if ok != nil {
			return createNewSetAndAdd(obj, ds.EncodingInt64, item)
		}
		ii := int64(i)
		return insertIntoTypedSet(obj, ii)
	case *set.TypedSet[float32]:
		i, ok := strconv.ParseFloat(item, 32)
		if ok != nil {
			return createNewSetAndAdd(obj, ds.EncodingFloat32, item)
		}
		ff := float32(i)
		return insertIntoTypedSet(obj, ff)
	case *set.TypedSet[float64]:
		i, ok := strconv.ParseFloat(item, 64)
		if ok != nil {
			return createNewSetAndAdd(obj, ds.EncodingFloat64, item)
		}
		ff := float64(i)
		return insertIntoTypedSet(obj, ff)
	default:
		return insertIntoTypedSet(obj, item)
	}
	return false
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

	ok := set.IsTypeTypedSet(obj)
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

	return MakeEvalResult(lenOfSet(obj))

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

	ok := set.IsTypeTypedSet(obj)
	if !ok {
		return MakeEvalError(ErrWrongTypeOperation)
	}
	return MakeEvalResult(getAllItemsFromSet(obj))
}

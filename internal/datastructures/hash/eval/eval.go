package eval

import (
	"strconv"

	"github.com/dicedb/dice/internal/clientio"
	"github.com/dicedb/dice/internal/cmd"
	ds "github.com/dicedb/dice/internal/datastructures"
	hash "github.com/dicedb/dice/internal/datastructures/hash"
	"github.com/dicedb/dice/internal/store"
	dstore "github.com/dicedb/dice/internal/store"
)

type Eval struct {
	store *store.Store
	op    *cmd.DiceDBCmd
}

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

func hSetGeneric(args []string, store *dstore.Store, cmd string) *EvalResponse {
	if len(args) < 3 {
		return MakeEvalError(ds.ErrWrongArgumentCount(cmd))
	}

	key := args[0]
	size := len(args) - 1
	if size < 0 || size%2 != 0 {
		return MakeEvalError(ds.ErrWrongArgumentCount(cmd))
	}

	obj := store.Get(key)

	if obj == nil {
		obj = hash.NewHash()
		store.Put(key, obj)
	}

	hash, ok := hash.GetIfTypeHash(obj)
	if !ok {
		return MakeEvalError(ds.ErrWrongTypeOperation)
	}

	var count int64 = 0

	for i := 1; i < len(args); i += 2 {
		field := args[i]
		value := args[i+1]
		if hash.Add(field, value, -1) {
			count++
		}

	}
	switch cmd {
	case "HSET":
		return MakeEvalResult(count)
	default:
		return MakeEvalResult(clientio.OK)
	}
}

func evalHSET(args []string, store *dstore.Store) *EvalResponse {
	return hSetGeneric(args, store, "HSET")
}

func evalHMSET(args []string, store *dstore.Store) *EvalResponse {
	return hSetGeneric(args, store, "HMSET")
}

func evalHGET(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 2 {
		return MakeEvalError(ds.ErrWrongArgumentCount("HGET"))
	}

	key := args[0]
	field := args[1]

	obj := store.Get(key)
	if obj == nil {
		return MakeEvalResult(clientio.NIL)
	}

	hash, ok := hash.GetIfTypeHash(obj)
	if !ok {
		return MakeEvalError(ds.ErrWrongTypeOperation)
	}

	value, ok := hash.Get(field)
	if !ok {
		return MakeEvalResult(clientio.NIL)
	}

	return MakeEvalResult(value)
}

func evalHMGET(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 2 {
		return MakeEvalError(ds.ErrWrongArgumentCount("HMGET"))
	}

	key := args[0]
	obj := store.Get(key)
	values := make([]interface{}, 0)

	if obj == nil {
		obj = hash.NewHash()
	}
	hash, ok := hash.GetIfTypeHash(obj)
	if !ok {
		return MakeEvalError(ds.ErrWrongTypeOperation)
	}

	for i := 1; i < len(args); i++ {
		value, ok := hash.Get(args[i])
		if ok {
			values = append(values, value)
		} else {
			values = append(values, nil)
		}
	}

	return MakeEvalResult(values)
}

func evalHKEYS(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 1 {
		return MakeEvalError(ds.ErrWrongArgumentCount("HKEYS"))
	}

	key := args[0]
	obj := store.Get(key)
	if obj == nil {
		return MakeEvalResult(clientio.EmptyArray)
	}

	hash, ok := hash.GetIfTypeHash(obj)
	if !ok {
		return MakeEvalError(ds.ErrWrongTypeOperation)
	}

	keys := make([]string, 0, len(hash.Value))
	for key, item := range hash.Value {
		if !item.Expired() {
			keys = append(keys, key)
		}
	}
	return MakeEvalResult(keys)
}

func evalHEXISTS(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 2 {
		return MakeEvalError(ds.ErrWrongArgumentCount("HEXISTS"))
	}

	key := args[0]
	field := args[1]

	obj := store.Get(key)
	if obj == nil {
		return MakeEvalResult(clientio.IntegerZero)
	}

	hash, ok := hash.GetIfTypeHash(obj)
	if !ok {
		return MakeEvalError(ds.ErrWrongTypeOperation)
	}

	if _, ok := hash.Get(field); ok {
		return MakeEvalResult(clientio.IntegerOne)
	}

	return MakeEvalResult(clientio.IntegerZero)
}

func evalHVALS(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 1 {
		return MakeEvalError(ds.ErrWrongArgumentCount("HVALS"))
	}

	key := args[0]
	obj := store.Get(key)
	if obj == nil {
		return MakeEvalResult(clientio.EmptyArray)
	}

	hash, ok := hash.GetIfTypeHash(obj)
	if !ok {
		return MakeEvalError(ds.ErrWrongTypeOperation)
	}

	values := make([]string, 0, len(hash.Value))
	for _, item := range hash.Value {
		val, ok := item.Get()
		if ok {
			values = append(values, val)
		}
	}
	return MakeEvalResult(values)
}

func evalHDEL(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 2 {
		return MakeEvalError(ds.ErrWrongArgumentCount("HDEL"))
	}

	key := args[0]
	obj := store.Get(key)
	if obj == nil {
		return MakeEvalResult(int64(0))
	}

	hash, ok := hash.GetIfTypeHash(obj)
	if !ok {
		return MakeEvalError(ds.ErrWrongTypeOperation)
	}

	var count int64 = 0
	for i := 1; i < len(args); i++ {
		if hash.Delete(args[i]) {
			count++
		}
	}

	return MakeEvalResult(count)
}

func evalHLEN(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 1 {
		return MakeEvalError(ds.ErrWrongArgumentCount("HLEN"))
	}

	key := args[0]
	obj := store.Get(key)
	if obj == nil {
		return MakeEvalResult(clientio.IntegerZero)
	}

	hash, ok := hash.GetIfTypeHash(obj)
	if !ok {
		return MakeEvalError(ds.ErrWrongTypeOperation)
	}

	count := 0
	for _, item := range hash.Value {
		if !item.Expired() {
			count++
		}
	}

	return MakeEvalResult(count)
}

func evalHSTRLEN(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 2 {
		return MakeEvalError(ds.ErrWrongArgumentCount("HSTRLEN"))
	}

	key := args[0]
	field := args[1]

	obj := store.Get(key)
	if obj == nil {
		return MakeEvalResult(clientio.IntegerZero)
	}

	hash, ok := hash.GetIfTypeHash(obj)
	if !ok {
		return MakeEvalError(ds.ErrWrongTypeOperation)
	}

	value, ok := hash.Get(field)
	if !ok {
		return MakeEvalResult(clientio.IntegerZero)
	}

	return MakeEvalResult(len(value))
}

func evalHGETALL(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 1 {
		return MakeEvalError(ds.ErrWrongArgumentCount("HGETALL"))
	}

	key := args[0]
	obj := store.Get(key)
	if obj == nil {
		return MakeEvalResult([]string{})
	}

	hash, ok := hash.GetIfTypeHash(obj)
	if !ok {
		return MakeEvalError(ds.ErrWrongTypeOperation)
	}

	values := make([]string, 0, len(hash.Value)*2)
	for key, item := range hash.Value {
		val, ok := item.Get()
		if ok {
			values = append(values, key)
			values = append(values, val)
		}
	}

	return MakeEvalResult(values)
}

func evalHINCRBY(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 3 {
		return MakeEvalError(ds.ErrWrongArgumentCount("HINCRBY"))
	}

	key := args[0]
	field := args[1]
	increment, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		return MakeEvalError(ds.ErrIntegerOutOfRange)
	}
	obj := store.Get(key)
	if obj == nil {
		obj = hash.NewHash()
	}

	hash, ok := hash.GetIfTypeHash(obj)
	if !ok {
		return MakeEvalError(ds.ErrWrongTypeOperation)
	}

	value, ok := hash.Get(field)
	if !ok {
		value = "0"
	}
	newVal, err := hashIncrementValue(&value, increment)
	if err != nil {
		return MakeEvalError(err)
	}
	hash.Add(field, strconv.FormatInt(newVal, 10), -1)
	return MakeEvalResult(newVal)
}

func evalHINCRBYFLOAT(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 3 {
		return MakeEvalError(ds.ErrWrongArgumentCount("HINCRBYFLOAT"))
	}

	key := args[0]
	field := args[1]
	increment, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		return MakeEvalError(ds.ErrInvalidNumberFormat)
	}
	obj := store.Get(key)
	if obj == nil {
		obj = hash.NewHash()
	}

	hash, ok := hash.GetIfTypeHash(obj)
	if !ok {
		return MakeEvalError(ds.ErrWrongTypeOperation)
	}

	value, ok := hash.Get(field)
	if !ok {
		value = "0"
	}
	newVal, err := hashIncrementFloatValue(&value, increment)
	if err != nil {
		return MakeEvalError(err)
	}
	hash.Add(field, *newVal, -1)
	return MakeEvalResult(*newVal)
}

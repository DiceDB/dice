package eval

import (
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
	}

	hash, ok := hash.GetIfTypeHash(obj)
	if !ok {
		return MakeEvalError(ds.ErrWrongTypeOperation)
	}

	count := 0

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
	if obj == nil {
		return MakeEvalResult(clientio.NIL)
	}

	hash, ok := hash.GetIfTypeHash(obj)
	if !ok {
		return MakeEvalError(ds.ErrWrongTypeOperation)
	}

	values := make([]interface{}, 0)
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

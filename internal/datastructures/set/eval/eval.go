package eval

import (
	"github.com/dicedb/dice/internal/cmd"
	"github.com/dicedb/dice/internal/comm"
	"github.com/dicedb/dice/internal/datastructures/set"
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

package eval

import (
	"github.com/dicedb/dice/internal/cmd"
	ds "github.com/dicedb/dice/internal/datastructures"
	"github.com/dicedb/dice/internal/datastructures/json"

	dstore "github.com/dicedb/dice/internal/store"
)

type Eval struct {
	store *dstore.Store
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

func evalJSONSET(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 2 {
		return MakeEvalError(ds.ErrWrongArgumentCount("JSONSET"))
	}

	key := args[0]
	value := args[1]

	obj := store.Get(key)

	if obj == nil {
		obj = json.NewJSONString()
		store.Put(key, obj)
	}

	jsonString, ok := json.GetIfTypeJSONString(obj)
	if !ok {
		return MakeEvalError(ds.ErrWrongTypeOperation)
	}

	jsonString.Value = value

	return MakeEvalResult(nil)
}

package eval

import (
	"github.com/dicedb/dice/internal/cmd"
	"github.com/dicedb/dice/internal/comm"
	"github.com/dicedb/dice/internal/store"
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

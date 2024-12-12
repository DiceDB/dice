package eval

import (
	"math"
	"strconv"
	"strings"

	"github.com/dicedb/dice/internal/clientio"
	"github.com/dicedb/dice/internal/cmd"
	ds "github.com/dicedb/dice/internal/datastructures"
	"github.com/dicedb/dice/internal/datastructures/sds"
	eval "github.com/dicedb/dice/internal/eval"
	"github.com/dicedb/dice/internal/server/utils"
	"github.com/dicedb/dice/internal/store"
	dstore "github.com/dicedb/dice/internal/store"
)

const (
	defaultExpiry        = -1
	EmptyStr             = ""
	Ex            string = "EX"
	Px            string = "PX"
	Pxat          string = "PXAT"
	Exat          string = "EXAT"
	XX            string = "XX"
	NX            string = "NX"
	KEEPTTL       string = "KEEPTTL"
)

type exDurationState int

const (
	Uninitialized exDurationState = iota
	Initialized
)

type Eval struct {
	store *store.Store
	op    *cmd.DiceDBCmd
}

func NewEval(store *store.Store, op *cmd.DiceDBCmd) *Eval {
	return &Eval{
		store: store,
		op:    op,
	}
}

func (e *Eval) Evaluate() *eval.EvalResponse {
	switch e.op.Cmd {
	case "SET":
		return e.evalSET()
	case "GET":
		return e.evalGET()
	case "DECR":
		return e.evalDECR()
	case "DECRBY":
		return e.evalDECRBY()
	case "INCR":
		return e.evalINCR()
	case "INCRBY":
		return e.evalINCRBY()
	default:
		return nil
	}
}

func (e *Eval) evalSET() *eval.EvalResponse {
	args := e.op.Args
	if len(args) <= 1 {
		return eval.MakeEvalError(ds.ErrWrongArgumentCount("SET"))
	}

	key := e.op.Args[0]
	value := e.op.Args[1]
	var exDurationMs int64 = -1
	var state exDurationState = Uninitialized
	var keepttl bool = false
	var oldVal *interface{}

	key, value = args[0], args[1]

	for i := 2; i < len(args); i++ {
		arg := strings.ToUpper(args[i])
		switch arg {
		case Ex, Px:
			if state != Uninitialized {
				return eval.MakeEvalError(ds.ErrSyntax)
			}
			if keepttl {
				return eval.MakeEvalError(ds.ErrSyntax)
			}
			i++
			if i == len(args) {
				return eval.MakeEvalError(ds.ErrSyntax)
			}

			exDuration, err := strconv.ParseInt(args[i], 10, 64)
			if err != nil {
				return eval.MakeEvalError(ds.ErrIntegerOutOfRange)
			}

			if exDuration <= 0 || exDuration >= ds.MaxExDuration {
				return eval.MakeEvalError(ds.ErrInvalidExpireTime("SET"))
			}

			// converting seconds to milliseconds
			if arg == Ex {
				exDuration *= 1000
			}
			exDurationMs = exDuration
			state = Initialized

		case Pxat, Exat:
			if state != Uninitialized {
				return eval.MakeEvalError(ds.ErrSyntax)
			}
			if keepttl {
				return eval.MakeEvalError(ds.ErrSyntax)
			}
			i++
			if i == len(args) {
				return eval.MakeEvalError(ds.ErrSyntax)
			}
			exDuration, err := strconv.ParseInt(args[i], 10, 64)
			if err != nil {
				return eval.MakeEvalError(ds.ErrIntegerOutOfRange)
			}

			if exDuration < 0 {
				return eval.MakeEvalError(ds.ErrInvalidExpireTime("SET"))
			}

			if arg == Exat {
				exDuration *= 1000
			}
			exDurationMs = exDuration - utils.GetCurrentTime().UnixMilli()
			// If the expiry time is in the past, set exDurationMs to 0
			// This will be used to signal immediate expiration
			if exDurationMs < 0 {
				exDurationMs = 0
			}
			state = Initialized

		case XX:
			// Get the key from the hash table
			obj := e.store.Get(key)

			// if key does not exist, return RESP encoded nil
			if obj == nil {
				return eval.MakeEvalResult(clientio.NIL)
			}
		case NX:
			obj := e.store.Get(key)
			if obj != nil {
				return eval.MakeEvalResult(clientio.NIL)
			}
		case ds.KeepTTL:
			if state != Uninitialized {
				return eval.MakeEvalError(ds.ErrSyntax)
			}
			keepttl = true
		case ds.GET:
			getResult := e.evalGET()
			if getResult.Error != nil {
				return eval.MakeEvalError(ds.ErrWrongTypeOperation)
			}
			oldVal = &getResult.Result
		default:
			return eval.MakeEvalError(ds.ErrSyntax)
		}
	}

	// putting the k and value in a Hash Table
	e.store.Put(key, sds.NewString(value), dstore.WithKeepTTL(keepttl))

	if oldVal != nil {
		return eval.MakeEvalResult(*oldVal)
	}
	return eval.MakeEvalResult(clientio.OK)
}

func (e *Eval) evalGET() *eval.EvalResponse {
	args := e.op.Args
	if len(args) != 1 {
		return eval.MakeEvalError(ds.ErrWrongArgumentCount("GET"))
	}

	key := args[0]

	obj := e.store.Get(key)

	// if key does not exist, return RESP encoded nil
	if obj == nil {
		return eval.MakeEvalResult(clientio.NIL)
	}

	if sds, ok := sds.GetIfTypeSDS(obj); ok {
		return eval.MakeEvalResult(sds.Get())
	}
	return eval.MakeEvalError(ds.ErrWrongTypeOperation)
}

func (e *Eval) evalDECR() *eval.EvalResponse {
	if len(e.op.Args) != 1 {
		return eval.MakeEvalError(ds.ErrWrongArgumentCount("DECR"))
	}

	return e.incrDecrCmd(-1)
}

func (e *Eval) evalDECRBY() *eval.EvalResponse {
	args := e.op.Args
	if len(args) != 2 {
		return eval.MakeEvalError(ds.ErrWrongArgumentCount("DECRBY"))
	}
	decrementAmount, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return eval.MakeEvalError(ds.ErrIntegerOutOfRange)
	}
	return e.incrDecrCmd(-decrementAmount)
}

func (e *Eval) evalINCR() *eval.EvalResponse {
	if len(e.op.Args) != 1 {
		return eval.MakeEvalError(ds.ErrWrongArgumentCount("INCR"))
	}

	return e.incrDecrCmd(1)
}

func (e *Eval) evalINCRBY() *eval.EvalResponse {
	args := e.op.Args
	if len(args) != 2 {
		return eval.MakeEvalError(ds.ErrWrongArgumentCount("INCRBY"))
	}
	incrAmount, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return eval.MakeEvalError(ds.ErrIntegerOutOfRange)
	}
	return e.incrDecrCmd(incrAmount)
}

func (e *Eval) incrDecrCmd(incr int64) *eval.EvalResponse {
	key := e.op.Args[0]
	obj := e.store.Get(key)
	if obj == nil {
		e.store.Put(key, sds.NewString("1"), dstore.WithKeepTTL(false))
	}

	// increment the value if it is an integer
	var sdsObj sds.SDSInterface
	var ok bool
	if sdsObj, ok = sds.GetIfTypeSDS(obj); !ok {
		return eval.MakeEvalError(ds.ErrWrongTypeOperation)
	}
	i, err := strconv.ParseInt(sdsObj.Get(), 10, 64)
	if err != nil {
		return eval.MakeEvalError(ds.ErrWrongTypeOperation)
	}

	// check overflow
	if (incr < 0 && i < 0 && incr < (math.MinInt64-i)) ||
		(incr > 0 && i > 0 && incr > (math.MaxInt64-i)) {
		return eval.MakeEvalError(ds.ErrOverflow)
	}

	i += incr
	err = sdsObj.Set(strconv.FormatInt(i, 10))
	return eval.MakeEvalResult(i)
}

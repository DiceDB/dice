package eval

import (
	"math"
	"strconv"
	"strings"
	"time"

	ds "github.com/dicedb/dice/internal/datastructures"
	"github.com/dicedb/dice/internal/store"
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
	store *store.Store[ds.DSInterface]
	op    *ops.Operation
}

func NewEval(store *store.Store[ds.DSInterface], op *ops.Operation) *Eval {
	return &Eval{
		store: store,
		op:    op,
	}
}

func (e *Eval) Evaluate() []byte {
	switch e.op.Cmd {
	case "SET":
		return e.set()
	case "GET":
		return e.get()
	case "INCR":
		return e.incrementBy()
	case "INCRBY":
		return e.increment()
	case "DECR":
		return e.decrement()
	case "DECRBY":
		return e.decrementBy()
	default:
		return nil
	}
}

func (e *Eval) set() []byte {
	// Ensure the number of arguments is correct
	if len(e.op.Args) <= 1 {
		return diceerrors.NewErrArity("GET")
	}

	state := Uninitialized
	var exDurationMs int64 = -1
	args := e.op.Args
	//var keepttl bool = false
	key := e.op.Args[0]
	value := e.op.Args[1]
	_, exists := e.store.Get(key)

	for i := 2; i < len(args); i++ {
		arg := strings.ToUpper(args[i])
		switch arg {
		case Ex, Px:
			if state != Uninitialized {
				return diceerrors.NewErrWithMessage(diceerrors.SyntaxErr)
			}
			i++
			if i == len(args) {
				return diceerrors.NewErrWithMessage(diceerrors.SyntaxErr)
			}

			exDuration, err := strconv.ParseInt(args[i], 10, 64)
			if err != nil {
				return diceerrors.NewErrWithMessage(diceerrors.IntOrOutOfRangeErr)
			}

			if exDuration <= 0 {
				return diceerrors.NewErrExpireTime("SET")
			}

			// converting seconds to milliseconds
			if arg == Ex {
				exDuration *= 1000
			}
			exDurationMs = exDuration
			state = Initialized

		case Pxat, Exat:
			if state != Uninitialized {
				return diceerrors.NewErrWithMessage(diceerrors.SyntaxErr)
			}
			i++
			if i == len(args) {
				return diceerrors.NewErrWithMessage(diceerrors.SyntaxErr)
			}
			exDuration, err := strconv.ParseInt(args[i], 10, 64)
			if err != nil {
				return diceerrors.NewErrWithMessage(diceerrors.IntOrOutOfRangeErr)
			}

			if exDuration < 0 {
				return diceerrors.NewErrExpireTime("SET")
			}

			if arg == Exat {
				exDuration *= 1000
			}
			exDurationMs = exDuration - time.Now().UnixMilli()
			// If the expiry time is in the past, set exDurationMs to 0
			// This will be used to signal immediate expiration
			if exDurationMs < 0 {
				exDurationMs = 0
			}
			state = Initialized

		case XX, NX:
			// if key does not exist, return RESP encoded nil
			if !exists {
				return eval.RespNIL
			}
		case KEEPTTL:
			//keepttl = true
		default:
			return diceerrors.NewErrWithMessage(diceerrors.SyntaxErr)
		}
	}

	// Cast the value properly based on the encoding type
	e.store.Put(key, NewString(value), exDurationMs)

	return eval.RespOK
}

func (e *Eval) get() []byte {
	args := e.op.Args
	if len(args) != 1 {
		return diceerrors.NewErrArity("GET")
	}

	var key = args[0]

	obj, ok := e.store.Get(key)

	// if key does not exist, return RESP encoded nil
	if !ok {
		return eval.RespNIL
	}

	// Decode and return the value based on its encoding
	if sds, ok := getIfTypeSDS(obj); ok {
		return []byte(eval.Encode(sds.Get(), false))
	}

	return diceerrors.NewErrWithMessage(diceerrors.WrongTypeErr)
}

func (e *Eval) decrement() []byte {
	if len(e.op.Args) != 1 {
		return diceerrors.NewErrArity("DECR")
	}

	return e.incrDecrCmd(-1)
}

func (e *Eval) decrementBy() []byte {
	args := e.op.Args
	if len(args) != 2 {
		return diceerrors.NewErrArity("DECRBY")
	}
	decrementAmount, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return diceerrors.NewErrWithMessage(diceerrors.IntOrOutOfRangeErr)
	}
	return e.incrDecrCmd(-decrementAmount)
}

func (e *Eval) increment() []byte {
	if len(e.op.Args) != 1 {
		return diceerrors.NewErrArity("DECR")
	}

	return e.incrDecrCmd(1)
}

func (e *Eval) incrementBy() []byte {
	args := e.op.Args
	if len(args) != 2 {
		return diceerrors.NewErrArity("INCRBY")
	}
	incrAmount, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return diceerrors.NewErrWithMessage(diceerrors.IntOrOutOfRangeErr)
	}
	return e.incrDecrCmd(incrAmount)
}

func (e *Eval) incrDecrCmd(incr int64) []byte {
	key := e.op.Args[0]
	obj, ok := e.store.Get(key)
	if !ok {
		e.store.Put(key, NewString("1"), defaultExpiry)
	}

	// increment the value if it is an integer
	var sds SDSInterface
	if sds, ok = getIfTypeSDS(obj); !ok {
		return diceerrors.NewErrWithMessage(diceerrors.WrongTypeErr)
	}
	i, err := strconv.ParseInt(sds.Get(), 10, 64)
	if err != nil {
		return diceerrors.NewErrWithMessage(diceerrors.WrongTypeErr)
	}

	// check overflow
	if (incr < 0 && i < 0 && incr < (math.MinInt64-i)) ||
		(incr > 0 && i > 0 && incr > (math.MaxInt64-i)) {
		return diceerrors.NewErrWithMessage(diceerrors.ValOutOfRangeErr)
	}

	i += incr
	err = sds.Set(strconv.FormatInt(i, 10))
	if err != nil {
		return diceerrors.NewErrWithMessage(diceerrors.ValOutOfRangeErr)
	}

	return eval.Encode(i, false)
}

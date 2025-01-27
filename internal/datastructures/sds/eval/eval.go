package eval

import (
	"fmt"
	"math"
	"math/bits"
	"strconv"
	"strings"

	"github.com/dicedb/dice/internal/clientio"
	ds "github.com/dicedb/dice/internal/datastructures"
	"github.com/dicedb/dice/internal/datastructures/sds"
	"github.com/dicedb/dice/internal/server/utils"
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

func (e *Eval) Evaluate() *ds.EvalResponse {
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
	case "SETBIT":
		return e.evalSETBIT()
	case "GETBIT":
		return e.evalGETBIT()
	case "BITCOUNT":
		return e.evalBITCOUNT()
	default:
		return nil
	}
}

func (e *Eval) evalSET() *ds.EvalResponse {
	args := e.op.Args
	if len(args) <= 1 {
		return ds.MakeEvalError(ds.ErrWrongArgumentCount("SET"))
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
				return ds.MakeEvalError(ds.ErrSyntax)
			}
			if keepttl {
				return ds.MakeEvalError(ds.ErrSyntax)
			}
			i++
			if i == len(args) {
				return ds.MakeEvalError(ds.ErrSyntax)
			}

			exDuration, err := strconv.ParseInt(args[i], 10, 64)
			if err != nil {
				return ds.MakeEvalError(ds.ErrIntegerOutOfRange)
			}

			if exDuration <= 0 || exDuration >= ds.MaxExDuration {
				return ds.MakeEvalError(ds.ErrInvalidExpireTime("SET"))
			}

			// converting seconds to milliseconds
			if arg == Ex {
				exDuration *= 1000
			}
			exDurationMs = exDuration
			state = Initialized

		case Pxat, Exat:
			if state != Uninitialized {
				return ds.MakeEvalError(ds.ErrSyntax)
			}
			if keepttl {
				return ds.MakeEvalError(ds.ErrSyntax)
			}
			i++
			if i == len(args) {
				return ds.MakeEvalError(ds.ErrSyntax)
			}
			exDuration, err := strconv.ParseInt(args[i], 10, 64)
			if err != nil {
				return ds.MakeEvalError(ds.ErrIntegerOutOfRange)
			}

			if exDuration < 0 {
				return ds.MakeEvalError(ds.ErrInvalidExpireTime("SET"))
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
				return ds.MakeEvalResult(clientio.NIL)
			}
		case NX:
			obj := e.store.Get(key)
			if obj != nil {
				return ds.MakeEvalResult(clientio.NIL)
			}
		case ds.KeepTTL:
			if state != Uninitialized {
				return ds.MakeEvalError(ds.ErrSyntax)
			}
			keepttl = true
		case ds.GET:
			getResult := e.evalGET()
			if getResult.Error != nil {
				return ds.MakeEvalError(ds.ErrWrongTypeOperation)
			}
			oldVal = &getResult.Result
		default:
			return ds.MakeEvalError(ds.ErrSyntax)
		}
	}

	// putting the k and value in a Hash Table
	e.store.Put(key, sds.NewString(value), dstore.WithKeepTTL(keepttl))

	if oldVal != nil {
		return ds.MakeEvalResult(*oldVal)
	}
	return ds.MakeEvalResult(clientio.OK)
}

func (e *Eval) evalGET() *ds.EvalResponse {
	args := e.op.Args
	if len(args) != 1 {
		return ds.MakeEvalError(ds.ErrWrongArgumentCount("GET"))
	}

	key := args[0]

	obj := e.store.Get(key)

	// if key does not exist, return RESP encoded nil
	if obj == nil {
		return ds.MakeEvalResult(clientio.NIL)
	}

	if sds, ok := sds.GetIfTypeSDS(obj); ok {
		return ds.MakeEvalResult(sds.Get())
	}
	return ds.MakeEvalError(ds.ErrWrongTypeOperation)
}

func (e *Eval) evalDECR() *ds.EvalResponse {
	if len(e.op.Args) != 1 {
		return ds.MakeEvalError(ds.ErrWrongArgumentCount("DECR"))
	}

	return e.incrDecrCmd(-1)
}

func (e *Eval) evalDECRBY() *ds.EvalResponse {
	args := e.op.Args
	if len(args) != 2 {
		return ds.MakeEvalError(ds.ErrWrongArgumentCount("DECRBY"))
	}
	decrementAmount, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return ds.MakeEvalError(ds.ErrIntegerOutOfRange)
	}
	return e.incrDecrCmd(-decrementAmount)
}

func (e *Eval) evalINCR() *ds.EvalResponse {
	if len(e.op.Args) != 1 {
		return ds.MakeEvalError(ds.ErrWrongArgumentCount("INCR"))
	}

	return e.incrDecrCmd(1)
}

func (e *Eval) evalINCRBY() *ds.EvalResponse {
	args := e.op.Args
	if len(args) != 2 {
		return ds.MakeEvalError(ds.ErrWrongArgumentCount("INCRBY"))
	}
	incrAmount, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return ds.MakeEvalError(ds.ErrIntegerOutOfRange)
	}
	return e.incrDecrCmd(incrAmount)
}

func (e *Eval) incrDecrCmd(incr int64) *ds.EvalResponse {
	key := e.op.Args[0]
	obj := e.store.Get(key)
	if obj == nil {
		e.store.Put(key, sds.NewString("1"), dstore.WithKeepTTL(false))
	}

	// increment the value if it is an integer
	var sdsObj sds.SDSInterface
	var ok bool
	if sdsObj, ok = sds.GetIfTypeSDS(obj); !ok {
		return ds.MakeEvalError(ds.ErrWrongTypeOperation)
	}
	i, err := strconv.ParseInt(sdsObj.Get(), 10, 64)
	if err != nil {
		return ds.MakeEvalError(ds.ErrWrongTypeOperation)
	}

	// check overflow
	if (incr < 0 && i < 0 && incr < (math.MinInt64-i)) ||
		(incr > 0 && i > 0 && incr > (math.MaxInt64-i)) {
		return ds.MakeEvalError(ds.ErrOverflow)
	}

	i += incr
	err = sdsObj.Set(strconv.FormatInt(i, 10))
	return ds.MakeEvalResult(i)
}

func (e *Eval) evalSETBIT() *ds.EvalResponse {
	var err error
	fmt.Println(e.op.Args)
	if len(e.op.Args) != 3 {
		return &ds.EvalResponse{
			Result: nil,
			Error:  ds.ErrWrongArgumentCount("SETBIT"),
		}
	}

	key := e.op.Args[0]
	offset, err := strconv.ParseInt(e.op.Args[1], 10, 64)
	if err != nil || offset < 0 {
		return &ds.EvalResponse{
			Result: nil,
			Error:  ds.ErrGeneral("bit offset is not an integer or out of range"),
		}
	}

	value, err := strconv.ParseBool(e.op.Args[2])

	if err != nil {
		return &ds.EvalResponse{
			Result: nil,
			Error:  ds.ErrGeneral("bit is not an integer or out of range"),
		}
	}

	obj := e.store.Get(key)
	requiredByteArraySize := offset>>3 + 1

	if obj == nil {
		bytes := make([]byte, requiredByteArraySize)
		str := sds.NewString(string(bytes))

		obj = e.store.NewObj(str, -1)
		e.store.Put(key, obj)
	}

	byteArray, ok := sds.GetIfTypeSDS(obj)

	if !ok {
		return ds.MakeEvalError(ds.ErrWrongTypeOperation)
	}

	byteArrayLength := int64(len(byteArray.Get()))

	// check whether resize required or not
	if requiredByteArraySize > byteArrayLength {
		// resize as per the offset
		byteArray.Resize(requiredByteArraySize)
	}
	resp := byteArray.GetBit(offset)
	byteArray.SetBit(offset, value)

	if err != nil {
		return ds.MakeEvalError(ds.ErrWrongTypeOperation)
	}

	if resp == 1 {
		return &ds.EvalResponse{
			Result: clientio.IntegerOne,
			Error:  nil,
		}
	}
	return &ds.EvalResponse{
		Result: clientio.IntegerZero,
		Error:  nil,
	}
}

func (e *Eval) evalGETBIT() *ds.EvalResponse {
	if len(e.op.Args) != 2 {
		return &ds.EvalResponse{
			Result: nil,
			Error:  ds.ErrWrongArgumentCount("GETBIT"),
		}
	}

	key := e.op.Args[0]
	offset, err := strconv.ParseInt(e.op.Args[1], 10, 64)
	if err != nil || offset < 0 {
		return &ds.EvalResponse{
			Result: nil,
			Error:  ds.ErrGeneral("bit offset is not an integer or out of range"),
		}
	}

	obj := e.store.Get(key)

	if obj == nil {
		return &ds.EvalResponse{
			Result: clientio.IntegerZero,
			Error:  nil,
		}
	}

	byteArray, ok := sds.GetIfTypeSDS(obj)

	if !ok {
		return ds.MakeEvalError(ds.ErrWrongTypeOperation)
	}

	resp := byteArray.GetBit(offset)

	if resp == 1 {
		return &ds.EvalResponse{
			Result: clientio.IntegerOne,
			Error:  nil,
		}
	}
	return &ds.EvalResponse{
		Result: clientio.IntegerZero,
		Error:  nil,
	}
}

func (e *Eval) evalBITCOUNT() *ds.EvalResponse {
	var err error
	if len(e.op.Args) == 0 || len(e.op.Args) > 4 {
		return ds.MakeEvalError(ds.ErrWrongArgumentCount("BITCOUNT"))
	}

	key := e.op.Args[0]
	obj := e.store.Get(key)

	if obj == nil {
		return ds.MakeEvalResult(clientio.IntegerZero)
	}

	byteArray, ok := sds.GetIfTypeSDS(obj)

	if !ok {
		return ds.MakeEvalError(ds.ErrWrongTypeOperation)
	}
	value := []byte(byteArray.Get())
	valueLength := int64(len(value))
	start, end := int64(0), valueLength-1

	unit := ds.BYTE

	if len(e.op.Args) > 1 {
		start, err = strconv.ParseInt(e.op.Args[1], 10, 64)
		if err != nil {
			return ds.MakeEvalError(ds.ErrIntegerOutOfRange)
		}
		if len(e.op.Args) <= 2 {
			return ds.MakeEvalError(ds.ErrSyntax)
		}

		end, err = strconv.ParseInt(e.op.Args[2], 10, 64)
		if err != nil {
			return ds.MakeEvalError(ds.ErrIntegerOutOfRange)
		}
	}

	if len(e.op.Args) > 3 {
		unit = strings.ToUpper(e.op.Args[3])
	}

	switch unit {
	case ds.BYTE:
		if start < 0 {
			start += valueLength
		}
		if end < 0 {
			end += valueLength
		}
		if start > end || start >= valueLength {
			return ds.MakeEvalResult(clientio.IntegerZero)
		}
		end = min(end, valueLength-1)
		bitCount := 0
		for i := start; i <= end; i++ {
			bitCount += bits.OnesCount8(value[i])
		}
		return ds.MakeEvalResult(bitCount)
	case ds.BIT:
		if start < 0 {
			start += valueLength * 8
		}
		if end < 0 {
			end += valueLength * 8
		}
		if start > end {
			return ds.MakeEvalResult(clientio.IntegerZero)
		}
		startByte, endByte := start/8, min(end/8, valueLength-1)
		startBitOffset, endBitOffset := start%8, end%8
		if endByte == valueLength-1 {
			endBitOffset = 7
		}
		if startByte >= valueLength {
			return ds.MakeEvalResult(clientio.IntegerZero)
		}

		bitCount := 0
		if startByte == endByte {
			mask := byte(0xFF >> startBitOffset)
			mask &= byte(0xFF << (7 - endBitOffset))
			bitCount = bits.OnesCount8(value[startByte] & mask)
		} else {
			firstByteMask := byte(0xFF >> startBitOffset)
			bitCount += bits.OnesCount8(value[startByte] & firstByteMask)

			for i := startByte + 1; i < endByte; i++ {
				bitCount += bits.OnesCount8(value[i])
			}

			lastByteMask := byte(0xFF << (7 - endBitOffset))
			bitCount += bits.OnesCount8(value[endByte] & lastByteMask)
		}
		return ds.MakeEvalResult(bitCount)
	default:
		return ds.MakeEvalError(ds.ErrSyntax)
	}
}

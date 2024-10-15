package eval

import (
	"math"
	"math/bits"
	"strconv"
	"strings"

	"github.com/dicedb/dice/internal/clientio"
	diceerrors "github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/eval/sortedset"
	"github.com/dicedb/dice/internal/object"
	"github.com/dicedb/dice/internal/server/utils"
	dstore "github.com/dicedb/dice/internal/store"
)

// evalSET puts a new <key, value> pair in db as in the args
// args must contain key and value.
// args can also contain multiple options -
//
//	EX or ex which will set the expiry time(in secs) for the key
//	PX or px which will set the expiry time(in milliseconds) for the key
//	EXAT or exat which will set the specified Unix time at which the key will expire, in seconds (a positive integer).
//	PXAT or PX which will the specified Unix time at which the key will expire, in milliseconds (a positive integer).
//	XX orr xx which will only set the key if it already exists.
//
// Returns encoded error response if at least a <key, value> pair is not part of args
// Returns encoded error response if expiry time value in not integer
// Returns encoded error response if both PX and EX flags are present
// Returns encoded OK RESP once new entry is added
// If the key already exists then the value will be overwritten and expiry will be discarded
func evalSET(args []string, store *dstore.Store) *EvalResponse {
	if len(args) <= 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("SET"),
		}
	}

	var key, value string
	var exDurationMs int64 = -1
	var state exDurationState = Uninitialized
	var keepttl bool = false

	key, value = args[0], args[1]
	oType, oEnc := deduceTypeEncoding(value)

	for i := 2; i < len(args); i++ {
		arg := strings.ToUpper(args[i])
		switch arg {
		case Ex, Px:
			if state != Uninitialized {
				return &EvalResponse{
					Result: nil,
					Error:  diceerrors.ErrSyntax,
				}
			}
			i++
			if i == len(args) {
				return &EvalResponse{
					Result: nil,
					Error:  diceerrors.ErrSyntax,
				}
			}

			exDuration, err := strconv.ParseInt(args[i], 10, 64)
			if err != nil {
				return &EvalResponse{
					Result: nil,
					Error:  diceerrors.ErrIntegerOutOfRange,
				}
			}

			if exDuration <= 0 || exDuration >= maxExDuration {
				return &EvalResponse{
					Result: nil,
					Error:  diceerrors.ErrInvalidExpireTime("SET"),
				}
			}

			// converting seconds to milliseconds
			if arg == Ex {
				exDuration *= 1000
			}
			exDurationMs = exDuration
			state = Initialized

		case Pxat, Exat:
			if state != Uninitialized {
				return &EvalResponse{
					Result: nil,
					Error:  diceerrors.ErrSyntax,
				}
			}
			i++
			if i == len(args) {
				return &EvalResponse{
					Result: nil,
					Error:  diceerrors.ErrSyntax,
				}
			}
			exDuration, err := strconv.ParseInt(args[i], 10, 64)
			if err != nil {
				return &EvalResponse{
					Result: nil,
					Error:  diceerrors.ErrIntegerOutOfRange,
				}
			}

			if exDuration < 0 {
				return &EvalResponse{
					Result: nil,
					Error:  diceerrors.ErrInvalidExpireTime("SET"),
				}
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
			obj := store.Get(key)

			// if key does not exist, return RESP encoded nil
			if obj == nil {
				return &EvalResponse{
					Result: clientio.NIL,
					Error:  nil,
				}
			}
		case NX:
			obj := store.Get(key)
			if obj != nil {
				return &EvalResponse{
					Result: clientio.NIL,
					Error:  nil,
				}
			}
		case KeepTTL:
			keepttl = true
		default:
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrSyntax,
			}
		}
	}

	// Cast the value properly based on the encoding type
	var storedValue interface{}
	switch oEnc {
	case object.ObjEncodingInt:
		storedValue, _ = strconv.ParseInt(value, 10, 64)
	case object.ObjEncodingEmbStr, object.ObjEncodingRaw:
		storedValue = value
	default:
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrUnsupportedEncoding(int(oEnc)),
		}
	}

	// putting the k and value in a Hash Table
	store.Put(key, store.NewObj(storedValue, exDurationMs, oType, oEnc), dstore.WithKeepTTL(keepttl))

	return &EvalResponse{
		Result: clientio.OK,
		Error:  nil,
	}
}

// evalGET returns the value for the queried key in args
// The key should be the only param in args
// The RESP value of the key is encoded and then returned
// evalGET returns response.clientio.NIL if key is expired or it does not exist
func evalGET(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("GET"),
		}
	}

	key := args[0]

	obj := store.Get(key)

	// if key does not exist, return RESP encoded nil
	if obj == nil {
		return &EvalResponse{
			Result: clientio.NIL,
			Error:  nil,
		}
	}

	// Decode and return the value based on its encoding
	switch _, oEnc := object.ExtractTypeEncoding(obj); oEnc {
	case object.ObjEncodingInt:
		// Value is stored as an int64, so use type assertion
		if val, ok := obj.Value.(int64); ok {
			return &EvalResponse{
				Result: val,
				Error:  nil,
			}
		}

		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrUnexpectedType("int64", obj.Value),
		}

	case object.ObjEncodingEmbStr, object.ObjEncodingRaw:
		// Value is stored as a string, use type assertion
		if val, ok := obj.Value.(string); ok {
			return &EvalResponse{
				Result: val,
				Error:  nil,
			}
		}
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrUnexpectedType("string", obj.Value),
		}

	case object.ObjEncodingByteArray:
		// Value is stored as a bytearray, use type assertion
		if val, ok := obj.Value.(*ByteArray); ok {
			return &EvalResponse{
				Result: string(val.data),
				Error:  nil,
			}
		}

		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}

	default:
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}
}

// GETSET atomically sets key to value and returns the old value stored at key.
// Returns an error when key exists but does not hold a string value.
// Any previous time to live associated with the key is
// discarded on successful SET operation.
//
// Returns:
// Bulk string reply: the old value stored at the key.
// Nil reply: if the key does not exist.
func evalGETSET(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 2 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("GETSET"),
		}
	}

	key, value := args[0], args[1]
	getResp := evalGET([]string{key}, store)
	// Check if it's an error resp from GET
	if getResp.Error != nil {
		return getResp
	}

	// Previous TTL needs to be reset
	setResp := evalSET([]string{key, value}, store)
	// Check if it's an error resp from SET
	if setResp.Error != nil {
		return setResp
	}

	return getResp
}

// evalSETEX puts a new <key, value> pair in db as in the args
// args must contain only  key , expiry and value
// Returns encoded error response if <key,exp,value> is not part of args
// Returns encoded error response if expiry time value in not integer
// Returns encoded OK RESP once new entry is added
// If the key already exists then the value and expiry will be overwritten
func evalSETEX(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 3 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("SETEX"),
		}
	}

	var key, value string
	key, value = args[0], args[2]

	exDuration, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrIntegerOutOfRange,
		}
	}
	if exDuration <= 0 || exDuration >= maxExDuration {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrInvalidExpireTime("SETEX"),
		}
	}
	newArgs := []string{key, value, Ex, args[1]}

	return evalSET(newArgs, store)
}

// evalZADD adds all the specified members with the specified scores to the sorted set stored at key.
// If a specified member is already a member of the sorted set, the score is updated and the element
// reinserted at the right position to ensure the correct ordering.
// If key does not exist, a new sorted set with the specified members as sole members is created.
func evalZADD(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 3 || len(args)%2 == 0 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("ZADD"),
		}
	}

	key := args[0]
	obj := store.Get(key)

	var sortedSet *sortedset.Set

	if obj != nil {
		var err []byte
		sortedSet, err = sortedset.FromObject(obj)
		if err != nil {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongTypeOperation,
			}
		}
	} else {
		sortedSet = sortedset.New()
	}

	added := 0
	for i := 1; i < len(args); i += 2 {
		scoreStr := args[i]
		member := args[i+1]

		score, err := strconv.ParseFloat(scoreStr, 64)
		if err != nil || math.IsNaN(score) {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrInvalidNumberFormat,
			}
		}

		wasInserted := sortedSet.Upsert(score, member)

		if wasInserted {
			added += 1
		}
	}

	obj = store.NewObj(sortedSet, -1, object.ObjTypeSortedSet, object.ObjEncodingBTree)
	store.Put(key, obj, dstore.WithPutCmd(dstore.ZAdd))

	return &EvalResponse{
		Result: added,
		Error:  nil,
	}
}

// evalZRANGE returns the specified range of elements in the sorted set stored at key.
// The elements are considered to be ordered from the lowest to the highest score.
func evalZRANGE(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 3 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("ZRANGE"),
		}
	}

	key := args[0]
	startStr := args[1]
	stopStr := args[2]

	withScores := false
	reverse := false
	for i := 3; i < len(args); i++ {
		arg := strings.ToUpper(args[i])
		if arg == WithScores {
			withScores = true
		} else if arg == REV {
			reverse = true
		} else {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrSyntax,
			}
		}
	}

	start, err := strconv.Atoi(startStr)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrInvalidNumberFormat,
		}
	}

	stop, err := strconv.Atoi(stopStr)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrInvalidNumberFormat,
		}
	}

	obj := store.Get(key)
	if obj == nil {
		return &EvalResponse{
			Result: []string{},
			Error:  nil,
		}
	}

	sortedSet, errMsg := sortedset.FromObject(obj)

	if errMsg != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	result := sortedSet.GetRange(start, stop, withScores, reverse)

	return &EvalResponse{
		Result: result,
		Error:  nil,
	}
}

// SETBIT key offset value
func evalSETBIT(args []string, store *dstore.Store) *EvalResponse {
	var err error

	if len(args) != 3 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("SETBIT"),
		}
	}

	key := args[0]
	offset, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral("bit offset is not an integer or out of range"),
		}
	}

	value, err := strconv.ParseBool(args[2])
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral("bit is not an integer or out of range"),
		}
	}

	obj := store.Get(key)
	requiredByteArraySize := offset>>3 + 1

	if obj == nil {
		obj = store.NewObj(NewByteArray(int(requiredByteArraySize)), -1, object.ObjTypeByteArray, object.ObjEncodingByteArray)
		store.Put(args[0], obj)
	}

	if object.AssertType(obj.TypeEncoding, object.ObjTypeByteArray) == nil ||
		object.AssertType(obj.TypeEncoding, object.ObjTypeString) == nil ||
		object.AssertType(obj.TypeEncoding, object.ObjTypeInt) == nil {
		var byteArray *ByteArray
		oType, oEnc := object.ExtractTypeEncoding(obj)

		switch oType {
		case object.ObjTypeByteArray:
			byteArray = obj.Value.(*ByteArray)
		case object.ObjTypeString, object.ObjTypeInt:
			byteArray, err = NewByteArrayFromObj(obj)
			if err != nil {
				return &EvalResponse{
					Result: nil,
					Error:  diceerrors.ErrWrongTypeOperation,
				}
			}
		default:
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongTypeOperation,
			}
		}

		// Perform the resizing check
		byteArrayLength := byteArray.Length

		// check whether resize required or not
		if requiredByteArraySize > byteArrayLength {
			// resize as per the offset
			byteArray = byteArray.IncreaseSize(int(requiredByteArraySize))
		}

		resp := byteArray.GetBit(int(offset))
		byteArray.SetBit(int(offset), value)

		// We are returning newObject here so it is thread-safe
		// Old will be removed by GC
		newObj, err := ByteSliceToObj(store, obj, byteArray.data, oType, oEnc)
		if err != nil {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongTypeOperation,
			}
		}

		exp, ok := dstore.GetExpiry(obj, store)
		var exDurationMs int64 = -1
		if ok {
			exDurationMs = int64(exp - uint64(utils.GetCurrentTime().UnixMilli()))
		}
		// newObj has bydefault expiry time -1 , we need to set it
		if exDurationMs > 0 {
			store.SetExpiry(newObj, exDurationMs)
		}

		store.Put(key, newObj)
		if resp {
			return &EvalResponse{
				Result: clientio.IntegerOne,
				Error:  nil,
			}
		}
		return &EvalResponse{
			Result: clientio.IntegerZero,
			Error:  nil,
		}
	}
	return &EvalResponse{
		Result: nil,
		Error:  diceerrors.ErrWrongTypeOperation,
	}
}

// GETBIT key offset
func evalGETBIT(args []string, store *dstore.Store) *EvalResponse {
	var err error

	if len(args) != 2 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("GETBIT"),
		}
	}

	key := args[0]
	offset, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrIntegerOutOfRange,
		}
	}

	obj := store.Get(key)
	if obj == nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	requiredByteArraySize := offset>>3 + 1
	switch oType, _ := object.ExtractTypeEncoding(obj); oType {
	case object.ObjTypeSet:
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	case object.ObjTypeByteArray:
		byteArray := obj.Value.(*ByteArray)
		byteArrayLength := byteArray.Length

		// check whether offset, length exists or not
		if requiredByteArraySize > byteArrayLength {
			return &EvalResponse{
				Result: clientio.IntegerZero,
				Error:  nil,
			}
		}
		value := byteArray.GetBit(int(offset))
		if value {
			return &EvalResponse{
				Result: clientio.IntegerOne,
				Error:  nil,
			}
		}
		return &EvalResponse{
			Result: clientio.IntegerZero,
			Error:  nil,
		}

	case object.ObjTypeString, object.ObjTypeInt:
		byteArray, err := NewByteArrayFromObj(obj)
		if err != nil {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongTypeOperation,
			}
		}
		if requiredByteArraySize > byteArray.Length {
			return &EvalResponse{
				Result: clientio.IntegerZero,
				Error:  nil,
			}
		}
		value := byteArray.GetBit(int(offset))
		if value {
			return &EvalResponse{
				Result: clientio.IntegerOne,
				Error:  nil,
			}
		}
		return &EvalResponse{
			Result: clientio.IntegerZero,
			Error:  nil,
		}

	default:
		return &EvalResponse{
			Result: clientio.IntegerZero,
			Error:  nil,
		}
	}
}

func evalBITCOUNT(args []string, store *dstore.Store) *EvalResponse {
	var err error

	// if no key is provided, return error
	if len(args) == 0 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("BITCOUNT"),
		}
	}

	// if more than 4 arguments are provided, return error
	if len(args) > 4 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrSyntax,
		}
	}

	// fetching value of the key
	key := args[0]
	obj := store.Get(key)
	if obj == nil {
		return &EvalResponse{
			Result: clientio.IntegerZero,
			Error:  nil,
		}
	}

	var value []byte
	var valueLength int64

	switch {
	case object.AssertType(obj.TypeEncoding, object.ObjTypeByteArray) == nil:
		byteArray := obj.Value.(*ByteArray)
		value = byteArray.data
		valueLength = byteArray.Length
	case object.AssertType(obj.TypeEncoding, object.ObjTypeString) == nil:
		value = []byte(obj.Value.(string))
		valueLength = int64(len(value))
	case object.AssertType(obj.TypeEncoding, object.ObjTypeInt) == nil:
		value = []byte(strconv.FormatInt(obj.Value.(int64), 10))
		valueLength = int64(len(value))
	default:
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	// defining constants of the function
	start, end := int64(0), valueLength-1
	unit := BYTE

	// checking which arguments are present and validating arguments
	if len(args) > 1 {
		start, err = strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrIntegerOutOfRange,
			}
		}
		if len(args) <= 2 {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrSyntax,
			}
		}
		end, err = strconv.ParseInt(args[2], 10, 64)
		if err != nil {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrIntegerOutOfRange,
			}
		}
	}
	if len(args) > 3 {
		unit = strings.ToUpper(args[3])
	}

	switch unit {
	case BYTE:
		if start < 0 {
			start += valueLength
		}
		if end < 0 {
			end += valueLength
		}
		if start > end || start >= valueLength {
			return &EvalResponse{
				Result: clientio.IntegerZero,
				Error:  nil,
			}
		}
		end = min(end, valueLength-1)
		bitCount := 0
		for i := start; i <= end; i++ {
			bitCount += bits.OnesCount8(value[i])
		}
		return &EvalResponse{
			Result: bitCount,
			Error:  nil,
		}
	case BIT:
		if start < 0 {
			start += valueLength * 8
		}
		if end < 0 {
			end += valueLength * 8
		}
		if start > end {
			return &EvalResponse{
				Result: clientio.IntegerZero,
				Error:  nil,
			}
		}
		startByte, endByte := start/8, min(end/8, valueLength-1)
		startBitOffset, endBitOffset := start%8, end%8

		if endByte == valueLength-1 {
			endBitOffset = 7
		}

		if startByte >= valueLength {
			return &EvalResponse{
				Result: clientio.IntegerZero,
				Error:  nil,
			}
		}

		bitCount := 0

		// Use bit masks to count the bits instead of a loop
		if startByte == endByte {
			mask := byte(0xFF >> startBitOffset)
			mask &= byte(0xFF << (7 - endBitOffset))
			bitCount = bits.OnesCount8(value[startByte] & mask)
		} else {
			// Handle first byte
			firstByteMask := byte(0xFF >> startBitOffset)
			bitCount += bits.OnesCount8(value[startByte] & firstByteMask)

			// Handle all the middle ones
			for i := startByte + 1; i < endByte; i++ {
				bitCount += bits.OnesCount8(value[i])
			}

			// Handle last byte
			lastByteMask := byte(0xFF << (7 - endBitOffset))
			bitCount += bits.OnesCount8(value[endByte] & lastByteMask)
		}
		return &EvalResponse{
			Result: bitCount,
			Error:  nil,
		}
	default:
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrSyntax,
		}
	}
}

// BITOP <AND | OR | XOR | NOT> destkey key [key ...]
func evalBITOP(args []string, store *dstore.Store) *EvalResponse {
	operation, destKey := args[0], args[1]
	operation = strings.ToUpper(operation)

	// get all the keys
	keys := args[2:]

	// validation of commands
	// if operation is not from enums, then error out
	if !(operation == AND || operation == OR || operation == XOR || operation == NOT) {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrSyntax,
		}
	}

	if operation == NOT {
		if len(keys) != 1 {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral("BITOP NOT must be called with a single source key."),
			}
		}
		key := keys[0]
		obj := store.Get(key)
		if obj == nil {
			return &EvalResponse{
				Result: clientio.IntegerZero,
				Error:  nil,
			}
		}

		var value []byte

		switch oType, _ := object.ExtractTypeEncoding(obj); oType {
		case object.ObjTypeByteArray:
			byteArray := obj.Value.(*ByteArray)
			byteArrayObject := *byteArray
			value = byteArrayObject.data
			// perform the operation
			result := make([]byte, len(value))
			for i := 0; i < len(value); i++ {
				result[i] = ^value[i]
			}

			// initialize result with byteArray
			operationResult := NewByteArray(len(result))
			operationResult.data = result
			operationResult.Length = int64(len(result))

			// resize the byte array if necessary
			operationResult.ResizeIfNecessary()

			// create object related to result
			obj = store.NewObj(operationResult, -1, object.ObjTypeByteArray, object.ObjEncodingByteArray)

			// store the result in destKey
			store.Put(destKey, obj)
			return &EvalResponse{
				Result: len(value),
				Error:  nil,
			}
		case object.ObjTypeString, object.ObjTypeInt:
			if oType == object.ObjTypeString {
				value = []byte(obj.Value.(string))
			} else {
				value = []byte(strconv.FormatInt(obj.Value.(int64), 10))
			}
			// perform the operation
			result := make([]byte, len(value))
			for i := 0; i < len(value); i++ {
				result[i] = ^value[i]
			}
			resOType, resOEnc := deduceTypeEncoding(string(result))
			var storedValue interface{}
			if resOType == object.ObjTypeInt {
				storedValue, _ = strconv.ParseInt(string(result), 10, 64)
			} else {
				storedValue = string(result)
			}
			store.Put(destKey, store.NewObj(storedValue, -1, resOType, resOEnc))
			return &EvalResponse{
				Result: len(value),
				Error:  nil,
			}
		default:
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongTypeOperation,
			}
		}
	}
	// if operation is AND, OR, XOR
	values := make([][]byte, len(keys))

	// get the values of all keys
	for i, key := range keys {
		obj := store.Get(key)
		if obj == nil {
			values[i] = make([]byte, 0)
		} else {
			// handle the case when it is byte array
			switch oType, _ := object.ExtractTypeEncoding(obj); oType {
			case object.ObjTypeByteArray:
				byteArray := obj.Value.(*ByteArray)
				byteArrayObject := *byteArray
				values[i] = byteArrayObject.data
			case object.ObjTypeString:
				value := obj.Value.(string)
				values[i] = []byte(value)
			case object.ObjTypeInt:
				value := strconv.FormatInt(obj.Value.(int64), 10)
				values[i] = []byte(value)
			default:
				return &EvalResponse{
					Result: nil,
					Error:  diceerrors.ErrWrongTypeOperation,
				}
			}
		}
	}
	// get the length of the largest value
	maxLength := 0
	minLength := len(values[0])
	maxKeyIterator := 0
	for keyIterator, value := range values {
		if len(value) > maxLength {
			maxLength = len(value)
			maxKeyIterator = keyIterator
		}
		minLength = min(minLength, len(value))
	}
	result := make([]byte, maxLength)
	if operation == AND {
		for i := 0; i < maxLength; i++ {
			result[i] = 0
			if i < minLength {
				result[i] = values[maxKeyIterator][i]
			}
		}
	} else {
		for i := 0; i < maxLength; i++ {
			result[i] = 0x00
		}
	}

	// perform the operation
	for _, value := range values {
		for i := 0; i < len(value); i++ {
			switch operation {
			case AND:
				result[i] &= value[i]
			case OR:
				result[i] |= value[i]
			case XOR:
				result[i] ^= value[i]
			}
		}
	}
	// initialize result with byteArray
	operationResult := NewByteArray(len(result))
	operationResult.data = result
	operationResult.Length = int64(len(result))

	// create object related to result
	operationResultObject := store.NewObj(operationResult, -1, object.ObjTypeByteArray, object.ObjEncodingByteArray)

	// store the result in destKey
	store.Put(destKey, operationResultObject)

	return &EvalResponse{
		Result: len(result),
		Error:  nil,
	}
}

// Generic method for both BITFIELD and BITFIELD_RO.
// isReadOnly method is true for BITFIELD_RO command.
func bitfieldEvalGeneric(args []string, store *dstore.Store, isReadOnly bool) *EvalResponse {
	var ops []utils.BitFieldOp
	ops, err2 := utils.ParseBitfieldOps(args, isReadOnly)

	if err2 != nil {
		return &EvalResponse{
			Result: nil,
			Error:  err2,
		}
	}

	key := args[0]
	obj := store.Get(key)
	if obj == nil {
		obj = store.NewObj(NewByteArray(1), -1, object.ObjTypeByteArray, object.ObjEncodingByteArray)
		store.Put(args[0], obj)
	}
	var value *ByteArray
	var err error

	switch oType, _ := object.ExtractTypeEncoding(obj); oType {
	case object.ObjTypeByteArray:
		value = obj.Value.(*ByteArray)
	case object.ObjTypeString, object.ObjTypeInt:
		value, err = NewByteArrayFromObj(obj)
		if err != nil {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral("value is not a valid byte array"),
			}
		}
	default:
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	result := executeBitfieldOps(value, ops)
	return &EvalResponse{
		Result: result,
		Error:  nil,
	}
}

// evalBITFIELD evaluates BITFIELD operations on a key store string, int or bytearray types
// it returns an array of results depending on the subcommands
// it allows mutation using SET and INCRBY commands
// returns arity error, offset type error, overflow type error, encoding type error, integer error, syntax error
// GET <encoding> <offset> -- Returns the specified bit field.
// SET <encoding> <offset> <value> -- Set the specified bit field
// and returns its old value.
// INCRBY <encoding> <offset> <increment> -- Increments or decrements
// (if a negative increment is given) the specified bit field and returns the new value.
// There is another subcommand that only changes the behavior of successive
// INCRBY and SET subcommands calls by setting the overflow behavior:
// OVERFLOW [WRAP|SAT|FAIL]`
func evalBITFIELD(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("BITFIELD"),
		}
	}

	return bitfieldEvalGeneric(args, store, false)
}

// Read-only variant of the BITFIELD command. It is like the original BITFIELD but only accepts GET subcommand and can safely be used in read-only replicas.
func evalBITFIELDRO(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("BITFIELD_RO"),
		}
	}

	return bitfieldEvalGeneric(args, store, true)
}

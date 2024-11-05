package eval

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/axiomhq/hyperloglog"
	"github.com/bytedance/sonic"
	"github.com/dicedb/dice/internal/clientio"
	diceerrors "github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/eval/sortedset"
	"github.com/dicedb/dice/internal/object"
	"github.com/dicedb/dice/internal/server/utils"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/gobwas/glob"
	"github.com/ohler55/ojg/jp"
)

// evalEXPIRE sets an expiry time(in secs) on the specified key in args
// args should contain 2 values, key and the expiry time to be set for the key
// The expiry time should be in integer format; if not, it returns encoded error response
// Returns clientio.IntegerOne if expiry was set on the key successfully.
// Once the time is lapsed, the key will be deleted automatically
func evalEXPIRE(args []string, store *dstore.Store) *EvalResponse {
	if len(args) <= 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("EXPIRE"),
		}
	}

	var key = args[0]
	exDurationSec, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrIntegerOutOfRange,
		}
	}

	if exDurationSec < 0 || exDurationSec > maxExDuration {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrInvalidExpireTime("EXPIRE"),
		}
	}

	obj := store.Get(key)

	// 0 if the timeout was not set. e.g. key doesn't exist, or operation skipped due to the provided arguments
	if obj == nil {
		return &EvalResponse{
			Result: clientio.IntegerZero,
			Error:  nil,
		}
	}
	isExpirySet, err2 := dstore.EvaluateAndSetExpiry(args[2:], utils.AddSecondsToUnixEpoch(exDurationSec), key, store)

	if isExpirySet {
		return &EvalResponse{
			Result: clientio.IntegerOne,
			Error:  nil,
		}
	} else if err2 != nil {
		return &EvalResponse{
			Result: nil,
			Error:  err2,
		}
	}

	return &EvalResponse{
		Result: clientio.IntegerZero,
		Error:  nil,
	}
}

// evalEXPIREAT sets a expiry time(in unix-time-seconds) on the specified key in args
// args should contain 2 values, key and the expiry time to be set for the key
// The expiry time should be in integer format; if not, it returns encoded error response
// Returns response.IntegerOne if expiry was set on the key successfully.
// Once the time is lapsed, the key will be deleted automatically
func evalEXPIREAT(args []string, store *dstore.Store) *EvalResponse {
	if len(args) <= 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("EXPIREAT"),
		}
	}

	var key = args[0]
	exUnixTimeSec, err := strconv.ParseInt(args[1], 10, 64)
	if exUnixTimeSec < 0 || exUnixTimeSec > maxExDuration {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrInvalidExpireTime("EXPIREAT"),
		}
	}

	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrIntegerOutOfRange,
		}
	}

	isExpirySet, err := dstore.EvaluateAndSetExpiry(args[2:], exUnixTimeSec, key, store)
	if isExpirySet {
		return &EvalResponse{
			Result: clientio.IntegerOne,
			Error:  nil,
		}
	} else if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  err,
		}
	}
	return &EvalResponse{
		Result: clientio.IntegerZero,
		Error:  nil,
	}
}

// evalEXPIRETIME returns the absolute Unix timestamp (since January 1, 1970) in seconds at which the given key will expire
// args should contain only 1 value, the key
// Returns expiration Unix timestamp in seconds.
// Returns -1 if the key exists but has no associated expiration time.
// Returns -2 if the key does not exist.
func evalEXPIRETIME(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("EXPIRETIME"),
		}
	}

	var key = args[0]

	obj := store.Get(key)

	// -2 if key doesn't exist
	if obj == nil {
		return &EvalResponse{
			Result: clientio.IntegerNegativeTwo,
			Error:  nil,
		}
	}

	exTimeMili, ok := dstore.GetExpiry(obj, store)
	// -1 if key doesn't have expiration time set
	if !ok {
		return &EvalResponse{
			Result: clientio.IntegerNegativeOne,
			Error:  nil,
		}
	}

	return &EvalResponse{
		Result: exTimeMili / 1000,
		Error:  nil,
	}
}

// evalSET puts a new <key, value> pair in db as in the args
// args must contain key and value.
// args can also contain multiple options -
//
//	EX or ex which will set the expiry time(in secs) for the key
//	PX or px which will set the expiry time(in milliseconds) for the key
//	EXAT or exat which will set the specified Unix time at which the key will expire, in seconds (a positive integer)
//	PXAT or PX which will the specified Unix time at which the key will expire, in milliseconds (a positive integer)
//	XX or xx which will only set the key if it already exists
//	NX or nx which will only set the key if it doesn not already exist
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
			if keepttl {
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
			if keepttl {
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
			if state != Uninitialized {
				return &EvalResponse{
					Result: nil,
					Error:  diceerrors.ErrSyntax,
				}
			}
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
		if IsInt64(obj.Value) {
			return &EvalResponse{
				Result: obj.Value,
				Error:  nil,
			}
		} else if IsString(obj.Value) {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrUnexpectedType("int64", "string"),
			}
		} else {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrUnexpectedType("int64", "unknown"),
			}
		}

	case object.ObjEncodingEmbStr, object.ObjEncodingRaw:
		// Value is stored as a string, use type assertion
		if IsString(obj.Value) {
			return &EvalResponse{
				Result: obj.Value,
				Error:  nil,
			}
		} else if IsInt64(obj.Value) {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrUnexpectedType("string", "int64"),
			}
		} else {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrUnexpectedType("string", "unknown"),
			}
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

// evalHEXISTS returns if field is an existing field in the hash stored at key.
//
// This command returns 0, if the specified field doesn't exist in the key and 1 if it exists.
//
// If key doesn't exist, it returns 0.
//
// Usage: HEXISTS key field
func evalHEXISTS(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 2 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("HEXISTS"),
		}
	}

	key := args[0]
	hmKey := args[1]
	obj := store.Get(key)

	var hashMap HashMap

	if obj == nil {
		return &EvalResponse{
			Result: clientio.IntegerZero,
			Error:  nil,
		}
	}
	if err := object.AssertTypeAndEncoding(obj.TypeEncoding, object.ObjTypeHashMap, object.ObjEncodingHashMap); err != nil {
		return &EvalResponse{
			Error:  diceerrors.ErrGeneral(diceerrors.WrongTypeErr),
			Result: nil,
		}
	}

	hashMap = obj.Value.(HashMap)

	_, ok := hashMap.Get(hmKey)
	if ok {
		return &EvalResponse{
			Result: clientio.IntegerOne,
			Error:  nil,
		}
	}
	// Return 0, if specified field doesn't exist in the HashMap.
	return &EvalResponse{
		Result: clientio.IntegerZero,
		Error:  nil,
	}
}

// evalHKEYS is used to retrieve all the keys(or field names) within a hash.
//
// This command returns empty array, if the specified key doesn't exist.
//
// Complexity is O(n) where n is the size of the hash.
//
// Usage: HKEYS key
func evalHKEYS(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("HKEYS"),
		}
	}

	key := args[0]
	obj := store.Get(key)

	var hashMap HashMap
	var result []string

	if obj != nil {
		if err := object.AssertTypeAndEncoding(obj.TypeEncoding, object.ObjTypeHashMap, object.ObjEncodingHashMap); err != nil {
			return &EvalResponse{
				Error:  diceerrors.ErrGeneral(diceerrors.WrongTypeErr),
				Result: nil,
			}
		}
		hashMap = obj.Value.(HashMap)
	} else {
		return &EvalResponse{
			Result: clientio.EmptyArray,
			Error:  nil,
		}
	}

	for hmKey := range hashMap {
		result = append(result, hmKey)
	}

	return &EvalResponse{
		Result: result,
		Error:  nil,
	}
}

// evalHKEYS is used to retrieve all the values within a hash.
//
// This command returns empty array, if the specified key doesn't exist.
//
// Complexity is O(n) where n is the size of the hash.
//
// Usage: HVALS key
func evalHVALS(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 1 {
		return &EvalResponse{Error: diceerrors.ErrWrongArgumentCount("HVALS"), Result: nil}
	}

	key := args[0]
	obj := store.Get(key)

	if obj == nil {
		// Return an empty array for non-existent keys
		return &EvalResponse{
			Result: clientio.EmptyArray,
			Error:  nil,
		}
	}

	if err := object.AssertTypeAndEncoding(obj.TypeEncoding, object.ObjTypeHashMap, object.ObjEncodingHashMap); err != nil {
		return &EvalResponse{
			Error:  diceerrors.ErrGeneral(diceerrors.WrongTypeErr),
			Result: nil,
		}
	}

	hashMap := obj.Value.(HashMap)
	results := make([]string, 0, len(hashMap))

	for _, value := range hashMap {
		results = append(results, value)
	}

	return &EvalResponse{
		Result: results,
		Error:  nil,
	}
}

// Key, start and end are mandatory args.
// Returns a substring from the key(if it's a string) from start -> end.
// Returns ""(empty string) if key is not present and if start > end.
func evalGETRANGE(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 3 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("GETRANGE"),
		}
	}

	key := args[0]
	obj := store.Get(key)
	if obj == nil {
		return &EvalResponse{
			Result: string(""),
			Error:  nil,
		}
	}

	start, err := strconv.Atoi(args[1])
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrIntegerOutOfRange,
		}
	}
	end, err := strconv.Atoi(args[2])
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrIntegerOutOfRange,
		}
	}

	var str string
	switch _, oEnc := object.ExtractTypeEncoding(obj); oEnc {
	case object.ObjEncodingEmbStr, object.ObjEncodingRaw:
		if val, ok := obj.Value.(string); ok {
			str = val
		} else {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral("expected string but got another type"),
			}
		}
	case object.ObjEncodingInt:
		str = strconv.FormatInt(obj.Value.(int64), 10)
	default:
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	if str == "" {
		return &EvalResponse{
			Result: string(""),
			Error:  nil,
		}
	}

	if start < 0 {
		start = len(str) + start
	}

	if end < 0 {
		end = len(str) + end
	}

	if start >= len(str) || end < 0 || start > end {
		return &EvalResponse{
			Result: string(""),
			Error:  nil,
		}
	}

	if start < 0 {
		start = 0
	}

	if end >= len(str) {
		end = len(str) - 1
	}

	return &EvalResponse{
		Result: str[start : end+1],
		Error:  nil,
	}
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

// The ZCOUNT command in DiceDB counts the number of members in a sorted set at the specified key
// whose scores fall within a given range. The command takes three arguments: the key of the sorted set
// the minimum score, and the maximum score.
func evalZCOUNT(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 3 {
		// 1. Check no of arguments
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("ZCOUNT"),
		}
	}

	key := args[0]
	minArg := args[1]
	maxArg := args[2]

	// 2. Parse the min and max score arguments
	minValue, errMin := strconv.ParseFloat(minArg, 64)
	maxValue, errMax := strconv.ParseFloat(maxArg, 64)
	if errMin != nil || errMax != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrInvalidNumberFormat,
		}
	}

	// 3. Retrieve the object from the store
	obj := store.Get(key)
	if obj == nil {
		// If the key does not exist, return 0 (no error)
		return &EvalResponse{
			Result: 0,
			Error:  nil,
		}
	}

	// 4. Ensure the object is a valid sorted set
	var sortedSet *sortedset.Set
	var err []byte
	sortedSet, err = sortedset.FromObject(obj)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	count := sortedSet.CountInRange(minValue, maxValue)

	return &EvalResponse{
		Result: count,
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

// evalZREM removes the specified members from the sorted set stored at key.
// Non-existing members are ignored.
// Returns the number of members removed.
func evalZREM(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 2 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("ZREM"),
		}
	}

	key := args[0]
	obj := store.Get(key)

	if obj == nil {
		return &EvalResponse{
			Result: clientio.IntegerZero,
			Error:  nil,
		}
	}

	sortedSet, err := sortedset.FromObject(obj)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	countRem := 0
	for i := 1; i < len(args); i++ {
		if sortedSet.Remove(args[i]) {
			countRem += 1
		}
	}

	return &EvalResponse{
		Result: int64(countRem),
		Error:  nil,
	}
}

// evalAPPEND takes two arguments: the key and the value to append to the key's current value.
// If the key does not exist, it creates a new key with the given value (so APPEND will be similar to SET in this special case)
// If key already exists and is a string (or integers stored as strings), this command appends the value at the end of the string
func evalAPPEND(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 2 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("APPEND"),
		}
	}

	key, value := args[0], args[1]
	obj := store.Get(key)

	if obj == nil {
		// Key does not exist path

		// check if the value starts with '0' and has more than 1 character to handle leading zeros
		if len(value) > 1 && value[0] == '0' {
			// treat as string if has leading zeros
			store.Put(key, store.NewObj(value, -1, object.ObjTypeString, object.ObjEncodingRaw))
			return &EvalResponse{
				Result: len(value),
				Error:  nil,
			}
		}

		// Deduce type and encoding based on the value if no leading zeros
		oType, oEnc := deduceTypeEncoding(value)

		var storedValue interface{}
		// Store the value with the appropriate encoding based on the type
		switch oEnc {
		case object.ObjEncodingInt:
			storedValue, _ = strconv.ParseInt(value, 10, 64)
		case object.ObjEncodingEmbStr, object.ObjEncodingRaw:
			storedValue = value
		default:
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongTypeOperation,
			}
		}

		store.Put(key, store.NewObj(storedValue, -1, oType, oEnc))

		return &EvalResponse{
			Result: len(value),
			Error:  nil,
		}
	}
	// Key exists path
	_, currentEnc := object.ExtractTypeEncoding(obj)

	var currentValueStr string
	switch currentEnc {
	case object.ObjEncodingInt:
		// If the encoding is an integer, convert the current value to a string for concatenation
		currentValueStr = strconv.FormatInt(obj.Value.(int64), 10)
	case object.ObjEncodingEmbStr, object.ObjEncodingRaw:
		// If the encoding is a string, retrieve the string value for concatenation
		currentValueStr = obj.Value.(string)
	default:
		// If the encoding is neither integer nor string, return a "wrong type" error
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	newValue := currentValueStr + value

	store.Put(key, store.NewObj(newValue, -1, object.ObjTypeString, object.ObjEncodingRaw))

	return &EvalResponse{
		Result: len(newValue),
		Error:  nil,
	}
}

// evalZRANK returns the rank of the member in the sorted set stored at key.
// The rank (or index) is 0-based, which means that the member with the lowest score has rank 0.
// If the 'WITHSCORE' option is specified, it returns both the rank and the score of the member.
// Returns nil if the key does not exist or the member is not a member of the sorted set.
func evalZRANK(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 2 || len(args) > 3 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("ZRANK"),
		}
	}

	key := args[0]
	member := args[1]
	withScore := false

	if len(args) == 3 {
		if !strings.EqualFold(args[2], WithScore) {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrSyntax,
			}
		}
		withScore = true
	}

	obj := store.Get(key)
	if obj == nil {
		return &EvalResponse{
			Result: clientio.NIL,
			Error:  nil,
		}
	}

	sortedSet, err := sortedset.FromObject(obj)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}
	rank, score := sortedSet.RankWithScore(member, false)
	if rank == -1 {
		return &EvalResponse{
			Result: clientio.NIL,
			Error:  nil,
		}
	}

	if withScore {
		scoreStr := strconv.FormatFloat(score, 'f', -1, 64)
		return &EvalResponse{
			Result: []interface{}{rank, scoreStr},
			Error:  nil,
		}
	}

	return &EvalResponse{
		Result: rank,
		Error:  nil,
	}
}

// evalZCARD returns the cardinality (number of elements) of the sorted set stored at key.
// Returns 0 if the key does not exist.
func evalZCARD(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 1 || len(args) > 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("ZCARD"),
		}
	}

	key := args[0]
	obj := store.Get(key)

	if obj == nil {
		return &EvalResponse{
			Result: clientio.IntegerZero,
			Error:  nil,
		}
	}

	sortedSet, err := sortedset.FromObject(obj)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}
	return &EvalResponse{
		Result: int64(sortedSet.Len()),
		Error:  nil,
	}
}

// evalJSONCLEAR Clear container values (arrays/objects) and set numeric values to 0,
// Already cleared values are ignored for empty containers and zero numbers
// args must contain at least the key;  (path unused in this implementation)
// Returns encoded error if key is expired, or it does not exist
// Returns encoded error response if incorrect number of arguments
// Returns an integer reply specifying the number of matching JSON arrays
// and objects cleared + number of matching JSON numerical values zeroed.
func evalJSONCLEAR(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("JSON.CLEAR"),
		}
	}
	key := args[0]

	// Default path is root if not specified
	path := defaultRootPath
	if len(args) > 1 {
		path = args[1]
	}

	// Retrieve the object from the database
	obj := store.Get(key)
	if obj == nil {
		return &EvalResponse{
			Result: nil,
			Error:  nil,
		}
	}

	errWithMessage := object.AssertTypeAndEncoding(obj.TypeEncoding, object.ObjTypeJSON, object.ObjEncodingJSON)
	if errWithMessage != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	jsonData := obj.Value

	_, err := sonic.Marshal(jsonData)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	var countClear int64 = 0
	if len(args) == 1 || path == defaultRootPath {
		if jsonData != struct{}{} {
			// If path is root and len(args) == 1, return it instantly
			newObj := store.NewObj(struct{}{}, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
			store.Put(key, newObj)
			countClear++
			return &EvalResponse{
				Result: countClear,
				Error:  nil,
			}
		}
	}

	expr, err := jp.ParseString(path)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrJSONPathNotFound(path),
		}
	}

	newData, err := expr.Modify(jsonData, func(element any) (altered any, changed bool) {
		switch utils.GetJSONFieldType(element) {
		case utils.IntegerType, utils.NumberType:
			if element != utils.NumberZeroValue {
				countClear++
				return utils.NumberZeroValue, true
			}
		case utils.ArrayType:
			if len(element.([]interface{})) != 0 {
				countClear++
				return []interface{}{}, true
			}
		case utils.ObjectType:
			if element != struct{}{} {
				countClear++
				return struct{}{}, true
			}
		default:
			return element, false
		}
		return
	})
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral(err.Error()),
		}
	}

	jsonData = newData
	obj.Value = jsonData
	return &EvalResponse{
		Result: countClear,
		Error:  nil,
	}
}

// PFADD Adds all the element arguments to the HyperLogLog data structure stored at the variable
// name specified as first argument.
//
// Returns:
// If the approximated cardinality estimated by the HyperLogLog changed after executing the command,
// returns 1, otherwise 0 is returned.
func evalPFADD(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("PFADD"),
		}
	}

	key := args[0]
	obj := store.Get(key)

	// If key doesn't exist prior initial cardinality changes hence return 1
	if obj == nil {
		hll := hyperloglog.New()
		for _, arg := range args[1:] {
			hll.Insert([]byte(arg))
		}

		obj = store.NewObj(hll, -1, object.ObjTypeString, object.ObjEncodingRaw)

		store.Put(key, obj)
		return &EvalResponse{
			Result: int64(1),
			Error:  nil,
		}
	}

	existingHll, ok := obj.Value.(*hyperloglog.Sketch)
	if !ok {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrInvalidHyperLogLogKey,
		}
	}
	initialCardinality := existingHll.Estimate()
	for _, arg := range args[1:] {
		existingHll.Insert([]byte(arg))
	}

	if newCardinality := existingHll.Estimate(); initialCardinality != newCardinality {
		return &EvalResponse{
			Result: int64(1),
			Error:  nil,
		}
	}

	return &EvalResponse{
		Result: int64(0),
		Error:  nil,
	}
}

// evalJSONSTRLEN Report the length of the JSON String at path in key
// Returns by recursive descent an array of integer replies for each path,
// the string's length, or nil, if the matching JSON value is not a string.
func evalJSONSTRLEN(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("JSON.STRLEN"),
		}
	}

	key := args[0]

	obj := store.Get(key)

	if obj == nil {
		return &EvalResponse{
			Result: nil,
			Error:  nil,
		}
	}

	if len(args) < 2 {
		// no recursive
		// making consistent with arrlen
		// to-do parsing
		jsonData := obj.Value

		jsonDataType := strings.ToLower(utils.GetJSONFieldType(jsonData))
		if jsonDataType == "number" {
			jsonDataFloat := jsonData.(float64)
			if jsonDataFloat == float64(int64(jsonDataFloat)) {
				jsonDataType = "integer"
			}
		}
		if jsonDataType != utils.StringType {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrUnexpectedJSONPathType("string", jsonDataType),
			}
		}
		return &EvalResponse{
			Result: int64(len(jsonData.(string))),
			Error:  nil,
		}
	}

	path := args[1]

	// Check if the object is of JSON type
	errWithMessage := object.AssertTypeAndEncoding(obj.TypeEncoding, object.ObjTypeJSON, object.ObjEncodingJSON)
	if errWithMessage != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	jsonData := obj.Value
	if path == defaultRootPath {
		defaultStringResult := make([]interface{}, 0, 1)
		if utils.GetJSONFieldType(jsonData) == utils.StringType {
			defaultStringResult = append(defaultStringResult, int64(len(jsonData.(string))))
		} else {
			defaultStringResult = append(defaultStringResult, nil)
		}

		return &EvalResponse{
			Result: defaultStringResult,
			Error:  nil,
		}
	}

	// Parse the JSONPath expression
	expr, err := jp.ParseString(path)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrJSONPathNotFound(path),
		}
	}
	// Execute the JSONPath query
	results := expr.Get(jsonData)
	if len(results) == 0 {
		return &EvalResponse{
			Result: []interface{}{},
			Error:  nil,
		}
	}
	strLenResults := make([]interface{}, 0, len(results))
	for _, result := range results {
		switch utils.GetJSONFieldType(result) {
		case utils.StringType:
			strLenResults = append(strLenResults, int64(len(result.(string))))
		default:
			strLenResults = append(strLenResults, nil)
		}
	}
	return &EvalResponse{
		Result: strLenResults,
		Error:  nil,
	}
}

func evalPFCOUNT(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("PFCOUNT"),
		}
	}

	unionHll := hyperloglog.New()

	for _, arg := range args {
		obj := store.Get(arg)
		if obj != nil {
			currKeyHll, ok := obj.Value.(*hyperloglog.Sketch)
			if !ok {
				return &EvalResponse{
					Result: nil,
					Error:  diceerrors.ErrInvalidHyperLogLogKey,
				}
			}
			err := unionHll.Merge(currKeyHll)
			if err != nil {
				return &EvalResponse{
					Result: nil,
					Error:  diceerrors.ErrCorruptedHyperLogLogObject,
				}
			}
		}
	}

	return &EvalResponse{
		Result: unionHll.Estimate(),
		Error:  nil,
	}
}

// evalJSONOBJLEN return the number of keys in the JSON object at path in key.
// Returns an array of integer replies, an integer for each matching value,
// which is the json objects length, or nil, if the matching value is not a json.
// Returns encoded error if the key doesn't exist or key is expired or the matching value is not an array.
// Returns encoded error response if incorrect number of arguments
func evalJSONOBJLEN(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("JSON.OBJLEN"),
		}
	}

	key := args[0]

	// Retrieve the object from the database
	obj := store.Get(key)
	if obj == nil {
		return &EvalResponse{
			Result: nil,
			Error:  nil,
		}
	}

	// check if the object is json
	errWithMessage := object.AssertTypeAndEncoding(obj.TypeEncoding, object.ObjTypeJSON, object.ObjEncodingJSON)
	if errWithMessage != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	// get the value & check for marsheling error
	jsonData := obj.Value
	_, err := sonic.Marshal(jsonData)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}
	if len(args) == 1 {
		// check if the value is of json type
		if utils.GetJSONFieldType(jsonData) == utils.ObjectType {
			if castedData, ok := jsonData.(map[string]interface{}); ok {
				return &EvalResponse{
					Result: int64(len(castedData)),
					Error:  nil,
				}
			}
			return &EvalResponse{
				Result: nil,
				Error:  nil,
			}
		}
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	path, isDefinitePath := utils.ParseInputJSONPath(args[1])

	expr, err := jp.ParseString(path)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrJSONPathNotFound(path),
		}
	}

	// get all values for matching paths
	results := expr.Get(jsonData)

	objectLen := make([]interface{}, 0, len(results))

	for _, result := range results {
		switch utils.GetJSONFieldType(result) {
		case utils.ObjectType:
			if castedResult, ok := result.(map[string]interface{}); ok {
				objectLen = append(objectLen, int64(len(castedResult)))
			} else {
				objectLen = append(objectLen, nil)
			}
		default:
			// If it is a definitePath, and the only value is not JSON, throw wrong type error
			if isDefinitePath {
				return &EvalResponse{
					Result: nil,
					Error:  diceerrors.ErrWrongTypeOperation,
				}
			}
			objectLen = append(objectLen, nil)
		}
	}

	// Must return a single integer if it is a definite Path
	if isDefinitePath {
		if len(objectLen) == 0 {
			return &EvalResponse{
				Result: nil,
				Error:  nil,
			}
		}
		return &EvalResponse{
			Result: objectLen[0],
			Error:  nil,
		}
	}

	return &EvalResponse{
		Result: objectLen,
		Error:  nil,
	}
}

func evalPFMERGE(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("PFMERGE"),
		}
	}

	var mergedHll *hyperloglog.Sketch
	destKey := args[0]
	obj := store.Get(destKey)

	// If destKey doesn't exist, create a new HLL, else fetch the existing
	if obj == nil {
		mergedHll = hyperloglog.New()
	} else {
		var ok bool
		mergedHll, ok = obj.Value.(*hyperloglog.Sketch)
		if !ok {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrInvalidHyperLogLogKey,
			}
		}
	}

	for _, arg := range args {
		obj := store.Get(arg)
		if obj != nil {
			currKeyHll, ok := obj.Value.(*hyperloglog.Sketch)
			if !ok {
				return &EvalResponse{
					Result: nil,
					Error:  diceerrors.ErrInvalidHyperLogLogKey,
				}
			}

			err := mergedHll.Merge(currKeyHll)
			if err != nil {
				return &EvalResponse{
					Result: nil,
					Error:  diceerrors.ErrCorruptedHyperLogLogObject,
				}
			}
		}
	}

	// Save the mergedHll
	obj = store.NewObj(mergedHll, -1, object.ObjTypeString, object.ObjEncodingRaw)
	store.Put(destKey, obj)

	return &EvalResponse{
		Result: clientio.OK,
		Error:  nil,
	}
}

// evalPTTL returns Time-to-Live in millisecs for the queried key in args
// The key should be the only param in args else returns with an error
// Returns	RESP encoded time (in secs) remaining for the key to expire
//
//	RESP encoded -2 stating key doesn't exist or key is expired
//	RESP encoded -1 in case no expiration is set on the key
func evalPTTL(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("PTTL"),
		}
	}

	key := args[0]

	obj := store.Get(key)

	if obj == nil {
		return &EvalResponse{
			Result: clientio.IntegerNegativeTwo,
			Error:  nil,
		}
	}

	exp, isExpirySet := dstore.GetExpiry(obj, store)

	if !isExpirySet {
		return &EvalResponse{
			Result: clientio.IntegerNegativeOne,
			Error:  nil,
		}
	}

	// compute the time remaining for the key to expire and
	// return the RESP encoded form of it
	durationMs := exp - uint64(utils.GetCurrentTime().UnixMilli())
	return &EvalResponse{
		Result: durationMs,
		Error:  nil,
	}
}

// evalTTL returns Time-to-Live in secs for the queried key in args
// The key should be the only param in args else returns with an error
// Returns	RESP encoded time (in secs) remaining for the key to expire
//
//	RESP encoded -2 stating key doesn't exist or key is expired
//	RESP encoded -1 in case no expiration is set on the key
func evalTTL(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("TTL"),
		}
	}

	var key = args[0]

	obj := store.Get(key)

	// if key does not exist, return RESP encoded -2 denoting key does not exist
	if obj == nil {
		return &EvalResponse{
			Result: clientio.IntegerNegativeTwo,
			Error:  nil,
		}
	}

	// if object exist, but no expiration is set on it then send -1
	exp, isExpirySet := dstore.GetExpiry(obj, store)
	if !isExpirySet {
		return &EvalResponse{
			Result: clientio.IntegerNegativeOne,
			Error:  nil,
		}
	}

	// compute the time remaining for the key to expire and
	// return the RESP encoded form of it
	durationMs := exp - uint64(utils.GetCurrentTime().UnixMilli())

	return &EvalResponse{
		Result: durationMs / 1000,
		Error:  nil,
	}
}

// Increments the number stored at field in the hash stored at key by increment.
//
// If key does not exist, a new key holding a hash is created.
// If field does not exist the value is set to the increment value passed
//
// The range of values supported by HINCRBY is limited to 64-bit signed integers.
//
// Usage: HINCRBY key field increment
func evalHINCRBY(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 3 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("HINCRBY"),
		}
	}

	increment, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrIntegerOutOfRange,
		}
	}
	var hashmap HashMap
	key := args[0]
	obj := store.Get(key)
	if obj != nil {
		if err := object.AssertTypeAndEncoding(obj.TypeEncoding, object.ObjTypeHashMap, object.ObjEncodingHashMap); err != nil {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongTypeOperation,
			}
		}
		hashmap = obj.Value.(HashMap)
	}

	if hashmap == nil {
		hashmap = make(HashMap)
	}

	field := args[1]
	numkey, err := hashmap.incrementValue(field, increment)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral(err.Error()),
		}
	}

	obj = store.NewObj(hashmap, -1, object.ObjTypeHashMap, object.ObjEncodingHashMap)
	store.Put(key, obj)

	return &EvalResponse{
		Result: numkey,
		Error:  nil,
	}
}

// Increments the number stored at field in the hash stored at key by the specified floating point increment.
//
// If key does not exist, a new key holding a hash is created.
// If field does not exist, the value is set to the increment passed before the operation is performed.
//
// The precision of the increment is not restricted to integers, allowing for floating point values.
//
// Usage: HINCRBYFLOAT key field increment
func evalHINCRBYFLOAT(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 3 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("HINCRBYFLOAT"),
		}
	}

	increment, err := strconv.ParseFloat(strings.TrimSpace(args[2]), 64)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrInvalidNumberFormat,
		}
	}

	key := args[0]
	obj := store.Get(key)
	var hashmap HashMap
	if obj != nil {
		if err := object.AssertTypeAndEncoding(obj.TypeEncoding, object.ObjTypeHashMap, object.ObjEncodingHashMap); err != nil {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongTypeOperation,
			}
		}
		hashmap = obj.Value.(HashMap)
	}

	if hashmap == nil {
		hashmap = make(HashMap)
	}

	field := args[1]
	numkey, err := hashmap.incrementFloatValue(field, increment)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral(err.Error()),
		}
	}

	obj = store.NewObj(hashmap, -1, object.ObjTypeHashMap, object.ObjEncodingHashMap)
	store.Put(key, obj)

	return &EvalResponse{
		Result: numkey,
		Error:  nil,
	}
}

// evalHRANDFIELD returns random fields from a hash stored at key.
// If only the key is provided, one random field is returned.
// If count is provided, it returns that many unique random fields. A negative count allows repeated selections.
// The "WITHVALUES" option returns both fields and values.
// Returns nil if the key doesn't exist or the hash is empty.
// Errors: arity error, type error for non-hash, syntax error for "WITHVALUES", or count format error.
func evalHRANDFIELD(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 1 || len(args) > 3 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("HRANDFIELD"),
		}
	}

	key := args[0]
	obj := store.Get(key)
	if obj == nil {
		return &EvalResponse{
			Result: clientio.NIL,
			Error:  nil,
		}
	}

	if err := object.AssertTypeAndEncoding(obj.TypeEncoding, object.ObjTypeHashMap, object.ObjEncodingHashMap); err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	hashMap := obj.Value.(HashMap)
	if len(hashMap) == 0 {
		return &EvalResponse{
			Result: clientio.EmptyArray,
			Error:  nil,
		}
	}

	count := 1
	withValues := false

	if len(args) > 1 {
		var err error
		// The second argument is the count.
		count, err = strconv.Atoi(args[1])
		if err != nil {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrIntegerOutOfRange,
			}
		}

		// The third argument is the "WITHVALUES" option.
		if len(args) == 3 {
			if !strings.EqualFold(args[2], WithValues) {
				return &EvalResponse{
					Result: nil,
					Error:  diceerrors.ErrSyntax,
				}
			}
			withValues = true
		}
	}

	return selectRandomFields(hashMap, count, withValues)
}

// evalINCR increments the value of the specified key in args by 1,
// if the key exists and the value is integer format.
// The key should be the only param in args.
// If the key does not exist, new key is created with value 0,
// the value of the new key is then incremented.
// The value for the queried key should be of integer format,
// if not evalINCR returns encoded error response.
// evalINCR returns the incremented value for the key if there are no errors.
func evalINCR(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("INCR"),
		}
	}

	return incrDecrCmd(args, 1, store)
}

// INCRBY increments the value of the specified key in args by increment integer specified,
// if the key exists and the value is integer format.
// The key and the increment integer should be the only param in args.
// If the key does not exist, new key is created with value 0,
// the value of the new key is then incremented.
// The value for the queried key should be of integer format,
// if not INCRBY returns error response.
// evalINCRBY returns the incremented value for the key if there are no errors.
func evalINCRBY(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 2 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("INCRBY"),
		}
	}

	incrAmount, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrIntegerOutOfRange,
		}
	}
	return incrDecrCmd(args, incrAmount, store)
}

// evalDECR decrements the value of the specified key in args by 1,
// if the key exists and the value is integer format.
// The key should be the only param in args.
// If the key does not exist, new key is created with value 0,
// the value of the new key is then decremented.
// The value for the queried key should be of integer format,
// if not evalDECR returns error response.
// evalDECR returns the decremented value for the key if there are no errors.
func evalDECR(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("DECR"),
		}
	}
	return incrDecrCmd(args, -1, store)
}

// evalDECRBY decrements the value of the specified key in args by the specified decrement,
// if the key exists and the value is integer format.
// The key should be the first parameter in args, and the decrement should be the second parameter.
// If the key does not exist, new key is created with value 0,
// the value of the new key is then decremented by specified decrement.
// The value for the queried key should be of integer format,
// if not evalDECRBY returns an error response.
// evalDECRBY returns the decremented value for the key after applying the specified decrement if there are no errors.
func evalDECRBY(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 2 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("DECRBY"),
		}
	}
	decrAmount, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrIntegerOutOfRange,
		}
	}
	return incrDecrCmd(args, -decrAmount, store)
}

func incrDecrCmd(args []string, incr int64, store *dstore.Store) *EvalResponse {
	key := args[0]
	obj := store.Get(key)
	if obj == nil {
		obj = store.NewObj(incr, -1, object.ObjTypeInt, object.ObjEncodingInt)
		store.Put(key, obj)
		return &EvalResponse{
			Result: incr,
			Error:  nil,
		}
	}
	// if the type is not KV : return wrong type error
	// if the encoding or type is not int : return value is not an int error
	errStr := object.AssertType(obj.TypeEncoding, object.ObjTypeString)
	if errStr == nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrIntegerOutOfRange,
		}
	}

	errTypeInt := object.AssertType(obj.TypeEncoding, object.ObjTypeInt)
	errEncInt := object.AssertEncoding(obj.TypeEncoding, object.ObjEncodingInt)
	if errEncInt != nil || errTypeInt != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}
	i, _ := obj.Value.(int64)
	if (incr < 0 && i < 0 && incr < (math.MinInt64-i)) ||
		(incr > 0 && i > 0 && incr > (math.MaxInt64-i)) {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrOverflow,
		}
	}

	i += incr
	obj.Value = i
	return &EvalResponse{
		Result: i,
		Error:  nil,
	}
}

// evalINCRBYFLOAT increments the value of the  key in args by the specified increment,
// if the key exists and the value is a number.
// The key should be the first parameter in args, and the increment should be the second parameter.
// If the key does not exist, a new key is created with increment's value.
// If the value at the key is a string, it should be parsable to float64,
// if not evalINCRBYFLOAT returns an  error response.
// evalINCRBYFLOAT returns the incremented value for the key after applying the specified increment if there are no errors.
func evalINCRBYFLOAT(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 2 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("INCRBYFLOAT"),
		}
	}
	incr, err := strconv.ParseFloat(strings.TrimSpace(args[1]), 64)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral("value is not a valid float"),
		}
	}
	return incrByFloatCmd(args, incr, store)
}

func incrByFloatCmd(args []string, incr float64, store *dstore.Store) *EvalResponse {
	key := args[0]
	obj := store.Get(key)

	if obj == nil {
		strValue := formatFloat(incr, false)
		oType, oEnc := deduceTypeEncoding(strValue)
		obj = store.NewObj(strValue, -1, oType, oEnc)
		store.Put(key, obj)
		return &EvalResponse{
			Result: strValue,
			Error:  nil,
		}
	}

	errString := object.AssertType(obj.TypeEncoding, object.ObjTypeString)
	errInt := object.AssertType(obj.TypeEncoding, object.ObjTypeInt)
	if errString != nil && errInt != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	value, err := floatValue(obj.Value)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral("value is not a valid float"),
		}
	}
	value += incr
	if math.IsInf(value, 0) {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrValueOutOfRange,
		}
	}
	strValue := formatFloat(value, true)

	oType, oEnc := deduceTypeEncoding(strValue)

	// Remove the trailing decimal for integer values
	// to maintain consistency with redis
	strValue = strings.TrimSuffix(strValue, ".0")

	obj.Value = strValue
	obj.TypeEncoding = oType | oEnc

	return &EvalResponse{
		Result: strValue,
		Error:  nil,
	}
}

// floatValue returns the float64 value for an interface which
// contains either a string or an int.
func floatValue(value interface{}) (float64, error) {
	switch raw := value.(type) {
	case string:
		parsed, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return 0, err
		}
		return parsed, nil
	case int64:
		return float64(raw), nil
	}

	return 0, fmt.Errorf(diceerrors.IntOrFloatErr)
}

// ZPOPMIN Removes and returns the member with the lowest score from the sorted set at the specified key.
// If multiple members have the same score, the one that comes first alphabetically is returned.
// You can also specify a count to remove and return multiple members at once.
// If the set is empty, it returns an empty result.
func evalZPOPMIN(args []string, store *dstore.Store) *EvalResponse {
	// Incorrect number of arguments should return error
	if len(args) < 1 || len(args) > 2 {
		return &EvalResponse{
			Result: clientio.NIL,
			Error:  diceerrors.ErrWrongArgumentCount("ZPOPMIN"),
		}
	}

	key := args[0]        // Key argument
	obj := store.Get(key) // Getting sortedSet object from store

	// If the sortedSet is nil, return an empty list
	if obj == nil {
		return &EvalResponse{
			Result: []string{},
			Error:  nil,
		}
	}

	sortedSet, err := sortedset.FromObject(obj)
	if err != nil {
		return &EvalResponse{
			Result: clientio.NIL,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	count := 1
	// Check if the count argument is provided.
	if len(args) == 2 {
		countArg, err := strconv.Atoi(args[1])
		if err != nil {
			// Return an error if the argument is not a valid integer
			return &EvalResponse{
				Result: clientio.NIL,
				Error:  diceerrors.ErrIntegerOutOfRange,
			}
		}
		count = countArg
	}

	// If count is less than 1, empty array is returned
	if count < 1 {
		return &EvalResponse{
			Result: []string{},
			Error:  nil,
		}
	}

	// If the count argument is present, return all the members with lowest score sorted in ascending order.
	// If there are multiple lowest scores with same score value, it sorts the members in lexographical order of member name
	results := sortedSet.GetMin(count)

	return &EvalResponse{
		Result: results,
		Error:  nil,
	}
}

// evalHLEN returns the number of fields contained in the hash stored at key.
//
// If key doesn't exist, it returns 0.
//
// Usage: HLEN key
func evalHLEN(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("HLEN"),
		}
	}

	key := args[0]

	obj := store.Get(key)

	if obj == nil {
		return &EvalResponse{
			Result: clientio.IntegerZero,
			Error:  nil,
		}
	}

	if err := object.AssertTypeAndEncoding(obj.TypeEncoding, object.ObjTypeHashMap, object.ObjEncodingHashMap); err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	hashMap := obj.Value.(HashMap)
	return &EvalResponse{
		Result: len(hashMap),
		Error:  nil,
	}
}

// evalHSTRLEN returns the length of value associated with field in the hash stored at key.
//
// This command returns 0, if the specified field doesn't exist in the key
//
// If key doesn't exist, it returns 0.
//
// Usage: HSTRLEN key field value
func evalHSTRLEN(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 2 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("HSTRLEN"),
		}
	}

	key := args[0]
	hmKey := args[1]
	obj := store.Get(key)

	var hashMap HashMap

	if obj != nil {
		if err := object.AssertTypeAndEncoding(obj.TypeEncoding, object.ObjTypeHashMap, object.ObjEncodingHashMap); err != nil {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongTypeOperation,
			}
		}
		hashMap = obj.Value.(HashMap)
	} else {
		return &EvalResponse{
			Result: clientio.IntegerZero,
			Error:  nil,
		}
	}

	val, ok := hashMap.Get(hmKey)
	// Return 0, if specified field doesn't exist in the HashMap.
	if ok {
		return &EvalResponse{
			Result: len(*val),
			Error:  nil,
		}
	}
	return &EvalResponse{
		Result: clientio.IntegerZero,
		Error:  nil,
	}
}

// evalHSCAN return a two element multi-bulk reply, where the first element is a string representing the cursor,
// and the second element is a multi-bulk with an array of elements.
//
// The array of elements contain two elements, a field and a value, for every returned element of the Hash.
//
// If key doesn't exist, it returns an array containing 0 and empty array.
//
// Usage: HSCAN key cursor [MATCH pattern] [COUNT count]
func evalHSCAN(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 2 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("HSCAN"),
		}
	}

	key := args[0]
	cursor, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrIntegerOutOfRange,
		}
	}

	obj := store.Get(key)
	if obj == nil {
		return &EvalResponse{
			Result: []interface{}{"0", []string{}},
			Error:  nil,
		}
	}

	if err := object.AssertTypeAndEncoding(obj.TypeEncoding, object.ObjTypeHashMap, object.ObjEncodingHashMap); err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	hashMap := obj.Value.(HashMap)
	pattern := "*"
	count := 10

	// Parse optional arguments
	for i := 2; i < len(args); i += 2 {
		switch strings.ToUpper(args[i]) {
		case "MATCH":
			if i+1 < len(args) {
				pattern = args[i+1]
			}
		case CountConst:
			if i+1 < len(args) {
				parsedCount, err := strconv.Atoi(args[i+1])
				if err != nil || parsedCount < 1 {
					return &EvalResponse{
						Result: nil,
						Error:  diceerrors.ErrIntegerOutOfRange,
					}
				}
				count = parsedCount
			}
		}
	}

	// Note that this implementation has a time complexity of O(N), where N is the number of keys in 'hashMap'.
	// This is in contrast to Redis, which implements HSCAN in O(1) time complexity by maintaining a cursor.
	keys := make([]string, 0, len(hashMap))
	for k := range hashMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	matched := 0
	results := make([]string, 0, count*2)
	newCursor := 0

	g, err := glob.Compile(pattern)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral(fmt.Sprintf("Invalid glob pattern: %s", err)),
		}
	}

	// Scan the keys and add them to the results if they match the pattern
	for i := int(cursor); i < len(keys); i++ {
		if g.Match(keys[i]) {
			results = append(results, keys[i], hashMap[keys[i]])
			matched++
			if matched >= count {
				newCursor = i + 1
				break
			}
		}
	}

	// If we've scanned all keys, reset cursor to 0
	if newCursor >= len(keys) {
		newCursor = 0
	}

	return &EvalResponse{
		Result: []interface{}{strconv.Itoa(newCursor), results},
		Error:  nil,
	}
}

// evalBF.RESERVE evaluates the BF.RESERVE command responsible for initializing a
// new bloom filter and allocation it's relevant parameters based on given inputs.
// If no params are provided, it uses defaults.
func evalBFRESERVE(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 3 {
		return makeEvalError(diceerrors.ErrWrongArgumentCount("BF.RESERVE"))
	}

	opts, err := newBloomOpts(args[1:])
	if err != nil {
		return makeEvalError(err)
	}

	_, err = CreateBloomFilter(args[0], store, opts)
	if err != nil {
		return makeEvalError(err)
	}
	return makeEvalResult(clientio.OK)
}

// evalBFADD evaluates the BF.ADD command responsible for adding an element to a bloom filter. If the filter does not
// exist, it will create a new one with default parameters.
func evalBFADD(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 2 {
		return makeEvalError(diceerrors.ErrWrongArgumentCount("BF.ADD"))
	}

	bloom, err := getOrCreateBloomFilter(args[0], store, nil)
	if err != nil {
		return makeEvalError(err)
	}

	result, err := bloom.add(args[1])
	if err != nil {
		return makeEvalError(err)
	}

	return makeEvalResult(result)
}

// evalBFEXISTS evaluates the BF.EXISTS command responsible for checking existence of an element in a bloom filter.
func evalBFEXISTS(args []string, store *dstore.Store) *EvalResponse {
	// todo must work with objects of
	if len(args) != 2 {
		return makeEvalError(diceerrors.ErrWrongArgumentCount("BF.EXISTS"))
	}

	bloom, err := GetBloomFilter(args[0], store)
	if err != nil {
		return makeEvalError(err)
	}
	if bloom == nil {
		return makeEvalResult(clientio.IntegerZero)
	}
	result, err := bloom.exists(args[1])
	if err != nil {
		return makeEvalError(err)
	}
	return makeEvalResult(result)
}

// evalBFINFO evaluates the BF.INFO command responsible for returning the
// parameters and metadata of an existing bloom filter.
func evalBFINFO(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 1 || len(args) > 2 {
		return makeEvalError(diceerrors.ErrWrongArgumentCount("BF.INFO"))
	}

	bloom, err := GetBloomFilter(args[0], store)

	if err != nil {
		return makeEvalError(err)
	}

	if bloom == nil {
		return makeEvalError(diceerrors.ErrGeneral("not found"))
	}
	opt := ""
	if len(args) == 2 {
		opt = args[1]
	}
	result, err := bloom.info(opt)

	if err != nil {
		return makeEvalError(err)
	}

	return makeEvalResult(result)
}

// This command removes the element with the maximum score from the sorted set.
// If two elements have the same score then the members are aligned in lexicographically and the lexicographically greater element is removed.
// There is a second optional element called count which specifies the number of element to be removed.
// Returns the removed elements from the sorted set.
func evalZPOPMAX(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 1 || len(args) > 2 {
		return &EvalResponse{
			Result: clientio.NIL,
			Error:  diceerrors.ErrWrongArgumentCount("ZPOPMAX"),
		}
	}

	key := args[0]
	obj := store.Get(key)

	count := 1
	if len(args) > 1 {
		ops, err := strconv.Atoi(args[1])
		if err != nil {
			return &EvalResponse{
				Result: clientio.NIL,
				Error:  diceerrors.ErrGeneral("value is out of range, must be positive"), // This error is thrown when then count argument is not an integer
			}
		}
		if ops <= 0 {
			return &EvalResponse{
				Result: []string{}, // Returns empty array when the count is less than or equal to  0
				Error:  nil,
			}
		}
		count = ops
	}

	if obj == nil {
		return &EvalResponse{
			Result: []string{}, // Returns empty array when the object with given key is not present in the store
			Error:  nil,
		}
	}

	var sortedSet *sortedset.Set
	sortedSet, err := sortedset.FromObject(obj)
	if err != nil {
		return &EvalResponse{
			Result: clientio.NIL,
			Error:  diceerrors.ErrWrongTypeOperation, // Returns this error when a key is present in the store but is not of type sortedset.Set
		}
	}

	var res []string = sortedSet.PopMax(count)

	return &EvalResponse{
		Result: res,
		Error:  nil,
	}
}

// evalJSONARRTRIM trim an array so that it contains only the specified inclusive range of elements
// an array of integer replies for each path, the array's new size, or nil, if the matching JSON value is not an array.
func evalJSONARRTRIM(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 4 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("JSON.ARRTRIM"),
		}
	}
	var err error

	start := args[2]
	stop := args[3]
	var startIdx, stopIdx int
	startIdx, err = strconv.Atoi(start)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrIntegerOutOfRange,
		}
	}
	stopIdx, err = strconv.Atoi(stop)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrIntegerOutOfRange,
		}
	}

	key := args[0]
	obj := store.Get(key)
	if obj == nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral("key does not exist"),
		}
	}

	errWithMessage := object.AssertTypeAndEncoding(obj.TypeEncoding, object.ObjTypeJSON, object.ObjEncodingJSON)
	if errWithMessage != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral(string(errWithMessage)),
		}
	}

	jsonData := obj.Value

	_, err = sonic.Marshal(jsonData)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral("Existing key has wrong Dice type"),
		}
	}

	path := args[1]
	expr, err := jp.ParseString(path)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrJSONPathNotFound(path),
		}
	}

	results := expr.Get(jsonData)
	if len(results) == 0 {
		return &EvalResponse{
			Result: clientio.RespEmptyArray,
			Error:  nil,
		}
	}

	var resultsArray []interface{}
	// Capture the modified data when modifying the root path
	newData, modifyErr := expr.Modify(jsonData, func(data any) (interface{}, bool) {
		arr, ok := data.([]interface{})
		if !ok {
			// Not an array
			resultsArray = append(resultsArray, nil)
			return data, false
		}

		updatedArray := trimElementAndUpdateArray(arr, startIdx, stopIdx)

		resultsArray = append(resultsArray, len(updatedArray))
		return updatedArray, true
	})
	if modifyErr != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral(fmt.Sprintf("ERR failed to modify JSON data: %v", modifyErr)),
		}
	}

	jsonData = newData
	obj.Value = jsonData

	return &EvalResponse{
		Result: resultsArray,
		Error:  nil,
	}
}

// evalJSONARRAPPEND appends the value(s) provided in the args to the given array path
// in the JSON object saved at key in arguments.
// Args must contain atleast a key, path and value.
// If the key does not exist or is expired, it returns response.NIL.
// If the object at given path is not an array, it returns response.NIL.
// Returns the new length of the array at path.
func evalJSONARRAPPEND(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 3 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("JSON.ARRAPPEND"),
		}
	}

	key := args[0]
	path := args[1]
	values := args[2:]

	obj := store.Get(key)
	if obj == nil {
		return &EvalResponse{
			Result: clientio.NIL,
			Error:  nil,
		}
	}
	errWithMessage := object.AssertTypeAndEncoding(obj.TypeEncoding, object.ObjTypeJSON, object.ObjEncodingJSON)
	if errWithMessage != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}
	jsonData := obj.Value

	expr, err := jp.ParseString(path)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrJSONPathNotFound(path),
		}
	}

	// Parse the input values as JSON
	parsedValues := make([]interface{}, len(values))
	for i, v := range values {
		var parsedValue interface{}
		err := sonic.UnmarshalString(v, &parsedValue)
		if err != nil {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral(err.Error()),
			}
		}
		parsedValues[i] = parsedValue
	}

	var resultsArray []interface{}
	modified := false

	// Capture the modified data when modifying the root path
	var newData interface{}
	var modifyErr error

	newData, modifyErr = expr.Modify(jsonData, func(data any) (interface{}, bool) {
		arr, ok := data.([]interface{})
		if !ok {
			// Not an array
			resultsArray = append(resultsArray, clientio.NIL)
			return data, false
		}

		// Append the parsed values to the array
		arr = append(arr, parsedValues...)

		resultsArray = append(resultsArray, int64(len(arr)))
		modified = true
		return arr, modified
	})

	if modifyErr != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral(modifyErr.Error()),
		}
	}

	if !modified {
		// If no modification was made, it means the path did not exist or was not an array
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrJSONPathNotFound(path),
		}
	}

	jsonData = newData
	obj.Value = jsonData

	return &EvalResponse{
		Result: resultsArray,
		Error:  nil,
	}
}

// evalJSONARRLEN return the length of the JSON array at path in key
// Returns an array of integer replies, an integer for each matching value,
// each is the array's length, or nil, if the matching value is not an array.
// Returns encoded error if the key doesn't exist or key is expired or the matching value is not an array.
// Returns encoded error response if incorrect number of arguments
func evalJSONARRLEN(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("JSON.ARRLEN"),
		}
	}
	key := args[0]

	// Retrieve the object from the database
	obj := store.Get(key)

	// If the object is not present in the store or if its nil, then we should simply return nil.
	if obj == nil {
		return &EvalResponse{
			Result: clientio.NIL,
			Error:  nil,
		}
	}

	errWithMessage := object.AssertTypeAndEncoding(obj.TypeEncoding, object.ObjTypeJSON, object.ObjEncodingJSON)
	if errWithMessage != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	jsonData := obj.Value

	_, err := sonic.Marshal(jsonData)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	// This is the case if only argument passed to JSON.ARRLEN is the key itself.
	// This is valid only if the key holds an array; otherwise, an error should be returned.
	if len(args) == 1 {
		if utils.GetJSONFieldType(jsonData) == utils.ArrayType {
			return &EvalResponse{
				Result: len(jsonData.([]interface{})),
				Error:  nil,
			}
		}
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	path := args[1] // Getting the path to find the length of the array
	expr, err := jp.ParseString(path)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrJSONPathNotFound(path),
		}
	}

	results := expr.Get(jsonData)

	// If there are no results, that means the JSONPath does not exist
	if len(results) == 0 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrJSONPathNotFound(path),
		}
	}

	// If the results are greater than one, we need to print them as a list
	// This condition should be updated in future when supporting Complex JSONPaths
	if len(results) > 1 {
		arrlenList := make([]interface{}, 0, len(results))
		for _, result := range results {
			switch utils.GetJSONFieldType(result) {
			case utils.ArrayType:
				arrlenList = append(arrlenList, len(result.([]interface{})))
			default:
				arrlenList = append(arrlenList, clientio.NIL)
			}
		}

		return &EvalResponse{
			Result: arrlenList,
			Error:  nil,
		}
	}

	// Single result should be printed as single integer instead of list
	jsonValue := results[0]

	if utils.GetJSONFieldType(jsonValue) == utils.ArrayType {
		return &EvalResponse{
			Result: len(jsonValue.([]interface{})),
			Error:  nil,
		}
	}

	// If execution reaches this point, the provided path either does not exist.
	return &EvalResponse{
		Result: nil,
		Error:  diceerrors.ErrJSONPathNotFound(path),
	}
}

// popElementAndUpdateArray removes an element at the given index
// Returns popped element, remaining array and error
func popElementAndUpdateArray(arr []any, index string) (popElem any, updatedArray []any, err error) {
	if len(arr) == 0 {
		return nil, nil, nil
	}

	var idx int
	// if index is empty, pop last element
	if index == "" {
		idx = len(arr) - 1
	} else {
		var err error
		idx, err = strconv.Atoi(index)
		if err != nil {
			return nil, nil, err
		}
		// convert index to a valid index
		idx = adjustIndex(idx, arr)
	}

	popElem = arr[idx]
	arr = append(arr[:idx], arr[idx+1:]...)

	return popElem, arr, nil
}

func evalJSONARRPOP(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("JSON.ARRPOP"),
		}
	}
	key := args[0]

	var path = defaultRootPath
	if len(args) >= 2 {
		path = args[1]
	}

	var index string
	if len(args) >= 3 {
		index = args[2]
	}

	// Retrieve the object from the database
	obj := store.Get(key)
	if obj == nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrKeyNotFound,
		}
	}

	errWithMessage := object.AssertTypeAndEncoding(obj.TypeEncoding, object.ObjTypeJSON, object.ObjEncodingJSON)
	if errWithMessage != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	jsonData := obj.Value
	_, err := sonic.Marshal(jsonData)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	if path == defaultRootPath {
		arr, ok := jsonData.([]any)
		// if value can not be converted to array, it is of another type
		// returns nil in this case similar to redis
		// also, return nil if array is empty
		if !ok || len(arr) == 0 {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongTypeOperation,
			}
		}
		popElem, arr, err := popElementAndUpdateArray(arr, index)
		if err != nil {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral(err.Error()),
			}
		}

		// save the remaining array
		newObj := store.NewObj(arr, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
		store.Put(key, newObj)

		return &EvalResponse{
			Result: popElem,
			Error:  nil,
		}
	}

	// if path is not root then extract value at path
	expr, err := jp.ParseString(path)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrJSONPathNotFound(path),
		}
	}
	results := expr.Get(jsonData)

	// process value at each path
	popArr := make([]any, 0, len(results))
	for _, result := range results {
		arr, ok := result.([]any)
		// if value can not be converted to array, it is of another type
		// returns nil in this case similar to redis
		// also, return nil if array is empty
		if !ok || len(arr) == 0 {
			popArr = append(popArr, clientio.NIL)
			continue
		}

		popElem, arr, err := popElementAndUpdateArray(arr, index)
		if err != nil {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral(err.Error()),
			}
		}

		// update array in place in the json object
		err = expr.Set(jsonData, arr)
		if err != nil {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral(err.Error()),
			}
		}

		popArr = append(popArr, popElem)
	}
	return &EvalResponse{
		Result: popArr,
		Error:  nil,
	}
}

// evalJSONARRINSERT insert the json values into the array at path before the index (shifts to the right)
// returns an array of integer replies for each path, the array's new size, or nil.
func evalJSONARRINSERT(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 4 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("JSON.ARRINSERT"),
		}
	}
	key := args[0]
	obj := store.Get(key)
	if obj == nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral("key does not exist"),
		}
	}

	errWithMessage := object.AssertTypeAndEncoding(obj.TypeEncoding, object.ObjTypeJSON, object.ObjEncodingJSON)
	if errWithMessage != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral(string(errWithMessage)),
		}
	}

	jsonData := obj.Value
	var err error
	_, err = sonic.Marshal(jsonData)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral("Existing key has wrong Dice type"),
		}
	}

	path := args[1]
	expr, err := jp.ParseString(path)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrJSONPathNotFound(path),
		}
	}

	results := expr.Get(jsonData)
	if len(results) == 0 {
		return &EvalResponse{
			Result: clientio.RespEmptyArray,
			Error:  nil,
		}
	}
	index := args[2]
	var idx int
	idx, err = strconv.Atoi(index)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrIntegerOutOfRange,
		}
	}

	values := args[3:]
	// Parse the input values as JSON
	parsedValues := make([]interface{}, len(values))
	for i, v := range values {
		var parsedValue interface{}
		err := sonic.UnmarshalString(v, &parsedValue)
		if err != nil {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral(err.Error()),
			}
		}
		parsedValues[i] = parsedValue
	}

	var resultsArray []interface{}
	// Capture the modified data when modifying the root path
	modified := false
	newData, modifyErr := expr.Modify(jsonData, func(data any) (interface{}, bool) {
		arr, ok := data.([]interface{})
		if !ok {
			// Not an array
			resultsArray = append(resultsArray, nil)
			return data, false
		}

		// Append the parsed values to the array
		updatedArray, insertErr := insertElementAndUpdateArray(arr, idx, parsedValues)
		if insertErr != nil {
			err = insertErr
			return data, false
		}
		modified = true
		resultsArray = append(resultsArray, len(updatedArray))
		return updatedArray, true
	})
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral(err.Error()),
		}
	}

	if modifyErr != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral(fmt.Sprintf("ERR failed to modify JSON data: %v", modifyErr)),
		}
	}

	if !modified {
		return &EvalResponse{
			Result: resultsArray,
			Error:  nil,
		}
	}

	jsonData = newData
	obj.Value = jsonData
	return &EvalResponse{
		Result: resultsArray,
		Error:  nil,
	}
}

// evalJSONOBJKEYS retrieves the keys of a JSON object stored at path specified.
// It takes two arguments: the key where the JSON document is stored, and an optional JSON path.
// It returns a list of keys from the object at the specified path or an error if the path is invalid.
func evalJSONOBJKEYS(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("JSON.OBJKEYS"),
		}
	}

	key := args[0]
	// Default path is root if not specified
	path := defaultRootPath
	if len(args) > 1 {
		path = args[1]
	}

	// Retrieve the object from the database
	obj := store.Get(key)
	if obj == nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral("could not perform this operation on a key that doesn't exist"),
		}
	}

	// Check if the object is of JSON type
	errWithMessage := object.AssertTypeAndEncoding(obj.TypeEncoding, object.ObjTypeJSON, object.ObjEncodingJSON)
	if errWithMessage != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral(string(errWithMessage)),
		}
	}

	jsonData := obj.Value
	_, err := sonic.Marshal(jsonData)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral("Existing key has wrong Dice type"),
		}
	}

	// If path is root, return all keys of the entire JSON
	if len(args) == 1 {
		if utils.GetJSONFieldType(jsonData) == utils.ObjectType {
			keys := make([]string, 0)
			for key := range jsonData.(map[string]interface{}) {
				keys = append(keys, key)
			}
			return &EvalResponse{
				Result: keys,
				Error:  nil,
			}
		}
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	// Parse the JSONPath expression
	expr, err := jp.ParseString(path)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral(err.Error()),
		}
	}

	// Execute the JSONPath query
	results := expr.Get(jsonData)
	if len(results) == 0 {
		return &EvalResponse{
			Result: clientio.RespEmptyArray,
			Error:  nil,
		}
	}

	keysList := make([]interface{}, 0, len(results))

	for _, result := range results {
		switch utils.GetJSONFieldType(result) {
		case utils.ObjectType:
			keys := make([]string, 0)
			for key := range result.(map[string]interface{}) {
				keys = append(keys, key)
			}
			keysList = append(keysList, keys)
		default:
			keysList = append(keysList, nil)
		}
	}

	return &EvalResponse{
		Result: keysList,
		Error:  nil,
	}
}

// GETEX key [EX seconds | PX milliseconds | EXAT unix-time-seconds |
// PXAT unix-time-milliseconds | PERSIST]
// Get the value of key and optionally set its expiration.
// GETEX is similar to GET, but is a write command with additional options.
// The GETEX command supports a set of options that modify its behavior:
// EX seconds -- Set the specified expire time, in seconds.
// PX milliseconds -- Set the specified expire time, in milliseconds.
// EXAT timestamp-seconds -- Set the specified Unix time at which the key will expire, in seconds.
// PXAT timestamp-milliseconds -- Set the specified Unix time at which the key will expire, in milliseconds.
// PERSIST -- Remove the time to live associated with the key.
// The RESP value of the key is encoded and then returned
// evalGET returns response.RespNIL if key is expired or it does not exist
func evalGETEX(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("GETEX"),
		}
	}

	var key = args[0]

	var exDurationMs int64 = -1
	var state = Uninitialized
	var persist = false
	for i := 1; i < len(args); i++ {
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
			if exDuration <= 0 || exDuration > maxExDuration {
				return &EvalResponse{
					Result: nil,
					Error:  diceerrors.ErrInvalidExpireTime("GETEX"),
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

			if exDuration < 0 || exDuration > maxExDuration {
				return &EvalResponse{
					Result: nil,
					Error:  diceerrors.ErrInvalidExpireTime("GETEX"),
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

		case Persist:
			if state != Uninitialized {
				return &EvalResponse{
					Result: nil,
					Error:  diceerrors.ErrSyntax,
				}
			}
			persist = true
			state = Initialized
		default:
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrIntegerOutOfRange,
			}
		}
	}

	// Get the key from the hash table
	obj := store.Get(key)

	if obj == nil {
		return &EvalResponse{
			Result: clientio.NIL,
			Error:  nil,
		}
	}

	if object.AssertType(obj.TypeEncoding, object.ObjTypeSet) == nil ||
		object.AssertType(obj.TypeEncoding, object.ObjTypeJSON) == nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	// Get EvalResponse with correct data type
	getResp := evalGET([]string{key}, store)

	// If there is an error return the error response
	if getResp.Error != nil {
		return getResp
	}

	if state == Initialized {
		if persist {
			dstore.DelExpiry(obj, store)
		} else {
			store.SetExpiry(obj, exDurationMs)
		}
	}

	// return an EvalResponse with the value
	return getResp
}

// evalGETDEL returns the value for the queried key in args
// The key should be the only param in args
// The RESP value of the key is encoded and then returned
// In evalGETDEL  If the key exists, it will be deleted before its value is returned.
// evalGETDEL returns response.RespNIL if key is expired or it does not exist
func evalGETDEL(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("GETDEL"),
		}
	}

	key := args[0]

	// getting the key based on previous touch value
	obj := store.GetNoTouch(key)

	// if key does not exist, return RESP encoded nil
	if obj == nil {
		return &EvalResponse{
			Result: clientio.NIL,
			Error:  nil,
		}
	}

	// If the object exists, check if it is a Set object.
	if err := object.AssertType(obj.TypeEncoding, object.ObjTypeSet); err == nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	// If the object exists, check if it is a JSON object.
	if err := object.AssertType(obj.TypeEncoding, object.ObjTypeJSON); err == nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	// Get the key from the hash table
	objVal := store.GetDel(key)

	// Decode and return the value based on its encoding
	switch _, oEnc := object.ExtractTypeEncoding(objVal); oEnc {
	case object.ObjEncodingInt:
		// Value is stored as an int64, so use type assertion
		if IsInt64(objVal.Value) {
			return &EvalResponse{
				Result: objVal.Value,
				Error:  nil,
			}
		} else if IsString(objVal.Value) {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrUnexpectedType("int64", "string"),
			}
		} else {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrUnexpectedType("int64", "unknown"),
			}
		}

	case object.ObjEncodingEmbStr, object.ObjEncodingRaw:
		// Value is stored as a string, use type assertion
		if IsString(objVal.Value) {
			return &EvalResponse{
				Result: objVal.Value,
				Error:  nil,
			}
		} else if IsInt64(objVal.Value) {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrUnexpectedType("string", "int64"),
			}
		} else {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrUnexpectedType("string", "unknown"),
			}
		}

	case object.ObjEncodingByteArray:
		// Value is stored as a bytearray, use type assertion
		if val, ok := objVal.Value.(*ByteArray); ok {
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

// helper function to insert key value in hashmap associated with the given hash
func insertInHashMap(args []string, store *dstore.Store) (int64, error) {
	key := args[0]

	obj := store.Get(key)

	var hashMap HashMap

	if obj != nil {
		if err := object.AssertTypeAndEncoding(obj.TypeEncoding, object.ObjTypeHashMap, object.ObjEncodingHashMap); err != nil {
			return 0, diceerrors.ErrWrongTypeOperation
		}
		hashMap = obj.Value.(HashMap)
	}

	keyValuePairs := args[1:]

	hashMap, numKeys, err := hashMapBuilder(keyValuePairs, hashMap)
	if err != nil {
		return 0, err
	}

	obj = store.NewObj(hashMap, -1, object.ObjTypeHashMap, object.ObjEncodingHashMap)
	store.Put(key, obj)

	return numKeys, nil
}

// evalHSET sets the specified fields to their
// respective values in a hashmap stored at key
//
// This command overwrites the values of specified
// fields that exist in the hash.
//
// If key doesn't exist, a new key holding a hash is created.
//
// Usage: HSET key field value [field value ...]
func evalHSET(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 3 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("HSET"),
		}
	}

	numKeys, err := insertInHashMap(args, store)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  err,
		}
	}

	return &EvalResponse{
		Result: numKeys,
		Error:  nil,
	}
}

// evalHMSET sets the specified fields to their
// respective values in a hashmap stored at key
//
// This command overwrites the values of specified
// fields that exist in the hash.
//
// If key doesn't exist, a new key holding a hash is created.
//
// Usage: HMSET key field value [field value ...]
func evalHMSET(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 3 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("HMSET"),
		}
	}

	_, err := insertInHashMap(args, store)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  err,
		}
	}

	return &EvalResponse{
		Result: clientio.OK,
		Error:  nil,
	}
}

// evalHMGET returns an array of values associated with the given fields,
// in the same order as they are requested.
// If a field does not exist, returns a corresponding nil value in the array.
// If the key does not exist, returns an array of nil values.
func evalHMGET(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 2 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("HMGET"),
		}
	}
	key := args[0]

	// Fetch the object from the store using the key
	obj := store.Get(key)

	// Initialize the results slice
	results := make([]interface{}, len(args[1:]))

	// If the object is nil, return empty results for all requested fields
	if obj == nil {
		for i := range results {
			results[i] = nil // Return nil for non-existent fields
		}
		return &EvalResponse{
			Result: results,
			Error:  nil,
		}
	}

	// Assert that the object is of type HashMap
	if err := object.AssertTypeAndEncoding(obj.TypeEncoding, object.ObjTypeHashMap, object.ObjEncodingHashMap); err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	hashMap := obj.Value.(HashMap)

	// Loop through the requested fields
	for i, hmKey := range args[1:] {
		hmValue, present := hashMap.Get(hmKey)
		if present {
			results[i] = *hmValue // Set the value if it exists
		} else {
			results[i] = nil // Set to nil if field does not exist
		}
	}

	// Return the results and no error
	return &EvalResponse{
		Result: results,
		Error:  nil,
	}
}

func evalHGET(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 2 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("HGET"),
		}
	}

	key := args[0]
	hmKey := args[1]

	response := getValueFromHashMap(key, hmKey, store)
	if response.Error != nil {
		return &EvalResponse{
			Result: nil,
			Error:  response.Error,
		}
	}

	return &EvalResponse{
		Result: response.Result,
		Error:  nil,
	}
}

func evalHSETNX(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 3 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("HSETNX"),
		}
	}

	key := args[0]
	hmKey := args[1]

	response := getValueFromHashMap(key, hmKey, store)
	if response.Error != nil {
		return &EvalResponse{
			Result: nil,
			Error:  response.Error,
		}
	}

	if response.Result != clientio.NIL {
		return &EvalResponse{
			Result: int64(0),
			Error:  nil,
		}
	}

	evalHSET(args, store)

	return &EvalResponse{
		Result: int64(1),
		Error:  nil,
	}
}

func evalHDEL(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 2 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("HDEL"),
		}
	}

	key := args[0]
	fields := args[1:]

	obj := store.Get(key)

	if obj == nil {
		return &EvalResponse{
			Result: int64(0),
			Error:  nil,
		}
	}

	if err := object.AssertTypeAndEncoding(obj.TypeEncoding, object.ObjTypeHashMap, object.ObjEncodingHashMap); err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	hashMap := obj.Value.(HashMap)
	count := int64(0)
	for _, field := range fields {
		if _, ok := hashMap[field]; ok {
			delete(hashMap, field)
			count++
		}
	}

	if count > 0 {
		store.Put(key, obj)
	}

	return &EvalResponse{
		Result: count,
		Error:  nil,
	}
}

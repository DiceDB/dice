package eval

import (
	"errors"
	"fmt"
	"math"
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
	"github.com/ohler55/ojg/jp"
)

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

func evalLPUSH(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 2 {
		return &EvalResponse{
			Result: clientio.NIL,
			Error:  diceerrors.ErrWrongArgumentCount("LPUSH"),
		}
	}

	obj := store.Get(args[0])
	if obj == nil {
		obj = store.NewObj(NewDeque(), -1, object.ObjTypeByteList, object.ObjEncodingDeque)
	}

	if err := object.AssertType(obj.TypeEncoding, object.ObjTypeByteList); err != nil {
		return &EvalResponse{
			Result: clientio.NIL,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	if err := object.AssertEncoding(obj.TypeEncoding, object.ObjEncodingDeque); err != nil {
		return &EvalResponse{
			Result: clientio.NIL,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	store.Put(args[0], obj)
	for i := 1; i < len(args); i++ {
		obj.Value.(*Deque).LPush(args[i])
	}

	deq := obj.Value.(*Deque)

	return &EvalResponse{
		Result: deq.Length,
		Error:  nil,
	}
}

func evalRPUSH(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 2 {
		return &EvalResponse{
			Result: clientio.NIL,
			Error:  diceerrors.ErrWrongArgumentCount("RPUSH"),
		}
	}

	obj := store.Get(args[0])
	if obj == nil {
		obj = store.NewObj(NewDeque(), -1, object.ObjTypeByteList, object.ObjEncodingDeque)
	}

	if err := object.AssertType(obj.TypeEncoding, object.ObjTypeByteList); err != nil {
		return &EvalResponse{
			Result: clientio.NIL,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	if err := object.AssertEncoding(obj.TypeEncoding, object.ObjEncodingDeque); err != nil {
		return &EvalResponse{
			Result: clientio.NIL,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	store.Put(args[0], obj)
	for i := 1; i < len(args); i++ {
		obj.Value.(*Deque).RPush(args[i])
	}

	deq := obj.Value.(*Deque)

	return &EvalResponse{
		Result: deq.Length,
		Error:  nil,
	}
}

func evalLPOP(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 1 {
		return &EvalResponse{
			Result: clientio.NIL,
			Error:  diceerrors.ErrWrongArgumentCount("LPOP"),
		}
	}

	obj := store.Get(args[0])
	if obj == nil {
		return &EvalResponse{
			Result: clientio.NIL,
			Error:  nil,
		}
	}

	if err := object.AssertType(obj.TypeEncoding, object.ObjTypeByteList); err != nil {
		return &EvalResponse{
			Result: clientio.NIL,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	if err := object.AssertEncoding(obj.TypeEncoding, object.ObjEncodingDeque); err != nil {
		return &EvalResponse{
			Result: clientio.NIL,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	deq := obj.Value.(*Deque)
	x, err := deq.LPop()
	if err != nil {
		if errors.Is(err, ErrDequeEmpty) {
			return &EvalResponse{
				Result: clientio.NIL,
				Error:  nil,
			}
		}
		panic(fmt.Sprintf("unknown error: %v", err))
	}

	return &EvalResponse{
		Result: x,
		Error:  nil,
	}
}

func evalRPOP(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 1 {
		return &EvalResponse{
			Result: clientio.NIL,
			Error:  diceerrors.ErrWrongArgumentCount("RPOP"),
		}
	}

	obj := store.Get(args[0])
	if obj == nil {
		return &EvalResponse{
			Result: clientio.NIL,
			Error:  nil,
		}
	}

	if err := object.AssertType(obj.TypeEncoding, object.ObjTypeByteList); err != nil {
		return &EvalResponse{
			Result: clientio.NIL,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	if err := object.AssertEncoding(obj.TypeEncoding, object.ObjEncodingDeque); err != nil {
		return &EvalResponse{
			Result: clientio.NIL,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	deq := obj.Value.(*Deque)
	x, err := deq.RPop()
	if err != nil {
		if errors.Is(err, ErrDequeEmpty) {
			return &EvalResponse{
				Result: clientio.NIL,
				Error:  nil,
			}
		}
		panic(fmt.Sprintf("unknown error: %v", err))
	}

	return &EvalResponse{
		Result: x,
		Error:  nil,
	}
}

func evalLLEN(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 1 {
		return &EvalResponse{
			Result: clientio.NIL,
			Error:  diceerrors.ErrWrongArgumentCount("LLEN"),
		}
	}

	obj := store.Get(args[0])
	if obj == nil {
		return &EvalResponse{
			Result: clientio.IntegerZero,
			Error:  nil,
		}
	}

	if err := object.AssertTypeAndEncoding(obj.TypeEncoding, object.ObjTypeByteList, object.ObjEncodingDeque); err != nil {
		return &EvalResponse{
			Result: clientio.NIL,
			Error:  diceerrors.ErrWrongTypeOperation,
		}
	}

	deq := obj.Value.(*Deque)
	return &EvalResponse{
		Result: deq.Length,
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

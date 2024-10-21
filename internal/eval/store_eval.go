package eval

import (
	"fmt"
	"math"
	"math/bits"
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
		return &EvalResponse{
			Result: []interface{}{rank, score},
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
			Error:  diceerrors.ErrGeneral(diceerrors.WrongTypeHllErr),
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
					Error:  diceerrors.ErrGeneral(diceerrors.WrongTypeHllErr),
				}
			}
			err := unionHll.Merge(currKeyHll)
			if err != nil {
				return &EvalResponse{
					Result: nil,
					Error:  diceerrors.ErrGeneral(diceerrors.InvalidHllErr),
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
				Error:  diceerrors.ErrGeneral(diceerrors.WrongTypeHllErr),
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
					Error:  diceerrors.ErrGeneral(diceerrors.WrongTypeHllErr),
				}
			}

			err := mergedHll.Merge(currKeyHll)
			if err != nil {
				return &EvalResponse{
					Result: nil,
					Error:  diceerrors.ErrGeneral(diceerrors.InvalidHllErr),
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

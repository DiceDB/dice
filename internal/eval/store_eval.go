package eval

import (
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

	path := args[1]

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
			objectLen = append(objectLen, nil)
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

package eval

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unsafe"

	"github.com/dicedb/dice/internal/object"
	"github.com/rs/xid"

	"github.com/dicedb/dice/internal/sql"

	"github.com/axiomhq/hyperloglog"
	"github.com/bytedance/sonic"
	"github.com/charmbracelet/log"
	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/internal/auth"
	"github.com/dicedb/dice/internal/clientio"
	"github.com/dicedb/dice/internal/comm"
	diceerrors "github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/querywatcher"
	"github.com/dicedb/dice/internal/server/utils"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/ohler55/ojg/jp"
)

type exDurationState int

const (
	Uninitialized exDurationState = iota
	Initialized
)

var (
	TxnCommands       map[string]bool
	serverID          string
	diceCommandsCount int
)

const defaultRootPath = "$"
const maxExDuration = 9223372036854775

func init() {
	diceCommandsCount = len(DiceCmds)
	TxnCommands = map[string]bool{"EXEC": true, "DISCARD": true}
	serverID = fmt.Sprintf("%s:%d", config.DiceConfig.Server.Addr, config.DiceConfig.Server.Port)
}

// evalPING returns with an encoded "PONG"
// If any message is added with the ping command,
// the message will be returned.
func evalPING(args []string, store *dstore.Store) []byte {
	var b []byte

	if len(args) >= 2 {
		return diceerrors.NewErrArity("PING")
	}

	if len(args) == 0 {
		b = clientio.Encode("PONG", true)
	} else {
		b = clientio.Encode(args[0], false)
	}

	return b
}

// EvalAUTH returns with an encoded "OK" if the user is authenticated
// If the user is not authenticated, it returns with an encoded error message
func EvalAUTH(args []string, c *comm.Client) []byte {
	var err error

	if config.DiceConfig.Auth.Password == "" {
		return diceerrors.NewErrWithMessage("AUTH <password> called without any password configured for the default user. Are you sure your configuration is correct?")
	}

	username := auth.DefaultUserName
	var password string

	if len(args) == 1 {
		password = args[0]
	} else if len(args) == 2 {
		username, password = args[0], args[1]
	} else {
		return diceerrors.NewErrArity("AUTH")
	}

	if err = c.Session.Validate(username, password); err != nil {
		return clientio.Encode(err, false)
	}
	return clientio.RespOK
}

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
func evalSET(args []string, store *dstore.Store) []byte {
	if len(args) <= 1 {
		return diceerrors.NewErrArity("SET")
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

			if exDuration <= 0 || exDuration >= maxExDuration {
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
				return clientio.RespNIL
			}
		case NX:
			obj := store.Get(key)
			if obj != nil {
				return clientio.RespNIL
			}
		case KEEPTTL, Keepttl:
			keepttl = true
		default:
			return diceerrors.NewErrWithMessage(diceerrors.SyntaxErr)
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
		return clientio.Encode(fmt.Errorf("ERR unsupported encoding: %d", oEnc), false)
	}

	// putting the k and value in a Hash Table
	store.Put(key, store.NewObj(storedValue, exDurationMs, oType, oEnc), dstore.WithKeepTTL(keepttl))

	return clientio.RespOK
}

// evalMSET puts multiple <key, value> pairs in db as in the args
// MSET is atomic, so all given keys are set at once.
// args must contain key and value pairs.

// Returns encoded error response if at least a <key, value> pair is not part of args
// Returns encoded OK RESP once new entries are added
// If the key already exists then the value will be overwritten and expiry will be discarded
func evalMSET(args []string, store *dstore.Store) []byte {
	if len(args) <= 1 || len(args)%2 != 0 {
		return diceerrors.NewErrArity("MSET")
	}

	// MSET does not have expiry support
	var exDurationMs int64 = -1

	insertMap := make(map[string]*object.Obj, len(args)/2)
	for i := 0; i < len(args); i += 2 {
		key, value := args[i], args[i+1]
		oType, oEnc := deduceTypeEncoding(value)
		var storedValue interface{}
		switch oEnc {
		case object.ObjEncodingInt:
			storedValue, _ = strconv.ParseInt(value, 10, 64)
		case object.ObjEncodingEmbStr, object.ObjEncodingRaw:
			storedValue = value
		default:
			return clientio.Encode(fmt.Errorf("ERR unsupported encoding: %d", oEnc), false)
		}
		insertMap[key] = store.NewObj(storedValue, exDurationMs, oType, oEnc)
	}

	store.PutAll(insertMap)
	return clientio.RespOK
}

// evalGET returns the value for the queried key in args
// The key should be the only param in args
// The RESP value of the key is encoded and then returned
// evalGET returns response.RespNIL if key is expired or it does not exist
func evalGET(args []string, store *dstore.Store) []byte {
	if len(args) != 1 {
		return diceerrors.NewErrArity("GET")
	}

	key := args[0]

	obj := store.Get(key)

	// if key does not exist, return RESP encoded nil
	if obj == nil {
		return clientio.RespNIL
	}

	// Decode and return the value based on its encoding
	switch _, oEnc := object.ExtractTypeEncoding(obj); oEnc {
	case object.ObjEncodingInt:
		// Value is stored as an int64, so use type assertion
		if val, ok := obj.Value.(int64); ok {
			return clientio.Encode(val, false)
		}
		return diceerrors.NewErrWithFormattedMessage("expected int64 but got another type: %s", obj.Value)

	case object.ObjEncodingEmbStr, object.ObjEncodingRaw:
		// Value is stored as a string, use type assertion
		if val, ok := obj.Value.(string); ok {
			return clientio.Encode(val, false)
		}
		return diceerrors.NewErrWithMessage("expected string but got another type")

	case object.ObjEncodingByteArray:
		// Value is stored as a bytearray, use type assertion
		if val, ok := obj.Value.(*ByteArray); ok {
			return clientio.Encode(string(val.data), false)
		}
		return diceerrors.NewErrWithMessage(diceerrors.WrongTypeErr)

	default:
		return diceerrors.NewErrWithMessage(diceerrors.WrongTypeErr)
	}
}

// evalDBSIZE returns the number of keys in the database.
func evalDBSIZE(args []string, store *dstore.Store) []byte {
	if len(args) > 0 {
		return diceerrors.NewErrArity("DBSIZE")
	}

	// return the RESP encoded value
	return clientio.Encode(dstore.KeyspaceStat[0]["keys"], false)
}

// evalGETDEL returns the value for the queried key in args
// The key should be the only param in args
// The RESP value of the key is encoded and then returned
// In evalGETDEL  If the key exists, it will be deleted before its value is returned.
// evalGETDEL returns response.RespNIL if key is expired or it does not exist
func evalGETDEL(args []string, store *dstore.Store) []byte {
	if len(args) != 1 {
		return diceerrors.NewErrArity("GETDEL")
	}

	key := args[0]

	// getting the key based on previous touch value
	obj := store.GetNoTouch(key)

	// if key does not exist, return RESP encoded nil
	if obj == nil {
		return clientio.RespNIL
	}

	// If the object exists, check if it is a Set object.
	if err := object.AssertType(obj.TypeEncoding, object.ObjTypeSet); err == nil {
		return diceerrors.NewErrWithFormattedMessage(diceerrors.WrongTypeErr)
	}

	// If the object exists, check if it is a JSON object.
	if err := object.AssertType(obj.TypeEncoding, object.ObjTypeJSON); err == nil {
		return diceerrors.NewErrWithFormattedMessage(diceerrors.WrongTypeErr)
	}

	// Get the key from the hash table
	objVal := store.GetDel(key)

	// Decode and return the value based on its encoding
	switch _, oEnc := object.ExtractTypeEncoding(objVal); oEnc {
	case object.ObjEncodingInt:
		// Value is stored as an int64, so use type assertion
		if val, ok := objVal.Value.(int64); ok {
			return clientio.Encode(val, false)
		}
		return diceerrors.NewErrWithFormattedMessage("expected int64 but got another type: %s", objVal.Value)

	case object.ObjEncodingEmbStr, object.ObjEncodingRaw:
		// Value is stored as a string, use type assertion
		if val, ok := objVal.Value.(string); ok {
			return clientio.Encode(val, false)
		}
		return diceerrors.NewErrWithMessage("expected string but got another type")

	case object.ObjEncodingByteArray:
		// Value is stored as a bytearray, use type assertion
		if val, ok := objVal.Value.(*ByteArray); ok {
			return clientio.Encode(string(val.data), false)
		}
		return diceerrors.NewErrWithMessage(diceerrors.WrongTypeErr)

	default:
		return diceerrors.NewErrWithMessage(diceerrors.WrongTypeErr)
	}
}

// evalJSONDEBUG reports value's memmory usage in bytes
// Returns arity error if subcommand is missing
// Supports only two subcommand as of now - HELP and MEMORY
func evalJSONDebug(args []string, store *dstore.Store) []byte {
	if len(args) < 1 {
		return diceerrors.NewErrArity("JSON.DEBUG")
	}
	subcommand := strings.ToUpper(args[0])
	switch subcommand {
	case Help:
		return evalJSONDebugHelp()
	case Memory:
		return evalJSONDebugMemory(args[1:], store)
	default:
		return diceerrors.NewErrWithFormattedMessage("unknown subcommand - try `JSON.DEBUG HELP`")
	}
}

// evalJSONDebugHelp implements HELP subcommand for evalJSONDebug
// It returns help text
// It ignore any other args
func evalJSONDebugHelp() []byte {
	memoryText := "MEMORY <key> [path] - reports memory usage"
	helpText := "HELP                - this message"
	message := []string{memoryText, helpText}
	return clientio.Encode(message, false)
}

// evalJSONDebugMemory implements MEMORY subcommand for evalJSONDebug
// It returns value's memory usage in bytes
func evalJSONDebugMemory(args []string, store *dstore.Store) []byte {
	if len(args) < 1 {
		return diceerrors.NewErrArity("json.debug")
	}
	key := args[0]

	// default path is root if not specified
	path := defaultRootPath
	if len(args) > 1 {
		path = args[1] // any more args are ignored for this command altogether
	}

	obj := store.Get(key)
	if obj == nil {
		return clientio.RespZero
	}

	// check if the object is a valid JSON
	errWithMessage := object.AssertTypeAndEncoding(obj.TypeEncoding, object.ObjTypeJSON, object.ObjEncodingJSON)
	if errWithMessage != nil {
		return errWithMessage
	}

	// handle root path
	if path == defaultRootPath {
		jsonData := obj.Value

		// memory used by json data
		size := calculateSizeInBytes(jsonData)
		if size == -1 {
			return diceerrors.NewErrWithMessage("unknown type")
		}

		// add memory used by storage object
		size += int(unsafe.Sizeof(obj)) + calculateSizeInBytes(obj.LastAccessedAt) + calculateSizeInBytes(obj.TypeEncoding)

		return clientio.Encode(size, false)
	}

	// handle nested paths
	var results []any = []any{}
	if path != defaultRootPath {
		// check if path is valid
		expr, err := jp.ParseString(path)
		if err != nil {
			return diceerrors.NewErrWithMessage("invalid JSON path")
		}

		results = expr.Get(obj.Value)

		// handle error cases
		if len(results) == 0 {
			// this block will return '[]' for out of bound index for an array json type
			// this will maintain consistency with redis
			isArray := utils.IsArray(obj.Value)
			if isArray {
				arr, ok := obj.Value.([]any)
				if !ok {
					return diceerrors.NewErrWithMessage("invalid array json")
				}

				// extract index from arg
				reg := regexp.MustCompile(`^\$\.?\[(\d+|\*)\]`)
				matches := reg.FindStringSubmatch(path)

				if len(matches) == 2 {
					// convert index to int
					index, err := strconv.Atoi(matches[1])
					if err != nil {
						return diceerrors.NewErrWithMessage("unable to extract index")
					}

					// if index is out of bound return empty array
					if index >= len(arr) {
						return clientio.RespEmptyArray
					}
				}
			}

			// for rest json types, throw error
			return diceerrors.NewErrWithFormattedMessage("Path '$.%v' does not exist", path)
		}
	}

	// get memory used by each path
	sizeList := make([]interface{}, 0, len(results))
	for _, result := range results {
		size := calculateSizeInBytes(result)
		sizeList = append(sizeList, size)
	}

	return clientio.Encode(sizeList, false)
}

func calculateSizeInBytes(value interface{}) int {
	switch convertedValue := value.(type) {
	case string:
		return int(unsafe.Sizeof(value)) + len(convertedValue)

	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool, nil:
		return int(unsafe.Sizeof(value))

	// object
	case map[string]interface{}:
		size := int(unsafe.Sizeof(value))
		for k, v := range convertedValue {
			size += int(unsafe.Sizeof(k)) + len(k) + calculateSizeInBytes(v)
		}
		return size

	// array
	case []interface{}:
		size := int(unsafe.Sizeof(value))
		for _, elem := range convertedValue {
			size += calculateSizeInBytes(elem)
		}
		return size

	// unknown type
	default:
		return -1
	}
}

// evaLJSONFORGET removes the field specified by the given JSONPath from the JSON document stored under the provided key.
// calls the evalJSONDEL() with the arguments passed
// Returns response.RespZero if key is expired or it does not exist
// Returns encoded error response if incorrect number of arguments
// If the JSONPath points to the root of the JSON document, the entire key is deleted from the store.
// Returns an integer reply specified as the number of paths deleted (0 or more)
func evalJSONFORGET(args []string, store *dstore.Store) []byte {
	if len(args) < 1 {
		return diceerrors.NewErrArity("JSON.FORGET")
	}

	return evalJSONDEL(args, store)
}

// evalJSONARRLEN return the length of the JSON array at path in key
// Returns an array of integer replies, an integer for each matching value,
// each is the array's length, or nil, if the matching value is not an array.
// Returns encoded error if the key doesn't exist or key is expired or the matching value is not an array.
// Returns encoded error response if incorrect number of arguments
func evalJSONARRLEN(args []string, store *dstore.Store) []byte {
	if len(args) < 1 {
		return diceerrors.NewErrArity("JSON.ARRLEN")
	}
	key := args[0]

	// Retrieve the object from the database
	obj := store.Get(key)
	if obj == nil {
		return diceerrors.NewErrWithMessage("Path '.' does not exist or not an array")
	}

	errWithMessage := object.AssertTypeAndEncoding(obj.TypeEncoding, object.ObjTypeJSON, object.ObjEncodingJSON)
	if errWithMessage != nil {
		return errWithMessage
	}

	jsonData := obj.Value

	_, err := sonic.Marshal(jsonData)
	if err != nil {
		return diceerrors.NewErrWithMessage("Existing key has wrong Dice type")
	}

	if len(args) == 1 {
		if utils.GetJSONFieldType(jsonData) == utils.ArrayType {
			return clientio.Encode(len(jsonData.([]interface{})), false)
		}
		return diceerrors.NewErrWithMessage("Path '.' does not exist or not an array")
	}

	path := args[1]
	expr, err := jp.ParseString(path)
	if err != nil {
		return diceerrors.NewErrWithMessage("invalid JSONPath")
	}

	results := expr.Get(jsonData)

	arrlenList := make([]interface{}, 0, len(results))
	for _, result := range results {
		switch utils.GetJSONFieldType(result) {
		case utils.ArrayType:
			arrlenList = append(arrlenList, len(result.([]interface{})))
		default:
			arrlenList = append(arrlenList, nil)
		}
	}

	return clientio.Encode(arrlenList, false)
}

func evalJSONARRPOP(args []string, store *dstore.Store) []byte {
	if len(args) < 1 {
		return diceerrors.NewErrArity("json.arrpop")
	}
	key := args[0]

	var path string = defaultRootPath
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
		return diceerrors.NewErrWithMessage("could not perform this operation on a key that doesn't exist")
	}

	errWithMessage := object.AssertTypeAndEncoding(obj.TypeEncoding, object.ObjTypeJSON, object.ObjEncodingJSON)
	if errWithMessage != nil {
		return errWithMessage
	}

	jsonData := obj.Value
	_, err := sonic.Marshal(jsonData)
	if err != nil {
		return diceerrors.NewErrWithMessage("Existing key has wrong Dice type")
	}

	if path == defaultRootPath {
		arr, ok := jsonData.([]any)
		// if value can not be converted to array, it is of another type
		// returns nil in this case similar to redis
		// also, return nil if array is empty
		if !ok || len(arr) == 0 {
			return diceerrors.NewErrWithMessage("Path '$' does not exist or not an array")
		}
		popElem, arr, err := popElementAndUpdateArray(arr, index)
		if err != nil {
			return diceerrors.NewErrWithFormattedMessage("error popping element: %v", err)
		}

		// save the remaining array
		newObj := store.NewObj(arr, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
		store.Put(key, newObj)

		return clientio.Encode(popElem, false)
	}

	// if path is not root then extract value at path
	expr, err := jp.ParseString(path)
	if err != nil {
		return diceerrors.NewErrWithMessage("invalid JSONPath")
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
			popElem := clientio.RespNIL
			popArr = append(popArr, popElem)
			continue
		}

		popElem, arr, err := popElementAndUpdateArray(arr, index)
		if err != nil {
			return diceerrors.NewErrWithFormattedMessage("error popping element: %v", err)
		}

		// update array in place in the json object
		err = expr.Set(jsonData, arr)
		if err != nil {
			return diceerrors.NewErrWithFormattedMessage("error saving updated json: %v", err)
		}

		popArr = append(popArr, popElem)
	}
	return clientio.Encode(popArr, false)
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

// adjustIndex will bound the array between 0 and len(arr) - 1
// It also handles negative indexes
func adjustIndex(idx int, arr []any) int {
	// if index is positive and out of bound, limit it to the last index
	if idx > len(arr) {
		idx = len(arr) - 1
	}

	// if index is negative, change it to equivalent positive index
	if idx < 0 {
		// if index is out of bound then limit it to the first index
		if idx < -len(arr) {
			idx = 0
		} else {
			idx = len(arr) + idx
		}
	}
	return idx
}

// evalJSONOBJLEN return the number of keys in the JSON object at path in key.
// Returns an array of integer replies, an integer for each matching value,
// which is the json objects length, or nil, if the matching value is not a json.
// Returns encoded error if the key doesn't exist or key is expired or the matching value is not an array.
// Returns encoded error response if incorrect number of arguments
func evalJSONOBJLEN(args []string, store *dstore.Store) []byte {
	if len(args) < 1 {
		return diceerrors.NewErrArity("JSON.OBJLEN")
	}

	key := args[0]

	// Retrieve the object from the database
	obj := store.Get(key)
	if obj == nil {
		return clientio.RespNIL
	}

	// check if the object is json
	errWithMessage := object.AssertTypeAndEncoding(obj.TypeEncoding, object.ObjTypeJSON, object.ObjEncodingJSON)
	if errWithMessage != nil {
		return errWithMessage
	}

	// get the value & check for marsheling error
	jsonData := obj.Value
	_, err := sonic.Marshal(jsonData)
	if err != nil {
		return diceerrors.NewErrWithMessage("Existing key has wrong Dice type")
	}
	if len(args) == 1 {
		// check if the value is of json type
		if utils.GetJSONFieldType(jsonData) == utils.ObjectType {
			if castedData, ok := jsonData.(map[string]interface{}); ok {
				return clientio.Encode(len(castedData), false)
			}
			return clientio.RespNIL
		}
		return diceerrors.NewErrWithFormattedMessage(diceerrors.WrongTypeErr)
	}

	path := args[1]

	expr, err := jp.ParseString(path)
	if err != nil {
		return diceerrors.NewErrWithMessage(err.Error())
	}

	// get all values for matching paths
	results := expr.Get(jsonData)

	objectLen := make([]interface{}, 0, len(results))

	for _, result := range results {
		switch utils.GetJSONFieldType(result) {
		case utils.ObjectType:
			if castedResult, ok := result.(map[string]interface{}); ok {
				objectLen = append(objectLen, len(castedResult))
			} else {
				objectLen = append(objectLen, nil)
			}
		default:
			objectLen = append(objectLen, nil)
		}
	}
	return clientio.Encode(objectLen, false)
}

// evalJSONDEL delete a value that the given json path include in.
// Returns response.RespZero if key is expired or it does not exist
// Returns encoded error response if incorrect number of arguments
// Returns an integer reply specified as the number of paths deleted (0 or more)
func evalJSONDEL(args []string, store *dstore.Store) []byte {
	if len(args) < 1 {
		return diceerrors.NewErrArity("JSON.DEL")
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
		return clientio.RespZero
	}

	errWithMessage := object.AssertTypeAndEncoding(obj.TypeEncoding, object.ObjTypeJSON, object.ObjEncodingJSON)
	if errWithMessage != nil {
		return errWithMessage
	}

	jsonData := obj.Value

	_, err := sonic.Marshal(jsonData)
	if err != nil {
		return diceerrors.NewErrWithMessage("Existing key has wrong Dice type")
	}

	if len(args) == 1 || path == defaultRootPath {
		store.Del(key)
		return clientio.RespOne
	}

	expr, err := jp.ParseString(path)
	if err != nil {
		return diceerrors.NewErrWithMessage("invalid JSONPath")
	}
	results := expr.Get(jsonData)
	err = expr.Del(jsonData)
	if err != nil {
		return diceerrors.NewErrWithMessage(err.Error())
	}
	// Create a new object with the updated JSON data
	newObj := store.NewObj(jsonData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
	store.Put(key, newObj)
	return clientio.Encode(len(results), false)
}

// evalJSONCLEAR Clear container values (arrays/objects) and set numeric values to 0,
// Already cleared values are ignored for empty containers and zero numbers
// args must contain at least the key;  (path unused in this implementation)
// Returns encoded error if key is expired or it does not exist
// Returns encoded error response if incorrect number of arguments
// Returns an integer reply specifying the number of matching JSON arrays
// and objects cleared + number of matching JSON numerical values zeroed.
func evalJSONCLEAR(args []string, store *dstore.Store) []byte {
	if len(args) < 1 {
		return diceerrors.NewErrArity("JSON.CLEAR")
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
		return diceerrors.NewErrWithMessage("could not perform this operation on a key that doesn't exist")
	}

	errWithMessage := object.AssertTypeAndEncoding(obj.TypeEncoding, object.ObjTypeJSON, object.ObjEncodingJSON)
	if errWithMessage != nil {
		return errWithMessage
	}

	jsonData := obj.Value

	_, err := sonic.Marshal(jsonData)
	if err != nil {
		return diceerrors.NewErrWithMessage("Existing key has wrong Dice type")
	}

	var countClear uint64 = 0
	if len(args) == 1 || path == defaultRootPath {
		if jsonData != struct{}{} {
			// If path is root and len(args) == 1, return it instantly
			newObj := store.NewObj(struct{}{}, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
			store.Put(key, newObj)
			countClear++
			return clientio.Encode(countClear, false)
		}
	}

	expr, err := jp.ParseString(path)
	if err != nil {
		return diceerrors.NewErrWithMessage("invalid JSONPath")
	}

	_, err = expr.Modify(jsonData, func(element any) (altered any, changed bool) {
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
		return diceerrors.NewErrWithMessage(err.Error())
	}
	// Create a new object with the updated JSON data
	newObj := store.NewObj(jsonData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
	store.Put(key, newObj)
	return clientio.Encode(countClear, false)
}

// evalJSONTYPE retrieves a JSON value type stored at the specified key
// args must contain at least the key;  (path unused in this implementation)
// Returns response.RespNIL if key is expired or it does not exist
// Returns encoded error response if incorrect number of arguments
// The RESP value of the key's value type is encoded and then returned
func evalJSONTYPE(args []string, store *dstore.Store) []byte {
	if len(args) < 1 {
		return diceerrors.NewErrArity("JSON.TYPE")
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
		return clientio.RespNIL
	}

	errWithMessage := object.AssertTypeAndEncoding(obj.TypeEncoding, object.ObjTypeJSON, object.ObjEncodingJSON)
	if errWithMessage != nil {
		return errWithMessage
	}

	jsonData := obj.Value

	if path == defaultRootPath {
		_, err := sonic.Marshal(jsonData)
		if err != nil {
			return diceerrors.NewErrWithMessage("could not serialize result")
		}
		// If path is root and len(args) == 1, return "object" instantly
		if len(args) == 1 {
			return clientio.Encode(utils.ObjectType, false)
		}
	}

	// Parse the JSONPath expression
	expr, err := jp.ParseString(path)
	if err != nil {
		return diceerrors.NewErrWithMessage("invalid JSONPath")
	}

	results := expr.Get(jsonData)
	if len(results) == 0 {
		return clientio.RespEmptyArray
	}

	typeList := make([]string, 0, len(results))
	for _, result := range results {
		jsonType := utils.GetJSONFieldType(result)
		typeList = append(typeList, jsonType)
	}
	return clientio.Encode(typeList, false)
}

// evalJSONGET retrieves a JSON value stored at the specified key
// args must contain at least the key;  (path unused in this implementation)
// Returns response.RespNIL if key is expired or it does not exist
// Returns encoded error response if incorrect number of arguments
// The RESP value of the key is encoded and then returned
func evalJSONGET(args []string, store *dstore.Store) []byte {
	if len(args) < 1 {
		return diceerrors.NewErrArity("JSON.GET")
	}

	key := args[0]
	// Default path is root if not specified
	path := defaultRootPath
	if len(args) > 1 {
		path = args[1]
	}
	result, err := jsonGETHelper(store, path, key)
	if err != nil {
		return err
	}
	return clientio.Encode(result, false)
}

// helper function used by evalJSONGET and evalJSONMGET to prepare the results
func jsonGETHelper(store *dstore.Store, path, key string) (result interface{}, err2 []byte) {
	// Retrieve the object from the database
	obj := store.Get(key)
	if obj == nil {
		return result, nil
	}

	// Check if the object is of JSON type
	errWithMessage := object.AssertTypeAndEncoding(obj.TypeEncoding, object.ObjTypeJSON, object.ObjEncodingJSON)
	if errWithMessage != nil {
		return result, errWithMessage
	}

	jsonData := obj.Value

	// If path is root, return the entire JSON
	if path == defaultRootPath {
		resultBytes, err := sonic.Marshal(jsonData)
		if err != nil {
			return result, diceerrors.NewErrWithMessage("could not serialize result")
		}
		return string(resultBytes), nil
	}

	// Parse the JSONPath expression
	expr, err := jp.ParseString(path)
	if err != nil {
		return result, diceerrors.NewErrWithMessage("invalid JSONPath")
	}

	// Execute the JSONPath query
	results := expr.Get(jsonData)
	if len(results) == 0 {
		return result, nil
	}

	// Serialize the result
	var resultBytes []byte
	if len(results) == 1 {
		resultBytes, err = sonic.Marshal(results[0])
	} else {
		resultBytes, err = sonic.Marshal(results)
	}
	if err != nil {
		return nil, diceerrors.NewErrWithMessage("could not serialize result")
	}
	return string(resultBytes), nil
}

// evalJSONMGET retrieves a JSON value stored for the multiple key
// args must contain at least the key and a path;
// Returns encoded error response if incorrect number of arguments
// The RESP value of the key is encoded and then returned
func evalJSONMGET(args []string, store *dstore.Store) []byte {
	if len(args) < 2 {
		return diceerrors.NewErrArity("JSON.MGET")
	}

	var results []interface{}

	// Default path is root if not specified
	argsLen := len(args)
	path := args[argsLen-1]

	for i := 0; i < (argsLen - 1); i++ {
		key := args[i]
		result, _ := jsonGETHelper(store, path, key)
		results = append(results, result)
	}

	var interfaceObj interface{} = results
	return clientio.Encode(interfaceObj, false)
}

// evalJSONTOGGLE toggles a boolean value stored at the specified key and path.
// args must contain at least the key and path (where the boolean is located).
// If the key does not exist or is expired, it returns response.RespNIL.
// If the field at the specified path is not a boolean, it returns an encoded error response.
// If the boolean is `true`, it toggles to `false` (returns :0), and if `false`, it toggles to `true` (returns :1).
// Returns an encoded error response if the incorrect number of arguments is provided.
func evalJSONTOGGLE(args []string, store *dstore.Store) []byte {
	if len(args) < 2 {
		return diceerrors.NewErrArity("JSON.TOGGLE")
	}
	key := args[0]
	path := args[1]

	obj := store.Get(key)
	if obj == nil {
		return diceerrors.NewErrWithFormattedMessage("-ERR could not perform this operation on a key that doesn't exist")
	}

	errWithMessage := object.AssertTypeAndEncoding(obj.TypeEncoding, object.ObjTypeJSON, object.ObjEncodingJSON)
	if errWithMessage != nil {
		return errWithMessage
	}

	jsonData := obj.Value
	expr, err := jp.ParseString(path)
	if err != nil {
		return diceerrors.NewErrWithMessage("invalid JSONPath")
	}

	toggleResults := []interface{}{}
	modified := false

	_, err = expr.Modify(jsonData, func(value interface{}) (interface{}, bool) {
		boolValue, ok := value.(bool)
		if !ok {
			toggleResults = append(toggleResults, nil)
			return value, false
		}
		newValue := !boolValue
		toggleResults = append(toggleResults, boolToInt(newValue))
		modified = true
		return newValue, true
	})

	if err != nil {
		return diceerrors.NewErrWithMessage("failed to toggle values")
	}

	if modified {
		obj.Value = jsonData
	}

	toggleResults = ReverseSlice(toggleResults)
	return clientio.Encode(toggleResults, false)
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// ReverseSlice takes a slice of any type and returns a new slice with the elements reversed.
func ReverseSlice[T any](slice []T) []T {
	reversed := make([]T, len(slice))
	for i, v := range slice {
		reversed[len(slice)-1-i] = v
	}
	return reversed
}

// evalJSONSET stores a JSON value at the specified key
// args must contain at least the key, path (unused in this implementation), and JSON string
// Returns encoded error response if incorrect number of arguments
// Returns encoded error if the JSON string is invalid
// Returns response.RespOK if the JSON value is successfully stored
func evalJSONSET(args []string, store *dstore.Store) []byte {
	// Check if there are enough arguments
	if len(args) < 3 {
		return diceerrors.NewErrArity("JSON.SET")
	}

	key := args[0]
	path := args[1]
	jsonStr := args[2]
	for i := 3; i < len(args); i++ {
		switch args[i] {
		case NX, Nx:
			if i != len(args)-1 {
				return diceerrors.NewErrWithMessage(diceerrors.SyntaxErr)
			}
			obj := store.Get(key)
			if obj != nil {
				return clientio.RespNIL
			}
		case XX, Xx:
			if i != len(args)-1 {
				return diceerrors.NewErrWithMessage(diceerrors.SyntaxErr)
			}
			obj := store.Get(key)
			if obj == nil {
				return clientio.RespNIL
			}

		default:
			return diceerrors.NewErrWithMessage(diceerrors.SyntaxErr)
		}
	}

	// Parse the JSON string
	var jsonValue interface{}
	if err := sonic.UnmarshalString(jsonStr, &jsonValue); err != nil {
		return diceerrors.NewErrWithFormattedMessage("invalid JSON: %v", err.Error())
	}

	// Retrieve existing object or create new one
	obj := store.Get(key)
	var rootData interface{}

	if obj == nil {
		// If the key doesn't exist, create a new object
		if path != defaultRootPath {
			rootData = make(map[string]interface{})
		} else {
			rootData = jsonValue
		}
	} else {
		// If the key exists, check if it's a JSON object
		err := object.AssertType(obj.TypeEncoding, object.ObjTypeJSON)
		if err != nil {
			return clientio.Encode(err, false)
		}
		err = object.AssertEncoding(obj.TypeEncoding, object.ObjEncodingJSON)
		if err != nil {
			return clientio.Encode(err, false)
		}
		rootData = obj.Value
	}

	// If path is not root, use JSONPath to set the value
	if path != defaultRootPath {
		expr, err := jp.ParseString(path)
		if err != nil {
			return diceerrors.NewErrWithMessage("invalid JSONPath")
		}

		err = expr.Set(rootData, jsonValue)
		if err != nil {
			return diceerrors.NewErrWithMessage("failed to set value")
		}
	} else {
		// If path is root, replace the entire JSON
		rootData = jsonValue
	}

	// Create a new object with the updated JSON data
	newObj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
	store.Put(key, newObj)
	return clientio.RespOK
}

// evalJSONARRAPPEND appends the value(s) provided in the args to the given array path
// in the JSON object saved at key in arguments.
// Args must contain atleast a key, path and value.
// If the key does not exist or is expired, it returns response.RespNIL.
// If the object at given path is not an array, it returns response.RespNIL.
// Returns the new length of the array at path.
func evalJSONARRAPPEND(args []string, store *dstore.Store) []byte {
	if len(args) < 3 {
		return diceerrors.NewErrArity("JSON.ARRAPPEND")
	}

	key := args[0]
	path := args[1]
	values := args[2:]

	obj := store.Get(key)
	if obj == nil {
		return diceerrors.NewErrWithMessage("ERR key does not exist")
	}

	errWithMessage := object.AssertTypeAndEncoding(obj.TypeEncoding, object.ObjTypeJSON, object.ObjEncodingJSON)
	if errWithMessage != nil {
		return errWithMessage
	}

	jsonData := obj.Value

	expr, err := jp.ParseString(path)
	if err != nil {
		return diceerrors.NewErrWithMessage(fmt.Sprintf("ERR Path '%s' does not exist or not an array", path))
	}

	// Parse the input values as JSON
	parsedValues := make([]interface{}, len(values))
	for i, v := range values {
		var parsedValue interface{}
		err := sonic.UnmarshalString(v, &parsedValue)
		if err != nil {
			return diceerrors.NewErrWithMessage(err.Error())
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
			resultsArray = append(resultsArray, nil)
			return data, false
		}

		// Append the parsed values to the array
		arr = append(arr, parsedValues...)

		resultsArray = append(resultsArray, int64(len(arr)))
		modified = true
		return arr, modified
	})

	if modifyErr != nil {
		return diceerrors.NewErrWithMessage(fmt.Sprintf("ERR failed to modify JSON data: %v", modifyErr))
	}

	if !modified {
		// If no modification was made, it means the path did not exist or was not an array
		return clientio.Encode([]interface{}{nil}, false)
	}

	jsonData = newData
	obj.Value = jsonData

	return clientio.Encode(resultsArray, false)
}

// evalJSONINGEST stores a value at a dynamically generated key
// The key is created using a provided key prefix combined with a unique identifier
// args must contains key_prefix and path and json value
// It will call to evalJSONSET internally.
// Returns encoded error response if incorrect number of arguments
// Returns encoded error if the JSON string is invalid
// Returns unique identifier if the JSON value is successfully stored
func evalJSONINGEST(args []string, store *dstore.Store) []byte {
	if len(args) < 3 {
		return diceerrors.NewErrArity("JSON.INGEST")
	}

	keyPrefix := args[0]

	uniqueID := xid.New()
	uniqueKey := keyPrefix + uniqueID.String()

	var setArgs []string
	setArgs = append(setArgs, uniqueKey)
	setArgs = append(setArgs, args[1:]...)

	result := evalJSONSET(setArgs, store)
	if bytes.Equal(result, clientio.RespOK) {
		return clientio.Encode(uniqueID.String(), true)
	}
	return result
}

// evalTTL returns Time-to-Live in secs for the queried key in args
// The key should be the only param in args else returns with an error
// Returns	RESP encoded time (in secs) remaining for the key to expire
//
//	RESP encoded -2 stating key doesn't exist or key is expired
//	RESP encoded -1 in case no expiration is set on the key
func evalTTL(args []string, store *dstore.Store) []byte {
	if len(args) != 1 {
		return diceerrors.NewErrArity("TTL")
	}

	var key string = args[0]

	obj := store.Get(key)

	// if key does not exist, return RESP encoded -2 denoting key does not exist
	if obj == nil {
		return clientio.RespMinusTwo
	}

	// if object exist, but no expiration is set on it then send -1
	exp, isExpirySet := dstore.GetExpiry(obj, store)
	if !isExpirySet {
		return clientio.RespMinusOne
	}

	// compute the time remaining for the key to expire and
	// return the RESP encoded form of it
	durationMs := exp - uint64(utils.GetCurrentTime().UnixMilli())

	return clientio.Encode(int64(durationMs/1000), false)
}

// evalDEL deletes all the specified keys in args list
// returns the count of total deleted keys after encoding
func evalDEL(args []string, store *dstore.Store) []byte {
	var countDeleted int = 0

	for _, key := range args {
		if ok := store.Del(key); ok {
			countDeleted++
		}
	}

	return clientio.Encode(countDeleted, false)
}

// evalEXPIRE sets a expiry time(in secs) on the specified key in args
// args should contain 2 values, key and the expiry time to be set for the key
// The expiry time should be in integer format; if not, it returns encoded error response
// Returns response.RespOne if expiry was set on the key successfully.
// Once the time is lapsed, the key will be deleted automatically
func evalEXPIRE(args []string, store *dstore.Store) []byte {
	if len(args) <= 1 {
		return diceerrors.NewErrArity("EXPIRE")
	}

	var key string = args[0]
	exDurationSec, err := strconv.ParseInt(args[1], 10, 64)
	if exDurationSec < 0 || exDurationSec > maxExDuration {
		return diceerrors.NewErrExpireTime("EXPIRE")
	}

	if err != nil {
		return diceerrors.NewErrExpireTime("EXPIRE")
	}

	obj := store.Get(key)

	// 0 if the timeout was not set. e.g. key doesn't exist, or operation skipped due to the provided arguments
	if obj == nil {
		return clientio.RespZero
	}
	isExpirySet, err2 := evaluateAndSetExpiry(args[2:], utils.AddSecondsToUnixEpoch(exDurationSec), key, store)

	if isExpirySet {
		return clientio.RespOne
	} else if err2 != nil {
		return err2
	}
	return clientio.RespZero
}

// evalEXPIRETIME returns the absolute Unix timestamp (since January 1, 1970) in seconds at which the given key will expire
// args should contain only 1 value, the key
// Returns expiration Unix timestamp in seconds.
// Returns -1 if the key exists but has no associated expiration time.
// Returns -2 if the key does not exist.
func evalEXPIRETIME(args []string, store *dstore.Store) []byte {
	if len(args) != 1 {
		return diceerrors.NewErrArity("EXPIRETIME")
	}

	var key string = args[0]

	obj := store.Get(key)

	// -2 if key doesn't exist
	if obj == nil {
		return clientio.RespMinusTwo
	}

	exTimeMili, ok := dstore.GetExpiry(obj, store)
	// -1 if key doesn't have expiration time set
	if !ok {
		return clientio.RespMinusOne
	}

	return clientio.Encode(int(exTimeMili/1000), false)
}

// evalEXPIREAT sets a expiry time(in unix-time-seconds) on the specified key in args
// args should contain 2 values, key and the expiry time to be set for the key
// The expiry time should be in integer format; if not, it returns encoded error response
// Returns response.RespOne if expiry was set on the key successfully.
// Once the time is lapsed, the key will be deleted automatically
func evalEXPIREAT(args []string, store *dstore.Store) []byte {
	if len(args) <= 1 {
		return clientio.Encode(errors.New("ERR wrong number of arguments for 'expireat' command"), false)
	}

	var key string = args[0]
	exUnixTimeSec, err := strconv.ParseInt(args[1], 10, 64)
	if exUnixTimeSec < 0 || exUnixTimeSec > maxExDuration {
		return diceerrors.NewErrExpireTime("EXPIREAT")
	}

	if err != nil {
		return clientio.Encode(errors.New("ERR value is not an integer or out of range"), false)
	}

	isExpirySet, err2 := evaluateAndSetExpiry(args[2:], exUnixTimeSec, key, store)
	if isExpirySet {
		return clientio.RespOne
	} else if err2 != nil {
		return err2
	}
	return clientio.RespZero
}

// NX: Set the expiration only if the key does not already have an expiration time.
// XX: Set the expiration only if the key already has an expiration time.
// GT: Set the expiration only if the new expiration time is greater than the current one.
// LT: Set the expiration only if the new expiration time is less than the current one.
// Returns Boolean True and error nil if expiry was set on the key successfully.
// Returns Boolean False and error nil if conditions didn't met.
// Returns Boolean False and error not-nil if invalid combination of subCommands or if subCommand is invalid
func evaluateAndSetExpiry(subCommands []string, newExpiry int64, key string,
	store *dstore.Store) (shouldSetExpiry bool, err []byte) {
	var newExpInMilli = newExpiry * 1000
	var prevExpiry *uint64 = nil
	var nxCmd, xxCmd, gtCmd, ltCmd bool

	obj := store.Get(key)
	//  key doesn't exist
	if obj == nil {
		return false, nil
	}
	shouldSetExpiry = true
	// if no condition exists
	if len(subCommands) == 0 {
		store.SetUnixTimeExpiry(obj, newExpiry)
		return shouldSetExpiry, nil
	}

	expireTime, ok := dstore.GetExpiry(obj, store)
	if ok {
		prevExpiry = &expireTime
	}

	for i := range subCommands {
		subCommand := strings.ToUpper(subCommands[i])

		switch subCommand {
		case NX:
			nxCmd = true
			if prevExpiry != nil {
				shouldSetExpiry = false
			}
		case XX:
			xxCmd = true
			if prevExpiry == nil {
				shouldSetExpiry = false
			}
		case GT:
			gtCmd = true
			if prevExpiry == nil || *prevExpiry > uint64(newExpInMilli) {
				shouldSetExpiry = false
			}
		case LT:
			ltCmd = true
			if prevExpiry != nil && *prevExpiry < uint64(newExpInMilli) {
				shouldSetExpiry = false
			}
		default:
			return false, diceerrors.NewErrWithMessage("Unsupported option " + subCommands[i])
		}
	}

	if !nxCmd && gtCmd && ltCmd {
		return false, diceerrors.NewErrWithMessage("GT and LT options at the same time are not compatible")
	}

	if nxCmd && (xxCmd || gtCmd || ltCmd) {
		return false, diceerrors.NewErrWithMessage("NX and XX," +
			" GT or LT options at the same time are not compatible")
	}

	if shouldSetExpiry {
		store.SetUnixTimeExpiry(obj, newExpiry)
	}
	return shouldSetExpiry, nil
}

func evalHELLO(args []string, store *dstore.Store) []byte {
	if len(args) > 1 {
		return diceerrors.NewErrArity("HELLO")
	}

	var resp []interface{}
	resp = append(resp,
		"proto", 2,
		"id", serverID,
		"mode", "standalone",
		"role", "master",
		"modules", []interface{}{})

	return clientio.Encode(resp, false)
}

// evalINCR increments the value of the specified key in args by 1,
// if the key exists and the value is integer format.
// The key should be the only param in args.
// If the key does not exist, new key is created with value 0,
// the value of the new key is then incremented.
// The value for the queried key should be of integer format,
// if not evalINCR returns encoded error response.
// evalINCR returns the incremented value for the key if there are no errors.
func evalINCR(args []string, store *dstore.Store) []byte {
	if len(args) != 1 {
		return diceerrors.NewErrArity("INCR")
	}
	return incrDecrCmd(args, 1, store)
}

// evalDECR decrements the value of the specified key in args by 1,
// if the key exists and the value is integer format.
// The key should be the only param in args.
// If the key does not exist, new key is created with value 0,
// the value of the new key is then decremented.
// The value for the queried key should be of integer format,
// if not evalDECR returns encoded error response.
// evalDECR returns the decremented value for the key if there are no errors.
func evalDECR(args []string, store *dstore.Store) []byte {
	if len(args) != 1 {
		return diceerrors.NewErrArity("DECR")
	}
	return incrDecrCmd(args, -1, store)
}

// evalDECRBY decrements the value of the specified key in args by the specified decrement,
// if the key exists and the value is integer format.
// The key should be the first parameter in args, and the decrement should be the second parameter.
// If the key does not exist, new key is created with value 0,
// the value of the new key is then decremented by specified decrement.
// The value for the queried key should be of integer format,
// if not evalDECRBY returns an encoded error response.
// evalDECRBY returns the decremented value for the key after applying the specified decrement if there are no errors.
func evalDECRBY(args []string, store *dstore.Store) []byte {
	if len(args) != 2 {
		return diceerrors.NewErrArity("DECRBY")
	}
	decrementAmount, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return diceerrors.NewErrWithMessage(diceerrors.IntOrOutOfRangeErr)
	}
	return incrDecrCmd(args, -decrementAmount, store)
}

func incrDecrCmd(args []string, incr int64, store *dstore.Store) []byte {
	key := args[0]
	obj := store.Get(key)
	if obj == nil {
		obj = store.NewObj(int64(0), -1, object.ObjTypeInt, object.ObjEncodingInt)
		store.Put(key, obj)
	}

	// If the object exists, check if it is a set object.
	if err := object.AssertType(obj.TypeEncoding, object.ObjTypeSet); err == nil {
		return diceerrors.NewErrWithFormattedMessage(diceerrors.WrongTypeErr)
	}

	if err := object.AssertType(obj.TypeEncoding, object.ObjTypeInt); err != nil {
		return diceerrors.NewErrWithMessage(err.Error())
	}

	if err := object.AssertEncoding(obj.TypeEncoding, object.ObjEncodingInt); err != nil {
		return diceerrors.NewErrWithMessage(err.Error())
	}

	i, _ := obj.Value.(int64)
	// check overflow
	if (incr < 0 && i < 0 && incr < (math.MinInt64-i)) ||
		(incr > 0 && i > 0 && incr > (math.MaxInt64-i)) {
		return diceerrors.NewErrWithMessage(diceerrors.ValOutOfRangeErr)
	}

	i += incr
	obj.Value = i

	return clientio.Encode(i, false)
}

// evalINFO creates a buffer with the info of total keys per db
// Returns the encoded buffer as response
func evalINFO(args []string, store *dstore.Store) []byte {
	var info []byte
	buf := bytes.NewBuffer(info)
	buf.WriteString("# Keyspace\r\n")
	for i := range dstore.KeyspaceStat {
		fmt.Fprintf(buf, "db%d:keys=%d,expires=0,avg_ttl=0\r\n", i, dstore.KeyspaceStat[i]["keys"])
	}
	return clientio.Encode(buf.String(), false)
}

// TODO: Placeholder to support monitoring
func evalCLIENT(args []string, store *dstore.Store) []byte {
	return clientio.RespOK
}

// TODO: Placeholder to support monitoring
func evalLATENCY(args []string, store *dstore.Store) []byte {
	return clientio.Encode([]string{}, false)
}

// evalLRU deletes all the keys from the LRU
// returns encoded RESP OK
func evalLRU(args []string, store *dstore.Store) []byte {
	dstore.EvictAllkeysLRUOrLFU(store)
	return clientio.RespOK
}

// evalSLEEP sets db to sleep for the specified number of seconds.
// The sleep time should be the only param in args.
// Returns error response if the time param in args is not of integer format.
// evalSLEEP returns response.RespOK after sleeping for mentioned seconds
func evalSLEEP(args []string, store *dstore.Store) []byte {
	if len(args) != 1 {
		return diceerrors.NewErrArity("SLEEP")
	}

	durationSec, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return diceerrors.NewErrWithMessage(diceerrors.IntOrOutOfRangeErr)
	}
	time.Sleep(time.Duration(durationSec) * time.Second)
	return clientio.RespOK
}

// evalMULTI marks the start of the transaction for the client.
// All subsequent commands fired will be queued for atomic execution.
// The commands will not be executed until EXEC is triggered.
// Once EXEC is triggered it executes all the commands in queue,
// and closes the MULTI transaction.
func evalMULTI(args []string, store *dstore.Store) []byte {
	return clientio.RespOK
}

// EvalQWATCH adds the specified key to the watch list for the caller client.
// Every time a key in the watch list is modified, the client will be sent a response
// containing the new value of the key along with the operation that was performed on it.
// Contains only one argument, the query to be watched.
func EvalQWATCH(args []string, clientFd int, store *dstore.Store) []byte {
	if len(args) != 1 {
		return diceerrors.NewErrArity("QWATCH")
	}

	// Parse and get the selection from the query.
	query, e := sql.ParseQuery( /*sql=*/ args[0])

	if e != nil {
		return clientio.Encode(e, false)
	}

	// use an unbuffered channel to ensure that we only proceed to query execution once the query watcher has built the cache
	cacheChannel := make(chan *[]struct {
		Key   string
		Value *object.Obj
	})
	querywatcher.WatchSubscriptionChan <- querywatcher.WatchSubscription{
		Subscribe: true,
		Query:     query,
		ClientFD:  clientFd,
		CacheChan: cacheChannel,
	}

	store.CacheKeysForQuery(query.Where, cacheChannel)

	// Return the result of the query.
	responseChan := make(chan querywatcher.AdhocQueryResult)
	querywatcher.AdhocQueryChan <- querywatcher.AdhocQuery{
		Query:        query,
		ResponseChan: responseChan,
	}

	queryResult := <-responseChan
	if queryResult.Err != nil {
		return clientio.Encode(queryResult.Err, false)
	}

	// TODO: We should return the list of all queries being watched by the client.
	return clientio.Encode(clientio.CreatePushResponse(&query, queryResult.Result), false)
}

// EvalQUNWATCH removes the specified key from the watch list for the caller client.
func EvalQUNWATCH(args []string, clientFd int) []byte {
	if len(args) != 1 {
		return diceerrors.NewErrArity("QUNWATCH")
	}
	query, e := sql.ParseQuery( /*sql=*/ args[0])
	if e != nil {
		return clientio.Encode(e, false)
	}

	querywatcher.WatchSubscriptionChan <- querywatcher.WatchSubscription{
		Subscribe: false,
		Query:     query,
		ClientFD:  clientFd,
	}

	return clientio.RespOK
}

// SETBIT key offset value
func evalSETBIT(args []string, store *dstore.Store) []byte {
	var err error

	if len(args) != 3 {
		return diceerrors.NewErrArity("SETBIT")
	}

	key := args[0]
	offset, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return diceerrors.NewErrWithMessage("bit offset is not an integer or out of range")
	}

	value, err := strconv.ParseBool(args[2])
	if err != nil {
		return diceerrors.NewErrWithMessage("bit is not an integer or out of range")
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
				return diceerrors.NewErrWithMessage(diceerrors.WrongTypeErr)
			}
		default:
			return diceerrors.NewErrWithMessage(diceerrors.WrongTypeErr)
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
			return diceerrors.NewErrWithMessage(diceerrors.WrongTypeErr)
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
			return clientio.Encode(int(1), true)
		}
		return clientio.Encode(int(0), true)
	}
	return diceerrors.NewErrWithMessage(diceerrors.WrongTypeErr)
}

// GETBIT key offset
func evalGETBIT(args []string, store *dstore.Store) []byte {
	var err error

	if len(args) != 2 {
		return diceerrors.NewErrArity("GETBIT")
	}

	key := args[0]
	offset, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return diceerrors.NewErrWithMessage("bit offset is not an integer or out of range")
	}

	obj := store.Get(key)
	if obj == nil {
		return clientio.Encode(0, true)
	}

	requiredByteArraySize := offset>>3 + 1
	switch oType, _ := object.ExtractTypeEncoding(obj); oType {
	case object.ObjTypeSet:
		return diceerrors.NewErrWithFormattedMessage(diceerrors.WrongTypeErr)
	case object.ObjTypeByteArray:
		byteArray := obj.Value.(*ByteArray)
		byteArrayLength := byteArray.Length

		// check whether offset, length exists or not
		if requiredByteArraySize > byteArrayLength {
			return clientio.Encode(0, true)
		}
		value := byteArray.GetBit(int(offset))
		if value {
			return clientio.Encode(1, true)
		}
		return clientio.Encode(0, true)
	case object.ObjTypeString, object.ObjTypeInt:
		byteArray, err := NewByteArrayFromObj(obj)
		if err != nil {
			return diceerrors.NewErrWithMessage(diceerrors.WrongTypeErr)
		}
		if requiredByteArraySize > byteArray.Length {
			return clientio.Encode(0, true)
		}
		value := byteArray.GetBit(int(offset))
		if value {
			return clientio.Encode(1, true)
		}
		return clientio.Encode(0, true)
	default:
		return clientio.Encode(0, true)
	}
}

func evalBITCOUNT(args []string, store *dstore.Store) []byte {
	var err error

	// if more than 4 arguments are provided, return error
	if len(args) > 4 {
		return diceerrors.NewErrWithMessage(diceerrors.SyntaxErr)
	}

	// fetching value of the key
	var key string = args[0]
	obj := store.Get(key)
	if obj == nil {
		return clientio.Encode(0, false)
	}

	// Check for the type of the object
	if object.AssertType(obj.TypeEncoding, object.ObjTypeSet) == nil {
		return diceerrors.NewErrWithFormattedMessage(diceerrors.WrongTypeErr)
	}

	valueInterface := obj.Value
	value := []byte{}
	valueLength := int64(0)

	if object.AssertType(obj.TypeEncoding, object.ObjTypeByteArray) == nil {
		byteArray := obj.Value.(*ByteArray)
		byteArrayObject := *byteArray
		value = byteArrayObject.data
		valueLength = byteArray.Length
	}

	if object.AssertType(obj.TypeEncoding, object.ObjTypeString) == nil {
		value = []byte(valueInterface.(string))
		valueLength = int64(len(value))
	}

	if object.AssertType(obj.TypeEncoding, object.ObjTypeInt) == nil {
		value = []byte(strconv.FormatInt(valueInterface.(int64), 10))
		valueLength = int64(len(value))
	}

	// defining constants of the function
	start := int64(0)
	end := valueLength - 1
	unit := BYTE

	// checking which arguments are present and according validating arguments
	if len(args) > 1 {
		start, err = strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			return diceerrors.NewErrWithMessage(diceerrors.IntOrOutOfRangeErr)
		}
		// Adjust start index if it is negative
		if start < 0 {
			start += valueLength
		}
	}
	if len(args) > 2 {
		end, err = strconv.ParseInt(args[2], 10, 64)
		if err != nil {
			return diceerrors.NewErrWithMessage(diceerrors.IntOrOutOfRangeErr)
		}

		// Adjust end index if it is negative
		if end < 0 {
			end += valueLength
		}
	}
	if len(args) > 3 {
		unit = strings.ToUpper(args[3])
		if unit != BYTE && unit != BIT {
			return diceerrors.NewErrWithMessage(diceerrors.SyntaxErr)
		}
	}
	if start > end {
		return clientio.Encode(0, true)
	}
	if start > valueLength && unit == BYTE {
		return clientio.Encode(0, true)
	}
	if end > valueLength && unit == BYTE {
		end = valueLength - 1
	}

	bitCount := 0
	if unit == BYTE {
		for i := start; i <= end; i++ {
			bitCount += int(popcount(value[i]))
		}
		return clientio.Encode(bitCount, true)
	}
	startBitRange := start / 8
	endBitRange := min(end/8, valueLength-1)
	for i := startBitRange; i <= endBitRange; i++ {
		if i == startBitRange {
			considerBits := start % 8
			for j := 8 - considerBits - 1; j >= 0; j-- {
				bitCount += int(popcount(byte(int(value[i]) & (1 << j))))
			}
		} else if i == endBitRange {
			considerBits := end % 8
			for j := considerBits; j >= 0; j-- {
				bitCount += int(popcount(byte(int(value[i]) & (1 << (8 - j - 1)))))
			}
		} else {
			bitCount += int(popcount(value[i]))
		}
	}
	return clientio.Encode(bitCount, true)
}

// BITOP <AND | OR | XOR | NOT> destkey key [key ...]
func evalBITOP(args []string, store *dstore.Store) []byte {
	operation, destKey := args[0], args[1]
	operation = strings.ToUpper(operation)

	// get all the keys
	keys := args[2:]

	// validation of commands
	// if operation is not from enums, then error out
	if !(operation == AND || operation == OR || operation == XOR || operation == NOT) {
		return diceerrors.NewErrWithMessage(diceerrors.SyntaxErr)
	}

	if operation == NOT {
		if len(keys) != 1 {
			return diceerrors.NewErrWithMessage("BITOP NOT must be called with a single source key.")
		}
		key := keys[0]
		obj := store.Get(key)
		if obj == nil {
			return clientio.Encode(0, true)
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
			return clientio.Encode(len(value), true)
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
			return clientio.Encode(len(value), true)
		default:
			return diceerrors.NewErrWithFormattedMessage(diceerrors.WrongTypeErr)
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
				return diceerrors.NewErrWithMessage("value is not a valid byte array")
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

	return clientio.Encode(len(result), true)
}

// evalCommand evaluates COMMAND <subcommand> command based on subcommand
// COUNT: return total count of commands in Dice.
func evalCommand(args []string, store *dstore.Store) []byte {
	if len(args) == 0 {
		return evalCommandDefault()
	}
	subcommand := strings.ToUpper(args[0])
	switch subcommand {
	case Count:
		return evalCommandCount()
	case GetKeys:
		return evalCommandGetKeys(args[1:])
	case List:
		return evalCommandList()
	case Help:
		return evalCommandHelp()
	default:
		return diceerrors.NewErrWithFormattedMessage("unknown subcommand '%s'. Try COMMAND HELP.", subcommand)
	}
}

// evalCommandHelp prints help message
func evalCommandHelp() []byte {
	format := "COMMAND <subcommand> [<arg> [value] [opt] ...]. Subcommands are:"
	noTitle := "(no subcommand)"
	noMessage := "    Return details about all Dice commands."
	countTitle := "COUNT"
	countMessage := "    Return the total number of commands in this Dice server."
	listTitle := "LIST"
	listMessage := "     Return a list of all commands in this Dice server."
	getKeysTitle := "GETKEYS <full-command>"
	getKeysMessage := "     Return the keys from a full Dice command."
	helpTitle := "HELP"
	helpMessage := "     Print this help."
	message := []string{
		format,
		noTitle,
		noMessage,
		countTitle,
		countMessage,
		listTitle,
		listMessage,
		getKeysTitle,
		getKeysMessage,
		helpTitle,
		helpMessage,
	}
	return clientio.Encode(message, false)
}

func evalCommandDefault() []byte {
	cmds := convertDiceCmdsMapToSlice()
	return clientio.Encode(cmds, false)
}

func evalCommandList() []byte {
	cmds := make([]string, 0, diceCommandsCount)
	for k := range DiceCmds {
		cmds = append(cmds, k)
	}
	return clientio.Encode(cmds, false)
}

// evalKeys returns the list of keys that match the pattern
// The pattern should be the only param in args
func evalKeys(args []string, store *dstore.Store) []byte {
	if len(args) != 1 {
		return diceerrors.NewErrArity("KEYS")
	}

	pattern := args[0]
	keys, err := store.Keys(pattern)
	if err != nil {
		return clientio.Encode(err, false)
	}

	return clientio.Encode(keys, false)
}

// evalCommandCount returns an number of commands supported by DiceDB
func evalCommandCount() []byte {
	return clientio.Encode(diceCommandsCount, false)
}

// evalCommandGetKeys helps identify which arguments in a redis command
// are interpreted as keys.
// This is useful in analying long commands / scripts
func evalCommandGetKeys(args []string) []byte {
	if len(args) == 0 {
		return diceerrors.NewErrArity("COMMAND|GETKEYS")
	}
	diceCmd, ok := DiceCmds[strings.ToUpper(args[0])]
	if !ok {
		return diceerrors.NewErrWithMessage("invalid command specified")
	}

	keySpecs := diceCmd.KeySpecs
	if keySpecs.BeginIndex == 0 {
		return diceerrors.NewErrWithMessage("the command has no key arguments")
	}

	arity := diceCmd.Arity
	if (arity < 0 && len(args) < -arity) ||
		(arity >= 0 && len(args) != arity) {
		return diceerrors.NewErrWithMessage("invalid number of arguments specified for command")
	}
	keys := make([]string, 0)
	step := max(keySpecs.Step, 1)
	lastIdx := keySpecs.BeginIndex
	if keySpecs.LastKey != 0 {
		lastIdx = len(args) + keySpecs.LastKey
	}
	for i := keySpecs.BeginIndex; i <= lastIdx; i += step {
		keys = append(keys, args[i])
	}
	return clientio.Encode(keys, false)
}

func evalRename(args []string, store *dstore.Store) []byte {
	if len(args) != 2 {
		return diceerrors.NewErrArity("RENAME")
	}
	sourceKey := args[0]
	destKey := args[1]

	// if Source key does not exist, return RESP encoded nil
	sourceObj := store.Get(sourceKey)
	if sourceObj == nil {
		return diceerrors.NewErrWithMessage(diceerrors.NoKeyErr)
	}

	// if Source and Destination Keys are same return RESP encoded ok
	if sourceKey == destKey {
		return clientio.RespOK
	}

	if ok := store.Rename(sourceKey, destKey); ok {
		return clientio.RespOK
	}
	return clientio.RespNIL
}

// The MGET command returns an array of RESP values corresponding to the provided keys.
// For each key, if the key is expired or does not exist, the response will be response.RespNIL;
// otherwise, the response will be the RESP value of the key.
// MGET is atomic, it retrieves all values at once
func evalMGET(args []string, store *dstore.Store) []byte {
	if len(args) < 1 {
		return diceerrors.NewErrArity("MGET")
	}
	values := store.GetAll(args)
	resp := make([]interface{}, len(args))
	for i, obj := range values {
		if obj == nil {
			resp[i] = clientio.RespNIL
		} else {
			resp[i] = obj.Value
		}
	}
	return clientio.Encode(resp, false)
}

func evalEXISTS(args []string, store *dstore.Store) []byte {
	if len(args) == 0 {
		return diceerrors.NewErrArity("EXISTS")
	}

	var count int
	for _, key := range args {
		if store.GetNoTouch(key) != nil {
			count++
		}
	}

	return clientio.Encode(count, false)
}

func evalPersist(args []string, store *dstore.Store) []byte {
	if len(args) != 1 {
		return diceerrors.NewErrArity("PERSIST")
	}

	key := args[0]

	obj := store.Get(key)

	// If the key does not exist, return RESP encoded 0 to denote the key does not exist
	if obj == nil {
		return clientio.RespZero
	}

	// If the object exists but no expiration is set on it, return -1
	_, isExpirySet := dstore.GetExpiry(obj, store)
	if !isExpirySet {
		return clientio.RespMinusOne
	}

	// If the object exists, remove the expiration time
	dstore.DelExpiry(obj, store)

	return clientio.RespOne
}

func evalCOPY(args []string, store *dstore.Store) []byte {
	if len(args) < 2 {
		return diceerrors.NewErrArity("COPY")
	}

	isReplace := false

	sourceKey := args[0]
	destinationKey := args[1]
	sourceObj := store.Get(sourceKey)
	if sourceObj == nil {
		return clientio.RespZero
	}

	for i := 2; i < len(args); i++ {
		arg := strings.ToUpper(args[i])
		if arg == "REPLACE" {
			isReplace = true
		}
	}

	if isReplace {
		store.Del(destinationKey)
	}

	destinationObj := store.Get(destinationKey)
	if destinationObj != nil {
		return clientio.RespZero
	}

	copyObj := sourceObj.DeepCopy()
	if copyObj == nil {
		return clientio.RespZero
	}

	exp, ok := dstore.GetExpiry(sourceObj, store)
	var exDurationMs int64 = -1
	if ok {
		exDurationMs = int64(exp - uint64(utils.GetCurrentTime().UnixMilli()))
	}

	store.Put(destinationKey, copyObj)

	if exDurationMs > 0 {
		store.SetExpiry(copyObj, exDurationMs)
	}
	return clientio.RespOne
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
func evalGETEX(args []string, store *dstore.Store) []byte {
	if len(args) < 1 {
		return diceerrors.NewErrArity("GETEX")
	}

	var key string = args[0]

	// Get the key from the hash table
	obj := store.Get(key)

	// if key does not exist, return RESP encoded nil
	if obj == nil {
		return clientio.RespNIL
	}

	// check if the object is set type if yes then return error
	if object.AssertType(obj.TypeEncoding, object.ObjTypeSet) == nil {
		return diceerrors.NewErrWithFormattedMessage(diceerrors.WrongTypeErr)
	}

	var exDurationMs int64 = -1
	var state exDurationState = Uninitialized
	var persist bool = false
	for i := 1; i < len(args); i++ {
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
			if exDuration <= 0 || exDuration > maxExDuration {
				return diceerrors.NewErrExpireTime("GETEX")
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

			if exDuration < 0 || exDuration > maxExDuration {
				return diceerrors.NewErrExpireTime("GETEX")
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

		case "PERSIST":
			if state != Uninitialized {
				return diceerrors.NewErrWithMessage(diceerrors.SyntaxErr)
			}
			persist = true
			state = Initialized
		default:
			return diceerrors.NewErrWithMessage(diceerrors.SyntaxErr)
		}
	}

	if state == Initialized {
		if persist {
			dstore.DelExpiry(obj, store)
		} else {
			store.SetExpiry(obj, exDurationMs)
		}
	}

	// return the RESP encoded value
	return clientio.Encode(obj.Value, false)
}

// evalPTTL returns Time-to-Live in millisecs for the queried key in args
// The key should be the only param in args else returns with an error
// Returns	RESP encoded time (in secs) remaining for the key to expire
//
//	RESP encoded -2 stating key doesn't exist or key is expired
//	RESP encoded -1 in case no expiration is set on the key
func evalPTTL(args []string, store *dstore.Store) []byte {
	if len(args) != 1 {
		return diceerrors.NewErrArity("PTTL")
	}

	key := args[0]

	obj := store.Get(key)

	if obj == nil {
		return clientio.RespMinusTwo
	}

	exp, isExpirySet := dstore.GetExpiry(obj, store)

	if !isExpirySet {
		return clientio.RespMinusOne
	}

	// compute the time remaining for the key to expire and
	// return the RESP encoded form of it
	durationMs := exp - uint64(utils.GetCurrentTime().UnixMilli())
	return clientio.Encode(int64(durationMs), false)
}

// evalHSET sets the specified fields to their
// respective values in an hashmap stored at key
//
// This command overwrites the values of specified
// fields that exist in the hash.
//
// If key doesn't exist, a new key holding a hash is created.
//
// Usage: HSET key field value [field value ...]
func evalHSET(args []string, store *dstore.Store) []byte {
	if len(args) < 3 {
		return diceerrors.NewErrArity("HSET")
	}

	key := args[0]

	obj := store.Get(key)

	var hashMap HashMap
	var numKeys int64

	if obj != nil {
		if err := object.AssertTypeAndEncoding(obj.TypeEncoding, object.ObjTypeHashMap, object.ObjEncodingHashMap); err != nil {
			return diceerrors.NewErrWithMessage(diceerrors.WrongTypeErr)
		}
		hashMap = obj.Value.(HashMap)
	}

	keyValuePairs := args[1:]
	hashMap, numKeys, err := hashMapBuilder(keyValuePairs, hashMap)
	if err != nil {
		return diceerrors.NewErrWithMessage(err.Error())
	}

	obj = store.NewObj(hashMap, -1, object.ObjTypeHashMap, object.ObjEncodingHashMap)

	store.Put(key, obj)

	return clientio.Encode(numKeys, false)
}

func evalHGETALL(args []string, store *dstore.Store) []byte {
	if len(args) != 1 {
		return diceerrors.NewErrArity("HGETALL")
	}

	key := args[0]

	obj := store.Get(key)

	var hashMap HashMap
	var results []string

	if obj != nil {
		if err := object.AssertTypeAndEncoding(obj.TypeEncoding, object.ObjTypeHashMap, object.ObjEncodingHashMap); err != nil {
			return diceerrors.NewErrWithMessage(diceerrors.WrongTypeErr)
		}
		hashMap = obj.Value.(HashMap)
	}

	for hmKey, hmValue := range hashMap {
		results = append(results, hmKey, hmValue)
	}

	return clientio.Encode(results, false)
}

func evalHGET(args []string, store *dstore.Store) []byte {
	if len(args) != 2 {
		return diceerrors.NewErrArity("HGET")
	}

	key := args[0]
	hmKey := args[1]

	val, errWithMessage := getValueFromHashMap(key, hmKey, store)
	if errWithMessage != nil {
		return errWithMessage
	}
	return val
}

func evalObjectIdleTime(key string, store *dstore.Store) []byte {
	obj := store.GetNoTouch(key)
	if obj == nil {
		return clientio.RespNIL
	}

	return clientio.Encode(int64(dstore.GetIdleTime(obj.LastAccessedAt)), true)
}

func evalOBJECT(args []string, store *dstore.Store) []byte {
	if len(args) < 2 {
		return diceerrors.NewErrArity("OBJECT")
	}

	subcommand := strings.ToUpper(args[0])
	key := args[1]

	switch subcommand {
	case "IDLETIME":
		return evalObjectIdleTime(key, store)
	default:
		return diceerrors.NewErrWithMessage(diceerrors.SyntaxErr)
	}
}

func evalTOUCH(args []string, store *dstore.Store) []byte {
	if len(args) == 0 {
		return diceerrors.NewErrArity("TOUCH")
	}

	count := 0
	for _, key := range args {
		if store.Get(key) != nil {
			count++
		}
	}

	return clientio.Encode(count, false)
}

func evalLPUSH(args []string, store *dstore.Store) []byte {
	if len(args) < 2 {
		return diceerrors.NewErrArity("LPUSH")
	}

	obj := store.Get(args[0])
	if obj == nil {
		obj = store.NewObj(NewDeque(), -1, object.ObjTypeByteList, object.ObjEncodingDeque)
	}

	// if object is a set type, return error
	if object.AssertType(obj.TypeEncoding, object.ObjTypeSet) == nil {
		return diceerrors.NewErrWithFormattedMessage(diceerrors.WrongTypeErr)
	}

	if err := object.AssertType(obj.TypeEncoding, object.ObjTypeByteList); err != nil {
		return clientio.Encode(err, false)
	}

	if err := object.AssertEncoding(obj.TypeEncoding, object.ObjEncodingDeque); err != nil {
		return clientio.Encode(err, false)
	}

	store.Put(args[0], obj)
	for i := 1; i < len(args); i++ {
		obj.Value.(*Deque).LPush(args[i])
	}

	return clientio.RespOK
}

func evalRPUSH(args []string, store *dstore.Store) []byte {
	if len(args) < 2 {
		return diceerrors.NewErrArity("RPUSH")
	}

	obj := store.Get(args[0])
	if obj == nil {
		obj = store.NewObj(NewDeque(), -1, object.ObjTypeByteList, object.ObjEncodingDeque)
	}

	// if object is a set type, return error
	if object.AssertType(obj.TypeEncoding, object.ObjTypeSet) == nil {
		return diceerrors.NewErrWithFormattedMessage(diceerrors.WrongTypeErr)
	}

	if err := object.AssertType(obj.TypeEncoding, object.ObjTypeByteList); err != nil {
		return clientio.Encode(err, false)
	}

	if err := object.AssertEncoding(obj.TypeEncoding, object.ObjEncodingDeque); err != nil {
		return clientio.Encode(err, false)
	}

	store.Put(args[0], obj)
	for i := 1; i < len(args); i++ {
		obj.Value.(*Deque).RPush(args[i])
	}

	return clientio.RespOK
}

func evalRPOP(args []string, store *dstore.Store) []byte {
	if len(args) != 1 {
		return diceerrors.NewErrArity("RPOP")
	}

	obj := store.Get(args[0])
	if obj == nil {
		return clientio.RespNIL
	}

	// if object is a set type, return error
	if object.AssertType(obj.TypeEncoding, object.ObjTypeSet) == nil {
		return diceerrors.NewErrWithFormattedMessage(diceerrors.WrongTypeErr)
	}

	if err := object.AssertType(obj.TypeEncoding, object.ObjTypeByteList); err != nil {
		return clientio.Encode(err, false)
	}

	if err := object.AssertEncoding(obj.TypeEncoding, object.ObjEncodingDeque); err != nil {
		return clientio.Encode(err, false)
	}

	deq := obj.Value.(*Deque)
	x, err := deq.RPop()
	if err != nil {
		if err == ErrDequeEmpty {
			return clientio.RespNIL
		}
		panic(fmt.Sprintf("unknown error: %v", err))
	}

	return clientio.Encode(x, false)
}

func evalLPOP(args []string, store *dstore.Store) []byte {
	if len(args) != 1 {
		return diceerrors.NewErrArity("LPOP")
	}

	obj := store.Get(args[0])
	if obj == nil {
		return clientio.RespNIL
	}

	// if object is a set type, return error
	if object.AssertType(obj.TypeEncoding, object.ObjTypeSet) == nil {
		return diceerrors.NewErrWithFormattedMessage(diceerrors.WrongTypeErr)
	}

	if err := object.AssertType(obj.TypeEncoding, object.ObjTypeByteList); err != nil {
		return clientio.Encode(err, false)
	}

	if err := object.AssertEncoding(obj.TypeEncoding, object.ObjEncodingDeque); err != nil {
		return clientio.Encode(err, false)
	}

	deq := obj.Value.(*Deque)
	x, err := deq.LPop()
	if err != nil {
		if err == ErrDequeEmpty {
			return clientio.RespNIL
		}
		panic(fmt.Sprintf("unknown error: %v", err))
	}

	return clientio.Encode(x, false)
}

func evalLLEN(args []string, store *dstore.Store) []byte {
	if len(args) != 1 {
		return diceerrors.NewErrArity("LLEN")
	}

	obj := store.Get(args[0])
	if obj == nil {
		return clientio.Encode(0, false)
	}

	if err := object.AssertTypeAndEncoding(obj.TypeEncoding, object.ObjTypeByteList, object.ObjEncodingDeque); err != nil {
		return err
	}

	deq := obj.Value.(*Deque)
	return clientio.Encode(deq.Length, false)
}

// GETSET atomically sets key to value and returns the old value stored at key.
// Returns an error when key exists but does not hold a string value.
// Any previous time to live associated with the key is
// discarded on successful SET operation.
//
// Returns:
// Bulk string reply: the old value stored at the key.
// Nil reply: if the key does not exist.
func evalGETSET(args []string, store *dstore.Store) []byte {
	if len(args) != 2 {
		return diceerrors.NewErrArity("GETSET")
	}

	key, value := args[0], args[1]
	getResp := evalGET([]string{key}, store)
	// Check if it's an error resp from GET
	if strings.HasPrefix(string(getResp), "-") {
		return getResp
	}

	// Previous TTL needs to be reset
	setResp := evalSET([]string{key, value}, store)
	// Check if it's an error resp from SET
	if strings.HasPrefix(string(setResp), "-") {
		return setResp
	}

	return getResp
}

func evalFLUSHDB(args []string, store *dstore.Store) []byte {
	log.Info(args)
	if len(args) > 1 {
		return diceerrors.NewErrArity("FLUSHDB")
	}

	flushType := Sync
	if len(args) == 1 {
		flushType = strings.ToUpper(args[0])
	}

	// TODO: Update this method to work with shared-nothing multithreaded implementation
	switch flushType {
	case Sync, Async:
		store.ResetStore()
	default:
		return diceerrors.NewErrWithMessage(diceerrors.SyntaxErr)
	}

	return clientio.RespOK
}

func evalSADD(args []string, store *dstore.Store) []byte {
	if len(args) < 2 {
		return diceerrors.NewErrArity("SADD")
	}
	key := args[0]

	// Get the set object from the store.
	obj := store.Get(key)
	lengthOfItems := len(args[1:])

	var count int = 0
	if obj == nil {
		var exDurationMs int64 = -1
		var keepttl bool = false
		// If the object does not exist, create a new set object.
		value := make(map[string]struct{}, lengthOfItems)
		// Create a new object.
		obj = store.NewObj(value, exDurationMs, object.ObjTypeSet, object.ObjEncodingSetStr)
		store.Put(key, obj, dstore.WithKeepTTL(keepttl))
	}

	if err := object.AssertType(obj.TypeEncoding, object.ObjTypeSet); err != nil {
		return diceerrors.NewErrWithFormattedMessage(diceerrors.WrongTypeErr)
	}

	if err := object.AssertEncoding(obj.TypeEncoding, object.ObjEncodingSetStr); err != nil {
		return diceerrors.NewErrWithFormattedMessage(diceerrors.WrongTypeErr)
	}

	// Get the set object.
	set := obj.Value.(map[string]struct{})

	for _, arg := range args[1:] {
		if _, ok := set[arg]; !ok {
			set[arg] = struct{}{}
			count++
		}
	}

	return clientio.Encode(count, false)
}

func evalSMEMBERS(args []string, store *dstore.Store) []byte {
	if len(args) != 1 {
		return diceerrors.NewErrArity("SMEMBERS")
	}
	key := args[0]

	// Get the set object from the store.
	obj := store.Get(key)

	if obj == nil {
		return clientio.Encode([]string{}, false)
	}

	// If the object exists, check if it is a set object.
	if err := object.AssertType(obj.TypeEncoding, object.ObjTypeSet); err != nil {
		return diceerrors.NewErrWithFormattedMessage(diceerrors.WrongTypeErr)
	}

	if err := object.AssertEncoding(obj.TypeEncoding, object.ObjEncodingSetStr); err != nil {
		return diceerrors.NewErrWithFormattedMessage(diceerrors.WrongTypeErr)
	}

	// Get the set object.
	set := obj.Value.(map[string]struct{})
	// Get the members of the set.
	members := make([]string, 0, len(set))
	for k := range set {
		members = append(members, k)
	}

	return clientio.Encode(members, false)
}

func evalSREM(args []string, store *dstore.Store) []byte {
	if len(args) < 2 {
		return diceerrors.NewErrArity("SREM")
	}
	key := args[0]

	// Get the set object from the store.
	obj := store.Get(key)

	var count int = 0
	if obj == nil {
		return clientio.Encode(count, false)
	}

	// If the object exists, check if it is a set object.
	if err := object.AssertType(obj.TypeEncoding, object.ObjTypeSet); err != nil {
		return diceerrors.NewErrWithFormattedMessage(diceerrors.WrongTypeErr)
	}

	if err := object.AssertEncoding(obj.TypeEncoding, object.ObjEncodingSetStr); err != nil {
		return diceerrors.NewErrWithFormattedMessage(diceerrors.WrongTypeErr)
	}

	// Get the set object.
	set := obj.Value.(map[string]struct{})

	for _, arg := range args[1:] {
		if _, ok := set[arg]; ok {
			delete(set, arg)
			count++
		}
	}

	return clientio.Encode(count, false)
}

func evalSCARD(args []string, store *dstore.Store) []byte {
	if len(args) != 1 {
		return diceerrors.NewErrArity("SCARD")
	}

	key := args[0]

	// Get the set object from the store.
	obj := store.Get(key)

	if obj == nil {
		return clientio.Encode(0, false)
	}

	// If the object exists, check if it is a set object.
	if err := object.AssertType(obj.TypeEncoding, object.ObjTypeSet); err != nil {
		return diceerrors.NewErrWithFormattedMessage(diceerrors.WrongTypeErr)
	}

	if err := object.AssertEncoding(obj.TypeEncoding, object.ObjEncodingSetStr); err != nil {
		return diceerrors.NewErrWithFormattedMessage(diceerrors.WrongTypeErr)
	}

	// Get the set object.
	count := len(obj.Value.(map[string]struct{}))
	return clientio.Encode(count, false)
}

func evalSDIFF(args []string, store *dstore.Store) []byte {
	if len(args) < 1 {
		return diceerrors.NewErrArity("SDIFF")
	}

	srcKey := args[0]
	obj := store.Get(srcKey)

	// if the source key does not exist, return an empty response
	if obj == nil {
		return clientio.Encode([]string{}, false)
	}

	if err := object.AssertType(obj.TypeEncoding, object.ObjTypeSet); err != nil {
		return diceerrors.NewErrWithFormattedMessage(diceerrors.WrongTypeErr)
	}

	if err := object.AssertEncoding(obj.TypeEncoding, object.ObjEncodingSetStr); err != nil {
		return diceerrors.NewErrWithFormattedMessage(diceerrors.WrongTypeErr)
	}

	// Get the set object from the store.
	// store the count as the number of elements in the first set
	srcSet := obj.Value.(map[string]struct{})
	count := len(srcSet)

	tmpSet := make(map[string]struct{}, count)
	for k := range srcSet {
		tmpSet[k] = struct{}{}
	}

	// we decrement the count as we find the elements in the other sets
	// if the count is 0, we skip further sets but still get them from
	// the store to check if they are set objects and update their last accessed time

	for _, arg := range args[1:] {
		// Get the set object from the store.
		obj := store.Get(arg)

		if obj == nil {
			continue
		}

		// If the object exists, check if it is a set object.
		if err := object.AssertType(obj.TypeEncoding, object.ObjTypeSet); err != nil {
			return diceerrors.NewErrWithFormattedMessage(diceerrors.WrongTypeErr)
		}

		if err := object.AssertEncoding(obj.TypeEncoding, object.ObjEncodingSetStr); err != nil {
			return diceerrors.NewErrWithFormattedMessage(diceerrors.WrongTypeErr)
		}

		// only if the count is greater than 0, we need to check the other sets
		if count > 0 {
			// Get the set object.
			set := obj.Value.(map[string]struct{})

			for k := range set {
				if _, ok := tmpSet[k]; ok {
					delete(tmpSet, k)
					count--
				}
			}
		}
	}

	if count == 0 {
		return clientio.Encode([]string{}, false)
	}

	// Get the members of the set.
	members := make([]string, 0, len(tmpSet))
	for k := range tmpSet {
		members = append(members, k)
	}
	return clientio.Encode(members, false)
}

func evalSINTER(args []string, store *dstore.Store) []byte {
	if len(args) < 2 {
		return diceerrors.NewErrArity("SINTER")
	}

	sets := make([]map[string]struct{}, 0, len(args))

	var empty int = 0

	for _, arg := range args {
		// Get the set object from the store.
		obj := store.Get(arg)

		if obj == nil {
			empty++
			continue
		}

		// If the object exists, check if it is a set object.
		if err := object.AssertType(obj.TypeEncoding, object.ObjTypeSet); err != nil {
			return diceerrors.NewErrWithFormattedMessage(diceerrors.WrongTypeErr)
		}

		if err := object.AssertEncoding(obj.TypeEncoding, object.ObjEncodingSetStr); err != nil {
			return diceerrors.NewErrWithFormattedMessage(diceerrors.WrongTypeErr)
		}

		// Get the set object.
		set := obj.Value.(map[string]struct{})
		sets = append(sets, set)
	}

	if empty > 0 {
		return clientio.Encode([]string{}, false)
	}

	// sort the sets by the number of elements in the set
	// we will iterate over the smallest set
	// and check if the element is present in all the other sets
	sort.Slice(sets, func(i, j int) bool {
		return len(sets[i]) < len(sets[j])
	})

	count := 0
	resultSet := make(map[string]struct{}, len(sets[0]))

	// init the result set with the first set
	// store the number of elements in the first set in count
	// we will decrement the count if we do not find the elements in the other sets
	for k := range sets[0] {
		resultSet[k] = struct{}{}
		count++
	}

	for i := 1; i < len(sets); i++ {
		if count == 0 {
			break
		}
		for k := range resultSet {
			if _, ok := sets[i][k]; !ok {
				delete(resultSet, k)
				count--
			}
		}
	}

	if count == 0 {
		return clientio.Encode([]string{}, false)
	}

	members := make([]string, 0, len(resultSet))
	for k := range resultSet {
		members = append(members, k)
	}
	return clientio.Encode(members, false)
}

// PFADD Adds all the element arguments to the HyperLogLog data structure stored at the variable
// name specified as first argument.
//
// Returns:
// If the approximated cardinality estimated by the HyperLogLog changed after executing the command,
// returns 1, otherwise 0 is returned.
func evalPFADD(args []string, store *dstore.Store) []byte {
	if len(args) < 1 {
		return diceerrors.NewErrArity("PFADD")
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
		return clientio.Encode(1, false)
	}

	existingHll, ok := obj.Value.(*hyperloglog.Sketch)
	if !ok {
		return diceerrors.NewErrWithMessage(diceerrors.WrongTypeHllErr)
	}
	initialCardinality := existingHll.Estimate()
	for _, arg := range args[1:] {
		existingHll.Insert([]byte(arg))
	}

	if newCardinality := existingHll.Estimate(); initialCardinality != newCardinality {
		return clientio.Encode(1, false)
	}

	return clientio.Encode(0, false)
}

func evalPFCOUNT(args []string, store *dstore.Store) []byte {
	if len(args) < 1 {
		return diceerrors.NewErrArity("PFCOUNT")
	}

	unionHll := hyperloglog.New()

	for _, arg := range args {
		obj := store.Get(arg)
		if obj != nil {
			currKeyHll, ok := obj.Value.(*hyperloglog.Sketch)
			if !ok {
				return diceerrors.NewErrWithMessage(diceerrors.WrongTypeHllErr)
			}
			err := unionHll.Merge(currKeyHll)
			if err != nil {
				return diceerrors.NewErrWithMessage(diceerrors.InvalidHllErr)
			}
		}
	}

	return clientio.Encode(unionHll.Estimate(), false)
}

func evalPFMERGE(args []string, store *dstore.Store) []byte {
	if len(args) < 1 {
		return diceerrors.NewErrArity("PFMERGE")
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
			return diceerrors.NewErrWithMessage(diceerrors.WrongTypeHllErr)
		}
	}

	for _, arg := range args {
		obj := store.Get(arg)
		if obj != nil {
			currKeyHll, ok := obj.Value.(*hyperloglog.Sketch)
			if !ok {
				return diceerrors.NewErrWithMessage(diceerrors.WrongTypeHllErr)
			}

			err := mergedHll.Merge(currKeyHll)
			if err != nil {
				return diceerrors.NewErrWithMessage(diceerrors.InvalidHllErr)
			}
		}
	}

	// Save the mergedHll
	obj = store.NewObj(mergedHll, -1, object.ObjTypeString, object.ObjEncodingRaw)
	store.Put(destKey, obj)

	return clientio.RespOK
}

func evalJSONSTRLEN(args []string, store *dstore.Store) []byte {
	if len(args) < 1 {
		return diceerrors.NewErrArity("JSON.STRLEN")
	}

	key := args[0]

	if len(args) < 2 {
		// no recursive
		// making consistent with arrlen
		// to-do parsing
		obj := store.Get(key)

		if obj == nil {
			return clientio.RespNIL
		}
		jsonData := obj.Value

		if utils.GetJSONFieldType(jsonData) != utils.StringType {
			return diceerrors.NewErrWithFormattedMessage(diceerrors.JSONPathValueTypeErr)
		}
		return clientio.Encode(len(jsonData.(string)), false)
	}

	path := args[1]

	obj := store.Get(key)

	if obj == nil {
		return clientio.RespNIL
	}

	// Check if the object is of JSON type
	errWithMessage := object.AssertTypeAndEncoding(obj.TypeEncoding, object.ObjTypeJSON, object.ObjEncodingJSON)
	if errWithMessage != nil {
		return errWithMessage
	}

	jsonData := obj.Value
	if path == defaultRootPath {
		defaultStringResult := make([]interface{}, 0, 1)
		if utils.GetJSONFieldType(jsonData) == utils.StringType {
			defaultStringResult = append(defaultStringResult, int64(len(jsonData.(string))))
		} else {
			defaultStringResult = append(defaultStringResult, clientio.RespNIL)
		}

		return clientio.Encode(defaultStringResult, false)
	}

	// Parse the JSONPath expression
	expr, err := jp.ParseString(path)
	if err != nil {
		return diceerrors.NewErrWithMessage("invalid JSONPath")
	}
	// Execute the JSONPath query
	results := expr.Get(jsonData)
	if len(results) == 0 {
		return clientio.Encode([]interface{}{}, false)
	}
	strLenResults := make([]interface{}, 0, len(results))
	for _, result := range results {
		switch utils.GetJSONFieldType(result) {
		case utils.StringType:
			strLenResults = append(strLenResults, int64(len(result.(string))))
		default:
			strLenResults = append(strLenResults, clientio.RespNIL)
		}
	}
	return clientio.Encode(strLenResults, false)
}

func evalHLEN(args []string, store *dstore.Store) []byte {
	if len(args) != 1 {
		return diceerrors.NewErrArity("HLEN")
	}

	key := args[0]

	obj := store.Get(key)

	if obj == nil {
		return clientio.RespZero
	}

	if err := object.AssertTypeAndEncoding(obj.TypeEncoding, object.ObjTypeHashMap, object.ObjEncodingHashMap); err != nil {
		return diceerrors.NewErrWithFormattedMessage(diceerrors.WrongTypeErr)
	}

	hashMap := obj.Value.(HashMap)
	return clientio.Encode(len(hashMap), false)
}

func evalSELECT(args []string, store *dstore.Store) []byte {
	if len(args) != 1 {
		return diceerrors.NewErrArity("SELECT")
	}

	return clientio.RespOK
}

func formatFloat(f float64, b bool) string {
	formatted := strconv.FormatFloat(f, 'f', -1, 64)
	if b {
		parts := strings.Split(formatted, ".")
		if len(parts) == 1 {
			formatted += ".0"
		}
	}
	return formatted
}

// takes original value, increment values (float or int), a flag representing if increment is float
// returns new value, string representation, a boolean representing if the value was modified
func incrementValue(value any, isIncrFloat bool, incrFloat float64, incrInt int64) (newVal interface{}, stringRepresentation string, isModified bool) {
	switch utils.GetJSONFieldType(value) {
	case utils.NumberType:
		oldVal := value.(float64)
		var newVal float64
		if isIncrFloat {
			newVal = oldVal + incrFloat
		} else {
			newVal = oldVal + float64(incrInt)
		}
		resultString := formatFloat(newVal, isIncrFloat)
		return newVal, resultString, true
	case utils.IntegerType:
		if isIncrFloat {
			oldVal := float64(value.(int64))
			newVal := oldVal + incrFloat
			resultString := formatFloat(newVal, isIncrFloat)
			return newVal, resultString, true
		} else {
			oldVal := value.(int64)
			newVal := oldVal + incrInt
			resultString := fmt.Sprintf("%d", newVal)
			return newVal, resultString, true
		}
	default:
		return value, null, false
	}
}

func evalJSONNUMINCRBY(args []string, store *dstore.Store) []byte {
	if len(args) < 3 {
		return diceerrors.NewErrArity("JSON.NUMINCRBY")
	}
	key := args[0]
	obj := store.Get(key)

	if obj == nil {
		return diceerrors.NewErrWithFormattedMessage("-ERR could not perform this operation on a key that doesn't exist")
	}

	// Check if the object is of JSON type
	errWithMessage := object.AssertTypeAndEncoding(obj.TypeEncoding, object.ObjTypeJSON, object.ObjEncodingJSON)
	if errWithMessage != nil {
		return errWithMessage
	}

	path := args[1]

	jsonData := obj.Value
	// Parse the JSONPath expression
	expr, err := jp.ParseString(path)

	if err != nil {
		return diceerrors.NewErrWithMessage("invalid JSONPath")
	}

	isIncrFloat := false

	for i, r := range args[2] {
		if !unicode.IsDigit(r) && r != '.' && r != '-' {
			if i == 0 {
				return diceerrors.NewErrWithFormattedMessage("-ERR expected value at line 1 column %d", i+1)
			}
			return diceerrors.NewErrWithFormattedMessage("-ERR trailing characters at line 1 column %d", i+1)
		}
		if r == '.' {
			isIncrFloat = true
		}
	}
	var incrFloat float64
	var incrInt int64
	if isIncrFloat {
		incrFloat, err = strconv.ParseFloat(args[2], 64)
		if err != nil {
			return diceerrors.NewErrWithMessage(diceerrors.IntOrOutOfRangeErr)
		}
	} else {
		incrInt, err = strconv.ParseInt(args[2], 10, 64)
		if err != nil {
			return diceerrors.NewErrWithMessage(diceerrors.IntOrOutOfRangeErr)
		}
	}
	results := expr.Get(jsonData)

	if len(results) == 0 {
		respString := "[]"
		return clientio.Encode(respString, false)
	}

	resultArray := make([]string, 0, len(results))

	if path == defaultRootPath {
		newValue, resultString, isModified := incrementValue(jsonData, isIncrFloat, incrFloat, incrInt)
		if isModified {
			jsonData = newValue
		}
		resultArray = append(resultArray, resultString)
	} else {
		// Execute the JSONPath query
		_, err := expr.Modify(jsonData, func(value any) (interface{}, bool) {
			newValue, resultString, isModified := incrementValue(value, isIncrFloat, incrFloat, incrInt)
			resultArray = append(resultArray, resultString)
			return newValue, isModified
		})
		if err != nil {
			return diceerrors.NewErrWithMessage("invalid JSONPath")
		}
	}

	resultString := `[` + strings.Join(resultArray, ",") + `]`

	obj.Value = jsonData
	return clientio.Encode(resultString, false)
}

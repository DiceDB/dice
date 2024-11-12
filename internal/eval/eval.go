package eval

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/dicedb/dice/internal/eval/geo"
	"github.com/dicedb/dice/internal/eval/sortedset"
	"github.com/dicedb/dice/internal/object"
	"github.com/rs/xid"

	"github.com/dicedb/dice/internal/sql"

	"github.com/bytedance/sonic"
	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/internal/clientio"
	"github.com/dicedb/dice/internal/comm"
	diceerrors "github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/querymanager"
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

// EvalResponse represents the response of an evaluation operation for a command from store.
// It contains the sequence ID, the result of the store operation, and any error encountered during the operation.
type EvalResponse struct {
	Result interface{} // Result holds the outcome of the Store operation. Currently, it is expected to be of type []byte, but this may change in the future.
	Error  error       // Error holds any error that occurred during the operation. If no error, it will be nil.
}

// Following functions should be used to create a new EvalResponse with the given result and error.
// These ensure that result and error are mutually exclusive.
// If result is nil, then error should be non-nil and vice versa.

// makeEvalResult creates a new EvalResponse with the given result and nil error.
// This is a helper function to create a new EvalResponse with the given result and nil error.
/**
 * @param {interface{}} result - The result of the store operation.
 * @returns {EvalResponse} A new EvalResponse with the given result and nil error.
 */
func makeEvalResult(result interface{}) *EvalResponse {
	return &EvalResponse{
		Result: result,
		Error:  nil,
	}
}

// makeEvalError creates a new EvalResponse with the given error and nil result.
// This is a helper function to create a new EvalResponse with the given error and nil result.
/**
 * @param {error} err - The error that occurred during the store operation.
 * @returns {EvalResponse} A new EvalResponse with the given error and nil result.
 */
func makeEvalError(err error) *EvalResponse {
	return &EvalResponse{
		Result: nil,
		Error:  err,
	}
}

type jsonOperation string

const (
	IncrBy = "INCRBY"
	MultBy = "MULTBY"
)

const (
	defaultRootPath = "$"
	maxExDuration   = 9223372036854775
	CountConst      = "COUNT"
)

func init() {
	diceCommandsCount = len(DiceCmds)
	TxnCommands = map[string]bool{"EXEC": true, "DISCARD": true}
	serverID = fmt.Sprintf("%s:%d", config.DiceConfig.AsyncServer.Addr, config.DiceConfig.AsyncServer.Port)
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

// evalECHO returns the argument passed by the user
func evalECHO(args []string, store *dstore.Store) []byte {
	if len(args) != 1 {
		return diceerrors.NewErrArity("ECHO")
	}

	return clientio.Encode(args[0], false)
}

// EvalAUTH returns with an encoded "OK" if the user is authenticated
// If the user is not authenticated, it returns with an encoded error message
func EvalAUTH(args []string, c *comm.Client) []byte {
	var err error

	if config.DiceConfig.Auth.Password == "" {
		return diceerrors.NewErrWithMessage("AUTH <password> called without any password configured for the default user. Are you sure your configuration is correct?")
	}

	username := config.DiceConfig.Auth.UserName
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

// evalDBSIZE returns the number of keys in the database.
func evalDBSIZE(args []string, store *dstore.Store) []byte {
	if len(args) > 0 {
		return diceerrors.NewErrArity("DBSIZE")
	}

	// Expired keys must be explicitly deleted since the cronFrequency for cleanup is configurable.
	// A longer delay may prevent timely cleanup, leading to incorrect DBSIZE results.
	dstore.DeleteExpiredKeys(store)
	// return the RESP encoded value
	return clientio.Encode(store.GetDBSize(), false)
}

// evalJSONDEBUG reports value's memory usage in bytes
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
		path = args[1] // anymore args are ignored for this command altogether
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
	var results []any
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

// trimElementAndUpdateArray trim the array between the given start and stop index
// Returns trimmed array
func trimElementAndUpdateArray(arr []any, start, stop int) []any {
	updatedArray := make([]any, 0)
	length := len(arr)
	if len(arr) == 0 {
		return updatedArray
	}
	var startIdx, stopIdx int

	if start >= length {
		return updatedArray
	}

	startIdx = adjustIndex(start, arr)
	stopIdx = adjustIndex(stop, arr)

	if startIdx > stopIdx {
		return updatedArray
	}

	updatedArray = arr[startIdx : stopIdx+1]
	return updatedArray
}

// insertElementAndUpdateArray add an element at the given index
// Returns remaining array and error
func insertElementAndUpdateArray(arr []any, index int, elements []interface{}) (updatedArray []any, err error) {
	length := len(arr)
	var idx int
	if index >= -length && index <= length {
		idx = adjustIndex(index, arr)
	} else {
		return nil, errors.New("index out of bounds")
	}
	before := arr[:idx]
	after := arr[idx:]

	elements = append(elements, after...)
	before = append(before, elements...)
	updatedArray = append(updatedArray, before...)
	return updatedArray, nil
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

// evalJSONTYPE retrieves a JSON value type stored at the specified key
// args must contain at least the key;  (path unused in this implementation)
// Returns response.RespNIL if key is expired, or it does not exist
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
// Returns response.RespNIL if key is expired, or it does not exist
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
		return result, diceerrors.NewErrWithMessage(fmt.Sprintf("Path '%s' does not exist", path))
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
		switch strings.ToUpper(args[i]) {
		case NX:
			if i != len(args)-1 {
				return diceerrors.NewErrWithMessage(diceerrors.SyntaxErr)
			}
			obj := store.Get(key)
			if obj != nil {
				return clientio.RespNIL
			}
		case XX:
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

// Parses and returns the input string as an int64 or float64
func parseFloatInt(input string) (result interface{}, err error) {
	// Try to parse as an integer
	if intValue, parseErr := strconv.ParseInt(input, 10, 64); parseErr == nil {
		result = intValue
		return
	}

	// Try to parse as a float
	if floatValue, parseErr := strconv.ParseFloat(input, 64); parseErr == nil {
		result = floatValue
		return
	}

	// If neither parsing succeeds, return an error
	err = errors.New("invalid input: not a valid int or float")
	return
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

// evalDEL deletes all the specified keys in args list
// returns the count of total deleted keys after encoding
func evalDEL(args []string, store *dstore.Store) []byte {
	countDeleted := 0

	if len(args) < 1 {
		return diceerrors.NewErrArity("DEL")
	}

	for _, key := range args {
		if ok := store.Del(key); ok {
			countDeleted++
		}
	}

	return clientio.Encode(countDeleted, false)
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

// evalINFO creates a buffer with the info of total keys per db
// Returns the encoded buffer as response
func evalINFO(args []string, store *dstore.Store) []byte {
	var info []byte
	buf := bytes.NewBuffer(info)
	buf.WriteString("# Keyspace\r\n")
	fmt.Fprintf(buf, "db0:keys=%d,expires=0,avg_ttl=0\r\n", store.GetKeyCount())
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
func EvalQWATCH(args []string, httpOp, websocketOp bool, client *comm.Client, store *dstore.Store) []byte {
	if len(args) != 1 {
		return diceerrors.NewErrArity("Q.WATCH")
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
	var watchSubscription querymanager.QuerySubscription

	if httpOp || websocketOp {
		watchSubscription = querymanager.QuerySubscription{
			Subscribe:          true,
			Query:              query,
			CacheChan:          cacheChannel,
			QwatchClientChan:   client.HTTPQwatchResponseChan,
			ClientIdentifierID: client.ClientIdentifierID,
		}
	} else {
		watchSubscription = querymanager.QuerySubscription{
			Subscribe: true,
			Query:     query,
			ClientFD:  client.Fd,
			CacheChan: cacheChannel,
		}
	}

	querymanager.QuerySubscriptionChan <- watchSubscription
	store.CacheKeysForQuery(query.Where, cacheChannel)

	// Return the result of the query.
	responseChan := make(chan querymanager.AdhocQueryResult)
	querymanager.AdhocQueryChan <- querymanager.AdhocQuery{
		Query:        query,
		ResponseChan: responseChan,
	}

	queryResult := <-responseChan
	if queryResult.Err != nil {
		return clientio.Encode(queryResult.Err, false)
	}

	// TODO: We should return the list of all queries being watched by the client.
	return clientio.Encode(querymanager.GenericWatchResponse(sql.Qwatch, query.String(), *queryResult.Result), false)
}

// EvalQUNWATCH removes the specified key from the watch list for the caller client.
func EvalQUNWATCH(args []string, httpOp bool, client *comm.Client) []byte {
	if len(args) != 1 {
		return diceerrors.NewErrArity("Q.UNWATCH")
	}
	query, e := sql.ParseQuery( /*sql=*/ args[0])
	if e != nil {
		return clientio.Encode(e, false)
	}

	if httpOp {
		querymanager.QuerySubscriptionChan <- querymanager.QuerySubscription{
			Subscribe:          false,
			Query:              query,
			QwatchClientChan:   client.HTTPQwatchResponseChan,
			ClientIdentifierID: client.ClientIdentifierID,
		}
	} else {
		querymanager.QuerySubscriptionChan <- querymanager.QuerySubscription{
			Subscribe: false,
			Query:     query,
			ClientFD:  client.Fd,
		}
	}

	return clientio.RespOK
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
				return diceerrors.NewErrWithFormattedMessage(diceerrors.WrongTypeErr)
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
		return evalCommandCount(args[1:])
	case GetKeys:
		return evalCommandGetKeys(args[1:])
	case List:
		return evalCommandList(args[1:])
	case Help:
		return evalCommandHelp(args[1:])
	case Info:
		return evalCommandInfo(args[1:])
	case Docs:
		return evalCommandDocs(args[1:])
	default:
		return diceerrors.NewErrWithFormattedMessage("unknown subcommand '%s'. Try COMMAND HELP.", subcommand)
	}
}

// evalCommandHelp prints help message
func evalCommandHelp(args []string) []byte {
	if len(args) > 0 {
		return diceerrors.NewErrArity("COMMAND|HELP")
	}

	format := "COMMAND <subcommand> [<arg> [value] [opt] ...]. Subcommands are:"
	noTitle := "(no subcommand)"
	noMessage := "     Return details about all DiceDB commands."
	countTitle := CountConst
	countMessage := "     Return the total number of commands in this DiceDB server."
	listTitle := "LIST"
	listMessage := "     Return a list of all commands in this DiceDB server."
	infoTitle := "INFO [<command-name> ...]"
	infoMessage := "     Return details about the specified DiceDB commands. If no command names are given, documentation details for all commands are returned."
	docsTitle := "DOCS [<command-name> ...]"
	docsMessage := "\tReturn documentation details about multiple diceDB commands.\n\tIf no command names are given, documentation details for all\n\tcommands are returned."
	getKeysTitle := "GETKEYS <full-command>"
	getKeysMessage := "     Return the keys from a full DiceDB command."
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
		infoTitle,
		infoMessage,
		docsTitle,
		docsMessage,
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

func evalCommandList(args []string) []byte {
	if len(args) > 0 {
		return diceerrors.NewErrArity("COMMAND|LIST")
	}

	cmds := make([]string, 0, diceCommandsCount)
	for k := range DiceCmds {
		cmds = append(cmds, k)
		for _, sc := range DiceCmds[k].SubCommands {
			cmds = append(cmds, fmt.Sprint(k, "|", sc))
		}
	}
	return clientio.Encode(cmds, false)
}

// evalKeys returns the list of keys that match the pattern should be the only param in args
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

// evalCommandCount returns a number of commands supported by DiceDB
func evalCommandCount(args []string) []byte {
	if len(args) > 0 {
		return diceerrors.NewErrArity("COMMAND|COUNT")
	}

	return clientio.Encode(diceCommandsCount, false)
}

// evalCommandGetKeys helps identify which arguments in a redis command
// are interpreted as keys.
// This is useful in analyzing long commands / scripts
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

func evalCommandDefaultDocs() []byte {
	cmds := convertDiceCmdsMapToDocs()
	return clientio.Encode(cmds, false)
}

func evalCommandInfo(args []string) []byte {
	if len(args) == 0 {
		return evalCommandDefault()
	}

	cmdMetaMap := make(map[string]interface{})
	for _, cmdMeta := range DiceCmds {
		cmdMetaMap[cmdMeta.Name] = convertCmdMetaToSlice(&cmdMeta)
	}

	var result []interface{}
	for _, arg := range args {
		arg = strings.ToUpper(arg)
		if cmdMeta, found := cmdMetaMap[arg]; found {
			result = append(result, cmdMeta)
		} else {
			result = append(result, clientio.RespNIL)
		}
	}

	return clientio.Encode(result, false)
}

func evalCommandDocs(args []string) []byte {
	if len(args) == 0 {
		return evalCommandDefaultDocs()
	}

	cmdMetaMap := make(map[string]interface{})
	for _, cmdMeta := range DiceCmds {
		cmdMetaMap[cmdMeta.Name] = convertCmdMetaToDocs(&cmdMeta)
	}

	var result []interface{}
	for _, arg := range args {
		arg = strings.ToUpper(arg)
		if cmdMeta, found := cmdMetaMap[arg]; found {
			result = append(result, cmdMeta)
		}
	}

	return clientio.Encode(result, false)
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

	// If the object exists but no expiration is set on it, return 0
	_, isExpirySet := dstore.GetExpiry(obj, store)
	if !isExpirySet {
		return clientio.RespZero
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

func evalObjectIdleTime(key string, store *dstore.Store) []byte {
	obj := store.GetNoTouch(key)
	if obj == nil {
		return clientio.RespNIL
	}

	return clientio.Encode(int64(dstore.GetIdleTime(obj.LastAccessedAt)), true)
}

func evalObjectEncoding(key string, store *dstore.Store) []byte {
	var encodingTypeStr string

	obj := store.GetNoTouch(key)
	if obj == nil {
		return clientio.RespNIL
	}

	oType, oEnc := object.ExtractTypeEncoding(obj)
	switch {
	case oType == object.ObjTypeString && oEnc == object.ObjEncodingRaw:
		encodingTypeStr = "raw"
		return clientio.Encode(encodingTypeStr, false)

	case oType == object.ObjTypeString && oEnc == object.ObjEncodingEmbStr:
		encodingTypeStr = "embstr"
		return clientio.Encode(encodingTypeStr, false)

	case oType == object.ObjTypeInt && oEnc == object.ObjEncodingInt:
		encodingTypeStr = "int"
		return clientio.Encode(encodingTypeStr, false)

	case oType == object.ObjTypeByteList && oEnc == object.ObjEncodingDeque:
		encodingTypeStr = "deque"
		return clientio.Encode(encodingTypeStr, false)

	case oType == object.ObjTypeBitSet && oEnc == object.ObjEncodingBF:
		encodingTypeStr = "bf"
		return clientio.Encode(encodingTypeStr, false)

	case oType == object.ObjTypeJSON && oEnc == object.ObjEncodingJSON:
		encodingTypeStr = "json"
		return clientio.Encode(encodingTypeStr, false)

	case oType == object.ObjTypeByteArray && oEnc == object.ObjEncodingByteArray:
		encodingTypeStr = "bytearray"
		return clientio.Encode(encodingTypeStr, false)

	case oType == object.ObjTypeSet && oEnc == object.ObjEncodingSetStr:
		encodingTypeStr = "setstr"
		return clientio.Encode(encodingTypeStr, false)

	case oType == object.ObjTypeSet && oEnc == object.ObjEncodingSetInt:
		encodingTypeStr = "setint"
		return clientio.Encode(encodingTypeStr, false)

	case oType == object.ObjTypeHashMap && oEnc == object.ObjEncodingHashMap:
		encodingTypeStr = "hashmap"
		return clientio.Encode(encodingTypeStr, false)

	case oType == object.ObjTypeSortedSet && oEnc == object.ObjEncodingBTree:
		encodingTypeStr = "btree"
		return clientio.Encode(encodingTypeStr, false)

	default:
		return diceerrors.NewErrWithFormattedMessage(diceerrors.WrongTypeErr)
	}
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
	case "ENCODING":
		return evalObjectEncoding(key, store)
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

	deq := obj.Value.(*Deque)

	return clientio.Encode(deq.Length, false)
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

	deq := obj.Value.(*Deque)

	return clientio.Encode(deq.Length, false)
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
		if errors.Is(err, ErrDequeEmpty) {
			return clientio.RespNIL
		}
		panic(fmt.Sprintf("unknown error: %v", err))
	}

	return clientio.Encode(x, false)
}

func evalLPOP(args []string, store *dstore.Store) []byte {
	// By default we pop only 1
	popNumber := 1

	// LPOP accepts 1 or 2 arguments only - LPOP key [count]
	if len(args) < 1 || len(args) > 2 {
		return diceerrors.NewErrArity("LPOP")
	}

	// to updated the number of pops
	if len(args) == 2 {
		nos, err := strconv.Atoi(args[1])
		if err != nil {
			return diceerrors.NewErrWithFormattedMessage(diceerrors.IntOrOutOfRangeErr)
		}
		if nos == 0 {
			// returns empty string if count given is 0
			return clientio.Encode([]string{}, false)
		}
		if nos < 0 {
			// returns an out of range err if count is negetive
			return diceerrors.NewErrWithFormattedMessage(diceerrors.ValOutOfRangeErr)
		}
		popNumber = nos
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

	// holds the elements popped
	var elements []string
	for iter := 0; iter < popNumber; iter++ {
		x, err := deq.LPop()
		if err != nil {
			if errors.Is(err, ErrDequeEmpty) {
				break
			}
			panic(fmt.Sprintf("unknown error: %v", err))
		}
		elements = append(elements, x)
	}

	if len(elements) == 0 {
		return clientio.RespNIL
	}

	if len(elements) == 1 {
		return clientio.Encode(elements[0], false)
	}

	return clientio.Encode(elements, false)
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

func evalFLUSHDB(args []string, store *dstore.Store) []byte {
	slog.Info("FLUSHDB called", slog.Any("args", args))
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
	if len(args) < 1 {
		return diceerrors.NewErrArity("SINTER")
	}

	sets := make([]map[string]struct{}, 0, len(args))

	empty := 0

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

func evalSELECT(args []string, store *dstore.Store) []byte {
	if len(args) != 1 {
		return diceerrors.NewErrArity("SELECT")
	}

	return clientio.RespOK
}

// formatFloat formats float64 as string.
// Optionally appends a decimal (.0) for whole numbers,
// if b is true.
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

func evalTYPE(args []string, store *dstore.Store) []byte {
	if len(args) != 1 {
		return diceerrors.NewErrArity("TYPE")
	}
	key := args[0]
	obj := store.Get(key)
	if obj == nil {
		return clientio.Encode("none", true)
	}

	var typeStr string
	switch oType, _ := object.ExtractTypeEncoding(obj); oType {
	case object.ObjTypeString, object.ObjTypeInt, object.ObjTypeByteArray:
		typeStr = "string"
	case object.ObjTypeByteList:
		typeStr = "list"
	case object.ObjTypeSet:
		typeStr = "set"
	case object.ObjTypeHashMap:
		typeStr = "hash"
	default:
		typeStr = "non-supported type"
	}

	return clientio.Encode(typeStr, true)
}

func evalJSONRESP(args []string, store *dstore.Store) []byte {
	if len(args) < 1 {
		return diceerrors.NewErrArity("json.resp")
	}
	key := args[0]

	path := defaultRootPath
	if len(args) > 1 {
		path = args[1]
	}

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
		resp := parseJSONStructure(jsonData, false)

		return clientio.Encode(resp, false)
	}

	// if path is not root then extract value at path
	expr, err := jp.ParseString(path)
	if err != nil {
		return diceerrors.NewErrWithMessage("invalid JSONPath")
	}
	results := expr.Get(jsonData)

	// process value at each path
	ret := []any{}
	for _, result := range results {
		resp := parseJSONStructure(result, false)
		ret = append(ret, resp)
	}

	return clientio.Encode(ret, false)
}

func parseJSONStructure(jsonData interface{}, nested bool) (resp []any) {
	switch json := jsonData.(type) {
	case string, bool:
		resp = append(resp, json)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, nil:
		resp = append(resp, json)
	case map[string]interface{}:
		resp = append(resp, "{")
		for key, value := range json {
			resp = append(resp, key)
			resp = append(resp, parseJSONStructure(value, true)...)
		}
		// wrap in another array to offset print
		if nested {
			resp = []interface{}{resp}
		}
	case []interface{}:
		resp = append(resp, "[")
		for _, value := range json {
			resp = append(resp, parseJSONStructure(value, true)...)
		}
		// wrap in another array to offset print
		if nested {
			resp = []interface{}{resp}
		}
	default:
		resp = append(resp, []byte("(unsupported type)"))
	}
	return resp
}

// This method executes each operation, contained in ops array, based on commands used.
func executeBitfieldOps(value *ByteArray, ops []utils.BitFieldOp) []interface{} {
	overflowType := WRAP
	var result []interface{}
	for _, op := range ops {
		switch op.Kind {
		case GET:
			res := value.getBits(int(op.Offset), int(op.EVal), op.EType == SIGNED)
			result = append(result, res)
		case SET:
			prevValue := value.getBits(int(op.Offset), int(op.EVal), op.EType == SIGNED)
			value.setBits(int(op.Offset), int(op.EVal), op.Value)
			result = append(result, prevValue)
		case INCRBY:
			res, err := value.incrByBits(int(op.Offset), int(op.EVal), op.Value, overflowType, op.EType == SIGNED)
			if err != nil {
				result = append(result, nil)
			} else {
				result = append(result, res)
			}
		case OVERFLOW:
			overflowType = op.EType
		}
	}
	return result
}
func evalGEOADD(args []string, store *dstore.Store) []byte {
	if len(args) < 4 {
		return diceerrors.NewErrArity("GEOADD")
	}

	key := args[0]
	var nx, xx bool
	startIdx := 1

	// Parse options
	for startIdx < len(args) {
		option := strings.ToUpper(args[startIdx])
		if option == NX {
			nx = true
			startIdx++
		} else if option == XX {
			xx = true
			startIdx++
		} else {
			break
		}
	}

	// Check if we have the correct number of arguments after parsing options
	if (len(args)-startIdx)%3 != 0 {
		return diceerrors.NewErrArity("GEOADD")
	}

	if xx && nx {
		return diceerrors.NewErrWithMessage("ERR XX and NX options at the same time are not compatible")
	}

	// Get or create sorted set
	obj := store.Get(key)
	var ss *sortedset.Set
	if obj != nil {
		var err []byte
		ss, err = sortedset.FromObject(obj)
		if err != nil {
			return err
		}
	} else {
		ss = sortedset.New()
	}

	added := 0
	for i := startIdx; i < len(args); i += 3 {
		longitude, err := strconv.ParseFloat(args[i], 64)
		if err != nil || math.IsNaN(longitude) || longitude < -180 || longitude > 180 {
			return diceerrors.NewErrWithMessage("ERR invalid longitude")
		}

		latitude, err := strconv.ParseFloat(args[i+1], 64)
		if err != nil || math.IsNaN(latitude) || latitude < -85.05112878 || latitude > 85.05112878 {
			return diceerrors.NewErrWithMessage("ERR invalid latitude")
		}

		member := args[i+2]
		_, exists := ss.Get(member)

		// Handle XX option: Only update existing elements
		if xx && !exists {
			continue
		}

		// Handle NX option: Only add new elements
		if nx && exists {
			continue
		}

		hash := geo.EncodeHash(latitude, longitude)

		wasInserted := ss.Upsert(hash, member)
		if wasInserted {
			added++
		}
	}

	obj = store.NewObj(ss, -1, object.ObjTypeSortedSet, object.ObjEncodingBTree)
	store.Put(key, obj)

	return clientio.Encode(added, false)
}

func evalGEODIST(args []string, store *dstore.Store) []byte {
	if len(args) < 3 || len(args) > 4 {
		return diceerrors.NewErrArity("GEODIST")
	}

	key := args[0]
	member1 := args[1]
	member2 := args[2]
	unit := "m"
	if len(args) == 4 {
		unit = strings.ToLower(args[3])
	}

	// Get the sorted set
	obj := store.Get(key)
	if obj == nil {
		return clientio.RespNIL
	}

	ss, err := sortedset.FromObject(obj)
	if err != nil {
		return err
	}

	// Get the scores (geohashes) for both members
	score1, ok := ss.Get(member1)
	if !ok {
		return clientio.RespNIL
	}
	score2, ok := ss.Get(member2)
	if !ok {
		return clientio.RespNIL
	}

	lat1, lon1 := geo.DecodeHash(score1)
	lat2, lon2 := geo.DecodeHash(score2)

	distance := geo.GetDistance(lon1, lat1, lon2, lat2)

	result, err := geo.ConvertDistance(distance, unit)
	if err != nil {
		return err
	}

	return clientio.Encode(utils.RoundToDecimals(result, 4), false)
}

// evalJSONSTRAPPEND appends a string value to the JSON string value at the specified path
// in the JSON object saved at the key in arguments.
// Args must contain at least a key and the string value to append.
// If the key does not exist or is expired, it returns an error response.
// If the value at the specified path is not a string, it returns an error response.
// Returns the new length of the string at the specified path if successful.
func evalJSONSTRAPPEND(args []string, store *dstore.Store) []byte {
	if len(args) != 3 {
		return diceerrors.NewErrArity("JSON.STRAPPEND")
	}

	key := args[0]
	path := args[1]
	value := args[2]

	obj := store.Get(key)
	if obj == nil {
		return diceerrors.NewErrWithMessage(diceerrors.NoKeyExistsErr)
	}

	errWithMessage := object.AssertTypeAndEncoding(obj.TypeEncoding, object.ObjTypeJSON, object.ObjEncodingJSON)
	if errWithMessage != nil {
		return diceerrors.NewErrWithFormattedMessage(diceerrors.WrongKeyTypeErr)
	}

	jsonData := obj.Value

	var resultsArray []interface{}

	if path == "$" {
		// Handle root-level string
		if str, ok := jsonData.(string); ok {
			unquotedValue := strings.Trim(value, "\"")
			newValue := str + unquotedValue
			resultsArray = append(resultsArray, int64(len(newValue)))
			jsonData = newValue
		} else {
			return clientio.RespEmptyArray
		}
	} else {
		expr, err := jp.ParseString(path)
		if err != nil {
			return clientio.RespEmptyArray
		}

		_, modifyErr := expr.Modify(jsonData, func(data any) (interface{}, bool) {
			switch v := data.(type) {
			case string:
				unquotedValue := strings.Trim(value, "\"")
				newValue := v + unquotedValue
				resultsArray = append([]interface{}{int64(len(newValue))}, resultsArray...)
				return newValue, true
			default:
				resultsArray = append([]interface{}{clientio.RespNIL}, resultsArray...)
				return data, false
			}
		})

		if modifyErr != nil {
			return clientio.RespEmptyArray
		}
	}

	if len(resultsArray) == 0 {
		return clientio.RespEmptyArray
	}

	obj.Value = jsonData
	return clientio.Encode(resultsArray, false)
}

package eval

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dicedb/dice/internal/object"

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
// TODO: Needs to be removed after http and websocket migrated to the multithreading
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
// TODO: Needs to be removed after http and websocket migrated to the multithreading
func evalMSET(args []string, store *dstore.Store) []byte {
	if len(args) <= 1 || len(args)%2 != 0 {
		return diceerrors.NewErrArity("MSET")
	}

	// MSET does not have expiry support
	var exDurationMs int64 = -1

	insertMap := make(map[string]*object.Obj, len(args)/2)
	for i := 0; i < len(args); i += 2 {
		key, value := args[i], args[i+1]
		oType := deduceType(value)
		var storedValue interface{}
		switch oType {
		case object.ObjTypeInt:
			storedValue, _ = strconv.ParseInt(value, 10, 64)
		case object.ObjTypeString:
			storedValue = value
		default:
			return clientio.Encode(fmt.Errorf("ERR unsupported type: %d", oType), false)
		}
		insertMap[key] = store.NewObj(storedValue, exDurationMs, oType)
	}

	store.PutAll(insertMap)
	return clientio.RespOK
}

// evalDBSIZE returns the number of keys in the database.
// TODO: Needs to be removed after http and websocket migrated to the multithreading
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

// evalJSONMGET retrieves a JSON value stored for the multiple key
// args must contain at least the key and a path;
// Returns encoded error response if incorrect number of arguments
// The RESP value of the key is encoded and then returned
// TODO: Needs to be removed after http and websocket migrated to the multithreading
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
		result, _ := jsonMGETHelper(store, path, key)
		results = append(results, result)
	}

	var interfaceObj interface{} = results
	return clientio.Encode(interfaceObj, false)
}

func jsonMGETHelper(store *dstore.Store, path, key string) (result interface{}, err2 []byte) {
	// Retrieve the object from the database
	obj := store.Get(key)
	if obj == nil {
		return result, nil
	}

	// Check if the object is of JSON type
	if errWithMessage := object.AssertType(obj.Type, object.ObjTypeJSON); errWithMessage != nil {
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

// ReverseSlice takes a slice of any type and returns a new slice with the elements reversed.
func ReverseSlice[T any](slice []T) []T {
	reversed := make([]T, len(slice))
	for i, v := range slice {
		reversed[len(slice)-1-i] = v
	}
	return reversed
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

func evalHELLO(args []string, store *dstore.Store) []byte {
	if len(args) > 1 {
		return diceerrors.NewErrArity("HELLO")
	}

	var resp []interface{}
	serverID = fmt.Sprintf("%s:%d", config.DiceConfig.RespServer.Addr, config.DiceConfig.RespServer.Port)
	resp = append(resp,
		"proto", 2,
		"id", serverID,
		"mode", "standalone",
		"role", "master",
		"modules", []interface{}{})

	return clientio.Encode(resp, false)
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

// evalKeys returns the list of keys that match the pattern should be the only param in args
// TODO: Needs to be removed after http and websocket migrated to the multithreading
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

// TODO: Needs to be removed after http and websocket migrated to the multithreading
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
// TODO: Needs to be removed after http and websocket migrated to the multithreading
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

// TODO: Needs to be removed after http and websocket migrated to the multithreading
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
		if arg == dstore.Replace {
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

// TODO: Needs to be removed after http and websocket migrated to the multithreading
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

	if err := object.AssertType(obj.Type, object.ObjTypeSet); err != nil {
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
		if err := object.AssertType(obj.Type, object.ObjTypeSet); err != nil {
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

// Migrated to the new eval, but kept for http and websocket
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
		if err := object.AssertType(obj.Type, object.ObjTypeSet); err != nil {
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

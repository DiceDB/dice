package core

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bytedance/sonic"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/core/bit"
	"github.com/ohler55/ojg/jp"
)

type exDurationState int

const (
	Uninitialized exDurationState = iota
	Initialized
)

var RESP_NIL []byte = []byte("$-1\r\n")
var RESP_OK []byte = []byte("+OK\r\n")
var RESP_AUTH_FAILURE []byte = []byte("+NOAUTH\r\n")
var RESP_QUEUED []byte = []byte("+QUEUED\r\n")
var RESP_ZERO []byte = []byte(":0\r\n")
var RESP_ONE []byte = []byte(":1\r\n")
var RESP_MINUS_1 []byte = []byte(":-1\r\n")
var RESP_MINUS_2 []byte = []byte(":-2\r\n")
var RESP_EMPTY_ARRAY []byte = []byte("*0\r\n")

var txnCommands map[string]bool
var serverID string
var diceCommandsCount int

const defaultRootPath = "$"

func init() {
	diceCommandsCount = len(diceCmds)
	txnCommands = map[string]bool{"EXEC": true, "DISCARD": true}
	serverID = fmt.Sprintf("%s:%d", config.Host, config.Port)
}

// evalPING returns with an encoded "PONG"
// If any message is added with the ping command,
// the message will be returned.
func evalPING(args []string) []byte {
	var b []byte

	if len(args) >= 2 {
		return Encode(errors.New("ERR wrong number of arguments for 'ping' command"), false)
	}

	if len(args) == 0 {
		b = Encode("PONG", true)
	} else {
		b = Encode(args[0], false)
	}

	return b
}

// evalAUTH returns with an encoded "OK" if the user is authenticated
// If the user is not authenticated, it returns with an encoded error message
func evalAUTH(args []string, c *Client) []byte {
	var (
		b   []byte
		err error
	)

	if len(args) < 1 || len(args) > 2 {
		return Encode(errors.New("ERR wrong number of arguments for 'AUTH' command"), false)
	}

	if len(args) == 1 {
		if err = c.Session.Validate(DefaultUserName, args[0]); err != nil {
			log.Println("AUTH failed: ", err)
			return Encode(errors.New("AUTH failed"), false)
		}
		return RESP_OK
	}

	if len(args) == 2 {
		if err = c.Session.Validate(args[0], args[1]); err != nil {
			log.Println("AUTH failed: ", err)
			return Encode(errors.New("AUTH failed"), false)
		}
		return RESP_OK
	}

	return b
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
func evalSET(args []string) []byte {
	if len(args) <= 1 {
		return Encode(errors.New("ERR wrong number of arguments for 'set' command"), false)
	}

	var key, value string
	var exDurationMs int64 = -1
	var state exDurationState = Uninitialized

	key, value = args[0], args[1]
	oType, oEnc := deduceTypeEncoding(value)

	for i := 2; i < len(args); i++ {
		arg := strings.ToUpper(args[i])
		switch arg {
		case "EX", "PX":
			if state != Uninitialized {
				return Encode(errors.New("ERR syntax error"), false)
			}
			i++
			if i == len(args) {
				return Encode(errors.New("ERR syntax error"), false)
			}

			exDuration, err := strconv.ParseInt(args[i], 10, 64)
			if err != nil {
				return Encode(errors.New("ERR value is not an integer or out of range"), false)
			}
			if exDuration <= 0 {
				return Encode(errors.New("ERR invalid expire time in 'set' command"), false)
			}

			// converting seconds to milliseconds
			if arg == "EX" {
				exDuration = exDuration * 1000
			}
			exDurationMs = exDuration
			state = Initialized

		case "PXAT", "EXAT":
			if state != Uninitialized {
				return Encode(errors.New("ERR syntax error"), false)
			}
			i++
			if i == len(args) {
				return Encode(errors.New("ERR syntax error"), false)
			}
			exDuration, err := strconv.ParseInt(args[i], 10, 64)
			if err != nil {
				return Encode(errors.New("ERR value is not an integer or out of range"), false)
			}

			if exDuration < 0 {
				return Encode(errors.New("ERR invalid expire time in 'set' command"), false)
			}

			if arg == "EXAT" {
				exDuration = exDuration * 1000
			}
			exDurationMs = exDuration - time.Now().UnixMilli()
			// If the expiry time is in the past, set exDurationMs to 0
			// This will be used to signal immediate expiration
			if exDurationMs < 0 {
				exDurationMs = 0
			}
			state = Initialized

		case "XX":
			// Get the key from the hash table
			obj := Get(key)

			// if key does not exist, return RESP encoded nil
			if obj == nil {
				return RESP_NIL
			}
		case "NX":
			obj := Get(key)
			if obj != nil {
				return RESP_NIL
			}
		default:
			return Encode(errors.New("ERR syntax error"), false)
		}
	}

	// putting the k and value in a Hash Table
	Put(key, NewObj(value, exDurationMs, oType, oEnc))
	return RESP_OK
}

// evalMSET puts multiple <key, value> pairs in db as in the args
// MSET is atomic, so all given keys are set at once.
// args must contain key and value pairs.

// Returns encoded error response if at least a <key, value> pair is not part of args
// Returns encoded OK RESP once new entries are added
// If the key already exists then the value will be overwritten and expiry will be discarded
func evalMSET(args []string) []byte {
	if len(args) <= 1 || len(args)%2 != 0 {
		return Encode(errors.New("ERR wrong number of arguments for 'mset' command"), false)
	}

	// MSET does not have expiry support
	var exDurationMs int64 = -1

	insertMap := make(map[string]*Obj, len(args)/2)
	for i := 0; i < len(args); i += 2 {
		key, value := args[i], args[i+1]
		oType, oEnc := deduceTypeEncoding(value)
		insertMap[key] = NewObj(value, exDurationMs, oType, oEnc)
	}

	PutAll(insertMap)
	return RESP_OK
}

// evalGET returns the value for the queried key in args
// The key should be the only param in args
// The RESP value of the key is encoded and then returned
// evalGET returns RESP_NIL if key is expired or it does not exist
func evalGET(args []string) []byte {
	if len(args) != 1 {
		return Encode(errors.New("ERR wrong number of arguments for 'get' command"), false)
	}

	var key string = args[0]

	// Get the key from the hash table
	obj := Get(key)

	// if key does not exist, return RESP encoded nil
	if obj == nil {
		return RESP_NIL
	}

	// return the RESP encoded value
	return Encode(obj.Value, false)
}

// evalJSONGET retrieves a JSON value stored at the specified key
// args must contain at least the key;  (path unused in this implementation)
// Returns RESP_NIL if key is expired or it does not exist
// Returns encoded error response if incorrect number of arguments
// The RESP value of the key is encoded and then returned
func evalJSONGET(args []string) []byte {
	if len(args) < 1 {
		return Encode(errors.New("ERR wrong number of arguments for 'JSON.GET' command"), false)
	}

	key := args[0]
	// Default path is root if not specified
	path := defaultRootPath
	if len(args) > 1 {
		path = args[1]
	}

	// Retrieve the object from the database
	obj := Get(key)
	if obj == nil {
		return RESP_NIL
	}

	// Check if the object is of JSON type
	err := assertType(obj.TypeEncoding, OBJ_TYPE_JSON)
	if err != nil {
		return Encode(err, false)
	}
	err = assertEncoding(obj.TypeEncoding, OBJ_ENCODING_JSON)
	if err != nil {
		return Encode(err, false)
	}

	jsonData := obj.Value

	// If path is root, return the entire JSON
	if path == defaultRootPath {
		resultBytes, err := sonic.Marshal(jsonData)
		if err != nil {
			return Encode(errors.New("ERR could not serialize result"), false)
		}
		return Encode(string(resultBytes), false)
	}

	// Parse the JSONPath expression
	expr, err := jp.ParseString(path)
	if err != nil {
		return Encode(errors.New("ERR invalid JSONPath"), false)
	}

	// Execute the JSONPath query
	results := expr.Get(jsonData)
	if len(results) == 0 {
		return RESP_NIL
	}

	// Serialize the result
	var resultBytes []byte
	if len(results) == 1 {
		resultBytes, err = sonic.Marshal(results[0])
	} else {
		resultBytes, err = sonic.Marshal(results)
	}
	if err != nil {
		return Encode(errors.New("ERR could not serialize result"), false)
	}
	return Encode(string(resultBytes), false)
}

// evalJSONSET stores a JSON value at the specified key
// args must contain at least the key, path (unused in this implementation), and JSON string
// Returns encoded error response if incorrect number of arguments
// Returns encoded error if the JSON string is invalid
// Returns RESP_OK if the JSON value is successfully stored
func evalJSONSET(args []string) []byte {
	// Check if there are enough arguments
	if len(args) < 3 {
		return Encode(errors.New("ERR wrong number of arguments for 'JSON.SET' command"), false)
	}

	key := args[0]
	path := args[1]
	jsonStr := args[2]

	for i := 3; i < len(args); i++ {
		switch args[i] {
		case "NX", "nx":
			if i != len(args)-1 {
				return Encode(errors.New("ERR syntax error"), false)
			}
			obj := Get(key)
			if obj != nil {
				return RESP_NIL
			}
		case "XX", "xx":
			if i != len(args)-1 {
				return Encode(errors.New("ERR syntax error"), false)
			}
			obj := Get(key)
			if obj == nil {
				return RESP_NIL
			}

		default:
			return Encode(errors.New("ERR syntax error"), false)
		}
	}

	// Parse the JSON string
	var jsonValue interface{}
	if err := sonic.UnmarshalString(jsonStr, &jsonValue); err != nil {
		return Encode(fmt.Errorf("ERR invalid JSON: %v", err.Error()), false)
	}

	// Retrieve existing object or create new one
	obj := Get(key)
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
		err := assertType(obj.TypeEncoding, OBJ_TYPE_JSON)
		if err != nil {
			return Encode(err, false)
		}
		err = assertEncoding(obj.TypeEncoding, OBJ_ENCODING_JSON)
		if err != nil {
			return Encode(err, false)
		}
		rootData = obj.Value
	}

	// If path is not root, use JSONPath to set the value
	if path != defaultRootPath {
		expr, err := jp.ParseString(path)
		if err != nil {
			return Encode(errors.New("ERR invalid JSONPath"), false)
		}

		err = expr.Set(rootData, jsonValue)
		if err != nil {
			return Encode(errors.New("ERR failed to set value"), false)
		}
	} else {
		// If path is root, replace the entire JSON
		rootData = jsonValue
	}

	// Create a new object with the updated JSON data
	newObj := NewObj(rootData, -1, OBJ_TYPE_JSON, OBJ_ENCODING_JSON)
	Put(key, newObj)

	return RESP_OK
}

// evalTTL returns Time-to-Live in secs for the queried key in args
// The key should be the only param in args else returns with an error
// Returns	RESP encoded time (in secs) remaining for the key to expire
//
//	RESP encoded -2 stating key doesn't exist or key is expired
//	RESP encoded -1 in case no expiration is set on the key
func evalTTL(args []string) []byte {
	if len(args) != 1 {
		return Encode(errors.New("ERR wrong number of arguments for 'ttl' command"), false)
	}

	var key string = args[0]

	obj := Get(key)

	// if key does not exist, return RESP encoded -2 denoting key does not exist
	if obj == nil {
		return RESP_MINUS_2
	}

	// if object exist, but no expiration is set on it then send -1
	exp, isExpirySet := getExpiry(obj)
	if !isExpirySet {
		return RESP_MINUS_1
	}

	// compute the time remaining for the key to expire and
	// return the RESP encoded form of it
	durationMs := exp - uint64(time.Now().UnixMilli())

	return Encode(int64(durationMs/1000), false)
}

// evalDEL deletes all the specified keys in args list
// returns the count of total deleted keys after encoding
func evalDEL(args []string) []byte {
	var countDeleted int = 0

	for _, key := range args {
		if ok := Del(key); ok {
			countDeleted++
		}
	}

	return Encode(countDeleted, false)
}

// evalEXPIRE sets a expiry time(in secs) on the specified key in args
// args should contain 2 values, key and the expiry time to be set for the key
// The expiry time should be in integer format; if not, it returns encoded error response
// Returns RESP_ONE if expiry was set on the key successfully.
// Once the time is lapsed, the key will be deleted automatically
func evalEXPIRE(args []string) []byte {
	if len(args) <= 1 {
		return Encode(errors.New("ERR wrong number of arguments for 'expire' command"), false)
	}

	var key string = args[0]
	exDurationSec, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return Encode(errors.New("ERR value is not an integer or out of range"), false)
	}

	obj := Get(key)

	// 0 if the timeout was not set. e.g. key doesn't exist, or operation skipped due to the provided arguments
	if obj == nil {
		return RESP_ZERO
	}

	setExpiry(obj, exDurationSec*1000)

	// 1 if the timeout was set.
	return RESP_ONE
}

func evalHELLO(args []string) []byte {
	if len(args) > 1 {
		return Encode(errors.New("ERR wrong number of arguments for 'hello' command"), false)
	}

	var response []interface{}
	response = append(response, "proto", 2)
	response = append(response, "id", serverID)
	response = append(response, "mode", "standalone")
	response = append(response, "role", "master")
	response = append(response, "modules", []interface{}{})

	return Encode(response, false)
}

/* Description - Spawn a background thread to persist the data via AOF technique. Current implementation is
based on CoW optimization and Fork */
// TODO: Implement Acknowledgement so that main process could know whether child has finished writing to its AOF file or not.
// TODO: Make it safe from failure, an stable policy would be to write the new flushes to a temporary files and then rename them to the main process's AOF file
// TODO: Add fsync() and fdatasync() to persist to AOF for above cases.
func evalBGREWRITEAOF(args []string) []byte {
	// Fork a child process, this child process would inherit all the uncommitted pages from main process.
	// This technique utilizes the CoW or copy-on-write, so while the main process is free to modify them
	// the child would save all the pages to disk.
	// Check details here -https://www.sobyte.net/post/2022-10/fork-cow/
	newChild, _, _ := syscall.Syscall(syscall.SYS_FORK, 0, 0, 0)
	if newChild == 0 {
		//We are inside child process now, so we'll start flushing to disk.
		if err := DumpAllAOF(); err != nil {
			return Encode(errors.New("ERR AOF failed"), false)
		}
		return []byte("")
	} else {
		//Back to main thread
		return RESP_OK
	}
}

// evalINCR increments the value of the specified key in args by 1,
// if the key exists and the value is integer format.
// The key should be the only param in args.
// If the key does not exist, new key is created with value 0,
// the value of the new key is then incremented.
// The value for the queried key should be of integer format,
// if not evalINCR returns encoded error response.
// evalINCR returns the incremented value for the key if there are no errors.
func evalINCR(args []string) []byte {
	if len(args) != 1 {
		return Encode(errors.New("ERR wrong number of arguments for 'incr' command"), false)
	}
	return incrDecrCmd(args, 1)
}

// evalDECR decrements the value of the specified key in args by 1,
// if the key exists and the value is integer format.
// The key should be the only param in args.
// If the key does not exist, new key is created with value 0,
// the value of the new key is then decremented.
// The value for the queried key should be of integer format,
// if not evalDECR returns encoded error response.
// evalDECR returns the decremented value for the key if there are no errors.
func evalDECR(args []string) []byte {
	if len(args) != 1 {
		return Encode(errors.New("ERR wrong number of arguments for 'decr' command"), false)
	}
	return incrDecrCmd(args, -1)
}

func incrDecrCmd(args []string, incr int64) []byte {
	var key string = args[0]
	obj := Get(key)
	if obj == nil {
		obj = NewObj("0", -1, OBJ_TYPE_STRING, OBJ_ENCODING_INT)
		Put(key, obj)
	}

	if err := assertType(obj.TypeEncoding, OBJ_TYPE_STRING); err != nil {
		return Encode(err, false)
	}

	if err := assertEncoding(obj.TypeEncoding, OBJ_ENCODING_INT); err != nil {
		return Encode(err, false)
	}

	i, _ := strconv.ParseInt(obj.Value.(string), 10, 64)
	// check overflow
	if (incr < 0 && i < 0 && incr < (math.MinInt64-i)) ||
		(incr > 0 && i > 0 && incr > (math.MaxInt64-i)) {
		return Encode(errors.New("ERR value is out of range"), false)
	}

	i += int64(incr)
	obj.Value = strconv.FormatInt(i, 10)

	return Encode(i, false)
}

// evalINFO creates a buffer with the info of total keys per db
// Returns the encoded buffer as response
func evalINFO(args []string) []byte {
	var info []byte
	buf := bytes.NewBuffer(info)
	buf.WriteString("# Keyspace\r\n")
	for i := range KeyspaceStat {
		buf.WriteString(fmt.Sprintf("db%d:keys=%d,expires=0,avg_ttl=0\r\n", i, KeyspaceStat[i]["keys"]))
	}
	return Encode(buf.String(), false)
}

// TODO: Placeholder to support monitoring
func evalCLIENT(args []string) []byte {
	return RESP_OK
}

// TODO: Placeholder to support monitoring
func evalLATENCY(args []string) []byte {
	return Encode([]string{}, false)
}

// evalLRU deletes all the keys from the LRU
// returns encoded RESP OK
func evalLRU(args []string) []byte {
	evictAllkeysLRU()
	return RESP_OK
}

// evalSLEEP sets db to sleep for the specified number of seconds.
// The sleep time should be the only param in args.
// Returns error response if the time param in args is not of integer format.
// evalSLEEP returns RESP_OK after sleeping for mentioned seconds
func evalSLEEP(args []string) []byte {
	if len(args) != 1 {
		return Encode(errors.New("ERR wrong number of arguments for 'SLEEP' command"), false)
	}

	durationSec, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return Encode(errors.New("ERR value is not an integer or out of range"), false)
	}
	time.Sleep(time.Duration(durationSec) * time.Second)
	return RESP_OK
}

// evalMULTI marks the start of the transaction for the client.
// All subsequent commands fired will be queued for atomic execution.
// The commands will not be executed until EXEC is triggered.
// Once EXEC is triggered it executes all the commands in queue,
// and closes the MULTI transaction.
func evalMULTI(args []string) []byte {
	return RESP_OK
}

// evalQINTINS inserts the provided integer in the key identified by key
// first argument will be the key, that should be of type `QINT`
// second argument will be the integer value
// if the key does not exist, evalQINTINS will also create the integer queue
func evalQINTINS(args []string) []byte {
	if len(args) != 2 {
		return Encode(errors.New("ERR invalid number of arguments for `QINTINS` command"), false)
	}

	x, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return Encode(errors.New("ERR only integer values can be inserted in QINT"), false)
	}

	obj := Get(args[0])
	if obj == nil {
		obj = NewObj(NewQueueInt(), -1, OBJ_TYPE_BYTELIST, OBJ_ENCODING_QINT)
	}

	if err := assertType(obj.TypeEncoding, OBJ_TYPE_BYTELIST); err != nil {
		return Encode(err, false)
	}

	if err := assertEncoding(obj.TypeEncoding, OBJ_ENCODING_QINT); err != nil {
		return Encode(err, false)
	}

	Put(args[0], obj)

	q := obj.Value.(*QueueInt)
	q.Insert(x)

	return RESP_OK
}

// evalSTACKINTPUSH pushes the provided integer in the key identified by key
// first argument will be the key, that should be of type `STACKINT`
// second argument will be the integer value
// if the key does not exist, evalSTACKINTPUSH will also create the integer stack
func evalSTACKINTPUSH(args []string) []byte {
	if len(args) != 2 {
		return Encode(errors.New("ERR invalid number of arguments for `STACKINTPUSH` command"), false)
	}

	x, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return Encode(errors.New("ERR only integer values can be inserted in STACKINT"), false)
	}

	obj := Get(args[0])
	if obj == nil {
		obj = NewObj(NewStackInt(), -1, OBJ_TYPE_BYTELIST, OBJ_ENCODING_STACKINT)
	}

	if err := assertType(obj.TypeEncoding, OBJ_TYPE_BYTELIST); err != nil {
		return Encode(err, false)
	}

	if err := assertEncoding(obj.TypeEncoding, OBJ_ENCODING_STACKINT); err != nil {
		return Encode(err, false)
	}

	Put(args[0], obj)

	s := obj.Value.(*StackInt)
	s.Push(x)

	return RESP_OK
}

// evalQINTREM removes the element from the QINT identified by key
// first argument will be the key, that should be of type `QINT`
// if the key does not exist, evalQINTREM returns nil otherwise it
// returns the integer value popped from the queue
// if we remove from the empty queue, nil is returned
func evalQINTREM(args []string) []byte {
	if len(args) != 1 {
		return Encode(errors.New("ERR invalid number of arguments for `QINTREM` command"), false)
	}

	obj := Get(args[0])
	if obj == nil {
		return RESP_NIL
	}

	if err := assertType(obj.TypeEncoding, OBJ_TYPE_BYTELIST); err != nil {
		return Encode(err, false)
	}

	if err := assertEncoding(obj.TypeEncoding, OBJ_ENCODING_QINT); err != nil {
		return Encode(err, false)
	}

	q := obj.Value.(*QueueInt)
	x, err := q.Remove()

	if err == ErrQueueEmpty {
		return RESP_NIL
	}

	return Encode(x, false)
}

// evalSTACKINTPOP pops the element from the STACKINT identified by key
// first argument will be the key, that should be of type `STACKINT`
// if the key does not exist, evalSTACKINTPOP returns nil otherwise it
// returns the integer value popped from the stack
// if we remove from the empty stack, nil is returned
func evalSTACKINTPOP(args []string) []byte {
	if len(args) != 1 {
		return Encode(errors.New("ERR invalid number of arguments for `STACKINTPOP` command"), false)
	}

	obj := Get(args[0])
	if obj == nil {
		return RESP_NIL
	}

	if err := assertType(obj.TypeEncoding, OBJ_TYPE_BYTELIST); err != nil {
		return Encode(err, false)
	}

	if err := assertEncoding(obj.TypeEncoding, OBJ_ENCODING_STACKINT); err != nil {
		return Encode(err, false)
	}

	s := obj.Value.(*StackInt)
	x, err := s.Pop()

	if err == ErrStackEmpty {
		return RESP_NIL
	}

	return Encode(x, false)
}

// evalQINTLEN returns the length of the QINT identified by key
// returns the integer value indicating the length of the queue
// if the key does not exist, the response is 0
func evalQINTLEN(args []string) []byte {
	if len(args) != 1 {
		return Encode(errors.New("ERR invalid number of arguments for `QINTLEN` command"), false)
	}

	obj := Get(args[0])
	if obj == nil {
		return RESP_ZERO
	}

	if err := assertType(obj.TypeEncoding, OBJ_TYPE_BYTELIST); err != nil {
		return Encode(err, false)
	}

	if err := assertEncoding(obj.TypeEncoding, OBJ_ENCODING_QINT); err != nil {
		return Encode(err, false)
	}

	q := obj.Value.(*QueueInt)
	return Encode(q.Length, false)
}

// evalSTACKINTLEN returns the length of the STACKINT identified by key
// returns the integer value indicating the length of the stack
// if the key does not exist, the response is 0
func evalSTACKINTLEN(args []string) []byte {
	if len(args) != 1 {
		return Encode(errors.New("ERR invalid number of arguments for `STACKINTLEN` command"), false)
	}

	obj := Get(args[0])
	if obj == nil {
		return RESP_ZERO
	}

	if err := assertType(obj.TypeEncoding, OBJ_TYPE_BYTELIST); err != nil {
		return Encode(err, false)
	}

	if err := assertEncoding(obj.TypeEncoding, OBJ_ENCODING_STACKINT); err != nil {
		return Encode(err, false)
	}

	s := obj.Value.(*StackInt)
	return Encode(s.Length, false)
}

// evalQINTPEEK peeks into the QINT and returns 5 elements without popping them
// returns the array of integers as the response.
// if the key does not exist, then we return an empty array
func evalQINTPEEK(args []string) []byte {
	var num int64 = 5
	var err error

	if len(args) > 2 {
		return Encode(errors.New("ERR invalid number of arguments for `QINTPEEK` command"), false)
	}

	if len(args) == 2 {
		num, err = strconv.ParseInt(args[1], 10, 32)
		if err != nil || num <= 0 || num > 100 {
			return Encode(errors.New("ERR number of elements to peek should be a positive number less than 100"), false)
		}
	}

	obj := Get(args[0])
	if obj == nil {
		return RESP_EMPTY_ARRAY
	}

	if err := assertType(obj.TypeEncoding, OBJ_TYPE_BYTELIST); err != nil {
		return Encode(err, false)
	}

	if err := assertEncoding(obj.TypeEncoding, OBJ_ENCODING_QINT); err != nil {
		return Encode(err, false)
	}

	q := obj.Value.(*QueueInt)
	return Encode(q.Iterate(int(num)), false)
}

// evalSTACKINTPEEK peeks into the DINT and returns 5 elements without popping them
// returns the array of integers as the response.
// if the key does not exist, then we return an empty array
func evalSTACKINTPEEK(args []string) []byte {
	var num int64 = 5
	var err error

	if len(args) > 2 {
		return Encode(errors.New("ERR invalid number of arguments for `STACKINTPEEK` command"), false)
	}

	if len(args) == 2 {
		num, err = strconv.ParseInt(args[1], 10, 32)
		if err != nil || num <= 0 || num > 100 {
			return Encode(errors.New("ERR number of elements to peek should be a positive number less than 100"), false)
		}
	}

	obj := Get(args[0])
	if obj == nil {
		return RESP_EMPTY_ARRAY
	}

	if err := assertType(obj.TypeEncoding, OBJ_TYPE_BYTELIST); err != nil {
		return Encode(err, false)
	}

	if err := assertEncoding(obj.TypeEncoding, OBJ_ENCODING_STACKINT); err != nil {
		return Encode(err, false)
	}

	s := obj.Value.(*StackInt)
	return Encode(s.Iterate(int(num)), false)
}

// evalQREFINS inserts the reference of the provided key identified by key
// first argument will be the key, that should be of type `QREF`
// second argument will be the key that needs to be added to the queueref
// if the queue does not exist, evalQREFINS will also create the queueref
// returns 1 if the key reference was inserted
// returns 0 otherwise
func evalQREFINS(args []string) []byte {
	if len(args) != 2 {
		return Encode(errors.New("ERR invalid number of arguments for `QREFINS` command"), false)
	}

	obj := Get(args[0])
	if obj == nil {
		obj = NewObj(NewQueueRef(), -1, OBJ_TYPE_BYTELIST, OBJ_ENCODING_QREF)
	}

	if err := assertType(obj.TypeEncoding, OBJ_TYPE_BYTELIST); err != nil {
		return Encode(err, false)
	}

	if err := assertEncoding(obj.TypeEncoding, OBJ_ENCODING_QREF); err != nil {
		return Encode(err, false)
	}

	Put(args[0], obj)

	q := obj.Value.(*QueueRef)
	if q.Insert(args[1]) {
		return Encode(1, false)
	}
	return Encode(0, false)
}

// evalSTACKREFPUSH inserts the reference of the provided key identified by key
// first argument will be the key, that should be of type `STACKREF`
// second argument will be the key that needs to be added to the stackref
// if the stack does not exist, evalSTACKREFPUSH will also create the stackref
// returns 1 if the key reference was inserted
// returns 0 otherwise
func evalSTACKREFPUSH(args []string) []byte {
	if len(args) != 2 {
		return Encode(errors.New("ERR invalid number of arguments for `STACKREFPUSH` command"), false)
	}

	obj := Get(args[0])
	if obj == nil {
		obj = NewObj(NewStackRef(), -1, OBJ_TYPE_BYTELIST, OBJ_ENCODING_STACKREF)
	}

	if err := assertType(obj.TypeEncoding, OBJ_TYPE_BYTELIST); err != nil {
		return Encode(err, false)
	}

	if err := assertEncoding(obj.TypeEncoding, OBJ_ENCODING_STACKREF); err != nil {
		return Encode(err, false)
	}

	Put(args[0], obj)

	s := obj.Value.(*StackRef)
	if s.Push(args[1]) {
		return Encode(1, false)
	}
	return Encode(0, false)
}

// evalQREFREM removes the element from the QREF identified by key
// first argument will be the key, that should be of type `QREF`
// if the key does not exist, evalQREFREM returns nil otherwise it
// returns the RESP encoded value of the key reference from the queue
// if we remove from the empty queue, nil is returned
func evalQREFREM(args []string) []byte {
	if len(args) != 1 {
		return Encode(errors.New("ERR invalid number of arguments for `QREFREM` command"), false)
	}

	obj := Get(args[0])
	if obj == nil {
		return RESP_NIL
	}

	if err := assertType(obj.TypeEncoding, OBJ_TYPE_BYTELIST); err != nil {
		return Encode(err, false)
	}

	if err := assertEncoding(obj.TypeEncoding, OBJ_ENCODING_QREF); err != nil {
		return Encode(err, false)
	}

	q := obj.Value.(*QueueRef)
	x, err := q.Remove()

	if err == ErrQueueEmpty {
		return RESP_NIL
	}

	return Encode(x, false)
}

// evalSTACKREFPOP removes the element from the DREF identified by key
// first argument will be the key, that should be of type `STACKREF`
// if the key does not exist, evalSTACKREFPOP returns nil otherwise it
// returns the RESP encoded value of the key reference from the stack
// if we remove from the empty stack, nil is returned
func evalSTACKREFPOP(args []string) []byte {
	if len(args) != 1 {
		return Encode(errors.New("ERR invalid number of arguments for `STACKREFPOP` command"), false)
	}

	obj := Get(args[0])
	if obj == nil {
		return RESP_NIL
	}

	if err := assertType(obj.TypeEncoding, OBJ_TYPE_BYTELIST); err != nil {
		return Encode(err, false)
	}

	if err := assertEncoding(obj.TypeEncoding, OBJ_ENCODING_STACKREF); err != nil {
		return Encode(err, false)
	}

	s := obj.Value.(*StackRef)
	x, err := s.Pop()

	if err == ErrStackEmpty {
		return RESP_NIL
	}

	return Encode(x, false)
}

// evalQREFLEN returns the length of the QREF identified by key
// returns the integer value indicating the length of the queue
// if the key does not exist, the response is 0
func evalQREFLEN(args []string) []byte {
	if len(args) != 1 {
		return Encode(errors.New("ERR invalid number of arguments for `QREFLEN` command"), false)
	}

	obj := Get(args[0])
	if obj == nil {
		return RESP_ZERO
	}

	if err := assertType(obj.TypeEncoding, OBJ_TYPE_BYTELIST); err != nil {
		return Encode(err, false)
	}

	if err := assertEncoding(obj.TypeEncoding, OBJ_ENCODING_QREF); err != nil {
		return Encode(err, false)
	}

	q := obj.Value.(*QueueRef)
	return Encode(q.qi.Length, false)
}

// evalSTACKREFLEN returns the length of the STACKREF identified by key
// returns the integer value indicating the length of the stack
// if the key does not exist, the response is 0
func evalSTACKREFLEN(args []string) []byte {
	if len(args) != 1 {
		return Encode(errors.New("ERR invalid number of arguments for `STACKREFLEN` command"), false)
	}

	obj := Get(args[0])
	if obj == nil {
		return RESP_ZERO
	}

	if err := assertType(obj.TypeEncoding, OBJ_TYPE_BYTELIST); err != nil {
		return Encode(err, false)
	}

	if err := assertEncoding(obj.TypeEncoding, OBJ_ENCODING_STACKREF); err != nil {
		return Encode(err, false)
	}

	s := obj.Value.(*StackRef)
	return Encode(s.si.Length, false)
}

// evalQREFPEEK peeks into the QREF and returns 5 elements without popping them
// returns the array of resp encoded values as the response.
// if the key does not exist, then we return an empty array
func evalQREFPEEK(args []string) []byte {
	var num int64 = 5
	var err error

	if len(args) == 0 {
		return Encode(errors.New("ERR invalid number of arguments for `QREFPEEK` command"), false)
	}

	if len(args) == 2 {
		num, err = strconv.ParseInt(args[1], 10, 32)
		if err != nil || num <= 0 || num > 100 {
			return Encode(errors.New("ERR number of elements to peek should be a positive number less than 100"), false)
		}
	}

	obj := Get(args[0])
	if obj == nil {
		return RESP_EMPTY_ARRAY
	}

	if err := assertType(obj.TypeEncoding, OBJ_TYPE_BYTELIST); err != nil {
		return Encode(err, false)
	}

	if err := assertEncoding(obj.TypeEncoding, OBJ_ENCODING_QREF); err != nil {
		return Encode(err, false)
	}

	q := obj.Value.(*QueueRef)
	return Encode(q.Iterate(int(num)), false)
}

// evalSTACKREFPEEK peeks into the STACKREF and returns 5 elements without popping them
// returns the array of resp encoded values as the response.
// if the key does not exist, then we return an empty array
func evalSTACKREFPEEK(args []string) []byte {
	var num int64 = 5
	var err error

	if len(args) == 0 {
		return Encode(errors.New("ERR invalid number of arguments for `STACKREFPEEK` command"), false)
	}

	if len(args) == 2 {
		num, err = strconv.ParseInt(args[1], 10, 32)
		if err != nil || num <= 0 || num > 100 {
			return Encode(errors.New("ERR number of elements to peek should be a positive number less than 100"), false)
		}
	}

	obj := Get(args[0])
	if obj == nil {
		return RESP_EMPTY_ARRAY
	}

	if err := assertType(obj.TypeEncoding, OBJ_TYPE_BYTELIST); err != nil {
		return Encode(err, false)
	}

	if err := assertEncoding(obj.TypeEncoding, OBJ_ENCODING_STACKREF); err != nil {
		return Encode(err, false)
	}

	s := obj.Value.(*StackRef)
	return Encode(s.Iterate(int(num)), false)
}

// evalQWATCH adds the specified key to the watch list for the caller client.
// Every time a key in the watch list is modified, the client will be sent a response
// containing the new value of the key along with the operation that was performed on it.
// Contains only one argument, the key to be watched.
func evalQWATCH(args []string, c *Client) []byte {
	if len(args) != 1 {
		return Encode(errors.New("ERR invalid number of arguments for `QWATCH` command (expected 1)"), false)
	}

	// Parse and get the selection from the query.
	query, error := ParseQuery( /*sql=*/ args[0])

	if error != nil {
		return Encode(error, false)
	}

	AddWatcher(query, c.Fd)

	// Return the result of the query.
	result, err := ExecuteQuery(query)
	if err != nil {
		return Encode(err, false)
	}

	return Encode(result, false)
}

// SETBIT key offset value
func evalSETBIT(args []string) []byte {
	var err error

	if len(args) != 3 {
		return Encode(errors.New("ERR wrong number of arguments for 'setbit' command"), false)
	}

	key := args[0]
	offset, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return Encode(errors.New("ERR bit offset is not an integer or out of range"), false)
	}

	value, err := strconv.ParseBool(args[2])
	if err != nil {
		return Encode(errors.New("ERR bit is not an integer or out of range"), false)
	}

	obj := Get(key)
	requiredByteArraySize := offset/8 + 1

	if obj == nil {
		obj = NewObj(NewByteArray(int(requiredByteArraySize)), -1, OBJ_TYPE_BYTEARRAY, OBJ_ENCODING_BYTEARRAY)
		Put(args[0], obj)
	}

	// handle the case when it is string
	if assertType(obj.TypeEncoding, OBJ_TYPE_STRING) == nil {
		return Encode(errors.New("ERR value is not a valid byte array"), false)
	}

	// handle the case when it is byte array
	if assertType(obj.TypeEncoding, OBJ_TYPE_BYTEARRAY) == nil {
		byteArray := obj.Value.(*ByteArray)
		byteArrayLength := byteArray.Length

		// check whether resize required or not
		if requiredByteArraySize > byteArrayLength {
			// resize as per the offset
			byteArray = byteArray.IncreaseSize(int(requiredByteArraySize))
		}

		response := byteArray.GetBit(int(offset))
		byteArray.SetBit(int(offset), value)

		// if earlier bit was 1 and the new bit is 0
		// propability is that, we can remove some space from the byte array
		if response && !value {
			byteArray.ResizeIfNecessary()
		}

		if response {
			return Encode(int(1), true)
		}
		return Encode(int(0), true)
	}

	return Encode(0, false)
}

// GETBIT key offset
func evalGETBIT(args []string) []byte {
	var err error

	if len(args) != 2 {
		return Encode(errors.New("ERR wrong number of arguments for 'setbit' command"), false)
	}

	key := args[0]
	offset, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return Encode(errors.New("ERR bit offset is not an integer or out of range"), false)
	}

	obj := Get(key)
	if obj == nil {
		return Encode(0, true)
	}

	requiredByteArraySize := offset/8 + 1

	// handle the case when it is string
	if assertType(obj.TypeEncoding, OBJ_TYPE_STRING) == nil {
		return Encode(errors.New("ERR value is not a valid byte array"), false)
	}

	// handle the case when it is byte array
	if assertType(obj.TypeEncoding, OBJ_TYPE_BYTEARRAY) == nil {
		byteArray := obj.Value.(*ByteArray)
		byteArrayLength := byteArray.Length

		// check whether offset, length exists or not
		if requiredByteArraySize > byteArrayLength {
			return Encode(0, true)
		} else {
			value := byteArray.GetBit(int(offset))
			if value {
				return Encode(1, true)
			}
			return Encode(0, true)
		}
	}

	return Encode(0, true)
}

func evalBITCOUNT(args []string) []byte {
	var err error

	// if more than 4 arguments are provided, return error
	if len(args) > 4 {
		return Encode(errors.New("ERR syntax error"), false)
	}

	// fetching value of the key
	var key string = args[0]
	var obj = Get(key)
	if obj == nil {
		return Encode(0, false)
	}

	var valueInterface = obj.Value
	value := []byte{}
	valueLength := int64(0)

	if assertType(obj.TypeEncoding, OBJ_TYPE_BYTEARRAY) == nil {
		byteArray := obj.Value.(*ByteArray)
		byteArrayObject := *byteArray
		value = byteArrayObject.data
		valueLength = byteArray.Length
	}

	if assertType(obj.TypeEncoding, OBJ_TYPE_STRING) == nil {
		value = []byte(valueInterface.(string))
		valueLength = int64(len(value))
	}

	// defining constants of the function
	start := int64(0)
	end := valueLength - 1
	var unit = bit.BYTE

	// checking which arguments are present and according validating arguments
	if len(args) > 1 {
		start, err = strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			return Encode(errors.New("ERR value is not an integer or out of range"), false)
		}
		// Adjust start index if it is negative
		if start < 0 {
			start += valueLength
		}
	}
	if len(args) > 2 {
		end, err = strconv.ParseInt(args[2], 10, 64)
		if err != nil {
			return Encode(errors.New("ERR value is not an integer or out of range"), false)
		}

		// Adjust end index if it is negative
		if end < 0 {
			end += valueLength
		}
	}
	if len(args) > 3 {
		unit = strings.ToUpper(args[3])
		if unit != bit.BYTE && unit != bit.BIT {
			return Encode(errors.New("ERR syntax error"), false)
		}
	}
	if start > end {
		return Encode(0, true)
	}
	if start > valueLength && unit == bit.BYTE {
		return Encode(0, true)
	}
	if end > valueLength && unit == bit.BYTE {
		end = valueLength - 1
	}

	bitCount := 0
	if unit == bit.BYTE {
		for i := start; i <= end; i++ {
			bitCount += int(popcount(value[i]))
		}
		return Encode(bitCount, true)
	} else {
		startBitRange := start / 8
		endBitRange := end / 8

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
		return Encode(bitCount, true)
	}
}

// BITOP <AND | OR | XOR | NOT> destkey key [key ...]
func evalBITOP(args []string) []byte {
	operation, destKey := args[0], args[1]
	operation = strings.ToUpper(operation)

	// get all the keys
	keys := args[2:]

	// validation of commands
	// if operation is not from enums, then error out
	if !(operation == "AND" || operation == "OR" || operation == "XOR" || operation == "NOT") {
		return Encode(errors.New("ERR syntax error"), false)
	}
	// if operation is not, then keys lenght should be only 1
	if operation == "NOT" && len(keys) != 1 {
		return Encode(errors.New("ERR BITOP NOT must be called with a single source key."), false)
	}

	if operation == "NOT" {
		obj := Get(keys[0])
		if obj == nil {
			return Encode(0, true)
		}

		var value []byte
		if assertType(obj.TypeEncoding, OBJ_TYPE_BYTEARRAY) == nil {
			byteArray := obj.Value.(*ByteArray)
			byteArrayObject := *byteArray
			value = byteArrayObject.data
		} else {
			return Encode(errors.New("ERR value is not a valid byte array"), false)
		}

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
		obj = NewObj(operationResult, -1, OBJ_TYPE_BYTEARRAY, OBJ_ENCODING_BYTEARRAY)

		// store the result in destKey
		Put(destKey, obj)
		return Encode(len(value), true)
	} else {
		// if operation is AND, OR, XOR
		values := make([][]byte, len(keys))

		// get the values of all keys
		for i, key := range keys {
			obj := Get(key)
			if obj == nil {
				values[i] = make([]byte, 0)
			} else {
				// handle the case when it is byte array
				if assertType(obj.TypeEncoding, OBJ_TYPE_BYTEARRAY) == nil {
					byteArray := obj.Value.(*ByteArray)
					byteArrayObject := *byteArray
					values[i] = byteArrayObject.data
				} else {
					return Encode(errors.New("ERR value is not a valid byte array"), false)
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
			if len(value) < minLength {
				minLength = len(value)
			}
		}

		result := make([]byte, maxLength)
		if operation == "AND" {
			for i := 0; i < maxLength; i++ {
				if i < minLength {
					result[i] = values[maxKeyIterator][i]
				} else {
					result[i] = 0
				}
			}
		}
		if operation == "XOR" || operation == "OR" {
			for i := 0; i < maxLength; i++ {
				result[i] = 0x00
			}
		}

		// perform the operation
		for _, value := range values {
			for i := 0; i < len(value); i++ {
				if operation == "AND" {
					result[i] &= value[i]
				} else if operation == "OR" {
					result[i] |= value[i]
				} else if operation == "XOR" {
					result[i] ^= value[i]
				}
			}
		}

		// initialize result with byteArray
		operationResult := NewByteArray(len(result))
		operationResult.data = result
		operationResult.Length = int64(len(result))

		// resize the byte array if necessary
		operationResult.ResizeIfNecessary()

		// create object related to result
		operationResultObject := NewObj(operationResult, -1, OBJ_TYPE_BYTEARRAY, OBJ_ENCODING_BYTEARRAY)

		// store the result in destKey
		Put(destKey, operationResultObject)

		return Encode(len(result), true)
	}
}

// evalCommand evaluates COMMAND <subcommand> command based on subcommand
// COUNT: return total count of commands in Dice.
func evalCommand(args []string) []byte {
	if len(args) == 0 {
		return Encode(errors.New("(error) ERR wrong number of arguments for 'command' command"), false)
	}
	subcommand := strings.ToUpper(args[0])
	switch subcommand {
	case "COUNT":
		return evalCommandCount(nil)
	case "GETKEYS":
		return evalCommandGetKeys(args[1:])
	default:
		return Encode(fmt.Errorf("ERR unknown subcommand '%s'. Try COMMAND HELP", subcommand), false)
	}
}

// evalKeys returns the list of keys that match the pattern
// The pattern should be the only param in args
func evalKeys(args []string) []byte {
	if len(args) != 1 {
		return Encode(errors.New("ERR wrong number of arguments for 'keys' command"), false)
	}

	pattern := args[0]
	keys, err := Keys(pattern)
	if err != nil {
		return Encode(err, false)
	}

	return Encode(keys, false)
}

// evalCommandCount returns an number of commands supported by DiceDB
func evalCommandCount(args []string) []byte {
	return Encode(diceCommandsCount, false)
}

func evalCommandGetKeys(args []string) []byte {
	diceCmd, ok := diceCmds[strings.ToUpper(args[0])]
	if !ok {
		return Encode(errors.New("ERR invalid command specified"), false)
	}

	keySpecs := diceCmd.KeySpecs
	if keySpecs.BeginIndex == 0 {
		return Encode(errors.New("ERR the command has no key arguments"), false)
	}

	arity := diceCmd.Arity
	if (arity < 0 && len(args) < -arity) ||
		(arity >= 0 && len(args) != arity) {
		return Encode(errors.New("ERR invalid number of arguments specified for command"), false)
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

	return Encode(keys, false)
}
func evalRename(args []string) []byte {

	if len(args) != 2 {
		return Encode(errors.New("ERR wrong number of arguments for 'RENAME' command"), false)
	}
	sourceKey := args[0]
	destKey := args[1]

	//if Source and Destination Keys are same return RESP encoded ok
	if sourceKey == destKey {
		return RESP_OK
	}

	// if Source key does not exist, return RESP encoded nil
	sourceObj := Get(sourceKey)
	if sourceObj == nil {
		return Encode("ERR no such key", false)
	}

	if ok := Rename(sourceKey, destKey); ok {
		return RESP_OK
	}
	return RESP_NIL

}

func executeCommand(cmd *RedisCmd, c *Client) []byte {
	diceCmd, ok := diceCmds[cmd.Cmd]
	if !ok {
		return Encode(fmt.Errorf("ERR unknown command '%s', with args beginning with: %s", cmd.Cmd, strings.Join(cmd.Args, " ")), false)
	}

	if diceCmd.Name == "SUBSCRIBE" || diceCmd.Name == "QWATCH" {
		return evalQWATCH(cmd.Args, c)
	}
	if diceCmd.Name == "MULTI" {
		c.TxnBegin()
		return diceCmd.Eval(cmd.Args)
	}
	if diceCmd.Name == AuthCmd {
		return evalAUTH(cmd.Args, c)
	}
	if diceCmd.Name == "EXEC" {
		if !c.isTxn {
			return Encode(errors.New("ERR EXEC without MULTI"), false)
		}
		return c.TxnExec()
	}
	if diceCmd.Name == "DISCARD" {
		if !c.isTxn {
			return Encode(errors.New("ERR DISCARD without MULTI"), false)
		}
		c.TxnDiscard()
		return RESP_OK
	}
	if diceCmd.Name == "ABORT" {
		return RESP_OK
	}

	return diceCmd.Eval(cmd.Args)
}

func executeCommandToBuffer(cmd *RedisCmd, buf *bytes.Buffer, c *Client) {
	buf.Write(executeCommand(cmd, c))
}

func EvalAndRespond(cmds RedisCmds, c *Client) {
	var response []byte
	buf := bytes.NewBuffer(response)

	for _, cmd := range cmds {
		// Check if the command has been authenticated
		if cmd.Cmd != AuthCmd && !c.Session.IsActive() {
			if _, err := c.Write(RESP_AUTH_FAILURE); err != nil {
				log.Println("Error writing to client:", err)
			}
			continue
		}
		// if txn is not in progress, then we can simply
		// execute the command and add the response to the buffer
		if !c.isTxn {
			executeCommandToBuffer(cmd, buf, c)
			continue
		}

		// if the txn is in progress, we enqueue the command
		// and add the QUEUED response to the buffer
		if !txnCommands[cmd.Cmd] {
			// if the command is queuable the enqueu
			c.TxnQueue(cmd)
			buf.Write(RESP_QUEUED)
		} else {
			// if txn is active and the command is non-queuable
			// ex: EXEC, DISCARD
			// we execute the command and gather the response in buffer
			executeCommandToBuffer(cmd, buf, c)
		}
	}

	if _, err := c.Write(buf.Bytes()); err != nil {
		log.Panic(err)
	}
}

func evalPersist(args []string) []byte {
	if len(args) != 1 {
		return Encode(errors.New("ERR wrong number of arguments for 'persist' command"), false)
	}

	key := args[0]

	obj := Get(key)

	// If the key does not exist, return RESP encoded 0 to denote the key does not exist
	if obj == nil {
		return RESP_ZERO
	}

	// If the object exists but no expiration is set on it, return -1
	_, isExpirySet := getExpiry(obj)
	if !isExpirySet {
		return RESP_MINUS_1
	}

	// If the object exists, remove the expiration time
	delExpiry(obj)

	return RESP_ONE
}

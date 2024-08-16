package core

import (
	"bytes"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"syscall"
	"time"

	DiceErrors "github.com/dicedb/dice/core/diceerrors"

	"github.com/bytedance/sonic"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/core/bit"
	"github.com/dicedb/dice/internal/constants"
	"github.com/ohler55/ojg/jp"
)

type exDurationState int

const (
	Uninitialized exDurationState = iota
	Initialized
)

var RespNIL []byte = []byte("$-1\r\n")
var RespOK []byte = []byte("+OK\r\n")
var RespAuthFailure []byte = []byte("+NOAUTH\r\n")
var RespQueued []byte = []byte("+QUEUED\r\n")
var RespZero []byte = []byte(":0\r\n")
var RespOne []byte = []byte(":1\r\n")
var RespMinusOne []byte = []byte(":-1\r\n")
var RespMinusTwo []byte = []byte(":-2\r\n")
var RespEmptyArray []byte = []byte("*0\r\n")

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
		return DiceErrors.NewErrArity("PING")
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
		err error
	)

	if config.RequirePass == "" {
		return DiceErrors.NewErrWithMessage("AUTH <password> called without any password configured for the default user. Are you sure your configuration is correct?")
	}

	var username = DefaultUserName
	var password string

	if len(args) == 1 {
		password = args[0]
	} else if len(args) == 2 {
		username, password = args[0], args[1]
	} else {
		return DiceErrors.NewErrArity("AUTH")
	}

	if err = c.Session.Validate(username, password); err != nil {
		return Encode(err, false)
	}
	return RespOK
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
		return DiceErrors.NewErrArity("SET")
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
		case constants.Ex, constants.Px:
			if state != Uninitialized {
				return DiceErrors.NewErrWithMessage(DiceErrors.SyntaxErr)
			}
			i++
			if i == len(args) {
				return DiceErrors.NewErrWithMessage(DiceErrors.SyntaxErr)
			}

			exDuration, err := strconv.ParseInt(args[i], 10, 64)
			if err != nil {
				return DiceErrors.NewErrWithMessage(DiceErrors.IntOrOutOfRangeErr)
			}

			if exDuration <= 0 {
				return DiceErrors.NewErrExpireTime("SET")
			}

			// converting seconds to milliseconds
			if arg == constants.Ex {
				exDuration *= 1000
			}
			exDurationMs = exDuration
			state = Initialized

		case constants.Pxat, constants.Exat:
			if state != Uninitialized {
				return DiceErrors.NewErrWithMessage(DiceErrors.SyntaxErr)
			}
			i++
			if i == len(args) {
				return DiceErrors.NewErrWithMessage(DiceErrors.SyntaxErr)
			}
			exDuration, err := strconv.ParseInt(args[i], 10, 64)
			if err != nil {
				return DiceErrors.NewErrWithMessage(DiceErrors.IntOrOutOfRangeErr)
			}

			if exDuration < 0 {
				return DiceErrors.NewErrExpireTime("SET")
			}

			if arg == constants.Exat {
				exDuration *= 1000
			}
			exDurationMs = exDuration - time.Now().UnixMilli()
			// If the expiry time is in the past, set exDurationMs to 0
			// This will be used to signal immediate expiration
			if exDurationMs < 0 {
				exDurationMs = 0
			}
			state = Initialized

		case constants.XX:
			// Get the key from the hash table
			obj := Get(key)

			// if key does not exist, return RESP encoded nil
			if obj == nil {
				return RespNIL
			}
		case constants.NX:
			obj := Get(key)
			if obj != nil {
				return RespNIL
			}
		case constants.KEEPTTL, constants.Keepttl:
			keepttl = true
		default:
			return DiceErrors.NewErrWithMessage(DiceErrors.SyntaxErr)
		}
	}

	// putting the k and value in a Hash Table
	Put(key, NewObj(value, exDurationMs, oType, oEnc), WithKeepTTL(keepttl))

	return RespOK
}

// evalMSET puts multiple <key, value> pairs in db as in the args
// MSET is atomic, so all given keys are set at once.
// args must contain key and value pairs.

// Returns encoded error response if at least a <key, value> pair is not part of args
// Returns encoded OK RESP once new entries are added
// If the key already exists then the value will be overwritten and expiry will be discarded
func evalMSET(args []string) []byte {
	if len(args) <= 1 || len(args)%2 != 0 {
		return DiceErrors.NewErrArity("MSET")
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
	return RespOK
}

// evalGET returns the value for the queried key in args
// The key should be the only param in args
// The RESP value of the key is encoded and then returned
// evalGET returns RespNIL if key is expired or it does not exist
func evalGET(args []string) []byte {
	if len(args) != 1 {
		return DiceErrors.NewErrArity("GET")
	}

	var key string = args[0]

	// Get the key from the hash table
	obj := Get(key)

	// if key does not exist, return RESP encoded nil
	if obj == nil {
		return RespNIL
	}

	// return the RESP encoded value
	return Encode(obj.Value, false)
}

// evalGETDEL returns the value for the queried key in args
// The key should be the only param in args
// The RESP value of the key is encoded and then returned
// In evalGETDEL  If the key exists, it will be deleted before its value is returned.
// evalGETDEL returns RespNIL if key is expired or it does not exist
func evalGETDEL(args []string) []byte {
	if len(args) != 1 {
		return DiceErrors.NewErrArity("GETDEL")
	}

	var key = args[0]

	// Get the key from the hash table
	obj := GetDel(key)

	// if key does not exist, return RESP encoded nil
	if obj == nil {
		return RespNIL
	}

	// return the RESP encoded value
	return Encode(obj.Value, false)
}

// evalJSONGET retrieves a JSON value stored at the specified key
// args must contain at least the key;  (path unused in this implementation)
// Returns RespNIL if key is expired or it does not exist
// Returns encoded error response if incorrect number of arguments
// The RESP value of the key is encoded and then returned
func evalJSONGET(args []string) []byte {
	if len(args) < 1 {
		return DiceErrors.NewErrArity("JSON.GET")
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
		return RespNIL
	}

	// Check if the object is of JSON type
	err := assertType(obj.TypeEncoding, ObjTypeJSON)
	if err != nil {
		return Encode(err, false)
	}
	err = assertEncoding(obj.TypeEncoding, ObjEncodingJSON)
	if err != nil {
		return Encode(err, false)
	}

	jsonData := obj.Value

	// If path is root, return the entire JSON
	if path == defaultRootPath {
		resultBytes, err := sonic.Marshal(jsonData)
		if err != nil {
			return DiceErrors.NewErrWithMessage("could not serialize result")
		}
		return Encode(string(resultBytes), false)
	}

	// Parse the JSONPath expression
	expr, err := jp.ParseString(path)
	if err != nil {
		return DiceErrors.NewErrWithMessage("invalid JSONPath")
	}

	// Execute the JSONPath query
	results := expr.Get(jsonData)
	if len(results) == 0 {
		return RespNIL
	}

	// Serialize the result
	var resultBytes []byte
	if len(results) == 1 {
		resultBytes, err = sonic.Marshal(results[0])
	} else {
		resultBytes, err = sonic.Marshal(results)
	}
	if err != nil {
		return DiceErrors.NewErrWithMessage("could not serialize result")
	}
	return Encode(string(resultBytes), false)
}

// evalJSONSET stores a JSON value at the specified key
// args must contain at least the key, path (unused in this implementation), and JSON string
// Returns encoded error response if incorrect number of arguments
// Returns encoded error if the JSON string is invalid
// Returns RespOK if the JSON value is successfully stored
func evalJSONSET(args []string) []byte {
	// Check if there are enough arguments
	if len(args) < 3 {
		return DiceErrors.NewErrArity("JSON.SET")
	}

	key := args[0]
	path := args[1]
	jsonStr := args[2]

	for i := 3; i < len(args); i++ {
		switch args[i] {
		case constants.NX, constants.Nx:
			if i != len(args)-1 {
				return DiceErrors.NewErrWithMessage(DiceErrors.SyntaxErr)
			}
			obj := Get(key)
			if obj != nil {
				return RespNIL
			}
		case constants.XX, constants.Xx:
			if i != len(args)-1 {
				return DiceErrors.NewErrWithMessage(DiceErrors.SyntaxErr)
			}
			obj := Get(key)
			if obj == nil {
				return RespNIL
			}

		default:
			return DiceErrors.NewErrWithMessage(DiceErrors.SyntaxErr)
		}
	}

	// Parse the JSON string
	var jsonValue interface{}
	if err := sonic.UnmarshalString(jsonStr, &jsonValue); err != nil {
		return DiceErrors.NewErrWithFormattedMessage("invalid JSON: %v", err.Error())
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
		err := assertType(obj.TypeEncoding, ObjTypeJSON)
		if err != nil {
			return Encode(err, false)
		}
		err = assertEncoding(obj.TypeEncoding, ObjEncodingJSON)
		if err != nil {
			return Encode(err, false)
		}
		rootData = obj.Value
	}

	// If path is not root, use JSONPath to set the value
	if path != defaultRootPath {
		expr, err := jp.ParseString(path)
		if err != nil {
			return DiceErrors.NewErrWithMessage("invalid JSONPath")
		}

		err = expr.Set(rootData, jsonValue)
		if err != nil {
			return DiceErrors.NewErrWithMessage("failed to set value")
		}
	} else {
		// If path is root, replace the entire JSON
		rootData = jsonValue
	}

	// Create a new object with the updated JSON data
	newObj := NewObj(rootData, -1, ObjTypeJSON, ObjEncodingJSON)
	Put(key, newObj)

	return RespOK
}

// evalTTL returns Time-to-Live in secs for the queried key in args
// The key should be the only param in args else returns with an error
// Returns	RESP encoded time (in secs) remaining for the key to expire
//
//	RESP encoded -2 stating key doesn't exist or key is expired
//	RESP encoded -1 in case no expiration is set on the key
func evalTTL(args []string) []byte {
	if len(args) != 1 {
		return DiceErrors.NewErrArity("TTL")
	}

	var key string = args[0]

	obj := Get(key)

	// if key does not exist, return RESP encoded -2 denoting key does not exist
	if obj == nil {
		return RespMinusTwo
	}

	// if object exist, but no expiration is set on it then send -1
	exp, isExpirySet := getExpiry(obj)
	if !isExpirySet {
		return RespMinusOne
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
// Returns RespOne if expiry was set on the key successfully.
// Once the time is lapsed, the key will be deleted automatically
func evalEXPIRE(args []string) []byte {
	if len(args) <= 1 {
		return DiceErrors.NewErrArity("EXPIRE")
	}

	var key string = args[0]
	exDurationSec, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return DiceErrors.NewErrWithMessage(DiceErrors.IntOrOutOfRangeErr)
	}

	obj := Get(key)

	// 0 if the timeout was not set. e.g. key doesn't exist, or operation skipped due to the provided arguments
	if obj == nil {
		return RespZero
	}

	setExpiry(obj, exDurationSec*1000)

	// 1 if the timeout was set.
	return RespOne
}

func evalHELLO(args []string) []byte {
	if len(args) > 1 {
		return DiceErrors.NewErrArity("HELLO")
	}

	var response []interface{}
	response = append(response,
		"proto", 2,
		"id", serverID,
		"mode", "standalone",
		"role", "master",
		"modules", []interface{}{})

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
		// We are inside child process now, so we'll start flushing to disk.
		if err := DumpAllAOF(); err != nil {
			return DiceErrors.NewErrWithMessage("AOF failed")
		}
		return []byte(constants.EmptyStr)
	}
	// Back to main thread
	return RespOK
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
		return DiceErrors.NewErrArity("INCR")
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
		return DiceErrors.NewErrArity("DECR")
	}
	return incrDecrCmd(args, -1)
}

// evalDECRBY decrements the value of the specified key in args by the specified decrement,
// if the key exists and the value is integer format.
// The key should be the first parameter in args, and the decrement should be the second parameter.
// If the key does not exist, new key is created with value 0,
// the value of the new key is then decremented by specified decrement.
// The value for the queried key should be of integer format,
// if not evalDECRBY returns an encoded error response.
// evalDECRBY returns the decremented value for the key after applying the specified decrement if there are no errors.
func evalDECRBY(args []string) []byte {
	if len(args) != 2 {
		return DiceErrors.NewErrArity("DECRBY")
	}
	decrementAmount, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return DiceErrors.NewErrWithMessage(DiceErrors.IntOrOutOfRangeErr)
	}
	return incrDecrCmd(args, -decrementAmount)
}

func incrDecrCmd(args []string, incr int64) []byte {
	var key string = args[0]
	obj := Get(key)
	if obj == nil {
		obj = NewObj("0", -1, ObjTypeString, ObjEncodingInt)
		Put(key, obj)
	}

	if err := assertType(obj.TypeEncoding, ObjTypeString); err != nil {
		return DiceErrors.NewErrWithMessage(err.Error())
	}

	if err := assertEncoding(obj.TypeEncoding, ObjEncodingInt); err != nil {
		return DiceErrors.NewErrWithMessage(err.Error())
	}

	i, _ := strconv.ParseInt(obj.Value.(string), 10, 64)
	// check overflow
	if (incr < 0 && i < 0 && incr < (math.MinInt64-i)) ||
		(incr > 0 && i > 0 && incr > (math.MaxInt64-i)) {
		return DiceErrors.NewErrWithMessage(DiceErrors.ValOutOfRangeErr)
	}

	i += incr
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
		fmt.Fprintf(buf, "db%d:keys=%d,expires=0,avg_ttl=0\r\n", i, KeyspaceStat[i]["keys"])
	}
	return Encode(buf.String(), false)
}

// TODO: Placeholder to support monitoring
func evalCLIENT(args []string) []byte {
	return RespOK
}

// TODO: Placeholder to support monitoring
func evalLATENCY(args []string) []byte {
	return Encode([]string{}, false)
}

// evalLRU deletes all the keys from the LRU
// returns encoded RESP OK
func evalLRU(args []string) []byte {
	evictAllkeysLRU()
	return RespOK
}

// evalSLEEP sets db to sleep for the specified number of seconds.
// The sleep time should be the only param in args.
// Returns error response if the time param in args is not of integer format.
// evalSLEEP returns RespOK after sleeping for mentioned seconds
func evalSLEEP(args []string) []byte {
	if len(args) != 1 {
		return DiceErrors.NewErrArity("SLEEP")
	}

	durationSec, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return DiceErrors.NewErrWithMessage(DiceErrors.IntOrOutOfRangeErr)
	}
	time.Sleep(time.Duration(durationSec) * time.Second)
	return RespOK
}

// evalMULTI marks the start of the transaction for the client.
// All subsequent commands fired will be queued for atomic execution.
// The commands will not be executed until EXEC is triggered.
// Once EXEC is triggered it executes all the commands in queue,
// and closes the MULTI transaction.
func evalMULTI(args []string) []byte {
	return RespOK
}

// evalQINTINS inserts the provided integer in the key identified by key
// first argument will be the key, that should be of type `QINT`
// second argument will be the integer value
// if the key does not exist, evalQINTINS will also create the integer queue
func evalQINTINS(args []string) []byte {
	if len(args) != 2 {
		return DiceErrors.NewErrArity("QINTINS")
	}

	x, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return DiceErrors.NewErrWithMessage("only integer values can be inserted in QINT")
	}

	obj := Get(args[0])
	if obj == nil {
		obj = NewObj(NewQueueInt(), -1, ObjTypeByteList, ObjEncodingQint)
	}

	if err := assertType(obj.TypeEncoding, ObjTypeByteList); err != nil {
		return Encode(err, false)
	}

	if err := assertEncoding(obj.TypeEncoding, ObjEncodingQint); err != nil {
		return Encode(err, false)
	}

	Put(args[0], obj)

	q := obj.Value.(*QueueInt)
	q.Insert(x)

	return RespOK
}

// evalSTACKINTPUSH pushes the provided integer in the key identified by key
// first argument will be the key, that should be of type `STACKINT`
// second argument will be the integer value
// if the key does not exist, evalSTACKINTPUSH will also create the integer stack
func evalSTACKINTPUSH(args []string) []byte {
	if len(args) != 2 {
		return DiceErrors.NewErrArity("STACKINTPUSH")
	}

	x, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return DiceErrors.NewErrWithMessage(DiceErrors.IntOrOutOfRangeErr)
	}

	obj := Get(args[0])
	if obj == nil {
		obj = NewObj(NewStackInt(), -1, ObjTypeByteList, ObjEncodingStackInt)
	}

	if err := assertType(obj.TypeEncoding, ObjTypeByteList); err != nil {
		return Encode(err, false)
	}

	if err := assertEncoding(obj.TypeEncoding, ObjEncodingStackInt); err != nil {
		return Encode(err, false)
	}

	Put(args[0], obj)

	s := obj.Value.(*StackInt)
	s.Push(x)

	return RespOK
}

// evalQINTREM removes the element from the QINT identified by key
// first argument will be the key, that should be of type `QINT`
// if the key does not exist, evalQINTREM returns nil otherwise it
// returns the integer value popped from the queue
// if we remove from the empty queue, nil is returned
func evalQINTREM(args []string) []byte {
	if len(args) != 1 {
		return DiceErrors.NewErrArity("QINTREM")
	}

	obj := Get(args[0])
	if obj == nil {
		return RespNIL
	}

	if err := assertType(obj.TypeEncoding, ObjTypeByteList); err != nil {
		return Encode(err, false)
	}

	if err := assertEncoding(obj.TypeEncoding, ObjEncodingQint); err != nil {
		return Encode(err, false)
	}

	q := obj.Value.(*QueueInt)
	x, err := q.Remove()

	if err == ErrQueueEmpty {
		return RespNIL
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
		return DiceErrors.NewErrArity("STACKINTPOP")
	}

	obj := Get(args[0])
	if obj == nil {
		return RespNIL
	}

	if err := assertType(obj.TypeEncoding, ObjTypeByteList); err != nil {
		return Encode(err, false)
	}

	if err := assertEncoding(obj.TypeEncoding, ObjEncodingStackInt); err != nil {
		return Encode(err, false)
	}

	s := obj.Value.(*StackInt)
	x, err := s.Pop()

	if err == ErrStackEmpty {
		return RespNIL
	}

	return Encode(x, false)
}

// evalQINTLEN returns the length of the QINT identified by key
// returns the integer value indicating the length of the queue
// if the key does not exist, the response is 0
func evalQINTLEN(args []string) []byte {
	if len(args) != 1 {
		return DiceErrors.NewErrArity("QINTLEN")
	}

	obj := Get(args[0])
	if obj == nil {
		return RespZero
	}

	if err := assertType(obj.TypeEncoding, ObjTypeByteList); err != nil {
		return Encode(err, false)
	}

	if err := assertEncoding(obj.TypeEncoding, ObjEncodingQint); err != nil {
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
		return DiceErrors.NewErrArity("STACKINTLEN")
	}

	obj := Get(args[0])
	if obj == nil {
		return RespZero
	}

	if err := assertType(obj.TypeEncoding, ObjTypeByteList); err != nil {
		return Encode(err, false)
	}

	if err := assertEncoding(obj.TypeEncoding, ObjEncodingStackInt); err != nil {
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
		return DiceErrors.NewErrArity("QINTPEEK")
	}

	if len(args) == 2 {
		num, err = strconv.ParseInt(args[1], 10, 32)
		if err != nil || num <= 0 || num > 100 {
			return DiceErrors.NewErrWithFormattedMessage(DiceErrors.ElementPeekErr, 100)
		}
	}

	obj := Get(args[0])
	if obj == nil {
		return RespEmptyArray
	}

	if err := assertType(obj.TypeEncoding, ObjTypeByteList); err != nil {
		return Encode(err, false)
	}

	if err := assertEncoding(obj.TypeEncoding, ObjEncodingQint); err != nil {
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
		return DiceErrors.NewErrArity("STACKINTPEEK")
	}

	if len(args) == 2 {
		num, err = strconv.ParseInt(args[1], 10, 32)
		if err != nil || num <= 0 || num > 100 {
			return DiceErrors.NewErrWithFormattedMessage(DiceErrors.ElementPeekErr, 100)
		}
	}

	obj := Get(args[0])
	if obj == nil {
		return RespEmptyArray
	}

	if err := assertType(obj.TypeEncoding, ObjTypeByteList); err != nil {
		return Encode(err, false)
	}

	if err := assertEncoding(obj.TypeEncoding, ObjEncodingStackInt); err != nil {
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
		return DiceErrors.NewErrArity("QREFINS")
	}

	obj := Get(args[0])
	if obj == nil {
		obj = NewObj(NewQueueRef(), -1, ObjTypeByteList, ObjEncodingQref)
	}

	if err := assertType(obj.TypeEncoding, ObjTypeByteList); err != nil {
		return Encode(err, false)
	}

	if err := assertEncoding(obj.TypeEncoding, ObjEncodingQref); err != nil {
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
		return DiceErrors.NewErrArity("STACKREFPUSH")
	}

	obj := Get(args[0])
	if obj == nil {
		obj = NewObj(NewStackRef(), -1, ObjTypeByteList, ObjEncodingStackRef)
	}

	if err := assertType(obj.TypeEncoding, ObjTypeByteList); err != nil {
		return Encode(err, false)
	}

	if err := assertEncoding(obj.TypeEncoding, ObjEncodingStackRef); err != nil {
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
		return DiceErrors.NewErrArity("QREFREM")
	}

	obj := Get(args[0])
	if obj == nil {
		return RespNIL
	}

	if err := assertType(obj.TypeEncoding, ObjTypeByteList); err != nil {
		return Encode(err, false)
	}

	if err := assertEncoding(obj.TypeEncoding, ObjEncodingQref); err != nil {
		return Encode(err, false)
	}

	q := obj.Value.(*QueueRef)
	x, err := q.Remove()

	if err == ErrQueueEmpty {
		return RespNIL
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
		return DiceErrors.NewErrArity("STACKREFPOP")
	}

	obj := Get(args[0])
	if obj == nil {
		return RespNIL
	}

	if err := assertType(obj.TypeEncoding, ObjTypeByteList); err != nil {
		return Encode(err, false)
	}

	if err := assertEncoding(obj.TypeEncoding, ObjEncodingStackRef); err != nil {
		return Encode(err, false)
	}

	s := obj.Value.(*StackRef)
	x, err := s.Pop()

	if err == ErrStackEmpty {
		return RespNIL
	}

	return Encode(x, false)
}

// evalQREFLEN returns the length of the QREF identified by key
// returns the integer value indicating the length of the queue
// if the key does not exist, the response is 0
func evalQREFLEN(args []string) []byte {
	if len(args) != 1 {
		return DiceErrors.NewErrArity("QREFLEN")
	}

	obj := Get(args[0])
	if obj == nil {
		return RespZero
	}

	if err := assertType(obj.TypeEncoding, ObjTypeByteList); err != nil {
		return Encode(err, false)
	}

	if err := assertEncoding(obj.TypeEncoding, ObjEncodingQref); err != nil {
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
		return DiceErrors.NewErrArity("STACKREFLEN")
	}

	obj := Get(args[0])
	if obj == nil {
		return RespZero
	}

	if err := assertType(obj.TypeEncoding, ObjTypeByteList); err != nil {
		return Encode(err, false)
	}

	if err := assertEncoding(obj.TypeEncoding, ObjEncodingStackRef); err != nil {
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
		return DiceErrors.NewErrArity("QREFPEEK")
	}

	if len(args) == 2 {
		num, err = strconv.ParseInt(args[1], 10, 32)
		if err != nil || num <= 0 || num > 100 {
			return DiceErrors.NewErrWithFormattedMessage(DiceErrors.ElementPeekErr, 100)
		}
	}

	obj := Get(args[0])
	if obj == nil {
		return RespEmptyArray
	}

	if err := assertType(obj.TypeEncoding, ObjTypeByteList); err != nil {
		return Encode(err, false)
	}

	if err := assertEncoding(obj.TypeEncoding, ObjEncodingQref); err != nil {
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
		return DiceErrors.NewErrArity("STACKREFPEEK")
	}

	if len(args) == 2 {
		num, err = strconv.ParseInt(args[1], 10, 32)
		if err != nil || num <= 0 || num > 100 {
			return DiceErrors.NewErrWithFormattedMessage(DiceErrors.ElementPeekErr, 100)
		}
	}

	obj := Get(args[0])
	if obj == nil {
		return RespEmptyArray
	}

	if err := assertType(obj.TypeEncoding, ObjTypeByteList); err != nil {
		return Encode(err, false)
	}

	if err := assertEncoding(obj.TypeEncoding, ObjEncodingStackRef); err != nil {
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
		return DiceErrors.NewErrArity("QWATCH")
	}

	// Parse and get the selection from the query.
	query, e := ParseQuery( /*sql=*/ args[0])

	if e != nil {
		return Encode(e, false)
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
		return DiceErrors.NewErrArity("SETBIT")
	}

	key := args[0]
	offset, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return DiceErrors.NewErrWithMessage("bit offset is not an integer or out of range")
	}

	value, err := strconv.ParseBool(args[2])
	if err != nil {
		return DiceErrors.NewErrWithMessage("bit is not an integer or out of range")
	}

	obj := Get(key)
	requiredByteArraySize := offset/8 + 1

	if obj == nil {
		obj = NewObj(NewByteArray(int(requiredByteArraySize)), -1, ObjTypeByteArray, ObjEncodingByteArray)
		Put(args[0], obj)
	}

	// handle the case when it is string
	if assertType(obj.TypeEncoding, ObjTypeString) == nil {
		return DiceErrors.NewErrWithMessage("value is not a valid byte array")
	}

	// handle the case when it is byte array
	if assertType(obj.TypeEncoding, ObjTypeByteArray) == nil {
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
		return DiceErrors.NewErrArity("GETBIT")
	}

	key := args[0]
	offset, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return DiceErrors.NewErrWithMessage("bit offset is not an integer or out of range")
	}

	obj := Get(key)
	if obj == nil {
		return Encode(0, true)
	}

	requiredByteArraySize := offset/8 + 1

	// handle the case when it is string
	if assertType(obj.TypeEncoding, ObjTypeString) == nil {
		return DiceErrors.NewErrWithMessage("value is not a valid byte array")
	}

	// handle the case when it is byte array
	if assertType(obj.TypeEncoding, ObjTypeByteArray) == nil {
		byteArray := obj.Value.(*ByteArray)
		byteArrayLength := byteArray.Length

		// check whether offset, length exists or not
		if requiredByteArraySize > byteArrayLength {
			return Encode(0, true)
		}
		value := byteArray.GetBit(int(offset))
		if value {
			return Encode(1, true)
		}
		return Encode(0, true)
	}

	return Encode(0, true)
}

func evalBITCOUNT(args []string) []byte {
	var err error

	// if more than 4 arguments are provided, return error
	if len(args) > 4 {
		return DiceErrors.NewErrWithMessage(DiceErrors.SyntaxErr)
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

	if assertType(obj.TypeEncoding, ObjTypeByteArray) == nil {
		byteArray := obj.Value.(*ByteArray)
		byteArrayObject := *byteArray
		value = byteArrayObject.data
		valueLength = byteArray.Length
	}

	if assertType(obj.TypeEncoding, ObjTypeString) == nil {
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
			return DiceErrors.NewErrWithMessage(DiceErrors.IntOrOutOfRangeErr)
		}
		// Adjust start index if it is negative
		if start < 0 {
			start += valueLength
		}
	}
	if len(args) > 2 {
		end, err = strconv.ParseInt(args[2], 10, 64)
		if err != nil {
			return DiceErrors.NewErrWithMessage(DiceErrors.IntOrOutOfRangeErr)
		}

		// Adjust end index if it is negative
		if end < 0 {
			end += valueLength
		}
	}
	if len(args) > 3 {
		unit = strings.ToUpper(args[3])
		if unit != bit.BYTE && unit != bit.BIT {
			return DiceErrors.NewErrWithMessage(DiceErrors.SyntaxErr)
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
	}
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

// BITOP <AND | OR | XOR | NOT> destkey key [key ...]
func evalBITOP(args []string) []byte {
	operation, destKey := args[0], args[1]
	operation = strings.ToUpper(operation)

	// get all the keys
	keys := args[2:]

	// validation of commands
	// if operation is not from enums, then error out
	if !(operation == constants.AND || operation == constants.OR || operation == constants.XOR || operation == constants.NOT) {
		return DiceErrors.NewErrWithMessage(DiceErrors.SyntaxErr)
	}
	// if operation is not, then keys length should be only 1
	if operation == constants.NOT && len(keys) != 1 {
		return DiceErrors.NewErrWithMessage("BITOP NOT must be called with a single source key.")
	}

	if operation == constants.NOT {
		obj := Get(keys[0])
		if obj == nil {
			return Encode(0, true)
		}

		var value []byte
		if assertType(obj.TypeEncoding, ObjTypeByteArray) == nil {
			byteArray := obj.Value.(*ByteArray)
			byteArrayObject := *byteArray
			value = byteArrayObject.data
		} else {
			return DiceErrors.NewErrWithMessage("value is not a valid byte array")
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
		obj = NewObj(operationResult, -1, ObjTypeByteArray, ObjEncodingByteArray)

		// store the result in destKey
		Put(destKey, obj)
		return Encode(len(value), true)
	}
	// if operation is AND, OR, XOR
	values := make([][]byte, len(keys))

	// get the values of all keys
	for i, key := range keys {
		obj := Get(key)
		if obj == nil {
			values[i] = make([]byte, 0)
		} else {
			// handle the case when it is byte array
			if assertType(obj.TypeEncoding, ObjTypeByteArray) == nil {
				byteArray := obj.Value.(*ByteArray)
				byteArrayObject := *byteArray
				values[i] = byteArrayObject.data
			} else {
				return DiceErrors.NewErrWithMessage("value is not a valid byte array")
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
	if operation == constants.AND {
		for i := 0; i < maxLength; i++ {
			if i < minLength {
				result[i] = values[maxKeyIterator][i]
			} else {
				result[i] = 0
			}
		}
	}
	if operation == constants.XOR || operation == constants.OR {
		for i := 0; i < maxLength; i++ {
			result[i] = 0x00
		}
	}

	// perform the operation
	for _, value := range values {
		for i := 0; i < len(value); i++ {
			if operation == constants.AND {
				result[i] &= value[i]
			} else if operation == constants.OR {
				result[i] |= value[i]
			} else if operation == constants.XOR {
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
	operationResultObject := NewObj(operationResult, -1, ObjTypeByteArray, ObjEncodingByteArray)

	// store the result in destKey
	Put(destKey, operationResultObject)

	return Encode(len(result), true)
}

// evalCommand evaluates COMMAND <subcommand> command based on subcommand
// COUNT: return total count of commands in Dice.
func evalCommand(args []string) []byte {
	if len(args) == 0 {
		return DiceErrors.NewErrArity("COMMAND")
	}
	subcommand := strings.ToUpper(args[0])
	switch subcommand {
	case "COUNT":
		return evalCommandCount()
	case "GETKEYS":
		return evalCommandGetKeys(args[1:])
	default:
		return DiceErrors.NewErrWithFormattedMessage("unknown subcommand '%s'. Try COMMAND HELP", subcommand)
	}
}

// evalKeys returns the list of keys that match the pattern
// The pattern should be the only param in args
func evalKeys(args []string) []byte {
	if len(args) != 1 {
		return DiceErrors.NewErrArity("KEYS")
	}

	pattern := args[0]
	keys, err := Keys(pattern)
	if err != nil {
		return Encode(err, false)
	}

	return Encode(keys, false)
}

// evalCommandCount returns an number of commands supported by DiceDB
func evalCommandCount() []byte {
	return Encode(diceCommandsCount, false)
}

func evalCommandGetKeys(args []string) []byte {
	diceCmd, ok := diceCmds[strings.ToUpper(args[0])]
	if !ok {
		return DiceErrors.NewErrWithMessage("invalid command specified")
	}

	keySpecs := diceCmd.KeySpecs
	if keySpecs.BeginIndex == 0 {
		return DiceErrors.NewErrWithMessage("the command has no key arguments")
	}

	arity := diceCmd.Arity
	if (arity < 0 && len(args) < -arity) ||
		(arity >= 0 && len(args) != arity) {
		return DiceErrors.NewErrWithMessage("invalid number of arguments specified for command")
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
		return DiceErrors.NewErrArity("RENAME")
	}
	sourceKey := args[0]
	destKey := args[1]

	// if Source and Destination Keys are same return RESP encoded ok
	if sourceKey == destKey {
		return RespOK
	}

	// if Source key does not exist, return RESP encoded nil
	sourceObj := Get(sourceKey)
	if sourceObj == nil {
		return DiceErrors.NewErrWithMessage(DiceErrors.NoKeyErr)
	}

	if ok := Rename(sourceKey, destKey); ok {
		return RespOK
	}
	return RespNIL
}

// The MGET command returns an array of RESP values corresponding to the provided keys.
// For each key, if the key is expired or does not exist, the response will be RespNIL;
// otherwise, the response will be the RESP value of the key.
// MGET is atomic, it retrieves all values at once
func evalMGET(args []string) []byte {
	if len(args) < 1 {
		return DiceErrors.NewErrArity("MGET")
	}
	values := GetAll(args)
	response := make([]interface{}, len(args))
	for i, obj := range values {
		if obj == nil {
			response[i] = RespNIL
		} else {
			response[i] = obj.Value
		}
	}
	return Encode(response, false)
}

func evalEXISTS(args []string) []byte {
	if len(args) == 0 {
		return DiceErrors.NewErrArity("EXISTS")
	}

	var count int
	for _, key := range args {
		if GetNoTouch(key) != nil {
			count++
		}
	}

	return Encode(count, false)
}

func executeCommand(cmd *RedisCmd, c *Client) []byte {
	diceCmd, ok := diceCmds[cmd.Cmd]
	if !ok {
		return DiceErrors.NewErrWithFormattedMessage("unknown command '%s', with args beginning with: %s", cmd.Cmd, strings.Join(cmd.Args, " "))
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
			return DiceErrors.NewErrWithMessage("EXEC without MULTI")
		}
		return c.TxnExec()
	}
	if diceCmd.Name == "DISCARD" {
		if !c.isTxn {
			return DiceErrors.NewErrWithMessage("DISCARD without MULTI")
		}
		c.TxnDiscard()
		return RespOK
	}
	if diceCmd.Name == "ABORT" {
		return RespOK
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
			if _, err := c.Write(RespAuthFailure); err != nil {
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
			buf.Write(RespQueued)
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
		return DiceErrors.NewErrArity("PERSIST")
	}

	key := args[0]

	obj := Get(key)

	// If the key does not exist, return RESP encoded 0 to denote the key does not exist
	if obj == nil {
		return RespZero
	}

	// If the object exists but no expiration is set on it, return -1
	_, isExpirySet := getExpiry(obj)
	if !isExpirySet {
		return RespMinusOne
	}

	// If the object exists, remove the expiration time
	delExpiry(obj)

	return RespOne
}

func evalCOPY(args []string) []byte {
	if len(args) < 2 {
		return DiceErrors.NewErrArity("COPY")
	}

	isReplace := false

	sourceKey := args[0]
	destinationKey := args[1]

	sourceObj := Get(sourceKey)
	if sourceObj == nil {
		return RespZero
	}

	for i := 2; i < len(args); i++ {
		arg := strings.ToUpper(args[i])
		if arg == "REPLACE" {
			isReplace = true
		}
	}

	if isReplace {
		Del(destinationKey)
	}

	destinationObj := Get(destinationKey)
	if destinationObj != nil {
		return RespZero
	}

	copyObj := sourceObj.DeepCopy()
	if copyObj == nil {
		return RespZero
	}

	exp, ok := getExpiry(sourceObj)
	var exDurationMs int64 = -1
	if ok {
		exDurationMs = int64(exp - uint64(time.Now().UnixMilli()))
	}

	Put(destinationKey, copyObj)

	if exDurationMs > 0 {
		setExpiry(copyObj, exDurationMs)
	}
	return RespOne
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
// evalGET returns RespNIL if key is expired or it does not exist
func evalGETEX(args []string) []byte {
	if len(args) < 1 {
		return DiceErrors.NewErrArity("GETEX")
	}

	var key string = args[0]

	// Get the key from the hash table
	obj := Get(key)

	// if key does not exist, return RESP encoded nil
	if obj == nil {
		return RespNIL
	}

	var exDurationMs int64 = -1
	var state exDurationState = Uninitialized
	var persist bool = false
	for i := 1; i < len(args); i++ {
		arg := strings.ToUpper(args[i])
		switch arg {
		case constants.Ex, constants.Px:
			if state != Uninitialized {
				return DiceErrors.NewErrWithMessage(DiceErrors.SyntaxErr)
			}
			i++
			if i == len(args) {
				return DiceErrors.NewErrWithMessage(DiceErrors.SyntaxErr)
			}

			exDuration, err := strconv.ParseInt(args[i], 10, 64)
			if err != nil {
				return DiceErrors.NewErrWithMessage(DiceErrors.IntOrOutOfRangeErr)
			}
			if exDuration <= 0 {
				return DiceErrors.NewErrExpireTime("GETEX")
			}

			// converting seconds to milliseconds
			if arg == constants.Ex {
				exDuration *= 1000
			}
			exDurationMs = exDuration
			state = Initialized

		case constants.Pxat, constants.Exat:
			if state != Uninitialized {
				return DiceErrors.NewErrWithMessage(DiceErrors.SyntaxErr)
			}
			i++
			if i == len(args) {
				return DiceErrors.NewErrWithMessage(DiceErrors.SyntaxErr)
			}
			exDuration, err := strconv.ParseInt(args[i], 10, 64)
			if err != nil {
				return DiceErrors.NewErrWithMessage(DiceErrors.IntOrOutOfRangeErr)
			}

			if exDuration < 0 {
				return DiceErrors.NewErrExpireTime("GETEX")
			}

			if arg == constants.Exat {
				exDuration *= 1000
			}
			exDurationMs = exDuration - time.Now().UnixMilli()
			// If the expiry time is in the past, set exDurationMs to 0
			// This will be used to signal immediate expiration
			if exDurationMs < 0 {
				exDurationMs = 0
			}
			state = Initialized

		case "PERSIST":
			if state != Uninitialized {
				return DiceErrors.NewErrWithMessage(DiceErrors.SyntaxErr)
			}
			persist = true
			state = Initialized
		default:
			return DiceErrors.NewErrWithMessage(DiceErrors.SyntaxErr)
		}
	}

	if state == Initialized {
		if persist {
			delExpiry(obj)
		} else {
			setExpiry(obj, exDurationMs)
		}
	}

	// return the RESP encoded value
	return Encode(obj.Value, false)
}

// evalPTTL returns Time-to-Live in millisecs for the queried key in args
// The key should be the only param in args else returns with an error
// Returns	RESP encoded time (in secs) remaining for the key to expire
//
//	RESP encoded -2 stating key doesn't exist or key is expired
//	RESP encoded -1 in case no expiration is set on the key
func evalPTTL(args []string) []byte {
	if len(args) != 1 {
		return DiceErrors.NewErrArity("PTTL")
	}

	key := args[0]

	obj := Get(key)

	if obj == nil {
		return RespMinusTwo
	}

	exp, isExpirySet := getExpiry(obj)

	if !isExpirySet {
		return RespMinusOne
	}

	// compute the time remaining for the key to expire and
	// return the RESP encoded form of it
	durationMs := exp - uint64(time.Now().UnixMilli())
	return Encode(int64(durationMs), false)
}

func evalObjectIdleTime(key string) []byte {
	obj := GetNoTouch(key)
	if obj == nil {
		return RespNIL
	}

	return Encode(int64(getIdleTime(obj.LastAccessedAt)), true)
}

func evalOBJECT(args []string) []byte {
	if len(args) < 2 {
		return DiceErrors.NewErrArity("OBJECT")
	}

	subcommand := strings.ToUpper(args[0])
	key := args[1]

	switch subcommand {
	case "IDLETIME":
		return evalObjectIdleTime(key)
	default:
		return DiceErrors.NewErrWithMessage(DiceErrors.SyntaxErr)
	}
}

func evalTOUCH(args []string) []byte {
	if len(args) == 0 {
		return DiceErrors.NewErrArity("TOUCH")
	}

	count := 0
	for _, key := range args {
		if Get(key) != nil {
			count++
		}
	}

	return Encode(count, false)
}

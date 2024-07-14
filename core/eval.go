package core

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"strconv"
	"syscall"
	"time"
)

var RESP_NIL []byte = []byte("$-1\r\n")
var RESP_OK []byte = []byte("+OK\r\n")
var RESP_QUEUED []byte = []byte("+QUEUED\r\n")
var RESP_ZERO []byte = []byte(":0\r\n")
var RESP_ONE []byte = []byte(":1\r\n")
var RESP_MINUS_1 []byte = []byte(":-1\r\n")
var RESP_MINUS_2 []byte = []byte(":-2\r\n")
var RESP_EMPTY_ARRAY []byte = []byte("*0\r\n")

var txnCommands map[string]bool

func init() {
	txnCommands = map[string]bool{"EXEC": true, "DISCARD": true}
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

// evalSET puts a new <key, value> pair in db as in the args
// args must contain key and value.
// args can also contain multiple options -
//
//	EX or ex which will set the expiry time(in secs) for the key
//
// Returns encoded error response if at least a <key, value> pair is not part of args
// Returns encoded error response if expiry tme value in not integer
// Returns encoded OK RESP once new entry is added
// If the key already exists then the value will be overwritten and expiry will be discarded
func evalSET(args []string) []byte {
	if len(args) <= 1 {
		return Encode(errors.New("ERR wrong number of arguments for 'set' command"), false)
	}

	var key, value string
	var exDurationMs int64 = -1

	key, value = args[0], args[1]
	oType, oEnc := deduceTypeEncoding(value)

	for i := 2; i < len(args); i++ {
		switch args[i] {
		case "EX", "ex":
			i++
			if i == len(args) {
				return Encode(errors.New("ERR syntax error"), false)
			}

			exDurationSec, err := strconv.ParseInt(args[i], 10, 64)
			if err != nil {
				return Encode(errors.New("ERR value is not an integer or out of range"), false)
			}
			exDurationMs = exDurationSec * 1000
		default:
			return Encode(errors.New("ERR syntax error"), false)
		}
	}

	// putting the k and value in a Hash Table
	Put(key, NewObj(value, exDurationMs, oType, oEnc))
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

	// if key already expired then return nil
	if hasExpired(obj) {
		return RESP_NIL
	}

	// return the RESP encoded value
	return Encode(obj.Value, false)
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

	// if key expired i.e. key does not exist hence return -2
	if exp < uint64(time.Now().UnixMilli()) {
		return RESP_MINUS_2
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
		DumpAllAOF()
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
	i++
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

	WatchListMutex.Lock()
	defer WatchListMutex.Unlock()
	if WatchList[query] == nil {
		WatchList[query] = make(map[int]struct{})
	}

	// Add the client to this key's watch list
	WatchList[query][c.Fd] = struct{}{}

	return RESP_OK
}

func executeCommand(cmd *RedisCmd, c *Client) []byte {
	switch cmd.Cmd {
	case "PING":
		return evalPING(cmd.Args)
	case "SET":
		return evalSET(cmd.Args)
	case "GET":
		return evalGET(cmd.Args)
	case "TTL":
		return evalTTL(cmd.Args)
	case "DEL":
		return evalDEL(cmd.Args)
	case "EXPIRE":
		return evalEXPIRE(cmd.Args)
	case "BGREWRITEAOF":
		return evalBGREWRITEAOF(cmd.Args)
	case "INCR":
		return evalINCR(cmd.Args)
	case "INFO":
		return evalINFO(cmd.Args)
	case "CLIENT":
		return evalCLIENT(cmd.Args)
	case "LATENCY":
		return evalLATENCY(cmd.Args)
	case "LRU":
		return evalLRU(cmd.Args)
	case "SLEEP":
		return evalSLEEP(cmd.Args)
	case "QINTINS":
		return evalQINTINS(cmd.Args)
	case "QINTREM":
		return evalQINTREM(cmd.Args)
	case "QINTLEN":
		return evalQINTLEN(cmd.Args)
	case "QINTPEEK":
		return evalQINTPEEK(cmd.Args)
	case "BFINIT":
		return evalBFINIT(cmd.Args)
	case "BFADD":
		return evalBFADD(cmd.Args)
	case "BFEXISTS":
		return evalBFEXISTS(cmd.Args)
	case "BFINFO":
		return evalBFINFO(cmd.Args)
	case "QREFINS":
		return evalQREFINS(cmd.Args)
	case "QREFREM":
		return evalQREFREM(cmd.Args)
	case "QREFLEN":
		return evalQREFLEN(cmd.Args)
	case "QREFPEEK":
		return evalQREFPEEK(cmd.Args)
	case "STACKINTPUSH":
		return evalSTACKINTPUSH(cmd.Args)
	case "STACKINTPOP":
		return evalSTACKINTPOP(cmd.Args)
	case "STACKINTLEN":
		return evalSTACKINTLEN(cmd.Args)
	case "STACKINTPEEK":
		return evalSTACKINTPEEK(cmd.Args)
	case "STACKREFPUSH":
		return evalSTACKREFPUSH(cmd.Args)
	case "STACKREFPOP":
		return evalSTACKREFPOP(cmd.Args)
	case "STACKREFLEN":
		return evalSTACKREFLEN(cmd.Args)
	case "STACKREFPEEK":
		return evalSTACKREFPEEK(cmd.Args)
	case "SUBSCRIBE": // TODO: Remove this override once we support QWATCH in dice-cli.
		return evalQWATCH(cmd.Args, c)
	case "QWATCH":
		return evalQWATCH(cmd.Args, c)
	case "MULTI":
		c.TxnBegin()
		return evalMULTI(cmd.Args)
	case "EXEC":
		if !c.isTxn {
			return Encode(errors.New("ERR EXEC without MULTI"), false)
		}
		return c.TxnExec()
	case "DISCARD":
		if !c.isTxn {
			return Encode(errors.New("ERR DISCARD without MULTI"), false)
		}
		c.TxnDiscard()
		return RESP_OK
	case "ABORT":
		return RESP_OK
	default:
		return evalPING(cmd.Args)
	}
}

func executeCommandToBuffer(cmd *RedisCmd, buf *bytes.Buffer, c *Client) {
	buf.Write(executeCommand(cmd, c))
}

func EvalAndRespond(cmds RedisCmds, c *Client) {
	var response []byte
	buf := bytes.NewBuffer(response)

	for _, cmd := range cmds {
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

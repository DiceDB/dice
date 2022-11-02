package core

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/dicedb/dice/eviction"
	"github.com/dicedb/dice/handlers"
	"github.com/dicedb/dice/object"
	"github.com/dicedb/dice/pool"
)

var RESP_NIL []byte = []byte("$-1\r\n")
var RESP_OK []byte = []byte("+OK\r\n")
var RESP_QUEUED []byte = []byte("+QUEUED\r\n")
var RESP_ZERO []byte = []byte(":0\r\n")
var RESP_ONE []byte = []byte(":1\r\n")
var RESP_MINUS_1 []byte = []byte(":-1\r\n")
var RESP_MINUS_2 []byte = []byte(":-2\r\n")

var txnCommands map[string]bool

func init() {
	txnCommands = map[string]bool{"EXEC": true, "DISCARD": true}
}

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

func evalSET(args []string, dh *handlers.DiceKVstoreHandler) []byte {
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
	dh.Put(key, object.NewObj(value, exDurationMs, oType, oEnc))
	return RESP_OK
}

func compileToken(regex string) (matcher *regexp.Regexp, err error) {
	// NOTE: It is assumed that users won't pass * as part of the
	// string but only as wildcard
	tokenNormalized := strings.ReplaceAll(regex, "*", ".*")
	if matcher, err = regexp.Compile(tokenNormalized); err != nil {
		matcher = nil
	}
	return
}

func evalFILTERKEYS(args []string, dh *handlers.DiceKVstoreHandler) []byte {
	if len(args) != 1 {
		return Encode(errors.New("ERR wrong number of arguments for 'FILTERKEYS' command"), false)
	}
	var tokenKey string = args[0]

	// Get the regex compiled on the token
	matcher, err := compileToken(tokenKey)
	if err != nil {
		return Encode(errors.New("ERR key is malformed in command"), false)
	}

	var allData sync.Map
	// Define the worker Job
	diceWorker := pool.NewDiceWorker(func(i interface{}) {
		buf := i.(object.DiceWorkerBuffer)
		// fmt.Printf("{Key: %v, Value: %v}\n", buf.Key, (*buf.Value).Value)
		key := buf.Key
		val := buf.Value.Value
		if matcher.MatchString(key) {
			allData.Store(key, val)
			// fmt.Printf("{Key: %v, Val: %v}\n", key, val)
		}
	})
	// Spawn the worker
	diceWorker.Work(dh)
	return Encode(allData, false)
}

func evalGET(args []string, dh *handlers.DiceKVstoreHandler) []byte {
	if len(args) != 1 {
		return Encode(errors.New("ERR wrong number of arguments for 'get' command"), false)
	}

	var key string = args[0]

	// Get the key from the hash table
	obj := dh.Get(key)

	// if key does not exist, return RESP encoded nil
	if obj == nil {
		return RESP_NIL
	}

	// if key already expired then return nil
	if object.GetDiceExpiryStore().HasExpired(obj) {
		return RESP_NIL
	}
	// return the RESP encoded value
	return Encode(obj.Value, false)
}

func evalTTL(args []string, dh *handlers.DiceKVstoreHandler) []byte {
	if len(args) != 1 {
		return Encode(errors.New("ERR wrong number of arguments for 'ttl' command"), false)
	}

	var key string = args[0]

	obj := dh.Get(key)

	// if key does not exist, return RESP encoded -2 denoting key does not exist
	if obj == nil {
		return RESP_MINUS_2
	}

	// if object exist, but no expiration is set on it then send -1
	exp, isExpirySet := object.GetDiceExpiryStore().GetExpiry(obj)
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

func evalDEL(args []string, dh *handlers.DiceKVstoreHandler) []byte {
	var countDeleted int = 0

	for _, key := range args {
		if ok := dh.Del(key); ok {
			countDeleted++
		}
	}

	return Encode(countDeleted, false)
}

func evalEXPIRE(args []string, dh *handlers.DiceKVstoreHandler) []byte {
	if len(args) <= 1 {
		return Encode(errors.New("ERR wrong number of arguments for 'expire' command"), false)
	}

	var key string = args[0]
	exDurationSec, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return Encode(errors.New("ERR value is not an integer or out of range"), false)
	}

	obj := dh.Get(key)

	// 0 if the timeout was not set. e.g. key doesn't exist, or operation skipped due to the provided arguments
	if obj == nil {
		return RESP_ZERO
	}

	object.GetDiceExpiryStore().SetExpiry(obj, exDurationSec*100)

	// 1 if the timeout was set.
	return RESP_ONE
}

/* Description - Spawn a background thread to persist the data via AOF technique. Current implementation is
based on CoW optimization and Fork */
// TODO: Implement Acknowledgement so that main process could know whether child has finished writing to its AOF file or not.
// TODO: Make it safe from failure, an stable policy would be to write the new flushes to a temporary files and then rename them to the main process's AOF file
// TODO: Add fsync() and fdatasync() to persist to AOF for above cases.
func evalBGREWRITEAOF(args []string, dh *handlers.DiceKVstoreHandler) []byte {
	// Fork a child process, this child process would inherit all the uncommitted pages from main process.
	// This technique utilizes the CoW or copy-on-write, so while the main process is free to modify them
	// the child would save all the pages to disk.
	// Check details here -https://www.sobyte.net/post/2022-10/fork-cow/
	newChild, _, _ := syscall.Syscall(syscall.SYS_FORK, 0, 0, 0)
	if newChild == 0 {
		//We are inside child process now, so we'll start flushing to disk.
		DumpAllAOF(dh)
		return []byte("")
	} else {
		//Back to main thread
		return RESP_OK
	}
}

func evalINCR(args []string, dh *handlers.DiceKVstoreHandler) []byte {
	if len(args) != 1 {
		return Encode(errors.New("ERR wrong number of arguments for 'incr' command"), false)
	}

	var key string = args[0]
	obj := dh.Get(key)
	if obj == nil {
		obj = object.NewObj("0", -1, object.OBJ_TYPE_STRING, object.OBJ_ENCODING_INT)
		dh.Put(key, obj)
	}

	if err := assertType(obj.TypeEncoding, object.OBJ_TYPE_STRING); err != nil {
		return Encode(err, false)
	}

	if err := assertEncoding(obj.TypeEncoding, object.OBJ_ENCODING_INT); err != nil {
		return Encode(err, false)
	}

	i, _ := strconv.ParseInt(obj.Value.(string), 10, 64)
	i++
	obj.Value = strconv.FormatInt(i, 10)

	return Encode(i, false)
}

func evalINFO(args []string) []byte {
	var info []byte
	buf := bytes.NewBuffer(info)
	buf.WriteString("# Keyspace\r\n")
	for i := range KeyspaceStat {
		buf.WriteString(fmt.Sprintf("db%d:keys=%d,expires=0,avg_ttl=0\r\n", i, KeyspaceStat[i]["keys"]))
	}
	return Encode(buf.String(), false)
}

func evalCLIENT(args []string) []byte {
	return RESP_OK
}

func evalLATENCY(args []string) []byte {
	return Encode([]string{}, false)
}

func evalLRU(args []string, dh *handlers.DiceKVstoreHandler) []byte {
	eviction.EvictAllkeysLRU(dh)
	return RESP_OK
}

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

func evalMULTI(args []string) []byte {
	return RESP_OK
}

func executeCommand(cmd *RedisCmd, c *Client, dh *handlers.DiceKVstoreHandler) []byte {
	switch cmd.Cmd {
	case "PING":
		return evalPING(cmd.Args)
	case "SET":
		return evalSET(cmd.Args, dh)
	case "GET":
		return evalGET(cmd.Args, dh)
	case "FILTERKEYS":
		return evalFILTERKEYS(cmd.Args, dh)
	case "TTL":
		return evalTTL(cmd.Args, dh)
	case "DEL":
		return evalDEL(cmd.Args, dh)
	case "EXPIRE":
		return evalEXPIRE(cmd.Args, dh)
	case "BGREWRITEAOF":
		return evalBGREWRITEAOF(cmd.Args, dh)
	case "INCR":
		return evalINCR(cmd.Args, dh)
	case "INFO":
		return evalINFO(cmd.Args)
	case "CLIENT":
		return evalCLIENT(cmd.Args)
	case "LATENCY":
		return evalLATENCY(cmd.Args)
	case "LRU":
		return evalLRU(cmd.Args, dh)
	case "SLEEP":
		return evalSLEEP(cmd.Args)
	case "MULTI":
		c.TxnBegin()
		return evalMULTI(cmd.Args)
	case "EXEC":
		if !c.isTxn {
			return Encode(errors.New("ERR EXEC without MULTI"), false)
		}
		return c.TxnExec(dh)
	case "DISCARD":
		if !c.isTxn {
			return Encode(errors.New("ERR DISCARD without MULTI"), false)
		}
		c.TxnDiscard()
		return RESP_OK
	default:
		return evalPING(cmd.Args)
	}
}

func executeCommandToBuffer(cmd *RedisCmd, buf *bytes.Buffer, c *Client, dh *handlers.DiceKVstoreHandler) {
	buf.Write(executeCommand(cmd, c, dh))
}

func EvalAndRespond(cmds RedisCmds, c *Client, dh *handlers.DiceKVstoreHandler) {
	var response []byte
	buf := bytes.NewBuffer(response)
	for _, cmd := range cmds {
		// if txn is not in progress, then we can simply
		// execute the command and add the response to the buffer
		if !c.isTxn {
			executeCommandToBuffer(cmd, buf, c, dh)
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
			executeCommandToBuffer(cmd, buf, c, dh)
		}
	}
	c.Write(buf.Bytes())
}

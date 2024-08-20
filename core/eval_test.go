package core

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/bytedance/sonic"
	"github.com/dicedb/dice/internal/constants"
	"gotest.tools/v3/assert"
)

type evalTestCase struct {
	setup     func()
	input     []string
	output    []byte
	validator func(output []byte)
}

func resetStore(store *Store) {
	store.store = nil
	store.keypool = nil
	store.expires = nil
	KeyspaceStat[0] = nil
}

func setupTest(store *Store) {
	resetStore(store)
	store.store = make(map[unsafe.Pointer]*Obj)
	store.keypool = make(map[string]unsafe.Pointer)
	store.expires = make(map[*Obj]uint64)
	KeyspaceStat[0] = make(map[string]int)
}

func TestEval(t *testing.T) {
	store := NewStore()

	testEvalMSET(t, store)
	testEvalPING(t, store)
	testEvalHELLO(t, store)
	testEvalSET(t, store)
	testEvalGET(t, store)
	testEvalJSONTYPE(t, store)
	testEvalJSONGET(t, store)
	testEvalJSONSET(t, store)
	testEvalTTL(t, store)
	testEvalDel(t, store)
	testEvalPersist(t, store)
	testEvalEXPIRE(t, store)
	testEvalEXPIRETIME(t, store)
	testEvalEXPIREAT(t, store)
	testEvalDbsize(t, store)
}

func testEvalPING(t *testing.T, store *Store) {
	tests := map[string]evalTestCase{
		"nil value":            {input: nil, output: []byte("+PONG\r\n")},
		"empty args":           {input: []string{}, output: []byte("+PONG\r\n")},
		"one value":            {input: []string{"HEY"}, output: []byte("$3\r\nHEY\r\n")},
		"more than one values": {input: []string{"HEY", "HELLO"}, output: []byte("-ERR wrong number of arguments for 'ping' command\r\n")},
	}

	runEvalTests(t, tests, evalPING, store)
}

func testEvalHELLO(t *testing.T, store *Store) {
	response := []interface{}{
		"proto", 2,
		"id", serverID,
		"mode", "standalone",
		"role", "master",
		"modules", []interface{}{},
	}

	tests := map[string]evalTestCase{
		"nil value":            {input: nil, output: Encode(response, false)},
		"empty args":           {input: []string{}, output: Encode(response, false)},
		"one value":            {input: []string{"HEY"}, output: Encode(response, false)},
		"more than one values": {input: []string{"HEY", "HELLO"}, output: []byte("-ERR wrong number of arguments for 'hello' command\r\n")},
	}

	runEvalTests(t, tests, evalHELLO, store)
}

func testEvalSET(t *testing.T, store *Store) {
	tests := map[string]evalTestCase{
		"nil value":                       {input: nil, output: []byte("-ERR wrong number of arguments for 'set' command\r\n")},
		"empty array":                     {input: []string{}, output: []byte("-ERR wrong number of arguments for 'set' command\r\n")},
		"one value":                       {input: []string{"KEY"}, output: []byte("-ERR wrong number of arguments for 'set' command\r\n")},
		"key val pair":                    {input: []string{"KEY", "VAL"}, output: RespOK},
		"key val pair and expiry key":     {input: []string{"KEY", "VAL", constants.Px}, output: []byte("-ERR syntax error\r\n")},
		"key val pair and EX no val":      {input: []string{"KEY", "VAL", constants.Ex}, output: []byte("-ERR syntax error\r\n")},
		"key val pair and valid EX":       {input: []string{"KEY", "VAL", constants.Ex, "2"}, output: RespOK},
		"key val pair and invalid EX":     {input: []string{"KEY", "VAL", constants.Ex, "invalid_expiry_val"}, output: []byte("-ERR value is not an integer or out of range\r\n")},
		"key val pair and valid PX":       {input: []string{"KEY", "VAL", constants.Px, "2000"}, output: RespOK},
		"key val pair and invalid PX":     {input: []string{"KEY", "VAL", constants.Px, "invalid_expiry_val"}, output: []byte("-ERR value is not an integer or out of range\r\n")},
		"key val pair and both EX and PX": {input: []string{"KEY", "VAL", constants.Ex, "2", constants.Px, "2000"}, output: []byte("-ERR syntax error\r\n")},
		"key val pair and PXAT no val":    {input: []string{"KEY", "VAL", constants.Pxat}, output: []byte("-ERR syntax error\r\n")},
		"key val pair and invalid PXAT":   {input: []string{"KEY", "VAL", constants.Pxat, "invalid_expiry_val"}, output: []byte("-ERR value is not an integer or out of range\r\n")},
		"key val pair and expired PXAT":   {input: []string{"KEY", "VAL", constants.Pxat, "2"}, output: RespOK},
		"key val pair and negative PXAT":  {input: []string{"KEY", "VAL", constants.Pxat, "-123456"}, output: []byte("-ERR invalid expire time in 'set' command\r\n")},
		"key val pair and valid PXAT":     {input: []string{"KEY", "VAL", constants.Pxat, strconv.FormatInt(time.Now().Add(2*time.Minute).UnixMilli(), 10)}, output: RespOK},
	}

	runEvalTests(t, tests, evalSET, store)
}

func testEvalMSET(t *testing.T, store *Store) {
	tests := map[string]evalTestCase{
		"nil value":         {input: nil, output: []byte("-ERR wrong number of arguments for 'mset' command\r\n")},
		"empty array":       {input: []string{}, output: []byte("-ERR wrong number of arguments for 'mset' command\r\n")},
		"one value":         {input: []string{"KEY"}, output: []byte("-ERR wrong number of arguments for 'mset' command\r\n")},
		"key val pair":      {input: []string{"KEY", "VAL"}, output: RespOK},
		"odd key val pair":  {input: []string{"KEY", "VAL", "KEY2"}, output: []byte("-ERR wrong number of arguments for 'mset' command\r\n")},
		"even key val pair": {input: []string{"KEY", "VAL", "KEY2", "VAL2"}, output: RespOK},
	}

	runEvalTests(t, tests, evalMSET, store)
}

func testEvalGET(t *testing.T, store *Store) {
	tests := map[string]evalTestCase{
		"nil value": {
			setup:  func() {},
			input:  nil,
			output: []byte("-ERR wrong number of arguments for 'get' command\r\n"),
		},
		"empty array": {
			setup:  func() {},
			input:  []string{},
			output: []byte("-ERR wrong number of arguments for 'get' command\r\n"),
		},
		"key does not exist": {
			setup:  func() {},
			input:  []string{"NONEXISTENT_KEY"},
			output: RespNIL,
		},
		"multiple arguments": {
			setup:  func() {},
			input:  []string{"KEY1", "KEY2"},
			output: []byte("-ERR wrong number of arguments for 'get' command\r\n"),
		},
		"key exists": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "mock_value"
				obj := &Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.store[unsafe.Pointer(obj)] = obj
				store.keypool[key] = unsafe.Pointer(obj)
			},
			input:  []string{"EXISTING_KEY"},
			output: Encode("mock_value", false),
		},
		"key exists but expired": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "mock_value"
				obj := &Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.store[unsafe.Pointer(obj)] = obj
				store.keypool[key] = unsafe.Pointer(obj)
				store.expires[obj] = uint64(time.Now().Add(-2 * time.Minute).Unix())
			},
			input:  []string{"EXISTING_KEY"},
			output: RespNIL,
		},
	}

	runEvalTests(t, tests, evalGET, store)
}


func testEvalEXPIRE(t *testing.T, store *Store) {
	tests := map[string]evalTestCase{
		"nil value": {
			input:  nil,
			output: []byte("-ERR wrong number of arguments for 'expire' command\r\n"),
		},
		"empty args": {
			input:  []string{},
			output: []byte("-ERR wrong number of arguments for 'expire' command\r\n"),
		},
		"wrong number of args": {
			input:  []string{"KEY1"},
			output: []byte("-ERR wrong number of arguments for 'expire' command\r\n"),
		},
		"key does not exist": {
			input:  []string{"NONEXISTENT_KEY", strconv.FormatInt(1, 10)},
			output: RespZero,
		},
		"key exists": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "mock_value"
				obj := &Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.store[unsafe.Pointer(obj)] = obj
				store.keypool[key] = unsafe.Pointer(obj)
			},
			input:  []string{"EXISTING_KEY", strconv.FormatInt(1, 10)},
			output: RespOne,
		},
	}

	runEvalTests(t, tests, evalEXPIRE, store)
}

func testEvalEXPIRETIME(t *testing.T, store *Store) {
	tests := map[string]evalTestCase{
		"wrong number of args": {
			input:  []string{"KEY1", "KEY2"},
			output: []byte("-ERR wrong number of arguments for 'expiretime' command\r\n"),
		},
		"key does not exist": {
			input:  []string{"NONEXISTENT_KEY"},
			output: RespMinusTwo,
		},
		"key exists without expiry": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "mock_value"
				obj := &Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.store[unsafe.Pointer(obj)] = obj
				store.keypool[key] = unsafe.Pointer(obj)
			},
			input:  []string{"EXISTING_KEY"},
			output: RespMinusOne,
		},
		"key exists with expiry": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "mock_value"
				obj := &Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.store[unsafe.Pointer(obj)] = obj
				store.keypool[key] = unsafe.Pointer(obj)
				store.expires[obj] = uint64(2724123456123)
			},
			input:  []string{"EXISTING_KEY"},
			output: []byte(fmt.Sprintf(":%d\r\n", 2724123456)),
		},
	}

	runEvalTests(t, tests, evalEXPIRETIME, store)
}

func testEvalEXPIREAT(t *testing.T, store *Store) {
	tests := map[string]evalTestCase{
		"nil value": {
			input:  nil,
			output: []byte("-ERR wrong number of arguments for 'expireat' command\r\n"),
		},
		"empty args": {
			input:  []string{},
			output: []byte("-ERR wrong number of arguments for 'expireat' command\r\n"),
		},
		"wrong number of args": {
			input:  []string{"KEY1"},
			output: []byte("-ERR wrong number of arguments for 'expireat' command\r\n"),
		},
		"key does not exist": {
			input:  []string{"NONEXISTENT_KEY", strconv.FormatInt(time.Now().Add(2*time.Minute).Unix(), 10)},
			output: RespZero,
		},
		"key exists": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "mock_value"
				obj := &Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.store[unsafe.Pointer(obj)] = obj
				store.keypool[key] = unsafe.Pointer(obj)
			},
			input:  []string{"EXISTING_KEY", strconv.FormatInt(time.Now().Add(2*time.Minute).Unix(), 10)},
			output: RespOne,
		},
	}

	runEvalTests(t, tests, evalEXPIREAT, store)
}

func testEvalJSONTYPE(t *testing.T, store *Store) {
	tests := map[string]evalTestCase{
		"nil value": {
			setup:  func() {},
			input:  nil,
			output: []byte("-ERR wrong number of arguments for 'JSON.TYPE' command\r\n"),
		},
		"empty array": {
			setup:  func() {},
			input:  []string{},
			output: []byte("-ERR wrong number of arguments for 'JSON.TYPE' command\r\n"),
		},
		"key does not exist": {
			setup:  func() {},
			input:  []string{"NONEXISTENT_KEY"},
			output: RespNIL,
		},
		"object type value": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"language\":[\"java\",\"go\",\"python\"]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, ObjTypeJSON, ObjEncodingJSON)
				store.store[unsafe.Pointer(obj)] = obj
				store.keypool[key] = unsafe.Pointer(obj)
			},

			input:  []string{"EXISTING_KEY"},
			output: []byte("$6\r\nobject\r\n"),
		},
		"array type value": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"language\":[\"java\",\"go\",\"python\"]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, ObjTypeJSON, ObjEncodingJSON)
				store.store[unsafe.Pointer(obj)] = obj
				store.keypool[key] = unsafe.Pointer(obj)
			},

			input:  []string{"EXISTING_KEY", "$.language"},
			output: []byte("$5\r\narray\r\n"),
		},
		"string type value": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"a\":\"test\"}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, ObjTypeJSON, ObjEncodingJSON)
				store.store[unsafe.Pointer(obj)] = obj
				store.keypool[key] = unsafe.Pointer(obj)
			},

			input:  []string{"EXISTING_KEY", "$.a"},
			output: []byte("$6\r\nstring\r\n"),
		},
		"boolean type value": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"flag\":true}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, ObjTypeJSON, ObjEncodingJSON)
				store.store[unsafe.Pointer(obj)] = obj
				store.keypool[key] = unsafe.Pointer(obj)
			},

			input:  []string{"EXISTING_KEY", "$.flag"},
			output: []byte("$7\r\nboolean\r\n"),
		},
		"number type value": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"price\":3}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, ObjTypeJSON, ObjEncodingJSON)
				store.store[unsafe.Pointer(obj)] = obj
				store.keypool[key] = unsafe.Pointer(obj)
			},

			input:  []string{"EXISTING_KEY", "$.price"},
			output: []byte("$6\r\nnumber\r\n"),
		},
		"null type value": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"price\":3.14}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, ObjTypeJSON, ObjEncodingJSON)
				store.store[unsafe.Pointer(obj)] = obj
				store.keypool[key] = unsafe.Pointer(obj)
			},

			input:  []string{"EXISTING_KEY", "$.language"},
			output: RespNIL,
		},
		"multi type value": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"name\":\"tom\",\"partner\":{\"name\":\"jerry\"}}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, ObjTypeJSON, ObjEncodingJSON)
				store.store[unsafe.Pointer(obj)] = obj
				store.keypool[key] = unsafe.Pointer(obj)
			},

			input:  []string{"EXISTING_KEY", "$..name"},
			output: []byte("*2\r\n$6\r\nstring\r\n$6\r\nstring\r\n"),
		},
	}

	runEvalTests(t, tests, evalJSONTYPE, store)
}

func testEvalJSONGET(t *testing.T, store *Store) {
	tests := map[string]evalTestCase{
		"nil value": {
			setup:  func() {},
			input:  nil,
			output: []byte("-ERR wrong number of arguments for 'JSON.GET' command\r\n"),
		},
		"empty array": {
			setup:  func() {},
			input:  []string{},
			output: []byte("-ERR wrong number of arguments for 'JSON.GET' command\r\n"),
		},
		"key does not exist": {
			setup:  func() {},
			input:  []string{"NONEXISTENT_KEY"},
			output: RespNIL,
		},
		"key exists invalid value": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"a\":2}"
				obj := &Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.store[unsafe.Pointer(obj)] = obj
				store.keypool[key] = unsafe.Pointer(obj)
			},
			input:  []string{"EXISTING_KEY"},
			output: []byte("-the operation is not permitted on this type\r\n"),
		},
		"key exists value": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"a\":2}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, ObjTypeJSON, ObjEncodingJSON)
				store.store[unsafe.Pointer(obj)] = obj
				store.keypool[key] = unsafe.Pointer(obj)
			},

			input:  []string{"EXISTING_KEY"},
			output: []byte("$7\r\n{\"a\":2}\r\n"),
		},
		"key exists but expired": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "mock_value"
				obj := &Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.store[unsafe.Pointer(obj)] = obj
				store.keypool[key] = unsafe.Pointer(obj)
				store.expires[obj] = uint64(time.Now().Add(-2 * time.Minute).Unix())
			},
			input:  []string{"EXISTING_KEY"},
			output: RespNIL,
		},
	}

	runEvalTests(t, tests, evalJSONGET, store)
}

func testEvalJSONSET(t *testing.T, store *Store) {
	tests := map[string]evalTestCase{
		"nil value": {
			setup:  func() {},
			input:  nil,
			output: []byte("-ERR wrong number of arguments for 'JSON.SET' command\r\n"),
		},
		"empty array": {
			setup:  func() {},
			input:  []string{},
			output: []byte("-ERR wrong number of arguments for 'JSON.SET' command\r\n"),
		},
		"insufficient args": {
			setup:  func() {},
			input:  []string{},
			output: []byte("-ERR wrong number of arguments for 'JSON.SET' command\r\n"),
		},
		"invalid json path": {
			setup:  func() {},
			input:  []string{"doc", "$", "{\"a\":}"},
			output: nil,
			validator: func(output []byte) {
				assert.Assert(t, output != nil)
				assert.Assert(t, strings.Contains(string(output), "-ERR invalid JSON:"))
			},
		},
		"valid json path": {
			setup: func() {
			},
			input:  []string{"doc", "$", "{\"a\":2}"},
			output: RespOK,
		},
	}

	runEvalTests(t, tests, evalJSONSET, store)
}

func testEvalTTL(t *testing.T, store *Store) {
	tests := map[string]evalTestCase{
		"nil value": {
			setup:  func() {},
			input:  nil,
			output: []byte("-ERR wrong number of arguments for 'ttl' command\r\n"),
		},
		"empty array": {
			setup:  func() {},
			input:  []string{},
			output: []byte("-ERR wrong number of arguments for 'ttl' command\r\n"),
		},
		"key does not exist": {
			setup:  func() {},
			input:  []string{"NONEXISTENT_KEY"},
			output: RespMinusTwo,
		},
		"multiple arguments": {
			setup:  func() {},
			input:  []string{"KEY1", "KEY2"},
			output: []byte("-ERR wrong number of arguments for 'ttl' command\r\n"),
		},
		"key exists expiry not set": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "mock_value"
				obj := &Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.store[unsafe.Pointer(obj)] = obj
				store.keypool[key] = unsafe.Pointer(obj)
			},
			input:  []string{"EXISTING_KEY"},
			output: RespMinusOne,
		},
		"key exists not expired": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "mock_value"
				obj := &Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.store[unsafe.Pointer(obj)] = obj
				store.keypool[key] = unsafe.Pointer(obj)
				store.expires[obj] = uint64(time.Now().Add(2 * time.Minute).UnixMilli())
			},
			input: []string{"EXISTING_KEY"},
			validator: func(output []byte) {
				assert.Assert(t, output != nil)
				assert.Assert(t, !bytes.Equal(output, RespMinusOne))
				assert.Assert(t, !bytes.Equal(output, RespMinusTwo))
			},
		},
		"key exists but expired": {
			setup: func() {
				key := "EXISTING_EXPIRED_KEY"
				value := "mock_value"
				obj := &Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.store[unsafe.Pointer(obj)] = obj
				store.keypool[key] = unsafe.Pointer(obj)
				store.expires[obj] = uint64(time.Now().Add(-2 * time.Minute).Unix())
			},
			input:  []string{"EXISTING_KEY"},
			output: RespMinusTwo,
		},
	}

	runEvalTests(t, tests, evalTTL, store)
}

func testEvalDel(t *testing.T, store *Store) {
	tests := map[string]evalTestCase{
		"nil value": {
			setup:  func() {},
			input:  nil,
			output: []byte(":0\r\n"),
		},
		"empty array": {
			setup:  func() {},
			input:  []string{},
			output: []byte(":0\r\n"),
		},
		"key does not exist": {
			setup:  func() {},
			input:  []string{"NONEXISTENT_KEY"},
			output: []byte(":0\r\n"),
		},
		"key exists": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "mock_value"
				obj := &Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.store[unsafe.Pointer(obj)] = obj
				store.keypool[key] = unsafe.Pointer(obj)
				KeyspaceStat[0]["keys"]++
			},
			input:  []string{"EXISTING_KEY"},
			output: []byte(":1\r\n"),
		},
	}

	runEvalTests(t, tests, evalDEL, store)
}

// TestEvalPersist tests the evalPersist function using table-driven tests.
func testEvalPersist(t *testing.T, store *Store) {
	// Define test cases
	tests := map[string]evalTestCase{
		"wrong number of arguments": {
			input:  []string{"key1", "key2"},
			output: Encode(errors.New("ERR wrong number of arguments for 'persist' command"), false),
		},
		"key does not exist": {
			input:  []string{"nonexistent"},
			output: RespZero,
		},
		"key exists but no expiration set": {
			input: []string{"existent_no_expiry"},
			setup: func() {
				evalSET([]string{"existent_no_expiry", "value"}, store)
			},
			output: RespMinusOne,
		},
		"key exists and expiration removed": {
			input: []string{"existent_with_expiry"},
			setup: func() {
				evalSET([]string{"existent_with_expiry", "value", constants.Ex, "1"}, store)
			},
			output: RespOne,
		},
		"key exists with expiration set and not expired": {
			input: []string{"existent_with_expiry_not_expired"},
			setup: func() {
				// Simulate setting a key with an expiration time that has not yet passed
				evalSET([]string{"existent_with_expiry_not_expired", "value", constants.Ex, "10000"}, store) // 10000 seconds in the future
			},
			output: RespOne,
		},
	}

	runEvalTests(t, tests, evalPersist, store)
}

func testEvalDbsize(t *testing.T, store *Store) {
	tests := map[string]evalTestCase{
		"DBSIZE command with invalid no of args": {
			input:  []string{"INVALID_ARG"},
			output: []byte("-ERR wrong number of arguments for 'DBSIZE' command\r\n"),
		},
		"no key in db": {
			input:  nil,
			output: []byte(":0\r\n"),
		},
		"one key exists in db": {
			setup: func() {
				key := "KEY"
				value := "VAL"
				obj := &Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.store[unsafe.Pointer(obj)] = obj
				store.keypool[key] = unsafe.Pointer(obj)
			},
			input:  nil,
			output: []byte(":1\r\n"),
		},
		"two keys exist in db": {
			setup: func() {
				key1 := "KEY1"
				value1 := "VAL1"
				obj1 := &Obj{
					Value:          value1,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.store[unsafe.Pointer(obj1)] = obj1
				store.keypool[key1] = unsafe.Pointer(obj1)

				key2 := "KEY2"
				value2 := "VAL2"
				obj2 := &Obj{
					Value:          value2,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.store[unsafe.Pointer(obj2)] = obj2
				store.keypool[key2] = unsafe.Pointer(obj2)
			},
			input:  nil,
			output: []byte(":2\r\n"),
		},
	}

	runEvalTests(t, tests, evalDBSIZE, store)
}

func runEvalTests(t *testing.T, tests map[string]evalTestCase, evalFunc func([]string, *Store) []byte, store *Store) {
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			setupTest(store)

			if tc.setup != nil {
				tc.setup()
			}

			output := evalFunc(tc.input, store)
			if tc.validator != nil {
				tc.validator(output)
			} else {
				assert.Equal(t, string(tc.output), string(output))
			}
		})
	}
}

func BenchmarkEvalMSET(b *testing.B) {
	store := NewStore()
	for i := 0; i < b.N; i++ {
		evalMSET([]string{"KEY", "VAL", "KEY2", "VAL2"}, store)
	}
}

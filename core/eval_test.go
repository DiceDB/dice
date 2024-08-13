package core

import (
	"bytes"
	"errors"
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

func resetStore() {
	store = nil
	keypool = nil
	expires = nil
	KeyspaceStat[0] = nil
}

func setupTest() {
	resetStore()
	store = make(map[unsafe.Pointer]*Obj)
	keypool = make(map[string]unsafe.Pointer)
	expires = make(map[*Obj]uint64)
	KeyspaceStat[0] = make(map[string]int)
}

func TestEval(t *testing.T) {
	testCases := map[string]func(*testing.T){
		"MSET":    testEvalMSET,
		"PING":    testEvalPING,
		"HELLO":   testEvalHELLO,
		"SET":     testEvalSET,
		"GET":     testEvalGET,
		"JSONGET": testEvalJSONGET,
		"JSONSET": testEvalJSONSET,
		"TTL":     testEvalTTL,
		"DEL":     testEvalDel,
		"PERSIST": TestEvalPersist,
	}

	for name, testFunc := range testCases {
		t.Run(name, testFunc)
	}
}

func testEvalPING(t *testing.T) {
	tests := map[string]evalTestCase{
		"nil value":            {input: nil, output: []byte("+PONG\r\n")},
		"empty args":           {input: []string{}, output: []byte("+PONG\r\n")},
		"one value":            {input: []string{"HEY"}, output: []byte("$3\r\nHEY\r\n")},
		"more than one values": {input: []string{"HEY", "HELLO"}, output: []byte("-ERR wrong number of arguments for 'ping' command\r\n")},
	}

	runEvalTests(t, tests, evalPING)
}

func testEvalHELLO(t *testing.T) {
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

	runEvalTests(t, tests, evalHELLO)
}

func testEvalSET(t *testing.T) {
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

	runEvalTests(t, tests, evalSET)
}

func testEvalMSET(t *testing.T) {
	tests := map[string]evalTestCase{
		"nil value":         {input: nil, output: []byte("-ERR wrong number of arguments for 'mset' command\r\n")},
		"empty array":       {input: []string{}, output: []byte("-ERR wrong number of arguments for 'mset' command\r\n")},
		"one value":         {input: []string{"KEY"}, output: []byte("-ERR wrong number of arguments for 'mset' command\r\n")},
		"key val pair":      {input: []string{"KEY", "VAL"}, output: RespOK},
		"odd key val pair":  {input: []string{"KEY", "VAL", "KEY2"}, output: []byte("-ERR wrong number of arguments for 'mset' command\r\n")},
		"even key val pair": {input: []string{"KEY", "VAL", "KEY2", "VAL2"}, output: RespOK},
	}

	runEvalTests(t, tests, evalMSET)
}

func testEvalGET(t *testing.T) {
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
				store[unsafe.Pointer(obj)] = obj
				keypool[key] = unsafe.Pointer(obj)
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
				store[unsafe.Pointer(obj)] = obj
				keypool[key] = unsafe.Pointer(obj)
				expires[obj] = uint64(time.Now().Add(-2 * time.Minute).Unix())
			},
			input:  []string{"EXISTING_KEY"},
			output: RespNIL,
		},
	}

	runEvalTests(t, tests, evalGET)
}

func testEvalJSONGET(t *testing.T) {
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
				store[unsafe.Pointer(obj)] = obj
				keypool[key] = unsafe.Pointer(obj)
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
				obj := NewObj(rootData, -1, ObjTypeJSON, ObjEncodingJSON)
				store[unsafe.Pointer(obj)] = obj
				keypool[key] = unsafe.Pointer(obj)
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
				store[unsafe.Pointer(obj)] = obj
				keypool[key] = unsafe.Pointer(obj)
				expires[obj] = uint64(time.Now().Add(-2 * time.Minute).Unix())
			},
			input:  []string{"EXISTING_KEY"},
			output: RespNIL,
		},
	}

	runEvalTests(t, tests, evalJSONGET)
}

func testEvalJSONSET(t *testing.T) {
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

	runEvalTests(t, tests, evalJSONSET)
}

func testEvalTTL(t *testing.T) {
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
				store[unsafe.Pointer(obj)] = obj
				keypool[key] = unsafe.Pointer(obj)
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
				store[unsafe.Pointer(obj)] = obj
				keypool[key] = unsafe.Pointer(obj)
				expires[obj] = uint64(time.Now().Add(2 * time.Minute).UnixMilli())
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
				store[unsafe.Pointer(obj)] = obj
				keypool[key] = unsafe.Pointer(obj)
				expires[obj] = uint64(time.Now().Add(-2 * time.Minute).Unix())
			},
			input:  []string{"EXISTING_KEY"},
			output: RespMinusTwo,
		},
	}

	runEvalTests(t, tests, evalTTL)
}

func testEvalDel(t *testing.T) {
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
				store[unsafe.Pointer(obj)] = obj
				keypool[key] = unsafe.Pointer(obj)
				KeyspaceStat[0]["keys"]++
			},
			input:  []string{"EXISTING_KEY"},
			output: []byte(":1\r\n"),
		},
	}

	runEvalTests(t, tests, evalDEL)
}

// TestEvalPersist tests the evalPersist function using table-driven tests.
func TestEvalPersist(t *testing.T) {
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
				evalSET([]string{"existent_no_expiry", "value"})
			},
			output: RespMinusOne,
		},
		"key exists and expiration removed": {
			input: []string{"existent_with_expiry"},
			setup: func() {
				evalSET([]string{"existent_with_expiry", "value", constants.Ex, "1"})
			},
			output: RespOne,
		},
		"key exists with expiration set and not expired": {
			input: []string{"existent_with_expiry_not_expired"},
			setup: func() {
				// Simulate setting a key with an expiration time that has not yet passed
				evalSET([]string{"existent_with_expiry_not_expired", "value", constants.Ex, "10000"}) // 10000 seconds in the future
			},
			output: RespOne,
		},
	}

	runEvalTests(t, tests, evalPersist)
}

func runEvalTests(t *testing.T, tests map[string]evalTestCase, evalFunc func([]string) []byte) {
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			setupTest()

			if tc.setup != nil {
				tc.setup()
			}

			output := evalFunc(tc.input)
			if tc.validator != nil {
				tc.validator(output)
			} else {
				assert.Equal(t, string(tc.output), string(output))
			}
		})
	}
}

func BenchmarkEvalMSET(b *testing.B) {
	for i := 0; i < b.N; i++ {
		evalMSET([]string{"KEY", "VAL", "KEY2", "VAL2"})
	}
}

package eval

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/bytedance/sonic"
	"github.com/dicedb/dice/internal/server/utils"
	"github.com/ohler55/ojg/jp"

	"github.com/axiomhq/hyperloglog"
	"github.com/dicedb/dice/internal/clientio"
	diceerrors "github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/object"
	dstore "github.com/dicedb/dice/internal/store"
	testifyAssert "github.com/stretchr/testify/assert"
	"gotest.tools/v3/assert"
)

type evalTestCase struct {
	name           string
	setup          func()
	input          []string
	output         []byte
	validator      func(output []byte)
	migratedOutput EvalResponse
}

func setupTest(store *dstore.Store) *dstore.Store {
	dstore.ResetStore(store)
	dstore.KeyspaceStat[0] = make(map[string]int)

	return store
}

func TestEval(t *testing.T) {
	store := dstore.NewStore(nil)

	testEvalMSET(t, store)
	testEvalECHO(t, store)
	testEvalHELLO(t, store)
	testEvalSET(t, store)
	testEvalGET(t, store)
	testEvalGETEX(t, store)
	testEvalDebug(t, store)
	testEvalJSONARRINSERT(t, store)
	testEvalJSONARRPOP(t, store)
	testEvalJSONARRLEN(t, store)
	testEvalJSONDEL(t, store)
	testEvalJSONFORGET(t, store)
	testEvalJSONCLEAR(t, store)
	testEvalJSONTYPE(t, store)
	testEvalJSONGET(t, store)
	testEvalJSONSET(t, store)
	testEvalJSONNUMMULTBY(t, store)
	testEvalJSONTOGGLE(t, store)
	testEvalJSONARRAPPEND(t, store)
	testEvalTTL(t, store)
	testEvalDel(t, store)
	testEvalPersist(t, store)
	testEvalEXPIRE(t, store)
	testEvalEXPIRETIME(t, store)
	testEvalEXPIREAT(t, store)
	testEvalDbsize(t, store)
	testEvalGETSET(t, store)
	testEvalHSET(t, store)
	testEvalPFADD(t, store)
	testEvalPFCOUNT(t, store)
	testEvalHGET(t, store)
	testEvalHSTRLEN(t, store)
	testEvalHDEL(t, store)
	testEvalPFMERGE(t, store)
	testEvalJSONSTRLEN(t, store)
	testEvalJSONOBJLEN(t, store)
	testEvalHLEN(t, store)
	testEvalSELECT(t, store)
	testEvalLLEN(t, store)
	testEvalGETEX(t, store)
	testEvalJSONNUMINCRBY(t, store)
	testEvalTYPE(t, store)
	testEvalCOMMAND(t, store)
	testEvalJSONOBJKEYS(t, store)
	testEvalGETRANGE(t, store)
	testEvalPING(t, store)
	testEvalSETEX(t, store)
	testEvalFLUSHDB(t, store)
	testEvalINCRBYFLOAT(t, store)
	testEvalBITOP(t, store)
}

func testEvalPING(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"nil value":            {input: nil, output: []byte("+PONG\r\n")},
		"empty args":           {input: []string{}, output: []byte("+PONG\r\n")},
		"one value":            {input: []string{"HEY"}, output: []byte("$3\r\nHEY\r\n")},
		"more than one values": {input: []string{"HEY", "HELLO"}, output: []byte("-ERR wrong number of arguments for 'ping' command\r\n")},
	}

	runEvalTests(t, tests, evalPING, store)
}

func testEvalECHO(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"nil value":            {input: nil, output: []byte("-ERR wrong number of arguments for 'echo' command\r\n")},
		"empty args":           {input: []string{}, output: []byte("-ERR wrong number of arguments for 'echo' command\r\n")},
		"one value":            {input: []string{"HEY"}, output: []byte("$3\r\nHEY\r\n")},
		"more than one values": {input: []string{"HEY", "HELLO"}, output: []byte("-ERR wrong number of arguments for 'echo' command\r\n")},
	}

	runEvalTests(t, tests, evalECHO, store)
}

func testEvalHELLO(t *testing.T, store *dstore.Store) {
	resp := []interface{}{
		"proto", 2,
		"id", serverID,
		"mode", "standalone",
		"role", "master",
		"modules", []interface{}{},
	}

	tests := map[string]evalTestCase{
		"nil value":            {input: nil, output: clientio.Encode(resp, false)},
		"empty args":           {input: []string{}, output: clientio.Encode(resp, false)},
		"one value":            {input: []string{"HEY"}, output: clientio.Encode(resp, false)},
		"more than one values": {input: []string{"HEY", "HELLO"}, output: []byte("-ERR wrong number of arguments for 'hello' command\r\n")},
	}

	runEvalTests(t, tests, evalHELLO, store)
}

func testEvalSET(t *testing.T, store *dstore.Store) {

	tests := []evalTestCase{
		{
			name:           "nil value",
			input:          nil,
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("-ERR wrong number of arguments for 'set' command\r\n")},
		},
		{
			name:           "empty array",
			input:          []string{},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("-ERR wrong number of arguments for 'set' command\r\n")},
		},
		{
			name:           "one value",
			input:          []string{"KEY"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("-ERR wrong number of arguments for 'set' command\r\n")},
		},
		{
			name:           "key val pair",
			input:          []string{"KEY", "VAL"},
			migratedOutput: EvalResponse{Result: clientio.RespOK, Error: nil},
		},
		{
			name:           "key val pair with int val",
			input:          []string{"KEY", "123456"},
			migratedOutput: EvalResponse{Result: clientio.RespOK, Error: nil},
		},
		{
			name:           "key val pair and expiry key",
			input:          []string{"KEY", "VAL", Px},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("-ERR syntax error\r\n")},
		},
		{
			name:           "key val pair and EX no val",
			input:          []string{"KEY", "VAL", Ex},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("-ERR syntax error\r\n")},
		},
		{
			name:           "key val pair and valid EX",
			input:          []string{"KEY", "VAL", Ex, "2"},
			migratedOutput: EvalResponse{Result: clientio.RespOK, Error: nil},
		},
		{
			name:           "key val pair and invalid EX",
			input:          []string{"KEY", "VAL", Ex, "invalid_expiry_val"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("-ERR value is not an integer or out of range\r\n")},
		},
		{
			name:           "key val pair and valid PX",
			input:          []string{"KEY", "VAL", Px, "2000"},
			migratedOutput: EvalResponse{Result: clientio.RespOK, Error: nil},
		},
		{
			name:           "key val pair and invalid PX",
			input:          []string{"KEY", "VAL", Px, "invalid_expiry_val"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("-ERR value is not an integer or out of range\r\n")},
		},
		{
			name:           "key val pair and both EX and PX",
			input:          []string{"KEY", "VAL", Ex, "2", Px, "2000"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("-ERR syntax error\r\n")},
		},
		{
			name:           "key val pair and PXAT no val",
			input:          []string{"KEY", "VAL", Pxat},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("-ERR syntax error\r\n")},
		},
		{
			name:           "key val pair and invalid PXAT",
			input:          []string{"KEY", "VAL", Pxat, "invalid_expiry_val"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("-ERR value is not an integer or out of range\r\n")},
		},
		{
			name:           "key val pair and expired PXAT",
			input:          []string{"KEY", "VAL", Pxat, "2"},
			migratedOutput: EvalResponse{Result: clientio.RespOK, Error: nil},
		},
		{
			name:           "key val pair and negative PXAT",
			input:          []string{"KEY", "VAL", Pxat, "-123456"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("-ERR invalid expire time in 'set' command\r\n")},
		},
		{
			name:           "key val pair and valid PXAT",
			input:          []string{"KEY", "VAL", Pxat, strconv.FormatInt(time.Now().Add(2*time.Minute).UnixMilli(), 10)},
			migratedOutput: EvalResponse{Result: clientio.RespOK, Error: nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := evalSET(tt.input, store)

			// Handle comparison for byte slices
			if b, ok := response.Result.([]byte); ok && tt.migratedOutput.Result != nil {
				if expectedBytes, ok := tt.migratedOutput.Result.([]byte); ok {
					testifyAssert.True(t, bytes.Equal(b, expectedBytes), "expected and actual byte slices should be equal")
				}
			} else {
				assert.Equal(t, tt.migratedOutput.Result, response.Result)
			}

			if tt.migratedOutput.Error != nil {
				testifyAssert.EqualError(t, response.Error, tt.migratedOutput.Error.Error())
			} else {
				testifyAssert.NoError(t, response.Error)
			}
		})
	}
}

func testEvalGETEX(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{

		"key val pair and valid EX": {
			setup: func() {
				key := "foo"
				value := "bar"
				obj := &object.Obj{
					Value: value,
				}
				store.Put(key, obj)
			},
			input:  []string{"foo", Ex, "10"},
			output: clientio.Encode("bar", false),
		},
		"key val pair and invalid EX": {
			setup: func() {
				key := "foo"
				value := "bar"
				obj := &object.Obj{
					Value: value,
				}
				store.Put(key, obj)
			},
			input:  []string{"foo", Ex, "10000000000000000"},
			output: []byte("-ERR invalid expire time in 'getex' command\r\n")},
		"key holding json type": {
			setup: func() {
				evalJSONSET([]string{"JSONKEY", "$", "1"}, store)

			},
			input:  []string{"JSONKEY"},
			output: []byte("-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"),
		},
		"key holding set type": {
			setup: func() {
				evalSADD([]string{"SETKEY", "FRUITS", "APPLE", "MANGO", "BANANA"}, store)

			},
			input:  []string{"SETKEY"},
			output: []byte("-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"),
		},
	}

	runEvalTests(t, tests, evalGETEX, store)
}

func testEvalMSET(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"nil value":         {input: nil, output: []byte("-ERR wrong number of arguments for 'mset' command\r\n")},
		"empty array":       {input: []string{}, output: []byte("-ERR wrong number of arguments for 'mset' command\r\n")},
		"one value":         {input: []string{"KEY"}, output: []byte("-ERR wrong number of arguments for 'mset' command\r\n")},
		"key val pair":      {input: []string{"KEY", "VAL"}, output: clientio.RespOK},
		"odd key val pair":  {input: []string{"KEY", "VAL", "KEY2"}, output: []byte("-ERR wrong number of arguments for 'mset' command\r\n")},
		"even key val pair": {input: []string{"KEY", "VAL", "KEY2", "VAL2"}, output: clientio.RespOK},
	}

	runEvalTests(t, tests, evalMSET, store)
}

func testEvalGET(t *testing.T, store *dstore.Store) {
	tests := []evalTestCase{
		{
			name:           "nil value",
			input:          nil,
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("-ERR wrong number of arguments for 'get' command\r\n")},
		},
		{
			name:           "empty array",
			input:          []string{},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("-ERR wrong number of arguments for 'get' command\r\n")},
		},
		{
			name:           "key does not exist",
			input:          []string{"NONEXISTENT_KEY"},
			migratedOutput: EvalResponse{Result: clientio.RespNIL, Error: nil},
		},
		{
			name:           "multiple arguments",
			input:          []string{"KEY1", "KEY2"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("-ERR wrong number of arguments for 'get' command\r\n")},
		},
		{
			name: "key exists",
			setup: func() {
				key := "EXISTING_KEY"
				value := "mock_value"
				obj := &object.Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)
			},
			input:          []string{"EXISTING_KEY"},
			migratedOutput: EvalResponse{Result: fmt.Sprintf("$%d\r\n%s\r\n", len("mock_value"), "mock_value"), Error: nil},
		},
		{
			name: "key exists but expired",
			setup: func() {
				key := "EXISTING_KEY"
				value := "mock_value"
				obj := &object.Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)
				store.SetExpiry(obj, int64(-2*time.Millisecond))
			},
			input:          []string{"EXISTING_KEY"},
			migratedOutput: EvalResponse{Result: clientio.RespNIL, Error: nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := evalGET(tt.input, store)

			// Handle comparison for byte slices
			if b, ok := response.Result.([]byte); ok && tt.migratedOutput.Result != nil {
				if expectedBytes, ok := tt.migratedOutput.Result.([]byte); ok {
					testifyAssert.True(t, bytes.Equal(b, expectedBytes), "expected and actual byte slices should be equal")
				}
			} else {
				assert.Equal(t, tt.migratedOutput.Result, response.Result)
			}

			if tt.migratedOutput.Error != nil {
				testifyAssert.EqualError(t, response.Error, tt.migratedOutput.Error.Error())
			} else {
				testifyAssert.NoError(t, response.Error)
			}
		})
	}
}

func testEvalGETSET(t *testing.T, store *dstore.Store) {
	tests := []evalTestCase{
		{
			name:           "GETSET with 1 arg",
			input:          []string{"HELLO"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("-ERR wrong number of arguments for 'getset' command\r\n")},
		},
		{
			name:           "GETSET with 3 args",
			input:          []string{"HELLO", "WORLD", "WORLD1"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("-ERR wrong number of arguments for 'getset' command\r\n")},
		},
		{
			name:           "GETSET key not exists",
			input:          []string{"HELLO", "WORLD"},
			migratedOutput: EvalResponse{Result: clientio.RespNIL, Error: nil},
		},
		{
			name: "GETSET key exists",
			setup: func() {
				key := "EXISTING_KEY"
				value := "mock_value"
				obj := &object.Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)
			},
			input:          []string{"EXISTING_KEY", "WORLD"},
			migratedOutput: EvalResponse{Result: fmt.Sprintf("$%d\r\n%s\r\n", len("mock_value"), "mock_value"), Error: nil},
		},
		{
			name: "GETSET key exists TTL should be reset",
			setup: func() {
				key := "EXISTING_KEY"
				value := "mock_value"
				obj := &object.Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)
			},
			input:          []string{"EXISTING_KEY", "WORLD"},
			migratedOutput: EvalResponse{Result: fmt.Sprintf("$%d\r\n%s\r\n", len("mock_value"), "mock_value"), Error: nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := evalGETSET(tt.input, store)

			// Handle comparison for byte slices
			if b, ok := response.Result.([]byte); ok && tt.migratedOutput.Result != nil {
				if expectedBytes, ok := tt.migratedOutput.Result.([]byte); ok {
					testifyAssert.True(t, bytes.Equal(b, expectedBytes), "expected and actual byte slices should be equal")
				}
			} else {
				assert.Equal(t, tt.migratedOutput.Result, response.Result)
			}

			if tt.migratedOutput.Error != nil {
				testifyAssert.EqualError(t, response.Error, tt.migratedOutput.Error.Error())
			} else {
				testifyAssert.NoError(t, response.Error)
			}
		})
	}
}

func testEvalEXPIRE(t *testing.T, store *dstore.Store) {
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
			output: clientio.RespZero,
		},
		"key exists": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "mock_value"
				obj := &object.Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)
			},
			input:  []string{"EXISTING_KEY", strconv.FormatInt(1, 10)},
			output: clientio.RespOne,
		},
		"invalid expiry time exists - very large integer": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "mock_value"
				obj := &object.Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)

			},
			input:  []string{"EXISTING_KEY", strconv.FormatInt(9223372036854776, 10)},
			output: []byte("-ERR invalid expire time in 'expire' command\r\n"),
		},

		"invalid expiry time exists - negative integer": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "mock_value"
				obj := &object.Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)

			},
			input:  []string{"EXISTING_KEY", strconv.FormatInt(-1, 10)},
			output: []byte("-ERR invalid expire time in 'expire' command\r\n"),
		},
	}

	runEvalTests(t, tests, evalEXPIRE, store)
}

func testEvalEXPIRETIME(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"wrong number of args": {
			input:  []string{"KEY1", "KEY2"},
			output: []byte("-ERR wrong number of arguments for 'expiretime' command\r\n"),
		},
		"key does not exist": {
			input:  []string{"NONEXISTENT_KEY"},
			output: clientio.RespMinusTwo,
		},
		"key exists without expiry": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "mock_value"
				obj := &object.Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)
			},
			input:  []string{"EXISTING_KEY"},
			output: clientio.RespMinusOne,
		},
		"key exists with expiry": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "mock_value"
				obj := &object.Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)

				store.SetUnixTimeExpiry(obj, 2724123456123)
			},
			input:  []string{"EXISTING_KEY"},
			output: []byte(fmt.Sprintf(":%d\r\n", 2724123456123)),
		},
	}

	runEvalTests(t, tests, evalEXPIRETIME, store)
}

func testEvalEXPIREAT(t *testing.T, store *dstore.Store) {
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
			output: clientio.RespZero,
		},
		"key exists": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "mock_value"
				obj := &object.Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)
			},
			input:  []string{"EXISTING_KEY", strconv.FormatInt(time.Now().Add(2*time.Minute).Unix(), 10)},
			output: clientio.RespOne,
		},
		"invalid expire time - very large integer": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "mock_value"
				obj := &object.Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)

			},
			input:  []string{"EXISTING_KEY", strconv.FormatInt(9223372036854776, 10)},
			output: []byte("-ERR invalid expire time in 'expireat' command\r\n"),
		},
		"invalid expire time - negative integer": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "mock_value"
				obj := &object.Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)

			},
			input:  []string{"EXISTING_KEY", strconv.FormatInt(-1, 10)},
			output: []byte("-ERR invalid expire time in 'expireat' command\r\n"),
		},
	}

	runEvalTests(t, tests, evalEXPIREAT, store)
}

func testEvalJSONARRINSERT(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"nil value": {
			setup:  func() {},
			input:  nil,
			output: []byte("-ERR wrong number of arguments for 'json.arrinsert' command\r\n"),
		},
		"key does not exist": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"a\":2}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"NONEXISTENT_KEY", "$.a", "0", "1"},
			output: []byte("-ERR could not perform this operation on a key that doesn't exist\r\n"),
		},
		"index is not integer": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"a\":2}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"EXISTING_KEY", "$.a", "a", "1"},
			output: []byte("-ERR Couldn't parse as integer\r\n"),
		},
		"index out of bounds": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "[1,2,3]"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"EXISTING_KEY", "$", "4", "\"a\"", "1"},
			output: []byte("-ERR index out of bounds\r\n"),
		},
		"root path is not array": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"a\":2}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"EXISTING_KEY", "$.a", "0", "6"},
			output: []byte("*1\r\n$-1\r\n"),
		},
		"root path is array": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "[1,2]"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"EXISTING_KEY", "$", "0", "6", "\"a\"", "3.14"},
			output: []byte("*1\r\n:5\r\n"),
		},
		"subpath array insert positive index": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"connection\":{\"wireless\":true,\"names\":[\"1\",\"2\"]},\"price\":99.98,\"names\":[3,4]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"EXISTING_KEY", "$..names", "2", "7", "8"},
			output: []byte("*2\r\n:4\r\n:4\r\n"),
		},
		"subpath array insert negative index": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"connection\":{\"wireless\":true,\"names\":[\"1\",\"2\"]},\"price\":99.98,\"names\":[3,4]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"EXISTING_KEY", "$..names", "-1", "7", "8"},
			output: []byte("*2\r\n:4\r\n:4\r\n"),
		},
		"array insert with multitype value": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"a\":[1,2,3]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"EXISTING_KEY", "$.a", "0", "1", "null", "3.14", "true", "{\"a\":123}"},
			output: []byte("*1\r\n:8\r\n"),
		},
	}
	runEvalTests(t, tests, evalJSONARRINSERT, store)
}

func testEvalJSONARRLEN(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"nil value": {
			setup:  func() {},
			input:  nil,
			output: []byte("-ERR wrong number of arguments for 'json.arrlen' command\r\n"),
		},
		"key does not exist": {
			setup:  func() {},
			input:  []string{"NONEXISTENT_KEY"},
			output: []byte("-ERR Path '.' does not exist or not an array\r\n"),
		},
		"root not array arrlen": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"age\":13,\"name\":\"a\"}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"EXISTING_KEY"},
			output: []byte("-ERR Path '.' does not exist or not an array\r\n"),
		},
		"root array arrlen": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "[1,2,3]"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"EXISTING_KEY"},
			output: []byte(":3\r\n"),
		},
		"wildcase no array arrlen": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"age\":13,\"high\":1.60,\"pet\":null,\"flag\":false, \"partner\":{\"name\":\"tom\"}}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},

			input:  []string{"EXISTING_KEY", "$.*"},
			output: []byte("*5\r\n$-1\r\n$-1\r\n$-1\r\n$-1\r\n$-1\r\n"),
		},
		"subpath array arrlen": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"age\":13,\"high\":1.60,\"pet\":null,\"language\":[\"python\",\"golang\"], " +
					"\"flag\":false, \"partner\":{\"name\":\"tom\"}}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},

			input:  []string{"EXISTING_KEY", "$.language"},
			output: []byte("*1\r\n:2\r\n"),
		},
	}
	runEvalTests(t, tests, evalJSONARRLEN, store)
}

func testEvalJSONOBJLEN(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"nil value": {
			setup:  func() {},
			input:  nil,
			output: []byte("-ERR wrong number of arguments for 'json.objlen' command\r\n"),
		},
		"empty args": {
			setup:  func() {},
			input:  []string{},
			output: []byte("-ERR wrong number of arguments for 'json.objlen' command\r\n"),
		},
		"key does not exist": {
			setup:  func() {},
			input:  []string{"NONEXISTENT_KEY"},
			output: clientio.RespNIL,
		},
		"root not object": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "[1,2,3]"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"EXISTING_KEY"},
			output: []byte("-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"),
		},
		"root object objlen": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"name\":\"John\",\"age\":30,\"city\":\"New York\"}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"EXISTING_KEY"},
			output: []byte(":3\r\n"),
		},
		"wildcard no object objlen": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"name\":\"John\",\"age\":30,\"pets\":null,\"languages\":[\"python\",\"golang\"],\"flag\":false}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"EXISTING_KEY", "$.*"},
			output: []byte("*5\r\n$-1\r\n$-1\r\n$-1\r\n$-1\r\n$-1\r\n"),
		},
		"subpath object objlen": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"person\":{\"name\":\"John\",\"age\":30},\"languages\":[\"python\",\"golang\"]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"EXISTING_KEY", "$.person"},
			output: []byte("*1\r\n:2\r\n"),
		},
		"invalid JSONPath": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"name\":\"John\",\"age\":30}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"EXISTING_KEY", "$invalid_path"},
			output: []byte("-ERR parse error at 2 in $invalid_path\r\n"),
		},
		"incomapitable type(int) objlen": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"person\":{\"name\":\"John\",\"age\":30},\"languages\":[\"python\",\"golang\"]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"EXISTING_KEY", "$.person.age"},
			output: []byte("*1\r\n$-1\r\n"),
		},
		"incomapitable type(string) objlen": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"person\":{\"name\":\"John\",\"age\":30},\"languages\":[\"python\",\"golang\"]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"EXISTING_KEY", "$.person.name"},
			output: []byte("*1\r\n$-1\r\n"),
		},
		"incomapitable type(array) objlen": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"person\":{\"name\":\"John\",\"age\":30},\"languages\":[\"python\",\"golang\"]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"EXISTING_KEY", "$.languages"},
			output: []byte("*1\r\n$-1\r\n"),
		},
	}

	runEvalTests(t, tests, evalJSONOBJLEN, store)
}

func BenchmarkEvalJSONOBJLEN(b *testing.B) {
	sizes := []int{0, 10, 100, 1000, 10000, 100000} // Various sizes of JSON objects
	store := dstore.NewStore(nil)

	for _, size := range sizes {
		b.Run(fmt.Sprintf("JSONObjectSize_%d", size), func(b *testing.B) {
			key := fmt.Sprintf("benchmark_json_obj_%d", size)

			// Create a large JSON object with the given size
			jsonObj := make(map[string]interface{})
			for i := 0; i < size; i++ {
				jsonObj[fmt.Sprintf("key%d", i)] = fmt.Sprintf("value%d", i)
			}

			// Set the JSON object in the store
			args := []string{key, "$", fmt.Sprintf("%v", jsonObj)}
			evalJSONSET(args, store)

			b.ResetTimer()
			b.ReportAllocs()

			// Benchmark the evalJSONOBJLEN function
			for i := 0; i < b.N; i++ {
				_ = evalJSONOBJLEN([]string{key, "$"}, store)
			}
		})
	}
}

func testEvalJSONDEL(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"nil value": {
			setup:  func() {},
			input:  nil,
			output: []byte("-ERR wrong number of arguments for 'json.del' command\r\n"),
		},
		"key does not exist": {
			setup:  func() {},
			input:  []string{"NONEXISTENT_KEY"},
			output: clientio.RespZero,
		},
		"root path del": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"age\":13,\"high\":1.60,\"pet\":null,\"language\":[\"python\",\"golang\"], " +
					"\"flag\":false, \"partner\":{\"name\":\"tom\",\"language\":[\"rust\"]}}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"EXISTING_KEY"},
			output: clientio.RespOne,
		},
		"part path del": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"age\":13,\"high\":1.60,\"pet\":null,\"language\":[\"python\",\"golang\"], " +
					"\"flag\":false, \"partner\":{\"name\":\"tom\",\"language\":[\"rust\"]}}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},

			input:  []string{"EXISTING_KEY", "$..language"},
			output: []byte(":2\r\n"),
		},
		"wildcard path del": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"age\":13,\"high\":1.60,\"pet\":null,\"language\":[\"python\",\"golang\"], " +
					"\"flag\":false, \"partner\":{\"name\":\"tom\",\"language\":[\"rust\"]}}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},

			input:  []string{"EXISTING_KEY", "$.*"},
			output: []byte(":6\r\n"),
		},
	}
	runEvalTests(t, tests, evalJSONDEL, store)
}

func testEvalJSONFORGET(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"nil value": {
			setup:  func() {},
			input:  nil,
			output: []byte("-ERR wrong number of arguments for 'json.forget' command\r\n"),
		},
		"key does not exist": {
			setup:  func() {},
			input:  []string{"NONEXISTENT_KEY"},
			output: clientio.RespZero,
		},
		"root path forget": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"age\":13,\"high\":1.60,\"pet\":null,\"language\":[\"python\",\"golang\"], " +
					"\"flag\":false, \"partner\":{\"name\":\"tom\",\"language\":[\"rust\"]}}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"EXISTING_KEY"},
			output: clientio.RespOne,
		},
		"part path forget": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"age\":13,\"high\":1.60,\"pet\":null,\"language\":[\"python\",\"golang\"], " +
					"\"flag\":false, \"partner\":{\"name\":\"tom\",\"language\":[\"rust\"]}}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},

			input:  []string{"EXISTING_KEY", "$..language"},
			output: []byte(":2\r\n"),
		},
		"wildcard path forget": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"age\":13,\"high\":1.60,\"pet\":null,\"language\":[\"python\",\"golang\"], " +
					"\"flag\":false, \"partner\":{\"name\":\"tom\",\"language\":[\"rust\"]}}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},

			input:  []string{"EXISTING_KEY", "$.*"},
			output: []byte(":6\r\n"),
		},
	}
	runEvalTests(t, tests, evalJSONFORGET, store)
}

func testEvalJSONCLEAR(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"nil value": {
			setup:  func() {},
			input:  nil,
			output: []byte("-ERR wrong number of arguments for 'json.clear' command\r\n"),
		},
		"empty array": {
			setup:  func() {},
			input:  []string{},
			output: []byte("-ERR wrong number of arguments for 'json.clear' command\r\n"),
		},
		"key does not exist": {
			setup:  func() {},
			input:  []string{"NONEXISTENT_KEY"},
			output: []byte("-ERR could not perform this operation on a key that doesn't exist\r\n"),
		},
		"root clear": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"age\":13,\"high\":1.60,\"language\":[\"python\",\"golang\"], \"flag\":false, " +
					"\"partner\":{\"name\":\"tom\",\"language\":[\"rust\"]}}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"EXISTING_KEY"},
			output: []byte(":1\r\n"),
		},
		"array type clear": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"array\":[1,2,3,\"s\",null]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},

			input:  []string{"EXISTING_KEY"},
			output: []byte(":1\r\n"),
		},
		"string type clear": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"a\":\"test\"}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},

			input:  []string{"EXISTING_KEY", "$.a"},
			output: []byte(":0\r\n"),
		},
		"integer type clear": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"age\":13}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},

			input:  []string{"EXISTING_KEY", "$.age"},
			output: []byte(":1\r\n"),
		},
		"number type clear": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"price\":3.14}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},

			input:  []string{"EXISTING_KEY", "$.price"},
			output: []byte(":1\r\n"),
		},
		"boolean type clear": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"flag\":false}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"EXISTING_KEY", "$.flag"},
			output: []byte(":0\r\n"),
		},
		"multi type clear": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"age\":13,\"high\":1.60,\"name\":\"jerry\",\"language\":[\"python\",\"golang\"]," +
					"\"partner\":{\"name\":\"tom\",\"language\":[\"rust\"]}}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"EXISTING_KEY", "$.*"},
			output: []byte(":4\r\n"),
		},
	}
	runEvalTests(t, tests, evalJSONCLEAR, store)
}

func testEvalJSONTYPE(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"nil value": {
			setup:  func() {},
			input:  nil,
			output: []byte("-ERR wrong number of arguments for 'json.type' command\r\n"),
		},
		"empty array": {
			setup:  func() {},
			input:  []string{},
			output: []byte("-ERR wrong number of arguments for 'json.type' command\r\n"),
		},
		"key does not exist": {
			setup:  func() {},
			input:  []string{"NONEXISTENT_KEY"},
			output: clientio.RespNIL,
		},
		"object type value": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"language\":[\"java\",\"go\",\"python\"]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
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
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},

			input:  []string{"EXISTING_KEY", "$.language"},
			output: []byte("*1\r\n$5\r\narray\r\n"),
		},
		"string type value": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"a\":\"test\"}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},

			input:  []string{"EXISTING_KEY", "$.a"},
			output: []byte("*1\r\n$6\r\nstring\r\n"),
		},
		"boolean type value": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"flag\":true}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},

			input:  []string{"EXISTING_KEY", "$.flag"},
			output: []byte("*1\r\n$7\r\nboolean\r\n"),
		},
		"number type value": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"price\":3}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},

			input:  []string{"EXISTING_KEY", "$.price"},
			output: []byte("*1\r\n$6\r\nnumber\r\n"),
		},
		"null type value": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"price\":3.14}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},

			input:  []string{"EXISTING_KEY", "$.language"},
			output: clientio.RespEmptyArray,
		},
		"multi type value": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"name\":\"tom\",\"partner\":{\"name\":\"jerry\"}}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},

			input:  []string{"EXISTING_KEY", "$..name"},
			output: []byte("*2\r\n$6\r\nstring\r\n$6\r\nstring\r\n"),
		},
	}

	runEvalTests(t, tests, evalJSONTYPE, store)
}

func testEvalJSONGET(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"nil value": {
			setup:  func() {},
			input:  nil,
			output: []byte("-ERR wrong number of arguments for 'json.get' command\r\n"),
		},
		"empty array": {
			setup:  func() {},
			input:  []string{},
			output: []byte("-ERR wrong number of arguments for 'json.get' command\r\n"),
		},
		"key does not exist": {
			setup:  func() {},
			input:  []string{"NONEXISTENT_KEY"},
			output: clientio.RespNIL,
		},
		"key exists invalid value": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"a\":2}"
				obj := &object.Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)
			},
			input:  []string{"EXISTING_KEY"},
			output: []byte("-ERR Existing key has wrong Dice type\r\n"),
		},
		"key exists value": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"a\":2}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},

			input:  []string{"EXISTING_KEY"},
			output: []byte("$7\r\n{\"a\":2}\r\n"),
		},
		"key exists but expired": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "mock_value"
				obj := &object.Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)

				store.SetExpiry(obj, int64(-2*time.Millisecond))
			},
			input:  []string{"EXISTING_KEY"},
			output: clientio.RespNIL,
		},
	}

	runEvalTests(t, tests, evalJSONGET, store)
}

func testEvalJSONSET(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"nil value": {
			setup:  func() {},
			input:  nil,
			output: []byte("-ERR wrong number of arguments for 'json.set' command\r\n"),
		},
		"empty array": {
			setup:  func() {},
			input:  []string{},
			output: []byte("-ERR wrong number of arguments for 'json.set' command\r\n"),
		},
		"insufficient args": {
			setup:  func() {},
			input:  []string{},
			output: []byte("-ERR wrong number of arguments for 'json.set' command\r\n"),
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
			output: clientio.RespOK,
		},
	}

	runEvalTests(t, tests, evalJSONSET, store)
}

func testEvalJSONNUMMULTBY(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"nil value": {
			setup:  func() {},
			input:  nil,
			output: []byte("-ERR wrong number of arguments for 'json.nummultby' command\r\n"),
		},
		"empty array": {
			setup:  func() {},
			input:  []string{},
			output: []byte("-ERR wrong number of arguments for 'json.nummultby' command\r\n"),
		},
		"insufficient args": {
			setup:  func() {},
			input:  []string{"doc"},
			output: []byte("-ERR wrong number of arguments for 'json.nummultby' command\r\n"),
		},
		"non-numeric multiplier on existing key": {
			setup: func() {
				key := "doc"
				value := "{\"a\":10,\"b\":[{\"a\":2}, {\"a\":5}, {\"a\":\"c\"}]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"doc", "$.a", "qwe"},
			output: []byte("-ERR expected value at line 1 column 1\r\n"),
		},
		"nummultby on non integer root fields": {
			setup: func() {
				key := "doc"
				value := "{\"a\": \"b\",\"b\":[{\"a\":2}, {\"a\":5}, {\"a\":\"c\"}]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"doc", "$.a", "2"},
			output: []byte("$6\r\n[null]\r\n"),
		},
		"nummultby on recursive fields": {
			setup: func() {
				key := "doc"
				value := "{\"a\": \"b\",\"b\":[{\"a\":2}, {\"a\":5}, {\"a\":\"c\"}]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"doc", "$..a", "2"},
			output: []byte("$16\r\n[4,10,null,null]\r\n"),
		},
		"nummultby on integer root fields": {
			setup: func() {
				key := "doc"
				value := "{\"a\":10,\"b\":[{\"a\":2}, {\"a\":5}, {\"a\":\"c\"}]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"doc", "$.a", "2"},
			output: []byte("$4\r\n[20]\r\n"),
		},
		"nummultby on non-existent key": {
			setup: func() {
				key := "doc"
				value := "{\"a\":10,\"b\":[{\"a\":2}, {\"a\":5}, {\"a\":\"c\"}]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"doc", "$..fe", "2"},
			output: []byte("$2\r\n[]\r\n"),
		},
	}
	runEvalTests(t, tests, evalJSONNUMMULTBY, store)
}

func testEvalJSONARRAPPEND(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"arr append to non array fields": {
			setup: func() {
				key := "array"
				value := "{\"a\":2}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"array", "$.a", "6"},
			output: []byte("*1\r\n$-1\r\n"),
		},
		"arr append single element to an array field": {
			setup: func() {
				key := "array"
				value := "{\"a\":[1,2]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"array", "$.a", "6"},
			output: []byte("*1\r\n:3\r\n"),
		},
		"arr append multiple elements to an array field": {
			setup: func() {
				key := "array"
				value := "{\"a\":[1,2]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"array", "$.a", "6", "7", "8"},
			output: []byte("*1\r\n:5\r\n"),
		},
		"arr append string value": {
			setup: func() {
				key := "array"
				value := "{\"b\":[\"b\",\"c\"]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"array", "$.b", `"d"`},
			output: []byte("*1\r\n:3\r\n"),
		},
		"arr append nested array value": {
			setup: func() {
				key := "array"
				value := "{\"a\":[[1,2]]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"array", "$.a", "[1,2,3]"},
			output: []byte("*1\r\n:2\r\n"),
		},
		"arr append with json value": {
			setup: func() {
				key := "array"
				value := "{\"a\":[{\"b\": 1}]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"array", "$.a", "{\"c\": 3}"},
			output: []byte("*1\r\n:2\r\n"),
		},
		"arr append to append on multiple fields": {
			setup: func() {
				key := "array"
				value := "{\"a\":[1,2],\"b\":{\"a\":[10]}}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"array", "$..a", "6"},
			output: []byte("*2\r\n:2\r\n:3\r\n"),
		},
		"arr append to append on root node": {
			setup: func() {
				key := "array"
				value := "[1,2,3]"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"array", "$", "6"},
			output: []byte("*1\r\n:4\r\n"),
		},
		"arr append to an array with different type": {
			setup: func() {
				key := "array"
				value := "{\"a\":[1,2]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"array", "$.a", `"blue"`},
			output: []byte("*1\r\n:3\r\n"),
		},
	}
	runEvalTests(t, tests, evalJSONARRAPPEND, store)
}

func testEvalJSONTOGGLE(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"nil value": {
			setup:  func() {},
			input:  nil,
			output: []byte("-ERR wrong number of arguments for 'json.toggle' command\r\n"),
		},
		"empty array": {
			setup:  func() {},
			input:  []string{},
			output: []byte("-ERR wrong number of arguments for 'json.toggle' command\r\n"),
		},
		"key does not exist": {
			setup:  func() {},
			input:  []string{"NONEXISTENT_KEY", ".active"},
			output: []byte("-ERR could not perform this operation on a key that doesn't exist\r\n"),
		},
		"key exists, toggling boolean true to false": {
			setup: func() {
				key := "EXISTING_KEY"
				value := `{"active":true}`
				var rootData interface{}
				err := sonic.Unmarshal([]byte(value), &rootData)
				if err != nil {
					fmt.Printf("Debug: Error unmarshaling JSON: %v\n", err)
				}
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)

			},
			input:  []string{"EXISTING_KEY", ".active"},
			output: clientio.Encode([]interface{}{0}, false),
		},
		"key exists, toggling boolean false to true": {
			setup: func() {
				key := "EXISTING_KEY"
				value := `{"active":false}`
				var rootData interface{}
				err := sonic.Unmarshal([]byte(value), &rootData)
				if err != nil {
					fmt.Printf("Debug: Error unmarshaling JSON: %v\n", err)
				}
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"EXISTING_KEY", ".active"},
			output: clientio.Encode([]interface{}{1}, false),
		},
		"key exists but expired": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"active\":true}"
				obj := &object.Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)
				store.SetExpiry(obj, int64(-2*time.Millisecond))
			},
			input:  []string{"EXISTING_KEY", ".active"},
			output: []byte("-ERR could not perform this operation on a key that doesn't exist\r\n"),
		},
		"nested JSON structure with multiple booleans": {
			setup: func() {
				key := "NESTED_KEY"
				value := `{"isSimple":true,"nested":{"isSimple":false}}`
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"NESTED_KEY", "$..isSimple"},
			output: clientio.Encode([]interface{}{0, 1}, false),
		},
		"deeply nested JSON structure with multiple matching fields": {
			setup: func() {
				key := "DEEP_NESTED_KEY"
				value := `{"field": true, "nested": {"field": false, "nested": {"field": true}}}`
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"DEEP_NESTED_KEY", "$..field"},
			output: clientio.Encode([]interface{}{0, 1, 0}, false),
		},
	}
	runEvalTests(t, tests, evalJSONTOGGLE, store)
}

func testEvalTTL(t *testing.T, store *dstore.Store) {
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
			output: clientio.RespMinusTwo,
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
				obj := &object.Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)
			},
			input:  []string{"EXISTING_KEY"},
			output: clientio.RespMinusOne,
		},
		"key exists not expired": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "mock_value"
				obj := &object.Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)

				store.SetExpiry(obj, int64(2*time.Millisecond))
			},
			input: []string{"EXISTING_KEY"},
			validator: func(output []byte) {
				assert.Assert(t, output != nil)
				assert.Assert(t, !bytes.Equal(output, clientio.RespMinusOne))
				assert.Assert(t, !bytes.Equal(output, clientio.RespMinusTwo))
			},
		},
		"key exists but expired": {
			setup: func() {
				key := "EXISTING_EXPIRED_KEY"
				value := "mock_value"
				obj := &object.Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)

				store.SetExpiry(obj, int64(-2*time.Millisecond))
			},
			input:  []string{"EXISTING_KEY"},
			output: clientio.RespMinusTwo,
		},
	}

	runEvalTests(t, tests, evalTTL, store)
}

func testEvalDel(t *testing.T, store *dstore.Store) {
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
				obj := &object.Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)

				dstore.KeyspaceStat[0]["keys"]++
			},
			input:  []string{"EXISTING_KEY"},
			output: []byte(":1\r\n"),
		},
	}

	runEvalTests(t, tests, evalDEL, store)
}

// TestEvalPersist tests the evalPersist function using table-driven tests.
func testEvalPersist(t *testing.T, store *dstore.Store) {
	// Define test cases
	tests := map[string]evalTestCase{
		"wrong number of arguments": {
			input:  []string{"key1", "key2"},
			output: clientio.Encode(errors.New("ERR wrong number of arguments for 'persist' command"), false),
		},
		"key does not exist": {
			input:  []string{"nonexistent"},
			output: clientio.RespZero,
		},
		"key exists but no expiration set": {
			input: []string{"existent_no_expiry"},
			setup: func() {
				evalSET([]string{"existent_no_expiry", "value"}, store)
			},
			output: clientio.RespMinusOne,
		},
		"key exists and expiration removed": {
			input: []string{"existent_with_expiry"},
			setup: func() {
				evalSET([]string{"existent_with_expiry", "value", Ex, "1"}, store)
			},
			output: clientio.RespOne,
		},
		"key exists with expiration set and not expired": {
			input: []string{"existent_with_expiry_not_expired"},
			setup: func() {
				// Simulate setting a key with an expiration time that has not yet passed
				evalSET([]string{"existent_with_expiry_not_expired", "value", Ex, "10000"}, store) // 10000 seconds in the future
			},
			output: clientio.RespOne,
		},
	}

	runEvalTests(t, tests, evalPersist, store)
}

func testEvalDbsize(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"DBSIZE command with invalid no of args": {
			input:  []string{"INVALID_ARG"},
			output: []byte("-ERR wrong number of arguments for 'dbsize' command\r\n"),
		},
		"no key in db": {
			input:  nil,
			output: []byte(":0\r\n"),
		},
		"one key exists in db": {
			setup: func() {
				evalSET([]string{"key", "val"}, store)
			},
			input:  nil,
			output: []byte(":1\r\n"),
		},
		"two keys exist in db": {
			setup: func() {
				evalSET([]string{"key1", "val1"}, store)
				evalSET([]string{"key2", "val2"}, store)
			},
			input:  nil,
			output: []byte(":2\r\n"),
		},
		"repeating keys shall result in same dbsize": {
			setup: func() {
				evalSET([]string{"key1", "val1"}, store)
				evalSET([]string{"key2", "val2"}, store)
				evalSET([]string{"key2", "val2"}, store)
			},
			input:  nil,
			output: []byte(":2\r\n"),
		},
		"deleted keys shall be reflected in dbsize": {
			setup: func() {
				evalSET([]string{"key1", "val1"}, store)
				evalSET([]string{"key2", "val2"}, store)
				evalDEL([]string{"key2"}, store)
			},
			input:  nil,
			output: []byte(":1\r\n"),
		},
	}

	runEvalTests(t, tests, evalDBSIZE, store)
}

func testEvalPFADD(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"nil value":           {input: nil, output: []byte("-ERR wrong number of arguments for 'pfadd' command\r\n")},
		"empty array":         {input: []string{}, output: []byte("-ERR wrong number of arguments for 'pfadd' command\r\n")},
		"one value":           {input: []string{"KEY"}, output: []byte(":1\r\n")},
		"key val pair":        {input: []string{"KEY", "VAL"}, output: []byte(":1\r\n")},
		"key multiple values": {input: []string{"KEY", "VAL", "VAL1", "VAL2"}, output: []byte(":1\r\n")},
		"Incorrect type provided": {
			setup: func() {
				key, value := "EXISTING_KEY", "VALUE"
				oType, oEnc := deduceTypeEncoding(value)
				var exDurationMs int64 = -1
				var keepttl bool = false

				store.Put(key, store.NewObj(value, exDurationMs, oType, oEnc), dstore.WithKeepTTL(keepttl))
			},
			input:  []string{"EXISTING_KEY", "1"},
			output: []byte("-WRONGTYPE Key is not a valid HyperLogLog string value.\r\n")},
	}

	runEvalTests(t, tests, evalPFADD, store)
}

func testEvalPFCOUNT(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"PFCOUNT with empty arg": {
			input:  []string{},
			output: []byte("-ERR wrong number of arguments for 'pfcount' command\r\n"),
		},
		"PFCOUNT key not exists": {
			input:  []string{"HELLO"},
			output: clientio.Encode(0, false),
		},
		"PFCOUNT key exists": {
			setup: func() {
				key := "EXISTING_KEY"
				value := hyperloglog.New()
				value.Insert([]byte("VALUE"))
				obj := &object.Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)
			},
			input:  []string{"EXISTING_KEY"},
			output: clientio.Encode(1, false),
		},
	}

	runEvalTests(t, tests, evalPFCOUNT, store)
}

func testEvalHGET(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"wrong number of args passed": {
			setup:  func() {},
			input:  nil,
			output: []byte("-ERR wrong number of arguments for 'hget' command\r\n"),
		},
		"only key passed": {
			setup:  func() {},
			input:  []string{"KEY"},
			output: []byte("-ERR wrong number of arguments for 'hget' command\r\n"),
		},
		"key doesn't exists": {
			setup:  func() {},
			input:  []string{"KEY", "field_name"},
			output: clientio.RespNIL,
		},
		"key exists but field_name doesn't exists": {
			setup: func() {
				key := "KEY_MOCK"
				field := "mock_field_name"
				newMap := make(HashMap)
				newMap[field] = "mock_field_value"

				obj := &object.Obj{
					TypeEncoding:   object.ObjTypeHashMap | object.ObjEncodingHashMap,
					Value:          newMap,
					LastAccessedAt: uint32(time.Now().Unix()),
				}

				store.Put(key, obj)
			},
			input:  []string{"KEY_MOCK", "non_existent_key"},
			output: clientio.RespNIL,
		},
		"both key and field_name exists": {
			setup: func() {
				key := "KEY_MOCK"
				field := "mock_field_name"
				newMap := make(HashMap)
				newMap[field] = "mock_field_value"

				obj := &object.Obj{
					TypeEncoding:   object.ObjTypeHashMap | object.ObjEncodingHashMap,
					Value:          newMap,
					LastAccessedAt: uint32(time.Now().Unix()),
				}

				store.Put(key, obj)
			},
			input:  []string{"KEY_MOCK", "mock_field_name"},
			output: clientio.Encode("mock_field_value", false),
		},
	}

	runEvalTests(t, tests, evalHGET, store)
}

func testEvalHSTRLEN(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"wrong number of args passed": {
			setup:  func() {},
			input:  nil,
			output: []byte("-ERR wrong number of arguments for 'hstrlen' command\r\n"),
		},
		"only key passed": {
			setup:  func() {},
			input:  []string{"KEY"},
			output: []byte("-ERR wrong number of arguments for 'hstrlen' command\r\n"),
		},
		"key doesn't exist": {
			setup:  func() {},
			input:  []string{"KEY", "field_name"},
			output: clientio.Encode(0, false),
		},
		"key exists but field_name doesn't exists": {
			setup: func() {
				key := "KEY_MOCK"
				field := "mock_field_name"
				newMap := make(HashMap)
				newMap[field] = "mock_field_value"

				obj := &object.Obj{
					TypeEncoding:   object.ObjTypeHashMap | object.ObjEncodingHashMap,
					Value:          newMap,
					LastAccessedAt: uint32(time.Now().Unix()),
				}

				store.Put(key, obj)
			},
			input:  []string{"KEY_MOCK", "non_existent_key"},
			output: clientio.Encode(0, false),
		},
		"both key and field_name exists": {
			setup: func() {
				key := "KEY_MOCK"
				field := "mock_field_name"
				newMap := make(HashMap)
				newMap[field] = "HelloWorld"

				obj := &object.Obj{
					TypeEncoding:   object.ObjTypeHashMap | object.ObjEncodingHashMap,
					Value:          newMap,
					LastAccessedAt: uint32(time.Now().Unix()),
				}

				store.Put(key, obj)
			},
			input:  []string{"KEY_MOCK", "mock_field_name"},
			output: clientio.Encode(10, false),
		},
	}

	runEvalTests(t, tests, evalHSTRLEN, store)
}

func testEvalHDEL(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"HDEL with wrong number of args": {
			input:  []string{"key"},
			output: []byte("-ERR wrong number of arguments for 'hdel' command\r\n"),
		},
		"HDEL with key does not exist": {
			input:  []string{"nonexistent", "field"},
			output: clientio.RespZero,
		},
		"HDEL with key exists but not a hash": {
			setup: func() {
				evalSET([]string{"string_key", "string_value"}, store)
			},
			input:  []string{"string_key", "field"},
			output: []byte("-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"),
		},
		"HDEL with delete existing fields": {
			setup: func() {
				evalHSET([]string{"hash_key", "field1", "value1", "field2", "value2"}, store)
			},
			input:  []string{"hash_key", "field1", "field2", "nonexistent"},
			output: clientio.Encode(int64(2), false),
		},
		"HDEL with delete non-existing fields": {
			setup: func() {
				evalHSET([]string{"hash_key", "field1", "value1"}, store)
			},
			input:  []string{"hash_key", "nonexistent1", "nonexistent2"},
			output: clientio.RespZero,
		},
	}

	runEvalTests(t, tests, evalHDEL, store)
}

func testEvalPFMERGE(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"nil value":   {input: nil, output: []byte("-ERR wrong number of arguments for 'pfmerge' command\r\n")},
		"empty array": {input: []string{}, output: []byte("-ERR wrong number of arguments for 'pfmerge' command\r\n")},
		"PFMERGE invalid hll object": {
			setup: func() {
				key := "INVALID_OBJ_DEST_KEY"
				value := "123"
				obj := &object.Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)
			},
			input:  []string{"INVALID_OBJ_DEST_KEY"},
			output: []byte("-WRONGTYPE Key is not a valid HyperLogLog string value.\r\n"),
		},
		"PFMERGE destKey doesn't exist": {
			input:  []string{"NON_EXISTING_DEST_KEY"},
			output: clientio.RespOK,
		},
		"PFMERGE destKey exist": {
			input:  []string{"NON_EXISTING_DEST_KEY"},
			output: clientio.RespOK,
		},
		"PFMERGE destKey exist srcKey doesn't exists": {
			setup: func() {
				key := "EXISTING_DEST_KEY"
				value := hyperloglog.New()
				value.Insert([]byte("VALUE"))
				obj := &object.Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)
			},
			input:  []string{"EXISTING_DEST_KEY", "NON_EXISTING_SRC_KEY"},
			output: clientio.RespOK,
		},
		"PFMERGE destKey exist srcKey exists": {
			setup: func() {
				key := "EXISTING_DEST_KEY"
				value := hyperloglog.New()
				value.Insert([]byte("VALUE"))
				obj := &object.Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)
			},
			input:  []string{"EXISTING_DEST_KEY", "NON_EXISTING_SRC_KEY"},
			output: clientio.RespOK,
		},
		"PFMERGE destKey exist multiple srcKey exist": {
			setup: func() {
				key := "EXISTING_DEST_KEY"
				value := hyperloglog.New()
				value.Insert([]byte("VALUE"))
				obj := &object.Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)
				srcKey := "EXISTING_SRC_KEY"
				srcValue := hyperloglog.New()
				value.Insert([]byte("SRC_VALUE"))
				srcKeyObj := &object.Obj{
					Value:          srcValue,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(srcKey, srcKeyObj)
			},
			input:  []string{"EXISTING_DEST_KEY", "EXISTING_SRC_KEY"},
			output: clientio.RespOK,
		},
	}

	runEvalTests(t, tests, evalPFMERGE, store)
}

func testEvalJSONSTRLEN(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"nil value": {
			setup:  func() {},
			input:  nil,
			output: []byte("-ERR wrong number of arguments for 'json.strlen' command\r\n"),
		},
		"key does not exist": {
			setup:  func() {},
			input:  []string{"NONEXISTENT_KEY"},
			output: []byte("$-1\r\n"),
		},
		"root not string strlen": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"age\":13,\"name\":\"a\"}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"EXISTING_KEY"},
			output: []byte("-WRONGTYPE wrong type of path value - expected string but found integer\r\n"),
		},
		"root array strlen": {
			setup: func() {
				key := "EXISTING_KEY"
				value := `"hello"`
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"EXISTING_KEY"},
			output: []byte(":5\r\n"),
		},
		"subpath string strlen": {
			setup: func() {
				key := "EXISTING_KEY"
				value := `{"partner":{"name":"tom","language":["rust"]}}`
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},

			input:  []string{"EXISTING_KEY", "$..name"},
			output: []byte("*1\r\n:3\r\n"),
		},
		"subpath not string strlen": {
			setup: func() {
				key := "EXISTING_KEY"
				value := `{"partner":{"name":21,"language":["rust"]}}`
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},

			input:  []string{"EXISTING_KEY", "$..name"},
			output: []byte("*1\r\n$-1\r\n"),
		},
	}
	runEvalTests(t, tests, evalJSONSTRLEN, store)
}

func testEvalLLEN(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"nil value": {
			input:  nil,
			output: []byte("-ERR wrong number of arguments for 'llen' command\r\n"),
		},
		"empty args": {
			input:  []string{},
			output: []byte("-ERR wrong number of arguments for 'llen' command\r\n"),
		},
		"wrong number of args": {
			input:  []string{"KEY1", "KEY2"},
			output: []byte("-ERR wrong number of arguments for 'llen' command\r\n"),
		},
		"key does not exist": {
			input:  []string{"NONEXISTENT_KEY"},
			output: clientio.RespZero,
		},
		"key exists": {
			setup: func() {
				evalLPUSH([]string{"EXISTING_KEY", "mock_value"}, store)
			},
			input:  []string{"EXISTING_KEY"},
			output: clientio.RespOne,
		},
		"key with different type": {
			setup: func() {
				evalSET([]string{"EXISTING_KEY", "mock_value"}, store)
			},
			input:  []string{"EXISTING_KEY"},
			output: []byte("-ERR Existing key has wrong Dice type\r\n"),
		},
	}

	runEvalTests(t, tests, evalLLEN, store)
}

func testEvalJSONNUMINCRBY(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"incr on numeric field": {
			setup: func() {
				key := "number"
				value := "{\"a\": 2}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"number", "$.a", "3"},
			output: []byte("$3\r\n[5]\r\n"),
		},

		"incr on float field": {
			setup: func() {
				key := "number"
				value := "{\"a\": 2.5}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"number", "$.a", "1.5"},
			output: []byte("$5\r\n[4.0]\r\n"),
		},

		"incr on multiple fields": {
			setup: func() {
				key := "number"
				value := "{\"a\": 2, \"b\": 10, \"c\": [15, {\"d\": 20}]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"number", "$..*", "5"},
			output: []byte("$22\r\n[25,20,null,7,15,null]\r\n"),
			validator: func(output []byte) {
				outPutString := string(output)
				startIndex := strings.Index(outPutString, "[")
				endIndex := strings.Index(outPutString, "]")
				arrayString := outPutString[startIndex+1 : endIndex]
				arr := strings.Split(arrayString, ",")
				testifyAssert.ElementsMatch(t, arr, []string{"25", "20", "7", "15", "null", "null"})
			},
		},

		"incr on array element": {
			setup: func() {
				key := "number"
				value := "{\"a\": [1, 2, 3]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"number", "$.a[1]", "5"},
			output: []byte("$3\r\n[7]\r\n"),
		},
		"incr on non-existent field": {
			setup: func() {
				key := "number"
				value := "{\"a\": 2}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"number", "$.b", "3"},
			output: []byte("$2\r\n[]\r\n"),
		},
		"incr with mixed fields": {
			setup: func() {
				key := "number"
				value := "{\"a\": 5, \"b\": \"not a number\", \"c\": [1, 2]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"number", "$..*", "2"},
			output: []byte("$17\r\n[3,4,null,7,null]\r\n"),
			validator: func(output []byte) {
				outPutString := string(output)
				startIndex := strings.Index(outPutString, "[")
				endIndex := strings.Index(outPutString, "]")
				arrayString := outPutString[startIndex+1 : endIndex]
				arr := strings.Split(arrayString, ",")
				testifyAssert.ElementsMatch(t, arr, []string{"3", "4", "7", "null", "null"})
			},
		},

		"incr on nested fields": {
			setup: func() {
				key := "number"
				value := "{\"a\": {\"b\": {\"c\": 10}}}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"number", "$..c", "5"},
			output: []byte("$4\r\n[15]\r\n"),
		},
	}

	runEvalTests(t, tests, evalJSONNUMINCRBY, store)
}

func runEvalTests(t *testing.T, tests map[string]evalTestCase, evalFunc func([]string, *dstore.Store) []byte, store *dstore.Store) {
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			store = setupTest(store)

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
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store := dstore.NewStore(nil)
		evalMSET([]string{"KEY", "VAL", "KEY2", "VAL2"}, store)
	}
}

func BenchmarkEvalHSET(b *testing.B) {
	store := dstore.NewStore(nil)
	for i := 0; i < b.N; i++ {
		evalHSET([]string{"KEY", fmt.Sprintf("FIELD_%d", i), fmt.Sprintf("VALUE_%d", i)}, store)
	}
}

func testEvalHSET(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"wrong number of args passed": {
			setup:  func() {},
			input:  nil,
			output: []byte("-ERR wrong number of arguments for 'hset' command\r\n"),
		},
		"only key passed": {
			setup:  func() {},
			input:  []string{"key"},
			output: []byte("-ERR wrong number of arguments for 'hset' command\r\n"),
		},
		"only key and field_name passed": {
			setup:  func() {},
			input:  []string{"KEY", "field_name"},
			output: []byte("-ERR wrong number of arguments for 'hset' command\r\n"),
		},
		"key, field and value passed": {
			setup:  func() {},
			input:  []string{"KEY1", "field_name", "value"},
			output: clientio.Encode(int64(1), false),
		},
		"key, field and value updated": {
			setup:  func() {},
			input:  []string{"KEY1", "field_name", "value_new"},
			output: clientio.Encode(int64(1), false),
		},
		"new set of key, field and value added": {
			setup:  func() {},
			input:  []string{"KEY2", "field_name_new", "value_new_new"},
			output: clientio.Encode(int64(1), false),
		},
		"apply with duplicate key, field and value names": {
			setup: func() {
				key := "KEY_MOCK"
				field := "mock_field_name"
				newMap := make(HashMap)
				newMap[field] = "mock_field_value"

				obj := &object.Obj{
					TypeEncoding:   object.ObjTypeHashMap | object.ObjEncodingHashMap,
					Value:          newMap,
					LastAccessedAt: uint32(time.Now().Unix()),
				}

				store.Put(key, obj)
			},
			input:  []string{"KEY_MOCK", "mock_field_name", "mock_field_value"},
			output: clientio.Encode(int64(0), false),
		},
		"same key -> update value, add new field and value": {
			setup: func() {
				key := "KEY_MOCK"
				field := "mock_field_name"
				mock_value := "mock_field_value"
				newMap := make(HashMap)
				newMap[field] = mock_value

				obj := &object.Obj{
					TypeEncoding:   object.ObjTypeHashMap | object.ObjEncodingHashMap,
					Value:          newMap,
					LastAccessedAt: uint32(time.Now().Unix()),
				}

				store.Put(key, obj)

				// Check if the map is saved correctly in the store
				res, err := getValueFromHashMap(key, field, store)

				assert.Assert(t, err == nil)
				assert.DeepEqual(t, res, clientio.Encode(mock_value, false))
			},
			input: []string{
				"KEY_MOCK",
				"mock_field_name",
				"mock_field_value_new",
				"mock_field_name_new",
				"mock_value_new",
			},
			output: clientio.Encode(int64(1), false),
		},
	}

	runEvalTests(t, tests, evalHSET, store)
}

func BenchmarkEvalPFCOUNT(b *testing.B) {
	store := *dstore.NewStore(nil)

	// Helper function to create and insert HLL objects
	createAndInsertHLL := func(key string, items []string) {
		hll := hyperloglog.New()
		for _, item := range items {
			hll.Insert([]byte(item))
		}
		obj := &object.Obj{
			Value:          hll,
			LastAccessedAt: uint32(time.Now().Unix()),
		}
		store.Put(key, obj)
	}

	// Create small HLLs (10000 items each)
	smallItems := make([]string, 10000)
	for i := 0; i < 100; i++ {
		smallItems[i] = fmt.Sprintf("SmallItem%d", i)
	}
	createAndInsertHLL("SMALL1", smallItems)
	createAndInsertHLL("SMALL2", smallItems)

	// Create medium HLLs (1000000 items each)
	mediumItems := make([]string, 1000000)
	for i := 0; i < 100; i++ {
		mediumItems[i] = fmt.Sprintf("MediumItem%d", i)
	}
	createAndInsertHLL("MEDIUM1", mediumItems)
	createAndInsertHLL("MEDIUM2", mediumItems)

	// Create large HLLs (1000000000 items each)
	largeItems := make([]string, 1000000000)
	for i := 0; i < 10000; i++ {
		largeItems[i] = fmt.Sprintf("LargeItem%d", i)
	}
	createAndInsertHLL("LARGE1", largeItems)
	createAndInsertHLL("LARGE2", largeItems)

	tests := []struct {
		name string
		args []string
	}{
		{"SingleSmallKey", []string{"SMALL1"}},
		{"TwoSmallKeys", []string{"SMALL1", "SMALL2"}},
		{"SingleMediumKey", []string{"MEDIUM1"}},
		{"TwoMediumKeys", []string{"MEDIUM1", "MEDIUM2"}},
		{"SingleLargeKey", []string{"LARGE1"}},
		{"TwoLargeKeys", []string{"LARGE1", "LARGE2"}},
		{"MixedSizes", []string{"SMALL1", "MEDIUM1", "LARGE1"}},
		{"ManySmallKeys", []string{"SMALL1", "SMALL2", "SMALL1", "SMALL2", "SMALL1"}},
		{"ManyMediumKeys", []string{"MEDIUM1", "MEDIUM2", "MEDIUM1", "MEDIUM2", "MEDIUM1"}},
		{"ManyLargeKeys", []string{"LARGE1", "LARGE2", "LARGE1", "LARGE2", "LARGE1"}},
		{"NonExistentKey", []string{"SMALL1", "NONEXISTENT", "LARGE1"}},
		{"AllKeys", []string{"SMALL1", "SMALL2", "MEDIUM1", "MEDIUM2", "LARGE1", "LARGE2"}},
	}

	b.ResetTimer()

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				evalPFCOUNT(tt.args, &store)
			}
		})
	}
}

func testEvalDebug(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{

		// invalid subcommand tests
		"no subcommand passed": {
			setup:  func() {},
			input:  nil,
			output: []byte("-ERR wrong number of arguments for 'json.debug' command\r\n"),
		},

		"wrong subcommand passed": {
			setup:  func() {},
			input:  []string{"WRONG_SUBCOMMAND"},
			output: []byte("-ERR unknown subcommand - try `JSON.DEBUG HELP`\r\n"),
		},

		// help subcommand tests
		"help no args": {
			setup:  func() {},
			input:  []string{"HELP"},
			output: []byte("*2\r\n$42\r\nMEMORY <key> [path] - reports memory usage\r\n$34\r\nHELP                - this message\r\n"),
		},

		"help with args": {
			setup:  func() {},
			input:  []string{"HELP", "EXTRA_ARG"},
			output: []byte("*2\r\n$42\r\nMEMORY <key> [path] - reports memory usage\r\n$34\r\nHELP                - this message\r\n"),
		},

		// memory subcommand tests
		"memory without args": {
			setup:  func() {},
			input:  []string{"MEMORY"},
			output: []byte("-ERR wrong number of arguments for 'json.debug' command\r\n"),
		},

		"memory nonexistant key": {
			setup:  func() {},
			input:  []string{"MEMORY", "NONEXISTANT_KEY"},
			output: clientio.RespZero,
		},

		// memory subcommand tests for existing key
		"no path": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"a\": 1}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"MEMORY", "EXISTING_KEY"},
			output: []byte(":89\r\n"),
		},

		"root path": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"a\": 1}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"MEMORY", "EXISTING_KEY", "$"},
			output: []byte(":89\r\n"),
		},

		"invalid path": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"a\": 1}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"MEMORY", "EXISTING_KEY", "INVALID_PATH"},
			output: []byte("-ERR Path '$.INVALID_PATH' does not exist\r\n"),
		},

		"valid path": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"a\": 1, \"b\": 2}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"MEMORY", "EXISTING_KEY", "$.a"},
			output: []byte("*1\r\n:16\r\n"),
		},

		// only the first path is picked whether it's valid or not for an object json
		// memory can be fetched only for one path in a command for an object json
		"multiple paths for object json": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"a\": 1, \"b\": \"dice\"}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"MEMORY", "EXISTING_KEY", "$.a", "$.b"},
			output: []byte("*1\r\n:16\r\n"),
		},

		"single index path for array json": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "[\"roll\", \"the\", \"dices\"]"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"MEMORY", "EXISTING_KEY", "$[1]"},
			output: []byte("*1\r\n:19\r\n"),
		},

		"multiple index paths for array json": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "[\"roll\", \"the\", \"dices\"]"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"MEMORY", "EXISTING_KEY", "$[1,2]"},
			output: []byte("*2\r\n:19\r\n:21\r\n"),
		},

		"index path out of range for array json": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "[\"roll\", \"the\", \"dices\"]"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"MEMORY", "EXISTING_KEY", "$[4]"},
			output: clientio.RespEmptyArray,
		},

		"multiple valid and invalid index paths": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "[\"roll\", \"the\", \"dices\"]"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"MEMORY", "EXISTING_KEY", "$[1,2,4]"},
			output: []byte("*2\r\n:19\r\n:21\r\n"),
		},

		"negative index path": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "[\"roll\", \"the\", \"dices\"]"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"MEMORY", "EXISTING_KEY", "$[-1]"},
			output: []byte("*1\r\n:21\r\n"),
		},

		"multiple negative indexe paths": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "[\"roll\", \"the\", \"dices\"]"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"MEMORY", "EXISTING_KEY", "$[-1,-2]"},
			output: []byte("*2\r\n:21\r\n:19\r\n"),
		},

		"negative index path out of bound": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "[\"roll\", \"the\", \"dices\"]"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"MEMORY", "EXISTING_KEY", "$[-4]"},
			output: []byte("-ERR Path '$.$[-4]' does not exist\r\n"),
		},

		"all paths with asterix for array json": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "[\"roll\", \"the\", \"dices\"]"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"MEMORY", "EXISTING_KEY", "$[*]"},
			output: []byte("*3\r\n:20\r\n:19\r\n:21\r\n"),
		},

		"all paths with semicolon for array json": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "[\"roll\", \"the\", \"dices\"]"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"MEMORY", "EXISTING_KEY", "$[:]"},
			output: []byte("*3\r\n:20\r\n:19\r\n:21\r\n"),
		},

		"array json with mixed types": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "[2, 3.5, true, null, \"dice\", {}, [], {\"a\": 1, \"b\": 2}, [7, 8, 0]]"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"MEMORY", "EXISTING_KEY", "$[:]"},
			output: []byte("*9\r\n:16\r\n:16\r\n:16\r\n:16\r\n:20\r\n:16\r\n:16\r\n:82\r\n:64\r\n"),
		},
	}

	runEvalTests(t, tests, evalJSONDebug, store)
}

func testEvalHLEN(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"wrong number of args": {
			input:  []string{},
			output: []byte("-ERR wrong number of arguments for 'hlen' command\r\n"),
		},
		"key does not exist": {
			input:  []string{"nonexistent_key"},
			output: clientio.RespZero,
		},
		"key exists but not a hash": {
			setup: func() {
				evalSET([]string{"string_key", "string_value"}, store)
			},
			input:  []string{"string_key"},
			output: []byte("-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"),
		},
		"empty hash": {
			setup:  func() {},
			input:  []string{"empty_hash"},
			output: clientio.RespZero,
		},
		"hash with elements": {
			setup: func() {
				evalHSET([]string{"hash_key", "field1", "value1", "field2", "value2", "field3", "value3"}, store)
			},
			input:  []string{"hash_key"},
			output: clientio.Encode(int64(3), false),
		},
	}

	runEvalTests(t, tests, evalHLEN, store)
}

func BenchmarkEvalHLEN(b *testing.B) {
	sizes := []int{0, 10, 100, 1000, 10000, 100000}
	store := dstore.NewStore(nil)

	for _, size := range sizes {
		b.Run(fmt.Sprintf("HashSize_%d", size), func(b *testing.B) {
			key := fmt.Sprintf("benchmark_hash_%d", size)

			args := []string{key}
			for i := 0; i < size; i++ {
				args = append(args, fmt.Sprintf("field%d", i), fmt.Sprintf("value%d", i))
			}
			evalHSET(args, store)

			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				evalHLEN([]string{key}, store)
			}
		})
	}
}

func testEvalSELECT(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"nil value": {
			setup:  func() {},
			input:  nil,
			output: []byte("-ERR wrong number of arguments for 'select' command\r\n"),
		},
		"database is specified": {
			setup:  func() {},
			input:  []string{"1"},
			output: clientio.RespOK,
		},
	}
	runEvalTests(t, tests, evalSELECT, store)
}

func testEvalJSONARRPOP(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"wrong number of args passed": {
			setup:  func() {},
			input:  nil,
			output: []byte("-ERR wrong number of arguments for 'json.arrpop' command\r\n"),
		},
		"key does not exist": {
			setup:  func() {},
			input:  []string{"NOTEXISTANT_KEY"},
			output: []byte("-ERR could not perform this operation on a key that doesn't exist\r\n"),
		},
		"empty array at root path": {
			setup: func() {
				key := "MOCK_KEY"
				value := "[]"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"MOCK_KEY"},
			output: []byte("-ERR Path '$' does not exist or not an array\r\n"),
		},
		"empty array at nested path": {
			setup: func() {
				key := "MOCK_KEY"
				value := "{\"a\": 1, \"b\": []}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"MOCK_KEY", "$.b"},
			output: []byte("*1\r\n$-1\r\n"),
		},
		"all paths with asterix": {
			setup: func() {
				key := "MOCK_KEY"
				value := "{\"a\": 1, \"b\": []}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"MOCK_KEY", "$.*"},
			output: []byte("*2\r\n$-1\r\n$-1\r\n"),
		},
		"array root path no index": {
			setup: func() {
				key := "MOCK_KEY"
				value := "[0, 1, 2, 3, 4, 5]"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"MOCK_KEY"},
			output: []byte(":5\r\n"),
		},
		"array root path valid positive index": {
			setup: func() {
				key := "MOCK_KEY"
				value := "[0, 1, 2, 3, 4, 5]"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"MOCK_KEY", "$", "2"},
			output: []byte(":2\r\n"),
		},
		"array root path out of bound positive index": {
			setup: func() {
				key := "MOCK_KEY"
				value := "[0, 1, 2, 3, 4, 5]"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"MOCK_KEY", "$", "10"},
			output: []byte(":5\r\n"),
		},
		"array root path valid negative index": {
			setup: func() {
				key := "MOCK_KEY"
				value := "[0, 1, 2, 3, 4, 5]"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"MOCK_KEY", "$", "-2"},
			output: []byte(":4\r\n"),
		},
		"array root path out of bound negative index": {
			setup: func() {
				key := "MOCK_KEY"
				value := "[0, 1, 2, 3, 4, 5]"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"MOCK_KEY", "$", "-10"},
			output: []byte(":0\r\n"),
		},
		"array at root path updated correctly": {
			setup: func() {
				key := "MOCK_KEY"
				value := "[0, 1, 2, 3, 4, 5]"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"MOCK_KEY", "$", "2"},
			output: []byte(":0\r\n"),
			validator: func(ouput []byte) {
				key := "MOCK_KEY"
				obj := store.Get(key)
				want := []interface{}{float64(0), float64(1), float64(3), float64(4), float64(5)}
				equal := reflect.DeepEqual(obj.Value, want)
				assert.Equal(t, equal, true)
			},
		},
		"nested array updated correctly": {
			setup: func() {
				key := "MOCK_KEY"
				value := "{\"a\": 2, \"b\": [0, 1, 2, 3, 4, 5]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"MOCK_KEY", "$.b", "2"},
			output: []byte("*1\r\n:2\r\n"),
			validator: func(ouput []byte) {
				key := "MOCK_KEY"
				path := "$.b"
				obj := store.Get(key)

				expr, err := jp.ParseString(path)
				assert.NilError(t, err, "error parsing path")

				results := expr.Get(obj.Value)
				assert.Equal(t, len(results), 1)

				want := []interface{}{float64(0), float64(1), float64(3), float64(4), float64(5)}

				equal := reflect.DeepEqual(results[0], want)
				assert.Equal(t, equal, true)
			},
		},
	}

	runEvalTests(t, tests, evalJSONARRPOP, store)
}

func testEvalTYPE(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"TYPE : incorrect number of arguments": {
			setup:  func() {},
			input:  []string{},
			output: diceerrors.NewErrArity("TYPE"),
		},
		"TYPE : key does not exist": {
			setup:  func() {},
			input:  []string{"nonexistent_key"},
			output: clientio.Encode("none", false),
		},
		"TYPE : key exists and is of type String": {
			setup: func() {
				store.Put("string_key", store.NewObj("value", -1, object.ObjTypeString, object.ObjEncodingRaw))
			},
			input:  []string{"string_key"},
			output: clientio.Encode("string", false),
		},
		"TYPE : key exists and is of type List": {
			setup: func() {
				store.Put("list_key", store.NewObj([]byte("value"), -1, object.ObjTypeByteList, object.ObjEncodingRaw))
			},
			input:  []string{"list_key"},
			output: clientio.Encode("list", false),
		},
		"TYPE : key exists and is of type Set": {
			setup: func() {
				store.Put("set_key", store.NewObj([]byte("value"), -1, object.ObjTypeSet, object.ObjEncodingRaw))
			},
			input:  []string{"set_key"},
			output: clientio.Encode("set", false),
		},
		"TYPE : key exists and is of type Hash": {
			setup: func() {
				store.Put("hash_key", store.NewObj([]byte("value"), -1, object.ObjTypeHashMap, object.ObjEncodingRaw))
			},
			input:  []string{"hash_key"},
			output: clientio.Encode("hash", false),
		},
	}
	runEvalTests(t, tests, evalTYPE, store)
}

func BenchmarkEvalTYPE(b *testing.B) {
	store := dstore.NewStore(nil)

	// Define different types of objects to benchmark
	objectTypes := map[string]func(){
		"String": func() {
			store.Put("string_key", store.NewObj("value", -1, object.ObjTypeString, object.ObjEncodingRaw))
		},
		"List": func() {
			store.Put("list_key", store.NewObj([]byte("value"), -1, object.ObjTypeByteList, object.ObjEncodingRaw))
		},
		"Set": func() {
			store.Put("set_key", store.NewObj([]byte("value"), -1, object.ObjTypeSet, object.ObjEncodingRaw))
		},
		"Hash": func() {
			store.Put("hash_key", store.NewObj([]byte("value"), -1, object.ObjTypeHashMap, object.ObjEncodingRaw))
		},
	}

	for objType, setupFunc := range objectTypes {
		b.Run(fmt.Sprintf("ObjectType_%s", objType), func(b *testing.B) {
			// Setup the object in the store
			setupFunc()

			b.ResetTimer()
			b.ReportAllocs()

			// Benchmark the evalTYPE function
			for i := 0; i < b.N; i++ {
				_ = evalTYPE([]string{fmt.Sprintf("%s_key", strings.ToLower(objType))}, store)
			}
		})
	}
}

func testEvalCOMMAND(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"command help": {
			input: []string{"HELP"},
			output: []byte("*11\r\n" +
				"$64\r\n" +
				"COMMAND <subcommand> [<arg> [value] [opt] ...]. Subcommands are:\r\n" +
				"$15\r\n" +
				"(no subcommand)\r\n" +
				"$43\r\n" +
				"    Return details about all Dice commands.\r\n" +
				"$5\r\n" +
				"COUNT\r\n" +
				"$60\r\n" +
				"    Return the total number of commands in this Dice server.\r\n" +
				"$4\r\n" +
				"LIST\r\n" +
				"$55\r\n" +
				"     Return a list of all commands in this Dice server.\r\n" +
				"$22\r\n" +
				"GETKEYS <full-command>\r\n" +
				"$46\r\n" +
				"     Return the keys from a full Dice command.\r\n" +
				"$4\r\n" +
				"HELP\r\n" +
				"$21\r\n" +
				"     Print this help.\r\n"),
		},
		"command info valid command SET": {
			input:  []string{"INFO", "SET"},
			output: []byte("*1\r\n*5\r\n$3\r\nSET\r\n:-3\r\n:1\r\n:0\r\n:0\r\n"),
		},
		"command info valid command GET": {
			input:  []string{"INFO", "GET"},
			output: []byte("*1\r\n*5\r\n$3\r\nGET\r\n:2\r\n:1\r\n:0\r\n:0\r\n"),
		},
		"command info valid command PING": {
			input:  []string{"INFO", "PING"},
			output: []byte("*1\r\n*5\r\n$4\r\nPING\r\n:-1\r\n:0\r\n:0\r\n:0\r\n"),
		},
		"command info multiple valid commands": {
			input:  []string{"INFO", "SET", "GET"},
			output: []byte("*2\r\n*5\r\n$3\r\nSET\r\n:-3\r\n:1\r\n:0\r\n:0\r\n*5\r\n$3\r\nGET\r\n:2\r\n:1\r\n:0\r\n:0\r\n"),
		},
		"command info invalid command": {
			input:  []string{"INFO", "INVALID_CMD"},
			output: []byte("*1\r\n$-1\r\n"),
		},
		"command info mixture of valid and invalid commands": {
			input:  []string{"INFO", "SET", "INVALID_CMD"},
			output: []byte("*2\r\n*5\r\n$3\r\nSET\r\n:-3\r\n:1\r\n:0\r\n:0\r\n$-1\r\n"),
		},
		"command unknown": {
			input:  []string{"UNKNOWN"},
			output: []byte("-ERR unknown subcommand 'UNKNOWN'. Try COMMAND HELP.\r\n"),
		},
	}

	runEvalTests(t, tests, evalCommand, store)
}

func testEvalJSONOBJKEYS(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"nil value": {
			setup:  func() {},
			input:  nil,
			output: []byte("-ERR wrong number of arguments for 'json.objkeys' command\r\n"),
		},
		"empty args": {
			setup:  func() {},
			input:  []string{},
			output: []byte("-ERR wrong number of arguments for 'json.objkeys' command\r\n"),
		},
		"key does not exist": {
			setup:  func() {},
			input:  []string{"NONEXISTENT_KEY"},
			output:  []byte("-ERR could not perform this operation on a key that doesn't exist\r\n"),
		},
		"root not object": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "[1]"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"EXISTING_KEY"},
			output: []byte("-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"),
		},
		"wildcard no object objkeys": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"name\":\"John\",\"age\":30,\"pets\":null,\"languages\":[\"python\",\"golang\"],\"flag\":false}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"EXISTING_KEY", "$.*"},
			output: []byte("*5\r\n$-1\r\n$-1\r\n$-1\r\n$-1\r\n$-1\r\n"),
		},
		"invalid JSONPath": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"name\":\"John\",\"age\":30}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"EXISTING_KEY", "$invalid_path"},
			output: []byte("-ERR parse error at 2 in $invalid_path\r\n"),
		},
		"incomapitable type(int)": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"person\":{\"name\":\"John\",\"age\":30},\"languages\":[\"python\",\"golang\"]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"EXISTING_KEY", "$.person.age"},
			output: []byte("*1\r\n$-1\r\n"),
		},
		"incomapitable type(string)": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"person\":{\"name\":\"John\",\"age\":30},\"languages\":[\"python\",\"golang\"]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"EXISTING_KEY", "$.person.name"},
			output: []byte("*1\r\n$-1\r\n"),
		},
		"incomapitable type(array)": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"person\":{\"name\":\"John\",\"age\":30},\"languages\":[\"python\",\"golang\"]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"EXISTING_KEY", "$.languages"},
			output: []byte("*1\r\n$-1\r\n"),
		},
	}

	runEvalTests(t, tests, evalJSONOBJKEYS, store)
}

func BenchmarkEvalJSONOBJKEYS(b *testing.B) {
	sizes := []int{0, 10, 100, 1000, 10000, 100000} // Various sizes of JSON objects
	store := dstore.NewStore(nil)

	for _, size := range sizes {
		b.Run(fmt.Sprintf("JSONObjectSize_%d", size), func(b *testing.B) {
			key := fmt.Sprintf("benchmark_json_objkeys_%d", size)

			// Create a large JSON object with the given size
			jsonObj := make(map[string]interface{})
			for i := 0; i < size; i++ {
				jsonObj[fmt.Sprintf("key%d", i)] = fmt.Sprintf("value%d", i)
			}

			// Set the JSON object in the store
			args := []string{key, "$", fmt.Sprintf("%v", jsonObj)}
			evalJSONSET(args, store)

			b.ResetTimer()
			b.ReportAllocs()

			// Benchmark the evalJSONOBJKEYS function
			for i := 0; i < b.N; i++ {
				_ = evalJSONOBJKEYS([]string{key, "$"}, store)
			}
		})
	}
}

func testEvalGETRANGE(t *testing.T, store *dstore.Store) {
	setupForStringValue := func() {
		store.Put("STRING_KEY", store.NewObj("Hello World", maxExDuration, object.ObjTypeString, object.ObjEncodingRaw))
	}
	setupForIntegerValue := func() {
		store.Put("INTEGER_KEY", store.NewObj("1234", maxExDuration, object.ObjTypeString, object.ObjEncodingRaw))
	}
	tests := map[string]evalTestCase{
		"GETRANGE against non-existing key": {
			setup:  func() {},
			input:  []string{"NON_EXISTING_KEY", "0", "-1"},
			output: clientio.Encode("", false),
		},
		"GETRANGE against wrong key type": {
			setup: func() {
				evalLPUSH([]string{"LKEY1", "list"}, store)
			},
			input:  []string{"LKEY1", "0", "-1"},
			output: diceerrors.NewErrWithFormattedMessage(diceerrors.WrongTypeErr),
		},
		"GETRANGE against string value: 0, 3": {
			setup:  setupForStringValue,
			input:  []string{"STRING_KEY", "0", "3"},
			output: clientio.Encode("Hell", false),
		},
		"GETRANGE against string value: 0, -1": {
			setup:  setupForStringValue,
			input:  []string{"STRING_KEY", "0", "-1"},
			output: clientio.Encode("Hello World", false),
		},
		"GETRANGE against string value: -4, -1": {
			setup:  setupForStringValue,
			input:  []string{"STRING_KEY", "-4", "-1"},
			output: clientio.Encode("orld", false),
		},
		"GETRANGE against string value: 5, 3": {
			setup:  setupForStringValue,
			input:  []string{"STRING_KEY", "5", "3"},
			output: clientio.Encode("", false),
		},
		"GETRANGE against string value: 5, 5000": {
			setup:  setupForStringValue,
			input:  []string{"STRING_KEY", "5", "5000"},
			output: clientio.Encode(" World", false),
		},
		"GETRANGE against string value: -5000, 10000": {
			setup:  setupForStringValue,
			input:  []string{"STRING_KEY", "-5000", "10000"},
			output: clientio.Encode("Hello World", false),
		},
		"GETRANGE against string value: 0, -100": {
			setup:  setupForStringValue,
			input:  []string{"STRING_KEY", "0", "-100"},
			output: clientio.Encode("", false),
		},
		"GETRANGE against string value: 1, -100": {
			setup:  setupForStringValue,
			input:  []string{"STRING_KEY", "1", "-100"},
			output: clientio.Encode("", false),
		},
		"GETRANGE against string value: -1, -100": {
			setup:  setupForStringValue,
			input:  []string{"STRING_KEY", "-1", "-100"},
			output: clientio.Encode("", false),
		},
		"GETRANGE against string value: -100, -100": {
			setup:  setupForStringValue,
			input:  []string{"STRING_KEY", "-100", "-100"},
			output: clientio.Encode("", false),
		},
		"GETRANGE against string value: -100, -101": {
			setup:  setupForStringValue,
			input:  []string{"STRING_KEY", "-100", "-101"},
			output: clientio.Encode("", false),
		},
		"GETRANGE against integer value: 0, 2": {
			setup:  setupForIntegerValue,
			input:  []string{"INTEGER_KEY", "0", "2"},
			output: clientio.Encode("123", false),
		},
		"GETRANGE against integer value: 0, -1": {
			setup:  setupForIntegerValue,
			input:  []string{"INTEGER_KEY", "0", "-1"},
			output: clientio.Encode("1234", false),
		},
		"GETRANGE against integer value: -3, -1": {
			setup:  setupForIntegerValue,
			input:  []string{"INTEGER_KEY", "-3", "-1"},
			output: clientio.Encode("234", false),
		},
		"GETRANGE against integer value: 5, 3": {
			setup:  setupForIntegerValue,
			input:  []string{"INTEGER_KEY", "5", "3"},
			output: clientio.Encode("", false),
		},
		"GETRANGE against integer value: 3, 5000": {
			setup:  setupForIntegerValue,
			input:  []string{"INTEGER_KEY", "3", "5000"},
			output: clientio.Encode("4", false),
		},

		"GETRANGE against integer value: -5000, 10000": {
			setup:  setupForIntegerValue,
			input:  []string{"INTEGER_KEY", "-5000", "10000"},
			output: clientio.Encode("1234", false),
		},
		"GETRANGE against integer value: 0, -100": {
			setup:  setupForIntegerValue,
			input:  []string{"INTEGER_KEY", "0", "-100"},
			output: clientio.Encode("", false),
		},
		"GETRANGE against integer value: 1, -100": {
			setup:  setupForIntegerValue,
			input:  []string{"INTEGER_KEY", "1", "-100"},
			output: clientio.Encode("", false),
		},
		"GETRANGE against integer value: -1, -100": {
			setup:  setupForIntegerValue,
			input:  []string{"INTEGER_KEY", "-1", "-100"},
			output: clientio.Encode("", false),
		},
		"GETRANGE against integer value: -100, -99": {
			setup:  setupForIntegerValue,
			input:  []string{"INTEGER_KEY", "-100", "-99"},
			output: clientio.Encode("", false),
		},
		"GETRANGE against integer value: -100, -100": {
			setup:  setupForIntegerValue,
			input:  []string{"INTEGER_KEY", "-100", "-100"},
			output: clientio.Encode("", false),
		},
		"GETRANGE against integer value: -100, -101": {
			setup:  setupForIntegerValue,
			input:  []string{"INTEGER_KEY", "-100", "-101"},
			output: clientio.Encode("", false),
		},
	}
	runEvalTests(t, tests, evalGETRANGE, store)
}

func BenchmarkEvalGETRANGE(b *testing.B) {
	store := dstore.NewStore(nil)
	store.Put("BENCHMARK_KEY", store.NewObj("Hello World", maxExDuration, object.ObjTypeString, object.ObjEncodingRaw))

	inputs := []struct {
		start string
		end   string
	}{
		{"0", "3"},
		{"0", "-1"},
		{"-4", "-1"},
		{"5", "3"},
		{"5", "5000"},
		{"-5000", "10000"},
	}

	for _, input := range inputs {
		b.Run(fmt.Sprintf("GETRANGE start=%s end=%s", input.start, input.end), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = evalGETRANGE([]string{"BENCHMARK_KEY", input.start, input.end}, store)
			}
		})
	}
}

func TestMSETConsistency(t *testing.T) {
	store := dstore.NewStore(nil)
	evalMSET([]string{"KEY", "VAL", "KEY2", "VAL2"}, store)

	assert.Equal(t, "VAL", store.Get("KEY").Value)
	assert.Equal(t, "VAL2", store.Get("KEY2").Value)
}

func testEvalSETEX(t *testing.T, store *dstore.Store) {
	mockTime := &utils.MockClock{CurrTime: time.Now()}
	utils.CurrentTime = mockTime

	tests := map[string]evalTestCase{
		"nil value":                              {input: nil, migratedOutput: EvalResponse{Result: nil, Error: errors.New("-ERR wrong number of arguments for 'setex' command\r\n")}},
		"empty array":                            {input: []string{}, migratedOutput: EvalResponse{Result: nil, Error: errors.New("-ERR wrong number of arguments for 'setex' command\r\n")}},
		"one value":                              {input: []string{"KEY"}, migratedOutput: EvalResponse{Result: nil, Error: errors.New("-ERR wrong number of arguments for 'setex' command\r\n")}},
		"key val pair":                           {input: []string{"KEY", "VAL"}, migratedOutput: EvalResponse{Result: nil, Error: errors.New("-ERR wrong number of arguments for 'setex' command\r\n")}},
		"key exp pair":                           {input: []string{"KEY", "123456"}, migratedOutput: EvalResponse{Result: nil, Error: errors.New("-ERR wrong number of arguments for 'setex' command\r\n")}},
		"key exp value pair":                     {input: []string{"KEY", "123", "VAL"}, migratedOutput: EvalResponse{Result: clientio.RespOK, Error: nil}},
		"key exp value pair with extra args":     {input: []string{"KEY", "123", "VAL", " "}, migratedOutput: EvalResponse{Result: nil, Error: errors.New("-ERR wrong number of arguments for 'setex' command\r\n")}},
		"key exp value pair with invalid exp":    {input: []string{"KEY", "0", "VAL"}, migratedOutput: EvalResponse{Result: nil, Error: errors.New("-ERR invalid expire time in 'setex' command\r\n")}},
		"key exp value pair with exp > maxexp":   {input: []string{"KEY", "9223372036854776", "VAL"}, migratedOutput: EvalResponse{Result: nil, Error: errors.New("-ERR invalid expire time in 'setex' command\r\n")}},
		"key exp value pair with exp > maxint64": {input: []string{"KEY", "92233720368547760000000", "VAL"}, migratedOutput: EvalResponse{Result: nil, Error: errors.New("-ERR value is not an integer or out of range\r\n")}},
		"key exp value pair with negative exp":   {input: []string{"KEY", "-23", "VAL"}, migratedOutput: EvalResponse{Result: nil, Error: errors.New("-ERR invalid expire time in 'setex' command\r\n")}},
		"key exp value pair with not-int exp":    {input: []string{"KEY", "12a", "VAL"}, migratedOutput: EvalResponse{Result: nil, Error: errors.New("-ERR value is not an integer or out of range\r\n")}},

		"set and get": {
			setup: func() {},
			input: []string{"TEST_KEY", "5", "TEST_VALUE"},
			validator: func(output []byte) {
				assert.Equal(t, string(clientio.RespOK), string(output))

				// Check if the key was set correctly
				getValue := evalGET([]string{"TEST_KEY"}, store)
				assert.Equal(t, string(clientio.Encode("TEST_VALUE", false)), string(getValue.Result.([]byte)))

				// Check if the TTL is set correctly (should be 5 seconds or less)
				ttlValue := evalTTL([]string{"TEST_KEY"}, store)
				ttl, err := strconv.Atoi(strings.TrimPrefix(strings.TrimSpace(string(ttlValue)), ":"))
				assert.NilError(t, err, "Failed to parse TTL")
				assert.Assert(t, ttl > 0 && ttl <= 5)

				// Wait for the key to expire
				mockTime.SetTime(mockTime.CurrTime.Add(6 * time.Second))

				// Check if the key has been deleted after expiry
				expiredValue := evalGET([]string{"TEST_KEY"}, store)
				assert.Equal(t, string(clientio.RespNIL), string(expiredValue.Result.([]byte)))
			},
		},
		"update existing key": {
			setup: func() {
				evalSET([]string{"EXISTING_KEY", "OLD_VALUE"}, store)
			},
			input: []string{"EXISTING_KEY", "10", "NEW_VALUE"},
			validator: func(output []byte) {
				assert.Equal(t, string(clientio.RespOK), string(output))

				// Check if the key was updated correctly
				getValue := evalGET([]string{"EXISTING_KEY"}, store)
				assert.Equal(t, string(clientio.Encode("NEW_VALUE", false)), string(getValue.Result.([]byte)))

				// Check if the TTL is set correctly
				ttlValue := evalTTL([]string{"EXISTING_KEY"}, store)
				ttl, err := strconv.Atoi(strings.TrimPrefix(strings.TrimSpace(string(ttlValue)), ":"))
				assert.NilError(t, err, "Failed to parse TTL")
				assert.Assert(t, ttl > 0 && ttl <= 10)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := evalSETEX(tt.input, store)

			if tt.validator != nil {
				if tt.migratedOutput.Error != nil {
					tt.validator([]byte(tt.migratedOutput.Error.Error()))
				} else {
					tt.validator(response.Result.([]byte))
				}
			} else {
				// Handle comparison for byte slices
				if b, ok := response.Result.([]byte); ok && tt.migratedOutput.Result != nil {
					if expectedBytes, ok := tt.migratedOutput.Result.([]byte); ok {
						testifyAssert.True(t, bytes.Equal(b, expectedBytes), "expected and actual byte slices should be equal")
					}
				} else {
					assert.Equal(t, tt.migratedOutput.Result, response.Result)
				}

				if tt.migratedOutput.Error != nil {
					testifyAssert.EqualError(t, response.Error, tt.migratedOutput.Error.Error())
				} else {
					testifyAssert.NoError(t, response.Error)
				}
			}
		})
	}
}

func BenchmarkEvalSETEX(b *testing.B) {
	store := dstore.NewStore(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key_%d", i)
		value := fmt.Sprintf("value_%d", i)
		expiry := "10" // 10 seconds expiry

		evalSETEX([]string{key, expiry, value}, store)
	}
}

func testEvalFLUSHDB(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"one key exists in db": {
			setup: func() {
				evalSET([]string{"key", "val"}, store)
			},
			input:  nil,
			output: clientio.RespOK,
		},
		"two keys exist in db": {
			setup: func() {
				evalSET([]string{"key1", "val1"}, store)
				evalSET([]string{"key2", "val2"}, store)
			},
			input:  nil,
			output: clientio.RespOK,
		},
	}
	runEvalTests(t, tests, evalFLUSHDB, store)
}

func testEvalINCRBYFLOAT(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"INCRBYFLOAT on a non existing key": {
			input:  []string{"float", "0.1"},
			output: clientio.Encode("0.1", false),
		},
		"INCRBYFLOAT on an existing key": {
			setup: func() {
				key := "key"
				value := "2.1"
				obj := &object.Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)
			},
			input:  []string{"key", "0.1"},
			output: clientio.Encode("2.2", false),
		},
		"INCRBYFLOAT on a key with integer value": {
			setup: func() {
				key := "key"
				value := "2"
				obj := &object.Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)
			},
			input:  []string{"key", "0.1"},
			output: clientio.Encode("2.1", false),
		},
		"INCRBYFLOAT by a negative increment": {
			setup: func() {
				key := "key"
				value := "2"
				obj := &object.Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)
			},
			input:  []string{"key", "-0.1"},
			output: clientio.Encode("1.9", false),
		},
		"INCRBYFLOAT by a scientific notation increment": {
			setup: func() {
				key := "key"
				value := "1"
				obj := &object.Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)
			},
			input:  []string{"key", "1e-2"},
			output: clientio.Encode("1.01", false),
		},
		"INCRBYFLOAT on a key holding a scientific notation value": {
			setup: func() {
				key := "key"
				value := "1e2"
				obj := &object.Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)
			},
			input:  []string{"key", "1e-1"},
			output: clientio.Encode("100.1", false),
		},
		"INCRBYFLOAT by an negative increment of the same value": {
			setup: func() {
				key := "key"
				value := "0.1"
				obj := &object.Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)
			},
			input:  []string{"key", "-0.1"},
			output: clientio.Encode("0", false),
		},
		"INCRBYFLOAT on a key with spaces": {
			setup: func() {
				key := "key"
				value := "   2   "
				obj := &object.Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)
			},
			input:  []string{"key", "0.1"},
			output: []byte("-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"),
		},
		"INCRBYFLOAT on a key with non numeric value": {
			setup: func() {
				key := "key"
				value := "string"
				obj := &object.Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)
			},
			input:  []string{"key", "0.1"},
			output: []byte("-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"),
		},
		"INCRBYFLOAT by a non numeric increment": {
			setup: func() {
				key := "key"
				value := "2.0"
				obj := &object.Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)
			},
			input:  []string{"key", "a"},
			output: []byte("-ERR value is not an integer or a float\r\n"),
		},
		"INCRBYFLOAT by a number that would turn float64 to Inf": {
			setup: func() {
				key := "key"
				value := "1e308"
				obj := &object.Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)
			},
			input:  []string{"key", "1e308"},
			output: []byte("-ERR value is out of range\r\n"),
		},
	}

	runEvalTests(t, tests, evalINCRBYFLOAT, store)
}

func BenchmarkEvalINCRBYFLOAT(b *testing.B) {
	store := dstore.NewStore(nil)
	store.Put("key1", store.NewObj("1", maxExDuration, object.ObjTypeString, object.ObjEncodingEmbStr))
	store.Put("key2", store.NewObj("1.2", maxExDuration, object.ObjTypeString, object.ObjEncodingEmbStr))

	inputs := []struct {
		key  string
		incr string
	}{
		{"key1", "0.1"},
		{"key1", "-0.1"},
		{"key2", "1000000.1"},
		{"key2", "-1000000.1"},
		{"key3", "-10.1234"},
	}

	for _, input := range inputs {
		b.Run(fmt.Sprintf("INCRBYFLOAT %s %s", input.key, input.incr), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = evalGETRANGE([]string{"INCRBYFLOAT", input.key, input.incr}, store)
			}
		})
	}
}

func testEvalBITOP(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"BITOP NOT (empty string)": {
			setup: func() {
				store.Put("s{t}", store.NewObj(&ByteArray{data: []byte("")}, maxExDuration, object.ObjTypeByteArray, object.ObjEncodingByteArray))
			},
			input:  []string{"NOT", "dest{t}", "s{t}"},
			output: clientio.Encode(0, true),
			validator: func(output []byte) {
				expectedResult := []byte{}
				assert.DeepEqual(t, expectedResult, store.Get("dest{t}").Value.(*ByteArray).data)
			},
		},
		"BITOP NOT (known string)": {
			setup: func() {
				store.Put("s{t}", store.NewObj(&ByteArray{data: []byte{0xaa, 0x00, 0xff, 0x55}}, maxExDuration, object.ObjTypeByteArray, object.ObjEncodingByteArray))
			},
			input:  []string{"NOT", "dest{t}", "s{t}"},
			output: clientio.Encode(4, true),
			validator: func(output []byte) {
				expectedResult := []byte{0x55, 0xff, 0x00, 0xaa}
				assert.DeepEqual(t, expectedResult, store.Get("dest{t}").Value.(*ByteArray).data)
			},
		},
		"BITOP where dest and target are the same key": {
			setup: func() {
				store.Put("s", store.NewObj(&ByteArray{data: []byte{0xaa, 0x00, 0xff, 0x55}}, maxExDuration, object.ObjTypeByteArray, object.ObjEncodingByteArray))
			},
			input:  []string{"NOT", "s", "s"},
			output: clientio.Encode(4, true),
			validator: func(output []byte) {
				expectedResult := []byte{0x55, 0xff, 0x00, 0xaa}
				assert.DeepEqual(t, expectedResult, store.Get("s").Value.(*ByteArray).data)
			},
		},
		"BITOP AND|OR|XOR don't change the string with single input key": {
			setup: func() {
				store.Put("a{t}", store.NewObj(&ByteArray{data: []byte{0x01, 0x02, 0xff}}, maxExDuration, object.ObjTypeByteArray, object.ObjEncodingByteArray))
			},
			input:  []string{"AND", "res1{t}", "a{t}"},
			output: clientio.Encode(3, true),
			validator: func(output []byte) {
				expectedResult := []byte{0x01, 0x02, 0xff}
				assert.DeepEqual(t, expectedResult, store.Get("res1{t}").Value.(*ByteArray).data)
			},
		},
		"BITOP missing key is considered a stream of zero": {
			setup: func() {
				store.Put("a{t}", store.NewObj(&ByteArray{data: []byte{0x01, 0x02, 0xff}}, maxExDuration, object.ObjTypeByteArray, object.ObjEncodingByteArray))
			},
			input:  []string{"AND", "res1{t}", "no-such-key{t}", "a{t}"},
			output: clientio.Encode(3, true),
			validator: func(output []byte) {
				expectedResult := []byte{0x00, 0x00, 0x00}
				assert.DeepEqual(t, expectedResult, store.Get("res1{t}").Value.(*ByteArray).data)
			},
		},
		"BITOP shorter keys are zero-padded to the key with max length": {
			setup: func() {
				store.Put("a{t}", store.NewObj(&ByteArray{data: []byte{0x01, 0x02, 0xff, 0xff}}, maxExDuration, object.ObjTypeByteArray, object.ObjEncodingByteArray))
				store.Put("b{t}", store.NewObj(&ByteArray{data: []byte{0x01, 0x02, 0xff}}, maxExDuration, object.ObjTypeByteArray, object.ObjEncodingByteArray))
			},
			input:  []string{"AND", "res1{t}", "a{t}", "b{t}"},
			output: clientio.Encode(4, true),
			validator: func(output []byte) {
				expectedResult := []byte{0x01, 0x02, 0xff, 0x00}
				assert.DeepEqual(t, expectedResult, store.Get("res1{t}").Value.(*ByteArray).data)
			},
		},
		"BITOP with non string source key": {
			setup: func() {
				store.Put("a{t}", store.NewObj("1", maxExDuration, object.ObjTypeString, object.ObjEncodingRaw))
				store.Put("b{t}", store.NewObj("2", maxExDuration, object.ObjTypeString, object.ObjEncodingRaw))
				store.Put("c{t}", store.NewObj([]byte("foo"), maxExDuration, object.ObjTypeByteList, object.ObjEncodingRaw))
			},
			input:  []string{"XOR", "dest{t}", "a{t}", "b{t}", "c{t}", "d{t}"},
			output: []byte("-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"),
		},
		"BITOP with empty string after non empty string": {
			setup: func() {
				store.Put("a{t}", store.NewObj(&ByteArray{data: []byte("\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00")}, -1, object.ObjTypeByteArray, object.ObjEncodingByteArray))
			},
			input:  []string{"OR", "x{t}", "a{t}", "b{t}"},
			output: clientio.Encode(32, true),
		},
	}

	runEvalTests(t, tests, evalBITOP, store)
}

func BenchmarkEvalBITOP(b *testing.B) {
	store := dstore.NewStore(nil)

	// Setup initial data for benchmarking
	store.Put("key1", store.NewObj(&ByteArray{data: []byte{0x01, 0x02, 0xff}}, maxExDuration, object.ObjTypeByteArray, object.ObjEncodingByteArray))
	store.Put("key2", store.NewObj(&ByteArray{data: []byte{0x01, 0x02, 0xff}}, maxExDuration, object.ObjTypeByteArray, object.ObjEncodingByteArray))

	// Define different operations to benchmark
	operations := []struct {
		name string
		op   string
	}{
		{"AND", "AND"},
		{"OR", "OR"},
		{"XOR", "XOR"},
		{"NOT", "NOT"},
	}

	for _, operation := range operations {
		b.Run(fmt.Sprintf("BITOP_%s", operation.name), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if operation.op == "NOT" {
					evalBITOP([]string{operation.op, "dest", "key1"}, store)
				} else {
					evalBITOP([]string{operation.op, "dest", "key1", "key2"}, store)
				}
			}
		})
	}
}

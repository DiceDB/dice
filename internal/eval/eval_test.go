package eval

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/bytedance/sonic"

	"github.com/dicedb/dice/internal/object"

	"github.com/axiomhq/hyperloglog"
	"github.com/dicedb/dice/internal/clientio"
	dstore "github.com/dicedb/dice/internal/store"
	"gotest.tools/v3/assert"
)

type evalTestCase struct {
	setup     func()
	input     []string
	output    []byte
	validator func(output []byte)
}

func setupTest(store *dstore.Store) *dstore.Store {
	dstore.ResetStore(store)
	dstore.KeyspaceStat[0] = make(map[string]int)

	return store
}

func TestEval(t *testing.T) {
	store := dstore.NewStore(nil)

	testEvalMSET(t, store)
	testEvalPING(t, store)
	testEvalHELLO(t, store)
	testEvalSET(t, store)
	testEvalGET(t, store)
	testEvalJSONARRLEN(t, store)
	testEvalJSONDEL(t, store)
	testEvalJSONFORGET(t, store)
	testEvalJSONCLEAR(t, store)
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
	testEvalGETSET(t, store)
	testEvalHSET(t, store)
	testEvalPFADD(t, store)
	testEvalPFCOUNT(t, store)
	testEvalHGET(t, store)
	testEvalPFMERGE(t, store)
	testEvalJSONSTRLEN(t, store)
	testEvalHLEN(t, store)
	testEvalSELECT(t, store)
	testEvalLLEN(t, store)
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

func testEvalHELLO(t *testing.T, store *dstore.Store) {
	resp := []interface{}{
		"proto", 2,
		"id", serverID,
		"mode", "standalone",
		"role", "master",
		"modules",
		[]interface{}{},
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
	tests := map[string]evalTestCase{
		"nil value":                       {input: nil, output: []byte("-ERR wrong number of arguments for 'set' command\r\n")},
		"empty array":                     {input: []string{}, output: []byte("-ERR wrong number of arguments for 'set' command\r\n")},
		"one value":                       {input: []string{"KEY"}, output: []byte("-ERR wrong number of arguments for 'set' command\r\n")},
		"key val pair":                    {input: []string{"KEY", "VAL"}, output: clientio.RespOK},
		"key val pair with int val":       {input: []string{"KEY", "123456"}, output: clientio.RespOK},
		"key val pair and expiry key":     {input: []string{"KEY", "VAL", Px}, output: []byte("-ERR syntax error\r\n")},
		"key val pair and EX no val":      {input: []string{"KEY", "VAL", Ex}, output: []byte("-ERR syntax error\r\n")},
		"key val pair and valid EX":       {input: []string{"KEY", "VAL", Ex, "2"}, output: clientio.RespOK},
		"key val pair and invalid EX":     {input: []string{"KEY", "VAL", Ex, "invalid_expiry_val"}, output: []byte("-ERR value is not an integer or out of range\r\n")},
		"key val pair and valid PX":       {input: []string{"KEY", "VAL", Px, "2000"}, output: clientio.RespOK},
		"key val pair and invalid PX":     {input: []string{"KEY", "VAL", Px, "invalid_expiry_val"}, output: []byte("-ERR value is not an integer or out of range\r\n")},
		"key val pair and both EX and PX": {input: []string{"KEY", "VAL", Ex, "2", Px, "2000"}, output: []byte("-ERR syntax error\r\n")},
		"key val pair and PXAT no val":    {input: []string{"KEY", "VAL", Pxat}, output: []byte("-ERR syntax error\r\n")},
		"key val pair and invalid PXAT":   {input: []string{"KEY", "VAL", Pxat, "invalid_expiry_val"}, output: []byte("-ERR value is not an integer or out of range\r\n")},
		"key val pair and expired PXAT":   {input: []string{"KEY", "VAL", Pxat, "2"}, output: clientio.RespOK},
		"key val pair and negative PXAT":  {input: []string{"KEY", "VAL", Pxat, "-123456"}, output: []byte("-ERR invalid expire time in 'set' command\r\n")},
		"key val pair and valid PXAT":     {input: []string{"KEY", "VAL", Pxat, strconv.FormatInt(time.Now().Add(2*time.Minute).UnixMilli(), 10)}, output: clientio.RespOK},
	}

	runEvalTests(t, tests, evalSET, store)
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
			output: clientio.RespNIL,
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
				obj := &object.Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)
			},
			input:  []string{"EXISTING_KEY"},
			output: clientio.Encode("mock_value", false),
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

	runEvalTests(t, tests, evalGET, store)
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
	}

	runEvalTests(t, tests, evalDBSIZE, store)
}

func testEvalGETSET(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"GETSET with 1 arg": {
			input:  []string{"HELLO"},
			output: []byte("-ERR wrong number of arguments for 'getset' command\r\n"),
		},
		"GETSET with 3 args": {
			input:  []string{"HELLO", "WORLD", "WORLD1"},
			output: []byte("-ERR wrong number of arguments for 'getset' command\r\n"),
		},
		"GETSET key not exists": {
			input:  []string{"HELLO", "WORLD"},
			output: clientio.RespNIL,
		},
		"GETSET key exists": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "mock_value"
				obj := &object.Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)
			},
			input:  []string{"EXISTING_KEY", "WORLD"},
			output: clientio.Encode("mock_value", false),
		},
		"GETSET key exists TTL should be reset": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "mock_value"
				obj := &object.Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)
			},
			input:  []string{"EXISTING_KEY", "WORLD"},
			output: clientio.Encode("mock_value", false),
		},
	}

	runEvalTests(t, tests, evalGETSET, store)
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
			output: []byte("-WRONGTYPE Key is not a valid HyperLogLog string value.\r\n"),
		},
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
				key := "EXISTING_KEY"
				value := "mock_value"
				obj := store.NewObj(NewDeque(), -1, object.ObjTypeByteList, object.ObjEncodingDeque)
				obj.Value.(*Deque).LPush(value)
				store.Put(key, obj)
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

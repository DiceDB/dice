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
	testEvalECHO(t, store)
	testEvalHELLO(t, store)
	testEvalSET(t, store)
	testEvalGET(t, store)
	testEvalDebug(t, store)
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
		input: []string{"doc", "$.a", "qwe"},
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
		input: []string{"doc", "$..fe", "2"},
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
	}

	runEvalTests(t, tests, evalCommand, store)
}
func TestMSETConsistency(t *testing.T) {
	store := dstore.NewStore(nil)
	evalMSET([]string{"KEY", "VAL", "KEY2", "VAL2"}, store)

	assert.Equal(t, "VAL", store.Get("KEY").Value)
	assert.Equal(t, "VAL2", store.Get("KEY2").Value)
}

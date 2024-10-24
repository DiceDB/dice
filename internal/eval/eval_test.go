package eval

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/dicedb/dice/internal/eval/sortedset"
	"github.com/dicedb/dice/internal/server/utils"

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
	name           string
	setup          func()
	input          []string
	output         []byte
	validator      func(output []byte)
	newValidator   func(output interface{})
	migratedOutput EvalResponse
}

func setupTest(store *dstore.Store) *dstore.Store {
	dstore.ResetStore(store)
	return store
}

func TestEval(t *testing.T) {
	store := dstore.NewStore(nil, nil)

	testEvalMSET(t, store)
	testEvalECHO(t, store)
	testEvalHELLO(t, store)
	testEvalSET(t, store)
	testEvalGET(t, store)
	testEvalGETEX(t, store)
	testEvalDebug(t, store)
	testEvalJSONARRTRIM(t, store)
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
	testEvalJSONRESP(t, store)
	testEvalTTL(t, store)
	testEvalDel(t, store)
	testEvalPersist(t, store)
	testEvalEXPIRE(t, store)
	testEvalEXPIRETIME(t, store)
	testEvalEXPIREAT(t, store)
	testEvalDbsize(t, store)
	testEvalGETSET(t, store)
	testEvalHSET(t, store)
	testEvalHMSET(t, store)
	testEvalHKEYS(t, store)
	testEvalPFADD(t, store)
	testEvalPFCOUNT(t, store)
	testEvalPFMERGE(t, store)
	testEvalHGET(t, store)
	testEvalHMGET(t, store)
	testEvalHSTRLEN(t, store)
	testEvalHEXISTS(t, store)
	testEvalHDEL(t, store)
	testEvalHSCAN(t, store)
	testEvalJSONSTRLEN(t, store)
	testEvalJSONOBJLEN(t, store)
	testEvalHLEN(t, store)
	testEvalSELECT(t, store)
	testEvalLPUSH(t, store)
	testEvalRPUSH(t, store)
	testEvalLPOP(t, store)
	testEvalRPOP(t, store)
	testEvalLLEN(t, store)
	testEvalGETEX(t, store)
	testEvalJSONNUMINCRBY(t, store)
	testEvalDUMP(t, store)
	testEvalTYPE(t, store)
	testEvalCOMMAND(t, store)
	testEvalHINCRBY(t, store)
	testEvalJSONOBJKEYS(t, store)
	testEvalGETRANGE(t, store)
	testEvalHSETNX(t, store)
	testEvalPING(t, store)
	testEvalSETEX(t, store)
	testEvalFLUSHDB(t, store)
	testEvalINCRBYFLOAT(t, store)
	testEvalBITOP(t, store)
	testEvalAPPEND(t, store)
	testEvalHRANDFIELD(t, store)
	testEvalZADD(t, store)
	testEvalZRANGE(t, store)
	testEvalZPOPMIN(t, store)
	testEvalZRANK(t, store)
	testEvalHVALS(t, store)
	testEvalBitField(t, store)
	testEvalHINCRBYFLOAT(t, store)
	testEvalBitFieldRO(t, store)
	testEvalGEOADD(t, store)
	testEvalGEODIST(t, store)
	testEvalSINTER(t, store)
	testEvalOBJECTENCODING(t, store)
	testEvalJSONSTRAPPEND(t, store)
	testEvalINCR(t, store)
	testEvalINCRBY(t, store)
	testEvalDECR(t, store)
	testEvalDECRBY(t, store)
	testEvalBFRESERVE(t, store)
	testEvalBFINFO(t, store)
	testEvalBFEXISTS(t, store)
	testEvalBFADD(t, store)
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
	tests := []evalTestCase{
		{
			name:           "nil value",
			input:          nil,
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR wrong number of arguments for 'set' command")},
		},
		{
			name:           "empty array",
			input:          []string{},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR wrong number of arguments for 'set' command")},
		},
		{
			name:           "one value",
			input:          []string{"KEY"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR wrong number of arguments for 'set' command")},
		},
		{
			name:           "key val pair",
			input:          []string{"KEY", "VAL"},
			migratedOutput: EvalResponse{Result: clientio.OK, Error: nil},
		},
		{
			name:           "key val pair with int val",
			input:          []string{"KEY", "123456"},
			migratedOutput: EvalResponse{Result: clientio.OK, Error: nil},
		},
		{
			name:           "key val pair and expiry key",
			input:          []string{"KEY", "VAL", Px},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR syntax error")},
		},
		{
			name:           "key val pair and EX no val",
			input:          []string{"KEY", "VAL", Ex},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR syntax error")},
		},
		{
			name:           "key val pair and valid EX",
			input:          []string{"KEY", "VAL", Ex, "2"},
			migratedOutput: EvalResponse{Result: clientio.OK, Error: nil},
		},
		{
			name:           "key val pair and invalid negative EX",
			input:          []string{"KEY", "VAL", Ex, "-2"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR invalid expire time in 'set' command")},
		},
		{
			name:           "key val pair and invalid float EX",
			input:          []string{"KEY", "VAL", Ex, "2.0"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR value is not an integer or out of range")},
		},
		{
			name:           "key val pair and invalid out of range int EX",
			input:          []string{"KEY", "VAL", Ex, "9223372036854775807"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR invalid expire time in 'set' command")},
		},
		{
			name:           "key val pair and invalid greater than max duration EX",
			input:          []string{"KEY", "VAL", Ex, "9223372036854775"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR invalid expire time in 'set' command")},
		},
		{
			name:           "key val pair and invalid EX",
			input:          []string{"KEY", "VAL", Ex, "invalid_expiry_val"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR value is not an integer or out of range")},
		},
		{
			name:           "key val pair and PX no val",
			input:          []string{"KEY", "VAL", Px},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR syntax error")},
		},
		{
			name:           "key val pair and valid PX",
			input:          []string{"KEY", "VAL", Px, "2000"},
			migratedOutput: EvalResponse{Result: clientio.OK, Error: nil},
		},
		{
			name:           "key val pair and invalid PX",
			input:          []string{"KEY", "VAL", Px, "invalid_expiry_val"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR value is not an integer or out of range")},
		},
		{
			name:           "key val pair and invalid negative PX",
			input:          []string{"KEY", "VAL", Px, "-2"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR invalid expire time in 'set' command")},
		},
		{
			name:           "key val pair and invalid float PX",
			input:          []string{"KEY", "VAL", Px, "2.0"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR value is not an integer or out of range")},
		},
		{
			name:           "key val pair and invalid out of range int PX",
			input:          []string{"KEY", "VAL", Px, "9223372036854775807"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR invalid expire time in 'set' command")},
		},
		{
			name:           "key val pair and invalid greater than max duration PX",
			input:          []string{"KEY", "VAL", Px, "9223372036854775"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR invalid expire time in 'set' command")},
		},
		{
			name:           "key val pair and both EX and PX",
			input:          []string{"KEY", "VAL", Ex, "2", Px, "2000"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR syntax error")},
		},
		{
			name:           "key val pair and PXAT no val",
			input:          []string{"KEY", "VAL", Pxat},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR syntax error")},
		},
		{
			name:           "key val pair and invalid PXAT",
			input:          []string{"KEY", "VAL", Pxat, "invalid_expiry_val"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR value is not an integer or out of range")},
		},
		{
			name:           "key val pair and expired PXAT",
			input:          []string{"KEY", "VAL", Pxat, "2"},
			migratedOutput: EvalResponse{Result: clientio.OK, Error: nil},
		},
		{
			name:           "key val pair and negative PXAT",
			input:          []string{"KEY", "VAL", Pxat, "-123456"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR invalid expire time in 'set' command")},
		},
		{
			name:           "key val pair and valid PXAT",
			input:          []string{"KEY", "VAL", Pxat, strconv.FormatInt(time.Now().Add(2*time.Minute).UnixMilli(), 10)},
			migratedOutput: EvalResponse{Result: clientio.OK, Error: nil},
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
			output: []byte("-ERR invalid expire time in 'getex' command\r\n"),
		},
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
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR wrong number of arguments for 'get' command")},
		},
		{
			name:           "empty array",
			input:          []string{},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR wrong number of arguments for 'get' command")},
		},
		{
			name:           "key does not exist",
			input:          []string{"NONEXISTENT_KEY"},
			migratedOutput: EvalResponse{Result: clientio.NIL, Error: nil},
		},
		{
			name:           "multiple arguments",
			input:          []string{"KEY1", "KEY2"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR wrong number of arguments for 'get' command")},
		},
		{
			name: "key exists",
			setup: func() {
				key := "diceKey"
				value := "diceVal"
				obj := &object.Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)
			},
			input:          []string{"diceKey"},
			migratedOutput: EvalResponse{Result: "diceVal", Error: nil},
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
			migratedOutput: EvalResponse{Result: clientio.NIL, Error: nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup the test store
			if tt.setup != nil {
				tt.setup()
			}

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
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR wrong number of arguments for 'getset' command")},
		},
		{
			name:           "GETSET with 3 args",
			input:          []string{"HELLO", "WORLD", "WORLD1"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR wrong number of arguments for 'getset' command")},
		},
		{
			name:           "GETSET key not exists",
			input:          []string{"HELLO", "WORLD"},
			migratedOutput: EvalResponse{Result: clientio.NIL, Error: nil},
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
			migratedOutput: EvalResponse{Result: "mock_value", Error: nil},
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
			migratedOutput: EvalResponse{Result: "mock_value", Error: nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup the test store
			if tt.setup != nil {
				tt.setup()
			}

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
		"invalid expiry time exists - empty string": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "mock_value"
				obj := &object.Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)
			},
			input:  []string{"EXISTING_KEY", ""},
			output: []byte("-ERR value is not an integer or out of range\r\n"),
		},
		"invalid expiry time exists - with float number": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "mock_value"
				obj := &object.Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)
			},
			input:  []string{"EXISTING_KEY", "0.456"},
			output: []byte("-ERR value is not an integer or out of range\r\n"),
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

func testEvalJSONARRTRIM(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"nil value": {
			setup:  func() {},
			input:  nil,
			output: []byte("-ERR wrong number of arguments for 'json.arrtrim' command\r\n"),
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
			input:  []string{"EXISTING_KEY", "$", "a", "1"},
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
			input:  []string{"EXISTING_KEY", "$", "0", "10"},
			output: []byte("*1\r\n:3\r\n"),
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
				value := "[1,2,3,4,5]"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"EXISTING_KEY", "$", "1", "3"},
			output: []byte("*1\r\n:3\r\n"),
		},
		"subpath array": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"connection\":{\"wireless\":true,\"names\":[0,1,2,3,4]},\"names\":[0,1,2,3,4]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"EXISTING_KEY", "$.names", "1", "3"},
			output: []byte("*1\r\n:3\r\n"),
		},
		"subpath two array": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"connection\":{\"wireless\":true,\"names\":[0,1,2,3,4]},\"names\":[0,1,2,3,4]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"EXISTING_KEY", "$..names", "1", "3"},
			output: []byte("*2\r\n:3\r\n:3\r\n"),
		},
		"subpath not array": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"connection\":{\"wireless\":true,\"names\":[0,1,2,3,4]},\"names\":[0,1,2,3,4]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"EXISTING_KEY", "$.connection", "1", "2"},
			output: []byte("*1\r\n$-1\r\n"),
		},
		"subpath array index negative": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"connection\":{\"wireless\":true,\"names\":[0,1,2,3,4]},\"names\":[0,1,2,3,4]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"EXISTING_KEY", "$.names", "-3", "-1"},
			output: []byte("*1\r\n:3\r\n"),
		},
		"index negative start larger than stop": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"connection\":{\"wireless\":true,\"names\":[0,1,2,3,4]},\"names\":[0,1,2,3,4]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"EXISTING_KEY", "$.names", "-1", "-3"},
			output: []byte("*1\r\n:0\r\n"),
		},
	}
	runEvalTests(t, tests, evalJSONARRTRIM, store)
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
			output: []byte("$-1\r\n"),
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
			output: []byte("-ERR Path '$' does not exist or not an array\r\n"),
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
			output: []byte(":2\r\n"),
		},
	}
	runEvalTests(t, tests, evalJSONARRLEN, store)
}

func testEvalJSONOBJLEN(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"jsonobjlen nil value": {
			name:  "jsonobjlen objlen nil value",
			setup: func() {},
			input: nil,
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("JSON.OBJLEN"),
			},
		},
		"jsonobjlen empty args": {
			name:  "jsonobjlen objlen empty args",
			setup: func() {},
			input: []string{},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("JSON.OBJLEN"),
			},
		},
		"jsonobjlen key does not exist": {
			name:  "jsonobjlen key does not exist",
			setup: func() {},
			input: []string{"NONEXISTENT_KEY"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  nil,
			},
		},
		"jsonobjlen root not object": {
			name: "jsonobjlen root not object",
			setup: func() {
				key := "EXISTING_KEY"
				value := "[1,2,3]"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input: []string{"EXISTING_KEY"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongTypeOperation,
			},
		},
		"jsonobjlen root object": {
			name: "jsonobjlen root object",
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"name\":\"John\",\"age\":30,\"city\":\"New York\"}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input: []string{"EXISTING_KEY"},
			migratedOutput: EvalResponse{
				Result: int64(3),
				Error:  nil,
			},
		},
		"jsonobjlen wildcard no object": {
			name: "jsonobjlen wildcard no object",
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"name\":\"John\",\"age\":30,\"pets\":null,\"languages\":[\"python\",\"golang\"],\"flag\":false}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input: []string{"EXISTING_KEY", "$.*"},
			migratedOutput: EvalResponse{
				Result: []interface{}{nil, nil, nil, nil, nil},
				Error:  nil,
			},
		},
		"jsonobjlen subpath object": {
			name: "jsonobjlen subpath object",
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"person\":{\"name\":\"John\",\"age\":30},\"languages\":[\"python\",\"golang\"]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input: []string{"EXISTING_KEY", "$.person"},
			migratedOutput: EvalResponse{
				Result: []interface{}{int64(2)},
				Error:  nil,
			},
		},
		"jsonobjlen invalid JSONPath": {
			name: "jsonobjlen invalid JSONPath",
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"name\":\"John\",\"age\":30}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input: []string{"EXISTING_KEY", "$invalid_path"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrJSONPathNotFound("$invalid_path"),
			},
		},
		"jsonobjlen incomapitable type(int)": {
			name: "jsonobjlen incomapitable type(int)",
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"person\":{\"name\":\"John\",\"age\":30},\"languages\":[\"python\",\"golang\"]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input: []string{"EXISTING_KEY", "$.person.age"},
			migratedOutput: EvalResponse{
				Result: []interface{}{nil},
				Error:  nil,
			},
		},
		"jsonobjlen incomapitable type(string)": {
			name: "jsonobjlen incomapitable type(string)",
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"person\":{\"name\":\"John\",\"age\":30},\"languages\":[\"python\",\"golang\"]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input: []string{"EXISTING_KEY", "$.person.name"},
			migratedOutput: EvalResponse{
				Result: []interface{}{nil},
				Error:  nil,
			},
		},
		"jsonobjlen incomapitable type(array)": {
			name: "jsonobjlen incomapitable type(array)",
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"person\":{\"name\":\"John\",\"age\":30},\"languages\":[\"python\",\"golang\"]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input: []string{"EXISTING_KEY", "$.languages"},
			migratedOutput: EvalResponse{
				Result: []interface{}{nil},
				Error:  nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store = setupTest(store)
			if tt.setup != nil {
				tt.setup()
			}

			response := evalJSONOBJLEN(tt.input, store)
			if tt.migratedOutput.Result != nil {
				if slice, ok := tt.migratedOutput.Result.([]interface{}); ok {
					assert.DeepEqual(t, slice, response.Result)
				} else {
					assert.Equal(t, tt.migratedOutput.Result, response.Result)
				}
			}

			if tt.migratedOutput.Error != nil {
				testifyAssert.EqualError(t, response.Error, tt.migratedOutput.Error.Error())
			} else {
				testifyAssert.NoError(t, response.Error)
			}
		})
	}
}

func BenchmarkEvalJSONOBJLEN(b *testing.B) {
	sizes := []int{0, 10, 100, 1000, 10000, 100000} // Various sizes of JSON objects
	store := dstore.NewStore(nil, nil)

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
		"jsonclear nil value": {
			name:  "jsonclear nil value",
			setup: func() {},
			input: nil,
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("JSON.CLEAR"),
			},
		},
		"jsonclear empty array": {
			name:  "jsonclear empty array",
			setup: func() {},
			input: []string{},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("JSON.CLEAR"),
			},
		},
		"jsonclear key does not exist": {
			name:  "jsonclear key does not exist",
			setup: func() {},
			input: []string{"NONEXISTENT_KEY"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  nil,
			},
		},
		"jsonclear root": {
			name: "jsonclear root",
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"age\":13,\"high\":1.60,\"language\":[\"python\",\"golang\"], \"flag\":false, " +
					"\"partner\":{\"name\":\"tom\",\"language\":[\"rust\"]}}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input: []string{"EXISTING_KEY"},
			migratedOutput: EvalResponse{
				Result: int64(1),
				Error:  nil,
			},
		},
		"jsonclear array type": {
			name: "jsonclear array type",
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"array\":[1,2,3,\"s\",null]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},

			input: []string{"EXISTING_KEY", "$.array"},
			migratedOutput: EvalResponse{
				Result: int64(1),
				Error:  nil,
			},
		},
		"jsonclear string type": {
			name: "jsonclear string type",
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"a\":\"test\"}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},

			input: []string{"EXISTING_KEY", "$.a"},
			migratedOutput: EvalResponse{
				Result: int64(0),
				Error:  nil,
			},
		},
		"jsonclear integer type": {
			name: "jsonclear integer type",
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"age\":13}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},

			input: []string{"EXISTING_KEY", "$.age"},
			migratedOutput: EvalResponse{
				Result: int64(1),
				Error:  nil,
			},
		},
		"jsonclear number type": {
			name: "jsonclear number type",
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"price\":3.14}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},

			input: []string{"EXISTING_KEY", "$.price"},
			migratedOutput: EvalResponse{
				Result: int64(1),
				Error:  nil,
			},
		},
		"jsonclear boolean type": {
			name: "jsonclear boolean type",
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"flag\":false}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input: []string{"EXISTING_KEY", "$.flag"},
			migratedOutput: EvalResponse{
				Result: int64(0),
				Error:  nil,
			},
		},
		"jsonclear multi type": {
			name: "jsonclear multi type",
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"age\":13,\"high\":1.60,\"name\":\"jerry\",\"language\":[\"python\",\"golang\"]," +
					"\"partner\":{\"name\":\"tom\",\"language\":[\"rust\"]}}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input: []string{"EXISTING_KEY", "$.*"},
			migratedOutput: EvalResponse{
				Result: int64(4),
				Error:  nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store = setupTest(store)
			if tt.setup != nil {
				tt.setup()
			}

			response := evalJSONCLEAR(tt.input, store)
			if tt.migratedOutput.Result != nil {
				if slice, ok := tt.migratedOutput.Result.([]interface{}); ok {
					assert.DeepEqual(t, slice, response.Result)
				} else {
					assert.Equal(t, tt.migratedOutput.Result, response.Result)
				}
			}

			if tt.migratedOutput.Error != nil {
				testifyAssert.EqualError(t, response.Error, tt.migratedOutput.Error.Error())
			} else {
				testifyAssert.NoError(t, response.Error)
			}
		})
	}
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
			output: clientio.Encode(errors.New("ERR wrong number of arguments for 'del' command"), false),
		},
		"empty array": {
			setup:  func() {},
			input:  []string{},
			output: clientio.Encode(errors.New("ERR wrong number of arguments for 'del' command"), false),
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

				store.IncrementKeyCount()
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
			output: clientio.RespZero,
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
		"PFADD nil value": {
			name:  "PFADD nil value",
			setup: func() {},
			input: nil,
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("PFADD"),
			},
		},
		"PFADD empty array": {
			name:  "PFADD empty array",
			setup: func() {},
			input: []string{},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("PFADD"),
			},
		},
		"PFADD one value": {
			name:  "PFADD one value",
			setup: func() {},
			input: []string{"KEY"},
			migratedOutput: EvalResponse{
				Result: int64(1),
				Error:  nil,
			},
		},
		"PFADD key val pair": {
			name:  "PFADD key val pair",
			setup: func() {},
			input: []string{"KEY", "VAL"},
			migratedOutput: EvalResponse{
				Result: int64(1),
				Error:  nil,
			},
		},
		"PFADD key multiple values": {
			name:  "PFADD key multiple values",
			setup: func() {},
			input: []string{"KEY", "VAL", "VAL1", "VAL2"},
			migratedOutput: EvalResponse{
				Result: int64(1),
				Error:  nil,
			},
		},
		"PFADD Incorrect type provided": {
			name: "PFADD Incorrect type provided",
			setup: func() {
				key, value := "EXISTING_KEY", "VALUE"
				oType, oEnc := deduceTypeEncoding(value)
				var exDurationMs int64 = -1
				keepttl := false

				store.Put(key, store.NewObj(value, exDurationMs, oType, oEnc), dstore.WithKeepTTL(keepttl))
			},
			input:  []string{"EXISTING_KEY", "1"},
			output: []byte("-WRONGTYPE Key is not a valid HyperLogLog string value"),
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrInvalidHyperLogLogKey,
			},
		},
	}

	runMigratedEvalTests(t, tests, evalPFADD, store)
}

func testEvalPFCOUNT(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"PFCOUNT with empty arg": {
			name:  "PFCOUNT with empty arg",
			setup: func() {},
			input: []string{},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("PFCOUNT"),
			},
		},
		"PFCOUNT key not exists": {
			name:  "PFCOUNT key not exists",
			setup: func() {},
			input: []string{"HELLO"},
			migratedOutput: EvalResponse{
				Result: uint64(0),
				Error:  nil,
			},
		},
		"PFCOUNT key exists": {
			name: "PFCOUNT key exists",
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
			input: []string{"EXISTING_KEY"},
			migratedOutput: EvalResponse{
				Result: uint64(1),
				Error:  nil,
			},
		},
	}

	runMigratedEvalTests(t, tests, evalPFCOUNT, store)
}

func testEvalPFMERGE(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"PFMERGE nil value": {
			name:  "PFMERGE nil value",
			setup: func() {},
			input: nil,
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("PFMERGE"),
			},
		},
		"PFMERGE empty array": {
			name:  "PFMERGE empty array",
			setup: func() {},
			input: []string{},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("PFMERGE"),
			},
		},
		"PFMERGE invalid hll object": {
			name: "PFMERGE invalid hll object",
			setup: func() {
				key := "INVALID_OBJ_DEST_KEY"
				value := "123"
				obj := &object.Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)
			},
			input: []string{"INVALID_OBJ_DEST_KEY"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrInvalidHyperLogLogKey,
			},
		},
		"PFMERGE destKey doesn't exist": {
			name:  "PFMERGE destKey doesn't exist",
			setup: func() {},
			input: []string{"NON_EXISTING_DEST_KEY"},
			migratedOutput: EvalResponse{
				Result: clientio.OK,
				Error:  nil,
			},
		},
		"PFMERGE destKey exist": {
			name:  "PFMERGE destKey exist",
			setup: func() {},
			input: []string{"NON_EXISTING_DEST_KEY"},
			migratedOutput: EvalResponse{
				Result: clientio.OK,
				Error:  nil,
			},
		},
		"PFMERGE destKey exist srcKey doesn't exists": {
			name: "PFMERGE destKey exist srcKey doesn't exists",
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
			input: []string{"EXISTING_DEST_KEY", "NON_EXISTING_SRC_KEY"},
			migratedOutput: EvalResponse{
				Result: clientio.OK,
				Error:  nil,
			},
		},
		"PFMERGE destKey exist srcKey exists": {
			name: "PFMERGE destKey exist srcKey exists",
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
			input: []string{"EXISTING_DEST_KEY", "NON_EXISTING_SRC_KEY"},
			migratedOutput: EvalResponse{
				Result: clientio.OK,
				Error:  nil,
			},
		},
		"PFMERGE destKey exist multiple srcKey exist": {
			name: "PFMERGE destKey exist multiple srcKey exist",
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
			input: []string{"EXISTING_DEST_KEY", "EXISTING_SRC_KEY"},
			migratedOutput: EvalResponse{
				Result: clientio.OK,
				Error:  nil,
			},
		},
	}

	runMigratedEvalTests(t, tests, evalPFMERGE, store)
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

func testEvalHMGET(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"wrong number of args passed": {
			setup:  func() {},
			input:  nil,
			output: []byte("-ERR wrong number of arguments for 'hmget' command\r\n"),
		},
		"only key passed": {
			setup:  func() {},
			input:  []string{"KEY"},
			output: []byte("-ERR wrong number of arguments for 'hmget' command\r\n"),
		},
		"key doesn't exists": {
			setup:  func() {},
			input:  []string{"KEY", "field_name"},
			output: clientio.Encode([]interface{}{nil}, false),
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
			output: clientio.Encode([]interface{}{nil}, false),
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
			output: clientio.Encode([]interface{}{"mock_field_value"}, false),
		},
		"some fields exist some do not": {
			setup: func() {
				key := "KEY_MOCK"
				newMap := HashMap{
					"field1": "value1",
					"field2": "value2",
				}
				obj := &object.Obj{
					TypeEncoding:   object.ObjTypeHashMap | object.ObjEncodingHashMap,
					Value:          newMap,
					LastAccessedAt: uint32(time.Now().Unix()),
				}

				store.Put(key, obj)
			},
			input:  []string{"KEY_MOCK", "field1", "field2", "field3", "field4"},
			output: clientio.Encode([]interface{}{"value1", "value2", nil, nil}, false),
		},
	}

	runEvalTests(t, tests, evalHMGET, store)
}

func testEvalHVALS(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"wrong number of args passed": {
			setup:  func() {},
			input:  nil,
			output: []byte("-ERR wrong number of arguments for 'hvals' command\r\n"),
		},
		"key doesn't exists": {
			setup:  func() {},
			input:  []string{"NONEXISTENTHVALSKEY"},
			output: clientio.Encode([]string{}, false),
		},
		"key exists": {
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
			input:  []string{"KEY_MOCK"},
			output: clientio.Encode([]string{"mock_field_value"}, false),
		},
	}

	runEvalTests(t, tests, evalHVALS, store)
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

func testEvalHEXISTS(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"wrong number of args passed": {
			setup:  func() {},
			input:  nil,
			output: []byte("-ERR wrong number of arguments for 'hexists' command\r\n"),
		},
		"only key passed": {
			setup:  func() {},
			input:  []string{"KEY"},
			output: []byte("-ERR wrong number of arguments for 'hexists' command\r\n"),
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
			output: clientio.Encode(1, false),
		},
	}

	runEvalTests(t, tests, evalHEXISTS, store)
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

func testEvalHSCAN(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"HSCAN with wrong number of args": {
			input:  []string{"key"},
			output: []byte("-ERR wrong number of arguments for 'hscan' command\r\n"),
		},
		"HSCAN with key does not exist": {
			input:  []string{"NONEXISTENT_KEY", "0"},
			output: []byte("*2\r\n$1\r\n0\r\n*0\r\n"),
		},
		"HSCAN with key exists but not a hash": {
			setup: func() {
				evalSET([]string{"string_key", "string_value"}, store)
			},
			input:  []string{"string_key", "0"},
			output: []byte("-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"),
		},
		"HSCAN with valid key and cursor": {
			setup: func() {
				evalHSET([]string{"hash_key", "field1", "value1", "field2", "value2"}, store)
			},
			input:  []string{"hash_key", "0"},
			output: []byte("*2\r\n$1\r\n0\r\n*4\r\n$6\r\nfield1\r\n$6\r\nvalue1\r\n$6\r\nfield2\r\n$6\r\nvalue2\r\n"),
		},
		"HSCAN with cursor at the end": {
			setup: func() {
				evalHSET([]string{"hash_key", "field1", "value1", "field2", "value2"}, store)
			},
			input:  []string{"hash_key", "2"},
			output: []byte("*2\r\n$1\r\n0\r\n*0\r\n"),
		},
		"HSCAN with cursor at the beginning": {
			setup: func() {
				evalHSET([]string{"hash_key", "field1", "value1", "field2", "value2"}, store)
			},
			input:  []string{"hash_key", "0"},
			output: []byte("*2\r\n$1\r\n0\r\n*4\r\n$6\r\nfield1\r\n$6\r\nvalue1\r\n$6\r\nfield2\r\n$6\r\nvalue2\r\n"),
		},
		"HSCAN with cursor in the middle": {
			setup: func() {
				evalHSET([]string{"hash_key", "field1", "value1", "field2", "value2"}, store)
			},
			input:  []string{"hash_key", "1"},
			output: []byte("*2\r\n$1\r\n0\r\n*2\r\n$6\r\nfield2\r\n$6\r\nvalue2\r\n"),
		},
		"HSCAN with MATCH argument": {
			setup: func() {
				evalHSET([]string{"hash_key", "field1", "value1", "field2", "value2", "field3", "value3"}, store)
			},
			input:  []string{"hash_key", "0", "MATCH", "field[12]*"},
			output: []byte("*2\r\n$1\r\n0\r\n*4\r\n$6\r\nfield1\r\n$6\r\nvalue1\r\n$6\r\nfield2\r\n$6\r\nvalue2\r\n"),
		},
		"HSCAN with COUNT argument": {
			setup: func() {
				evalHSET([]string{"hash_key", "field1", "value1", "field2", "value2", "field3", "value3"}, store)
			},
			input:  []string{"hash_key", "0", "COUNT", "2"},
			output: []byte("*2\r\n$1\r\n2\r\n*4\r\n$6\r\nfield1\r\n$6\r\nvalue1\r\n$6\r\nfield2\r\n$6\r\nvalue2\r\n"),
		},
		"HSCAN with MATCH and COUNT arguments": {
			setup: func() {
				evalHSET([]string{"hash_key", "field1", "value1", "field2", "value2", "field3", "value3", "field4", "value4"}, store)
			},
			input:  []string{"hash_key", "0", "MATCH", "field[13]*", "COUNT", "1"},
			output: []byte("*2\r\n$1\r\n1\r\n*2\r\n$6\r\nfield1\r\n$6\r\nvalue1\r\n"),
		},
		"HSCAN with invalid MATCH pattern": {
			setup: func() {
				evalHSET([]string{"hash_key", "field1", "value1", "field2", "value2"}, store)
			},
			input:  []string{"hash_key", "0", "MATCH", "[invalid"},
			output: []byte("-ERR Invalid glob pattern: unexpected end of input\r\n"),
		},
		"HSCAN with invalid COUNT value": {
			setup: func() {
				evalHSET([]string{"hash_key", "field1", "value1", "field2", "value2"}, store)
			},
			input:  []string{"hash_key", "0", "COUNT", "invalid"},
			output: []byte("-ERR value is not an integer or out of range\r\n"),
		},
	}

	runEvalTests(t, tests, evalHSCAN, store)
}

func testEvalJSONSTRLEN(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"jsonstrlen nil value": {
			name:  "jsonstrlen nil value",
			setup: func() {},
			input: nil,
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("JSON.STRLEN"),
			},
		},
		"jsonstrlen key does not exist": {
			name:  "jsonstrlen key does not exist",
			setup: func() {},
			input: []string{"NONEXISTENT_KEY"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  nil,
			},
		},
		"jsonstrlen root not string(object)": {
			name: "jsonstrlen root not string(object)",
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"name\":\"Bhima\",\"age\":10}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input: []string{"EXISTING_KEY"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrUnexpectedJSONPathType("string", "object"),
			},
		},
		"jsonstrlen root not string(number)": {
			name: "jsonstrlen root not string(number)",
			setup: func() {
				key := "EXISTING_KEY"
				value := "10.9"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input: []string{"EXISTING_KEY"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrUnexpectedJSONPathType("string", "number"),
			},
		},
		"jsonstrlen root not string(integer)": {
			name: "jsonstrlen root not string(integer)",
			setup: func() {
				key := "EXISTING_KEY"
				value := "10"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input: []string{"EXISTING_KEY"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrUnexpectedJSONPathType("string", "integer"),
			},
		},
		"jsonstrlen not string(array)": {
			name: "jsonstrlen not string(array)",
			setup: func() {
				key := "EXISTING_KEY"
				value := "[\"age\", \"name\"]"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input: []string{"EXISTING_KEY"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrUnexpectedJSONPathType("string", "array"),
			},
		},
		"jsonstrlen not string(boolean)": {
			name: "jsonstrlen not string(boolean)",
			setup: func() {
				key := "EXISTING_KEY"
				value := "true"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input: []string{"EXISTING_KEY"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrUnexpectedJSONPathType("string", "boolean"),
			},
		},
		"jsonstrlen root array": {
			name: "jsonstrlen root array",
			setup: func() {
				key := "EXISTING_KEY"
				value := `"hello"`
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input: []string{"EXISTING_KEY"},
			migratedOutput: EvalResponse{
				Result: int64(5),
				Error:  nil,
			},
		},
		"jsonstrlen subpath string": {
			name: "jsonstrlen subpath string",
			setup: func() {
				key := "EXISTING_KEY"
				value := `{"partner":{"name":"tom","language":["rust"]}}`
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},

			input: []string{"EXISTING_KEY", "$..name"},
			migratedOutput: EvalResponse{
				Result: []interface{}{int64(3)},
				Error:  nil,
			},
		},
		"jsonstrlen subpath not string": {
			name: "jsonstrlen subpath not string",
			setup: func() {
				key := "EXISTING_KEY"
				value := `{"partner":{"name":21,"language":["rust"]}}`
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},

			input: []string{"EXISTING_KEY", "$..name"},
			migratedOutput: EvalResponse{
				Result: []interface{}{nil},
				Error:  nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store = setupTest(store)
			if tt.setup != nil {
				tt.setup()
			}

			response := evalJSONSTRLEN(tt.input, store)
			if tt.migratedOutput.Result != nil {
				if slice, ok := tt.migratedOutput.Result.([]interface{}); ok {
					assert.DeepEqual(t, slice, response.Result)
				} else {
					assert.Equal(t, tt.migratedOutput.Result, response.Result)
				}
			}

			if tt.migratedOutput.Error != nil {
				testifyAssert.EqualError(t, response.Error, tt.migratedOutput.Error.Error())
			} else {
				testifyAssert.NoError(t, response.Error)
			}
		})
	}
}

func testEvalLPUSH(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"nil value": {
			input:          nil,
			migratedOutput: EvalResponse{Result: clientio.NIL, Error: errors.New("ERR wrong number of arguments for 'lpush' command")},
		},
		"empty args": {
			input:          []string{},
			migratedOutput: EvalResponse{Result: clientio.NIL, Error: errors.New("ERR wrong number of arguments for 'lpush' command")},
		},
		"wrong number of args": {
			input:          []string{"KEY1"},
			migratedOutput: EvalResponse{Result: clientio.NIL, Error: errors.New("ERR wrong number of arguments for 'lpush' command")},
		},
		"key with different type": {
			setup: func() {
				evalSET([]string{"EXISTING_KEY", "mock_value"}, store)
			},
			input:          []string{"EXISTING_KEY", "value_1"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")},
		},
		"key does not exist": {
			input:          []string{"NONEXISTENT_KEY", "value_1"},
			migratedOutput: EvalResponse{Result: int64(1), Error: nil},
		},
		"key exists": {
			setup: func() {
				evalLPUSH([]string{"EXISTING_KEY", "mock_value"}, store)
			},
			input:          []string{"EXISTING_KEY", "value_1", "value_2"},
			migratedOutput: EvalResponse{Result: int64(3), Error: nil},
		},
	}
	runMigratedEvalTests(t, tests, evalLPUSH, store)
}
func testEvalRPUSH(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"nil value": {
			input:          nil,
			migratedOutput: EvalResponse{Result: clientio.NIL, Error: errors.New("ERR wrong number of arguments for 'rpush' command")},
		},
		"empty args": {
			input:          []string{},
			migratedOutput: EvalResponse{Result: clientio.NIL, Error: errors.New("ERR wrong number of arguments for 'rpush' command")},
		},
		"wrong number of args": {
			input:          []string{"KEY1"},
			migratedOutput: EvalResponse{Result: clientio.NIL, Error: errors.New("ERR wrong number of arguments for 'rpush' command")},
		},
		"key with different type": {
			setup: func() {
				evalSET([]string{"EXISTING_KEY", "mock_value"}, store)
			},
			input:          []string{"EXISTING_KEY", "value_1"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")},
		},
		"key does not exist": {
			input:          []string{"NONEXISTENT_KEY", "value_1"},
			migratedOutput: EvalResponse{Result: int64(1), Error: nil},
		},
		"key exists": {
			setup: func() {
				evalLPUSH([]string{"EXISTING_KEY", "mock_value"}, store)
			},
			input:          []string{"EXISTING_KEY", "value_1", "value_2"},
			migratedOutput: EvalResponse{Result: int64(3), Error: nil},
		},
	}
	runMigratedEvalTests(t, tests, evalRPUSH, store)
}
func testEvalLPOP(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"nil value": {
			input:          nil,
			migratedOutput: EvalResponse{Result: clientio.NIL, Error: errors.New("ERR wrong number of arguments for 'lpop' command")},
		},
		"empty args": {
			input:          []string{},
			migratedOutput: EvalResponse{Result: clientio.NIL, Error: errors.New("ERR wrong number of arguments for 'lpop' command")},
		},
		"wrong number of args": {
			input:          []string{"KEY1", "KEY2"},
			migratedOutput: EvalResponse{Result: clientio.NIL, Error: errors.New("ERR wrong number of arguments for 'lpop' command")},
		},
		"key with different type": {
			setup: func() {
				evalSET([]string{"EXISTING_KEY", "mock_value"}, store)
			},
			input:          []string{"EXISTING_KEY"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")},
		},
		"key does not exist": {
			input:          []string{"NONEXISTENT_KEY"},
			migratedOutput: EvalResponse{Result: clientio.NIL, Error: nil},
		},
		"key exists with 1 value": {
			setup: func() {
				evalLPUSH([]string{"EXISTING_KEY", "mock_value"}, store)
			},
			input:          []string{"EXISTING_KEY"},
			migratedOutput: EvalResponse{Result: "mock_value", Error: nil},
		},
		"key exists with multiple values": {
			setup: func() {
				evalLPUSH([]string{"EXISTING_KEY", "value_1", "value_2"}, store)
				evalRPUSH([]string{"EXISTING_KEY", "value_3", "value_4"}, store)
			},
			input:          []string{"EXISTING_KEY"},
			migratedOutput: EvalResponse{Result: "value_2", Error: nil},
		},
	}
	runMigratedEvalTests(t, tests, evalLPOP, store)
}
func testEvalRPOP(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"nil value": {
			input:          nil,
			migratedOutput: EvalResponse{Result: clientio.NIL, Error: errors.New("ERR wrong number of arguments for 'rpop' command")},
		},
		"empty args": {
			input:          []string{},
			migratedOutput: EvalResponse{Result: clientio.NIL, Error: errors.New("ERR wrong number of arguments for 'rpop' command")},
		},
		"wrong number of args": {
			input:          []string{"KEY1", "KEY2"},
			migratedOutput: EvalResponse{Result: clientio.NIL, Error: errors.New("ERR wrong number of arguments for 'rpop' command")},
		},
		"key with different type": {
			setup: func() {
				evalSET([]string{"EXISTING_KEY", "mock_value"}, store)
			},
			input:          []string{"EXISTING_KEY"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")},
		},
		"key does not exist": {
			input:          []string{"NONEXISTENT_KEY"},
			migratedOutput: EvalResponse{Result: clientio.NIL, Error: nil},
		},
		"key exists with 1 value": {
			setup: func() {
				evalLPUSH([]string{"EXISTING_KEY", "mock_value"}, store)
			},
			input:          []string{"EXISTING_KEY"},
			migratedOutput: EvalResponse{Result: "mock_value", Error: nil},
		},
		"key exists with multiple values": {
			setup: func() {
				evalLPUSH([]string{"EXISTING_KEY", "value_1", "value_2"}, store)
				evalRPUSH([]string{"EXISTING_KEY", "value_3", "value_4"}, store)
			},
			input:          []string{"EXISTING_KEY"},
			migratedOutput: EvalResponse{Result: "value_4", Error: nil},
		},
	}
	runMigratedEvalTests(t, tests, evalRPOP, store)
}

func testEvalLLEN(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"nil value": {
			input:          nil,
			migratedOutput: EvalResponse{Result: clientio.NIL, Error: errors.New("ERR wrong number of arguments for 'llen' command")},
		},
		"empty args": {
			input:          []string{},
			migratedOutput: EvalResponse{Result: clientio.NIL, Error: errors.New("ERR wrong number of arguments for 'llen' command")},
		},
		"wrong number of args": {
			input:          []string{"KEY1", "KEY2"},
			migratedOutput: EvalResponse{Result: clientio.NIL, Error: errors.New("ERR wrong number of arguments for 'llen' command")},
		},
		"key does not exist": {
			input:          []string{"NONEXISTENT_KEY"},
			migratedOutput: EvalResponse{Result: clientio.IntegerZero, Error: nil},
		},
		"key exists": {
			setup: func() {
				evalLPUSH([]string{"EXISTING_KEY", "mock_value"}, store)
			},
			input:          []string{"EXISTING_KEY"},
			migratedOutput: EvalResponse{Result: int64(1), Error: nil},
		},
		"key with different type": {
			setup: func() {
				evalSET([]string{"EXISTING_KEY", "mock_value"}, store)
			},
			input:          []string{"EXISTING_KEY"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")},
		},
	}

	runMigratedEvalTests(t, tests, evalLLEN, store)
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

func runMigratedEvalTests(t *testing.T, tests map[string]evalTestCase, evalFunc func([]string, *dstore.Store) *EvalResponse, store *dstore.Store) {
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			store = setupTest(store)

			if tc.setup != nil {
				tc.setup()
			}

			output := evalFunc(tc.input, store)

			if tc.newValidator != nil {
				if tc.migratedOutput.Error != nil {
					tc.newValidator(tc.migratedOutput.Error)
				} else {
					tc.newValidator(output.Result)
				}
				return
			}

			if tc.migratedOutput.Error != nil {
				testifyAssert.EqualError(t, output.Error, tc.migratedOutput.Error.Error())
				return
			}

			// Handle comparison for byte slices and string slices
			// TODO: Make this generic so that all kind of slices can be handled
			if b, ok := output.Result.([]byte); ok && tc.migratedOutput.Result != nil {
				if expectedBytes, ok := tc.migratedOutput.Result.([]byte); ok {
					testifyAssert.True(t, bytes.Equal(b, expectedBytes), "expected and actual byte slices should be equal")
				}
			} else if a, ok := output.Result.([]string); ok && tc.migratedOutput.Result != nil {
				if expectedStringSlice, ok := tc.migratedOutput.Result.([]string); ok {
					testifyAssert.ElementsMatch(t, a, expectedStringSlice)
				}
			} else {
				testifyAssert.Equal(t, tc.migratedOutput.Result, output.Result)
			}

			testifyAssert.NoError(t, output.Error)
		})
	}
}

func BenchmarkEvalMSET(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store := dstore.NewStore(nil, nil)
		evalMSET([]string{"KEY", "VAL", "KEY2", "VAL2"}, store)
	}
}

func BenchmarkEvalHSET(b *testing.B) {
	store := dstore.NewStore(nil, nil)
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
				mockValue := "mock_field_value"
				newMap := make(HashMap)
				newMap[field] = mockValue

				obj := &object.Obj{
					TypeEncoding:   object.ObjTypeHashMap | object.ObjEncodingHashMap,
					Value:          newMap,
					LastAccessedAt: uint32(time.Now().Unix()),
				}

				store.Put(key, obj)

				// Check if the map is saved correctly in the store
				res, err := getValueFromHashMap(key, field, store)

				assert.Assert(t, err == nil)
				assert.DeepEqual(t, res, clientio.Encode(mockValue, false))
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

func testEvalHMSET(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"wrong number of args passed": {
			setup:  func() {},
			input:  nil,
			output: []byte("-ERR wrong number of arguments for 'hmset' command\r\n"),
		},
		"only key passed": {
			setup:  func() {},
			input:  []string{"key"},
			output: []byte("-ERR wrong number of arguments for 'hmset' command\r\n"),
		},
		"only key and field_name passed": {
			setup:  func() {},
			input:  []string{"KEY", "field_name"},
			output: []byte("-ERR wrong number of arguments for 'hmset' command\r\n"),
		},
		"key, field and value passed": {
			setup:  func() {},
			input:  []string{"KEY1", "field_name", "value"},
			output: clientio.RespOK,
		},
		"key, field and value updated": {
			setup:  func() {},
			input:  []string{"KEY1", "field_name", "value_new"},
			output: clientio.RespOK,
		},
		"new set of key, field and value added": {
			setup:  func() {},
			input:  []string{"KEY2", "field_name_new", "value_new_new"},
			output: clientio.RespOK,
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
			output: clientio.RespOK,
		},
		"same key -> update value, add new field and value": {
			setup: func() {
				key := "KEY_MOCK"
				field := "mock_field_name"
				mockValue := "mock_field_value"
				newMap := make(HashMap)
				newMap[field] = mockValue

				obj := &object.Obj{
					TypeEncoding:   object.ObjTypeHashMap | object.ObjEncodingHashMap,
					Value:          newMap,
					LastAccessedAt: uint32(time.Now().Unix()),
				}

				store.Put(key, obj)

				// Check if the map is saved correctly in the store
				res, err := getValueFromHashMap(key, field, store)

				assert.Assert(t, err == nil)
				assert.DeepEqual(t, res, clientio.Encode(mockValue, false))
			},
			input: []string{
				"KEY_MOCK",
				"mock_field_name",
				"mock_field_value_new",
				"mock_field_name_new",
				"mock_value_new",
			},
			output: clientio.RespOK,
		},
	}

	runEvalTests(t, tests, evalHMSET, store)
}

func testEvalHKEYS(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"wrong number of args passed": {
			setup:  func() {},
			input:  nil,
			output: []byte("-ERR wrong number of arguments for 'hkeys' command\r\n"),
		},
		"key doesn't exist": {
			setup:  func() {},
			input:  []string{"KEY"},
			output: clientio.Encode([]string{}, false),
		},
		"key exists but not a hash": {
			setup: func() {
				evalSET([]string{"string_key", "string_value"}, store)
			},
			input:  []string{"string_key"},
			output: []byte("-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"),
		},
		"key exists and is a hash": {
			setup: func() {
				key := "KEY_MOCK"
				field1 := "mock_field_name"
				newMap := make(HashMap)
				newMap[field1] = "HelloWorld"

				obj := &object.Obj{
					TypeEncoding:   object.ObjTypeHashMap | object.ObjEncodingHashMap,
					Value:          newMap,
					LastAccessedAt: uint32(time.Now().Unix()),
				}

				store.Put(key, obj)
			},
			input:  []string{"KEY_MOCK"},
			output: clientio.Encode([]string{"mock_field_name"}, false),
		},
	}

	runEvalTests(t, tests, evalHKEYS, store)
}

func BenchmarkEvalHKEYS(b *testing.B) {
	store := dstore.NewStore(nil, nil)

	for i := 0; i < b.N; i++ {
		evalHSET([]string{"KEY", fmt.Sprintf("FIELD_%d", i), fmt.Sprintf("VALUE_%d", i)}, store)
	}
	// Benchmark HKEYS
	for i := 0; i < b.N; i++ {
		evalHKEYS([]string{"KEY"}, store)
	}
}

func BenchmarkEvalPFCOUNT(b *testing.B) {
	store := *dstore.NewStore(nil, nil)

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

		"memory nonexistent key": {
			setup:  func() {},
			input:  []string{"MEMORY", "NONEXISTENT_KEY"},
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
	store := dstore.NewStore(nil, nil)

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
			validator: func(output []byte) {
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
			validator: func(output []byte) {
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
			output: []byte("+none\r\n"),
		},
		"TYPE : key exists and is of type String": {
			setup: func() {
				store.Put("string_key", store.NewObj("value", -1, object.ObjTypeString, object.ObjEncodingRaw))
			},
			input:  []string{"string_key"},
			output: []byte("+string\r\n"),
		},
		"TYPE : key exists and is of type List": {
			setup: func() {
				store.Put("list_key", store.NewObj([]byte("value"), -1, object.ObjTypeByteList, object.ObjEncodingRaw))
			},
			input:  []string{"list_key"},
			output: []byte("+list\r\n"),
		},
		"TYPE : key exists and is of type Set": {
			setup: func() {
				store.Put("set_key", store.NewObj([]byte("value"), -1, object.ObjTypeSet, object.ObjEncodingRaw))
			},
			input:  []string{"set_key"},
			output: []byte("+set\r\n"),
		},
		"TYPE : key exists and is of type Hash": {
			setup: func() {
				store.Put("hash_key", store.NewObj([]byte("value"), -1, object.ObjTypeHashMap, object.ObjEncodingRaw))
			},
			input:  []string{"hash_key"},
			output: []byte("+hash\r\n"),
		},
	}
	runEvalTests(t, tests, evalTYPE, store)
}

func BenchmarkEvalTYPE(b *testing.B) {
	store := dstore.NewStore(nil, nil)

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
			// Set up the object in the store
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
			output: []byte("*13\r\n" +
				"$64\r\n" +
				"COMMAND <subcommand> [<arg> [value] [opt] ...]. Subcommands are:\r\n" +
				"$15\r\n" +
				"(no subcommand)\r\n" +
				"$46\r\n" +
				"     Return details about all DiceDB commands.\r\n" +
				"$5\r\n" +
				"COUNT\r\n" +
				"$63\r\n" +
				"     Return the total number of commands in this DiceDB server.\r\n" +
				"$4\r\n" +
				"LIST\r\n" +
				"$57\r\n" +
				"     Return a list of all commands in this DiceDB server.\r\n" +
				"$25\r\n" +
				"INFO [<command-name> ...]\r\n" +
				"$140\r\n" +
				"     Return details about the specified DiceDB commands. If no command names are given, documentation details for all commands are returned.\r\n" +
				"$22\r\n" +
				"GETKEYS <full-command>\r\n" +
				"$48\r\n" +
				"     Return the keys from a full DiceDB command.\r\n" +
				"$4\r\n" +
				"HELP\r\n" +
				"$21\r\n" +
				"     Print this help.\r\n"),
		},
		"command help with wrong number of arguments": {
			input:  []string{"HELP", "EXTRA-ARGS"},
			output: []byte("-ERR wrong number of arguments for 'command|help' command\r\n"),
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
		"command count with wrong number of arguments": {
			input:  []string{"COUNT", "EXTRA-ARGS"},
			output: []byte("-ERR wrong number of arguments for 'command|count' command\r\n"),
		},
		"command list with wrong number of arguments": {
			input:  []string{"LIST", "EXTRA-ARGS"},
			output: []byte("-ERR wrong number of arguments for 'command|list' command\r\n"),
		},
		"command unknown": {
			input:  []string{"UNKNOWN"},
			output: []byte("-ERR unknown subcommand 'UNKNOWN'. Try COMMAND HELP.\r\n"),
		},
		"command getkeys with incorrect number of arguments": {
			input:  []string{"GETKEYS"},
			output: []byte("-ERR wrong number of arguments for 'command|getkeys' command\r\n"),
		},
		"command getkeys with unknown command": {
			input:  []string{"GETKEYS", "UNKNOWN"},
			output: []byte("-ERR invalid command specified\r\n"),
		},
		"command getkeys with a command that accepts no key arguments": {
			input:  []string{"GETKEYS", "FLUSHDB"},
			output: []byte("-ERR the command has no key arguments\r\n"),
		},
		"command getkeys with an invalid number of arguments for a command": {
			input:  []string{"GETKEYS", "MSET", "key1"},
			output: []byte("-ERR invalid number of arguments specified for command\r\n"),
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
			output: []byte("-ERR could not perform this operation on a key that doesn't exist\r\n"),
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
	store := dstore.NewStore(nil, nil)

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
			input: []string{"NON_EXISTING_KEY", "0", "-1"},
			migratedOutput: EvalResponse{
				Result: "",
				Error:  nil,
			},
		},
		"GETRANGE against wrong key type": {
			setup: func() {
				evalLPUSH([]string{"LKEY1", "list"}, store)
			},
			input: []string{"LKEY1", "0", "-1"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongTypeOperation,
			},
		},
		"GETRANGE against string value: 0, 3": {
			setup: setupForStringValue,
			input: []string{"STRING_KEY", "0", "3"},
			migratedOutput: EvalResponse{
				Result: "Hell",
				Error:  nil,
			},
		},
		"GETRANGE against string value: 0, -1": {
			setup: setupForStringValue,
			input: []string{"STRING_KEY", "0", "-1"},
			migratedOutput: EvalResponse{
				Result: "Hello World",
				Error:  nil,
			},
		},
		"GETRANGE against string value: -4, -1": {
			setup: setupForStringValue,
			input: []string{"STRING_KEY", "-4", "-1"},
			migratedOutput: EvalResponse{
				Result: "orld",
				Error:  nil,
			},
		},
		"GETRANGE against string value: 5, 3": {
			setup: setupForStringValue,
			input: []string{"STRING_KEY", "5", "3"},
			migratedOutput: EvalResponse{
				Result: "",
				Error:  nil,
			},
		},
		"GETRANGE against string value: -5000, 10000": {
			setup: setupForStringValue,
			input: []string{"STRING_KEY", "-5000", "10000"},
			migratedOutput: EvalResponse{
				Result: "Hello World",
				Error:  nil,
			},
		},
		"GETRANGE against string value: 0, -100": {
			setup: setupForStringValue,
			input: []string{"STRING_KEY", "0", "-100"},
			migratedOutput: EvalResponse{
				Result: "",
				Error:  nil,
			},
		},
		"GETRANGE against string value: 1, -100": {
			setup: setupForStringValue,
			input: []string{"STRING_KEY", "1", "-100"},
			migratedOutput: EvalResponse{
				Result: "",
				Error:  nil,
			},
		},
		"GETRANGE against string value: -1, -100": {
			setup: setupForStringValue,
			input: []string{"STRING_KEY", "-1", "-100"},
			migratedOutput: EvalResponse{
				Result: "",
				Error:  nil,
			},
		},
		"GETRANGE against string value: -100, -100": {
			setup: setupForStringValue,
			input: []string{"STRING_KEY", "-100", "-100"},
			migratedOutput: EvalResponse{
				Result: "",
				Error:  nil,
			},
		},
		"GETRANGE against string value: -100, -101": {
			setup: setupForStringValue,
			input: []string{"STRING_KEY", "-100", "-101"},
			migratedOutput: EvalResponse{
				Result: "",
				Error:  nil,
			},
		},
		"GETRANGE against integer value: 0, 2": {
			setup: setupForIntegerValue,
			input: []string{"INTEGER_KEY", "0", "2"},
			migratedOutput: EvalResponse{
				Result: "123",
				Error:  nil,
			},
		},
		"GETRANGE against integer value: 0, -1": {
			setup: setupForIntegerValue,
			input: []string{"INTEGER_KEY", "0", "-1"},
			migratedOutput: EvalResponse{
				Result: "1234",
				Error:  nil,
			},
		},
		"GETRANGE against integer value: -3, -1": {
			setup: setupForIntegerValue,
			input: []string{"INTEGER_KEY", "-3", "-1"},
			migratedOutput: EvalResponse{
				Result: "234",
				Error:  nil,
			},
		},
		"GETRANGE against integer value: 5, 3": {
			setup: setupForIntegerValue,
			input: []string{"INTEGER_KEY", "5", "3"},
			migratedOutput: EvalResponse{
				Result: "",
				Error:  nil,
			},
		},
		"GETRANGE against integer value: 3, 5000": {
			setup: setupForIntegerValue,
			input: []string{"INTEGER_KEY", "3", "5000"},
			migratedOutput: EvalResponse{
				Result: "4",
				Error:  nil,
			},
		},
		"GETRANGE against integer value: -5000, 10000": {
			setup: setupForIntegerValue,
			input: []string{"INTEGER_KEY", "-5000", "10000"},
			migratedOutput: EvalResponse{
				Result: "1234",
				Error:  nil,
			},
		},
		"GETRANGE against integer value: 0, -100": {
			setup: setupForIntegerValue,
			input: []string{"INTEGER_KEY", "0", "-100"},
			migratedOutput: EvalResponse{
				Result: "",
				Error:  nil,
			},
		},
		"GETRANGE against integer value: 1, -100": {
			setup: setupForIntegerValue,
			input: []string{"INTEGER_KEY", "1", "-100"},
			migratedOutput: EvalResponse{
				Result: "",
				Error:  nil,
			},
		},
		"GETRANGE against integer value: -1, -100": {
			setup: setupForIntegerValue,
			input: []string{"INTEGER_KEY", "-1", "-100"},
			migratedOutput: EvalResponse{
				Result: "",
				Error:  nil,
			},
		},
		"GETRANGE against integer value: -100, -99": {
			setup: setupForIntegerValue,
			input: []string{"INTEGER_KEY", "-100", "-99"},
			migratedOutput: EvalResponse{
				Result: "",
				Error:  nil,
			},
		},
		"GETRANGE against integer value: -100, -100": {
			setup: setupForIntegerValue,
			input: []string{"INTEGER_KEY", "-100", "-100"},
			migratedOutput: EvalResponse{
				Result: "",
				Error:  nil,
			},
		},
		"GETRANGE against integer value: -100, -101": {
			setup: setupForIntegerValue,
			input: []string{"INTEGER_KEY", "-100", "-101"},
			migratedOutput: EvalResponse{
				Result: "",
				Error:  nil,
			},
		},
	}

	runMigratedEvalTests(t, tests, evalGETRANGE, store)
}

func BenchmarkEvalGETRANGE(b *testing.B) {
	store := dstore.NewStore(nil, nil)
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

func BenchmarkEvalHSETNX(b *testing.B) {
	store := dstore.NewStore(nil, nil)
	for i := 0; i < b.N; i++ {
		evalHSETNX([]string{"KEY", fmt.Sprintf("FIELD_%d", i/2), fmt.Sprintf("VALUE_%d", i)}, store)
	}
}

func testEvalHSETNX(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"no args passed": {
			setup:  func() {},
			input:  nil,
			output: []byte("-ERR wrong number of arguments for 'hsetnx' command\r\n"),
		},
		"only key passed": {
			setup:  func() {},
			input:  []string{"key"},
			output: []byte("-ERR wrong number of arguments for 'hsetnx' command\r\n"),
		},
		"only key and field_name passed": {
			setup:  func() {},
			input:  []string{"KEY", "field_name"},
			output: []byte("-ERR wrong number of arguments for 'hsetnx' command\r\n"),
		},
		"more than one field and value passed": {
			setup:  func() {},
			input:  []string{"KEY", "field1", "value1", "field2", "value2"},
			output: []byte("-ERR wrong number of arguments for 'hsetnx' command\r\n"),
		},
		"key, field and value passed": {
			setup:  func() {},
			input:  []string{"KEY1", "field_name", "value"},
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
			input:  []string{"KEY_MOCK", "mock_field_name", "mock_field_value_2"},
			output: clientio.Encode(int64(0), false),
		},
	}

	runEvalTests(t, tests, evalHSETNX, store)
}

func TestMSETConsistency(t *testing.T) {
	store := dstore.NewStore(nil, nil)
	evalMSET([]string{"KEY", "VAL", "KEY2", "VAL2"}, store)

	assert.Equal(t, "VAL", store.Get("KEY").Value)
	assert.Equal(t, "VAL2", store.Get("KEY2").Value)
}

func BenchmarkEvalHINCRBY(b *testing.B) {
	store := dstore.NewStore(nil, nil)

	// creating new fields
	for i := 0; i < b.N; i++ {
		evalHINCRBY([]string{"KEY", fmt.Sprintf("FIELD_%d", i), fmt.Sprintf("%d", i)}, store)
	}

	// updating the existing fields
	for i := 0; i < b.N; i++ {
		evalHINCRBY([]string{"KEY", fmt.Sprintf("FIELD_%d", i), fmt.Sprintf("%d", i*10)}, store)
	}
}

func testEvalHINCRBY(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"invalid number of args passed": {
			setup:          func() {},
			input:          nil,
			migratedOutput: EvalResponse{Result: nil, Error: diceerrors.ErrWrongArgumentCount("HINCRBY")},
		},
		"only key is passed in args": {
			setup:          func() {},
			input:          []string{"key"},
			migratedOutput: EvalResponse{Result: nil, Error: diceerrors.ErrWrongArgumentCount("HINCRBY")},
		},
		"only key and field is passed in args": {
			setup:          func() {},
			input:          []string{"key field"},
			migratedOutput: EvalResponse{Result: nil, Error: diceerrors.ErrWrongArgumentCount("HINCRBY")},
		},
		"key, field and increment passed in args": {
			setup:          func() {},
			input:          []string{"key", "field", "10"},
			migratedOutput: EvalResponse{Result: int64(10), Error: nil},
		},
		"update the already existing field in the key": {
			setup: func() {
				key := "key"
				field := "field"
				h := make(HashMap)
				h[field] = "10"
				obj := &object.Obj{
					TypeEncoding:   object.ObjTypeHashMap | object.ObjEncodingHashMap,
					Value:          h,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)
			},
			input:          []string{"key", "field", "10"},
			migratedOutput: EvalResponse{Result: int64(20), Error: nil},
		},
		"increment value is not int64": {
			setup:          func() {},
			input:          []string{"key", "field", "hello"},
			migratedOutput: EvalResponse{Result: nil, Error: diceerrors.ErrIntegerOutOfRange},
		},
		"increment value is greater than the bound of int64": {
			setup:          func() {},
			input:          []string{"key", "field", "99999999999999999999999999999999999999999999999999999"},
			migratedOutput: EvalResponse{Result: nil, Error: diceerrors.ErrIntegerOutOfRange},
		},
		"update the existing field whose datatype is not int64": {
			setup: func() {
				key := "new_key"
				field := "new_field"
				newMap := make(HashMap)
				newMap[field] = "new_value"
				obj := &object.Obj{
					TypeEncoding:   object.ObjTypeHashMap | object.ObjEncodingHashMap,
					Value:          newMap,
					LastAccessedAt: uint32(time.Now().Unix()),
				}

				store.Put(key, obj)
			},
			input:          []string{"new_key", "new_field", "10"},
			migratedOutput: EvalResponse{Result: nil, Error: diceerrors.ErrHashValueNotInteger},
		},
		"update the existing field which has spaces": {
			setup: func() {
				key := "key"
				field := "field"
				h := make(HashMap)
				h[field] = " 10  "

				obj := &object.Obj{
					TypeEncoding:   object.ObjTypeHashMap | object.ObjEncodingHashMap,
					Value:          h,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)
			},
			input:          []string{"key", "field", "10"},
			migratedOutput: EvalResponse{Result: nil, Error: diceerrors.ErrHashValueNotInteger},
		},
		"updating the new field with negative value": {
			setup:          func() {},
			input:          []string{"key", "field", "-10"},
			migratedOutput: EvalResponse{Result: int64(-10), Error: nil},
		},
		"update the existing field with negative value": {
			setup: func() {
				key := "key"
				field := "field"
				h := make(HashMap)

				h[field] = "-10"
				obj := &object.Obj{
					TypeEncoding:   object.ObjTypeHashMap | object.ObjEncodingHashMap,
					Value:          h,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)
			},
			input:          []string{"key", "field", "-10"},
			migratedOutput: EvalResponse{Result: int64(-20), Error: nil},
		},
		"updating the existing field which would lead to positive overflow": {
			setup: func() {
				key := "key"
				field := "field"
				h := make(HashMap)

				h[field] = fmt.Sprintf("%v", math.MaxInt64)
				obj := &object.Obj{
					TypeEncoding:   object.ObjTypeHashMap | object.ObjEncodingHashMap,
					Value:          h,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)
			},
			input:          []string{"key", "field", "10"},
			migratedOutput: EvalResponse{Result: nil, Error: diceerrors.ErrOverflow},
		},
		"updating the existing field which would lead to negative overflow": {
			setup: func() {
				key := "key"
				field := "field"
				h := make(HashMap)

				h[field] = fmt.Sprintf("%v", math.MinInt64)
				obj := &object.Obj{
					TypeEncoding:   object.ObjTypeHashMap | object.ObjEncodingHashMap,
					Value:          h,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)
			},
			input:          []string{"key", "field", "-10"},
			migratedOutput: EvalResponse{Result: nil, Error: diceerrors.ErrOverflow},
		},
	}

	runMigratedEvalTests(t, tests, evalHINCRBY, store)
}

func testEvalSETEX(t *testing.T, store *dstore.Store) {
	mockTime := &utils.MockClock{CurrTime: time.Now()}
	utils.CurrentTime = mockTime

	tests := map[string]evalTestCase{
		"nil value":                              {input: nil, migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR wrong number of arguments for 'setex' command")}},
		"empty array":                            {input: []string{}, migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR wrong number of arguments for 'setex' command")}},
		"one value":                              {input: []string{"KEY"}, migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR wrong number of arguments for 'setex' command")}},
		"key val pair":                           {input: []string{"KEY", "VAL"}, migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR wrong number of arguments for 'setex' command")}},
		"key exp pair":                           {input: []string{"KEY", "123456"}, migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR wrong number of arguments for 'setex' command")}},
		"key exp value pair":                     {input: []string{"KEY", "123", "VAL"}, migratedOutput: EvalResponse{Result: clientio.OK, Error: nil}},
		"key exp value pair with extra args":     {input: []string{"KEY", "123", "VAL", " "}, migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR wrong number of arguments for 'setex' command")}},
		"key exp value pair with invalid exp":    {input: []string{"KEY", "0", "VAL"}, migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR invalid expire time in 'setex' command")}},
		"key exp value pair with exp > maxexp":   {input: []string{"KEY", "9223372036854776", "VAL"}, migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR invalid expire time in 'setex' command")}},
		"key exp value pair with exp > maxint64": {input: []string{"KEY", "92233720368547760000000", "VAL"}, migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR value is not an integer or out of range")}},
		"key exp value pair with negative exp":   {input: []string{"KEY", "-23", "VAL"}, migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR invalid expire time in 'setex' command")}},
		"key exp value pair with not-int exp":    {input: []string{"KEY", "12a", "VAL"}, migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR value is not an integer or out of range")}},

		"set and get": {
			setup: func() {},
			input: []string{"TEST_KEY", "5", "TEST_VALUE"},
			newValidator: func(output interface{}) {
				assert.Equal(t, clientio.OK, output)

				// Check if the key was set correctly
				getValue := evalGET([]string{"TEST_KEY"}, store)
				assert.Equal(t, "TEST_VALUE", getValue.Result)

				// Check if the TTL is set correctly (should be 5 seconds or less)
				ttlValue := evalTTL([]string{"TEST_KEY"}, store)
				ttl, err := strconv.Atoi(strings.TrimPrefix(strings.TrimSpace(string(ttlValue)), ":"))
				assert.NilError(t, err, "Failed to parse TTL")
				assert.Assert(t, ttl > 0 && ttl <= 5)

				// Wait for the key to expire
				mockTime.SetTime(mockTime.CurrTime.Add(6 * time.Second))

				// Check if the key has been deleted after expiry
				expiredValue := evalGET([]string{"TEST_KEY"}, store)
				assert.Equal(t, clientio.NIL, expiredValue.Result)
			},
		},
		"update existing key": {
			setup: func() {
				evalSET([]string{"EXISTING_KEY", "OLD_VALUE"}, store)
			},
			input: []string{"EXISTING_KEY", "10", "NEW_VALUE"},
			newValidator: func(output interface{}) {
				assert.Equal(t, clientio.OK, output)

				// Check if the key was updated correctly
				getValue := evalGET([]string{"EXISTING_KEY"}, store)
				assert.Equal(t, "NEW_VALUE", getValue.Result)

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

			if tt.newValidator != nil {
				if tt.migratedOutput.Error != nil {
					tt.newValidator(tt.migratedOutput.Error)
				} else {
					tt.newValidator(response.Result)
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
	store := dstore.NewStore(nil, nil)

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
	tests := []evalTestCase{
		{
			name:           "INCRBYFLOAT on a non existing key",
			input:          []string{"float", "0.1"},
			migratedOutput: EvalResponse{Result: "0.1", Error: nil},
		},
		{
			name: "INCRBYFLOAT on an existing key",
			setup: func() {
				key := "key"
				value := "2.1"
				obj := store.NewObj(value, -1, object.ObjTypeString, object.ObjEncodingRaw)
				store.Put(key, obj)
			},
			input:          []string{"key", "0.1"},
			migratedOutput: EvalResponse{Result: "2.2", Error: nil},
		},
		{
			name: "INCRBYFLOAT on a key with integer value",
			setup: func() {
				key := "key"
				value := "2"
				obj := store.NewObj(value, -1, object.ObjTypeInt, object.ObjEncodingInt)
				store.Put(key, obj)
			},
			input:          []string{"key", "0.1"},
			migratedOutput: EvalResponse{Result: "2.1", Error: nil},
		},
		{
			name: "INCRBYFLOAT by a negative increment",
			setup: func() {
				key := "key"
				value := "2"
				obj := store.NewObj(value, -1, object.ObjTypeInt, object.ObjEncodingInt)
				store.Put(key, obj)
			},
			input:          []string{"key", "-0.1"},
			migratedOutput: EvalResponse{Result: "1.9", Error: nil},
		},
		{
			name: "INCRBYFLOAT by a scientific notation increment",
			setup: func() {
				key := "key"
				value := "1"
				obj := store.NewObj(value, -1, object.ObjTypeInt, object.ObjEncodingInt)
				store.Put(key, obj)
			},
			input:          []string{"key", "1e-2"},
			migratedOutput: EvalResponse{Result: "1.01", Error: nil},
		},
		{
			name: "INCRBYFLOAT on a key holding a scientific notation value",
			setup: func() {
				key := "key"
				value := "1e2"
				obj := store.NewObj(value, -1, object.ObjTypeString, object.ObjEncodingEmbStr)
				store.Put(key, obj)
			},
			input:          []string{"key", "1e-1"},
			migratedOutput: EvalResponse{Result: "100.1", Error: nil},
		},
		{
			name: "INCRBYFLOAT by an negative increment of the same value",
			setup: func() {
				key := "key"
				value := "0.1"
				obj := store.NewObj(value, -1, object.ObjTypeString, object.ObjEncodingEmbStr)
				store.Put(key, obj)
			},
			input:          []string{"key", "-0.1"},
			migratedOutput: EvalResponse{Result: "0", Error: nil},
		},
		{
			name: "INCRBYFLOAT on a key with spaces",
			setup: func() {
				key := "key"
				value := "   2   "
				obj := store.NewObj(value, -1, object.ObjTypeString, object.ObjEncodingEmbStr)
				store.Put(key, obj)
			},
			input:          []string{"key", "0.1"},
			migratedOutput: EvalResponse{Result: nil, Error: diceerrors.ErrGeneral("value is not a valid float")},
		},
		{
			name: "INCRBYFLOAT on a key with non numeric value",
			setup: func() {
				key := "key"
				value := "string"
				obj := store.NewObj(value, -1, object.ObjTypeString, object.ObjEncodingEmbStr)
				store.Put(key, obj)
			},
			input:          []string{"key", "0.1"},
			migratedOutput: EvalResponse{Result: nil, Error: diceerrors.ErrGeneral("value is not a valid float")},
		},
		{
			name: "INCRBYFLOAT by a non numeric increment",
			setup: func() {
				key := "key"
				value := "2.0"
				obj := store.NewObj(value, -1, object.ObjTypeString, object.ObjEncodingEmbStr)
				store.Put(key, obj)
			},
			input:          []string{"key", "a"},
			migratedOutput: EvalResponse{Result: nil, Error: diceerrors.ErrGeneral("value is not a valid float")},
		},
		{
			name: "INCRBYFLOAT by a number that would turn float64 to Inf",
			setup: func() {
				key := "key"
				value := "1e308"
				obj := store.NewObj(value, -1, object.ObjTypeString, object.ObjEncodingEmbStr)
				store.Put(key, obj)
			},
			input:          []string{"key", "1e308"},
			migratedOutput: EvalResponse{Result: nil, Error: diceerrors.ErrValueOutOfRange},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store = setupTest(store)
			if tt.setup != nil {
				tt.setup()
			}

			response := evalINCRBYFLOAT(tt.input, store)
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

func BenchmarkEvalINCRBYFLOAT(b *testing.B) {
	store := dstore.NewStore(nil, nil)
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
	store := dstore.NewStore(nil, nil)

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

func testEvalHRANDFIELD(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"wrong number of args passed": {
			setup:          func() {},
			input:          nil,
			migratedOutput: EvalResponse{Result: nil, Error: diceerrors.ErrWrongArgumentCount("HRANDFIELD")},
		},
		"key doesn't exist": {
			setup: func() {},
			input: []string{"KEY"},
			migratedOutput: EvalResponse{
				Result: clientio.RespType(0),
				Error:  nil,
			},
		},
		"key exists with fields and no count argument": {
			setup: func() {
				key := "KEY_MOCK"
				newMap := make(HashMap)
				newMap["field1"] = "Value1"
				newMap["field2"] = "Value2"

				obj := &object.Obj{
					TypeEncoding:   object.ObjTypeHashMap | object.ObjEncodingHashMap,
					Value:          newMap,
					LastAccessedAt: uint32(time.Now().Unix()),
				}

				store.Put(key, obj)
			},
			input: []string{"KEY_MOCK"},
			newValidator: func(output interface{}) {
				assert.Assert(t, output != nil)
				stringSlice, ok := output.([]string)
				if !ok {
					testifyAssert.Error(t, diceerrors.ErrUnexpectedType("[]string", reflect.TypeOf(output)))
				}
				resultString := strings.Join(stringSlice, " ")
				assert.Assert(t,
					resultString == "field1" || resultString == "field2",
					"Unexpected field returned: %s", resultString)
			},
		},
		"key exists with fields and count argument": {
			setup: func() {
				key := "KEY_MOCK"
				newMap := make(HashMap)
				newMap["field1"] = "value1"
				newMap["field2"] = "value2"
				newMap["field3"] = "value3"

				obj := &object.Obj{
					TypeEncoding:   object.ObjTypeHashMap | object.ObjEncodingHashMap,
					Value:          newMap,
					LastAccessedAt: uint32(time.Now().Unix()),
				}

				store.Put(key, obj)
			},
			input: []string{"KEY_MOCK", "2"},
			newValidator: func(output interface{}) {
				assert.Assert(t, output != nil)
				stringSlice, ok := output.([]string)
				if !ok {
					testifyAssert.Error(t, diceerrors.ErrUnexpectedType("[]string", reflect.TypeOf(output)))
				}
				decodedResult := strings.Join(stringSlice, " ")
				fields := []string{"field1", "field2", "field3"}
				count := 0

				for _, field := range fields {
					if strings.Contains(decodedResult, field) {
						count++
					}
				}

				assert.Assert(t, count == 2)
			},
		},
		"key exists with count and WITHVALUES argument": {
			setup: func() {
				key := "KEY_MOCK"
				newMap := make(HashMap)
				newMap["field1"] = "value1"
				newMap["field2"] = "value2"
				newMap["field3"] = "value3"

				obj := &object.Obj{
					TypeEncoding:   object.ObjTypeHashMap | object.ObjEncodingHashMap,
					Value:          newMap,
					LastAccessedAt: uint32(time.Now().Unix()),
				}

				store.Put(key, obj)
			},
			input: []string{"KEY_MOCK", "2", WithValues},
			newValidator: func(output interface{}) {
				assert.Assert(t, output != nil)
				stringSlice, ok := output.([]string)
				if !ok {
					testifyAssert.Error(t, diceerrors.ErrUnexpectedType("[]string", reflect.TypeOf(output)))
				}
				decodedResult := strings.Join(stringSlice, " ")
				fieldsAndValues := []string{"field1", "value1", "field2", "value2", "field3", "value3"}
				count := 0
				for _, item := range fieldsAndValues {
					if strings.Contains(decodedResult, item) {
						count++
					}
				}

				assert.Equal(t, 4, count, "Expected 4 fields and values, found %d", count)
			},
		},
	}

	runMigratedEvalTests(t, tests, evalHRANDFIELD, store)
}

func testEvalAPPEND(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"nil value": {
			setup:          func() {},
			input:          nil,
			migratedOutput: EvalResponse{Result: nil, Error: diceerrors.ErrWrongArgumentCount("APPEND")},
		},
		"append invalid number of arguments": {
			setup: func() {
				store.Del("key")
			},
			input:          []string{"key", "val", "val2"},
			migratedOutput: EvalResponse{Result: nil, Error: diceerrors.ErrWrongArgumentCount("APPEND")},
		},
		"append to non-existing key": {
			setup: func() {
				store.Del("key")
			},
			input:          []string{"key", "val"},
			migratedOutput: EvalResponse{Result: 3, Error: nil},
		},
		"append string value to existing key having string value": {
			setup: func() {
				key := "key"
				value := "val"
				obj := store.NewObj(value, -1, object.ObjTypeString, object.ObjEncodingRaw)
				store.Put(key, obj)
			},
			input:          []string{"key", "val"},
			migratedOutput: EvalResponse{Result: 6, Error: nil},
		},
		"append integer value to non existing key": {
			setup: func() {
				store.Del("key")
			},
			input:          []string{"key", "123"},
			migratedOutput: EvalResponse{Result: 3, Error: nil},
			validator: func(output []byte) {
				obj := store.Get("key")
				_, enc := object.ExtractTypeEncoding(obj)
				if enc != object.ObjEncodingInt {
					t.Errorf("unexpected encoding")
				}
			},
		},
		"append string value to existing key having integer value": {
			setup: func() {
				key := "key"
				value := "123"
				storedValue, _ := strconv.ParseInt(value, 10, 64)
				obj := store.NewObj(storedValue, -1, object.ObjTypeInt, object.ObjEncodingInt)
				store.Put(key, obj)
			},
			input:          []string{"key", "val"},
			migratedOutput: EvalResponse{Result: 6, Error: nil},
		},
		"append empty string to non-existing key": {
			setup: func() {
				store.Del("key")
			},
			input:          []string{"key", ""},
			migratedOutput: EvalResponse{Result: 0, Error: nil},
		},
		"append empty string to existing key having empty string": {
			setup: func() {
				key := "key"
				value := ""
				obj := store.NewObj(value, -1, object.ObjTypeString, object.ObjEncodingRaw)
				store.Put(key, obj)
			},
			input:          []string{"key", ""},
			migratedOutput: EvalResponse{Result: 0, Error: nil},
		},
		"append empty string to existing key": {
			setup: func() {
				key := "key"
				value := "val"
				obj := store.NewObj(value, -1, object.ObjTypeString, object.ObjEncodingRaw)
				store.Put(key, obj)
			},
			input:          []string{"key", ""},
			migratedOutput: EvalResponse{Result: 3, Error: nil},
		},
		"append modifies the encoding from int to raw": {
			setup: func() {
				store.Del("key")
				storedValue, _ := strconv.ParseInt("1", 10, 64)
				obj := store.NewObj(storedValue, -1, object.ObjTypeInt, object.ObjEncodingInt)
				store.Put("key", obj)
			},
			input:          []string{"key", "2"},
			migratedOutput: EvalResponse{Result: 2, Error: nil},
			validator: func(output []byte) {
				obj := store.Get("key")
				_, enc := object.ExtractTypeEncoding(obj)
				if enc != object.ObjEncodingRaw {
					t.Errorf("unexpected encoding")
				}
			},
		},
		"append to key created using LPUSH": {
			setup: func() {
				key := "listKey"
				value := "val"
				// Create a new list object
				obj := store.NewObj(NewDeque(), -1, object.ObjTypeByteList, object.ObjEncodingDeque)
				store.Put(key, obj)
				obj.Value.(*Deque).LPush(value)
			},
			input:          []string{"listKey", "val"},
			migratedOutput: EvalResponse{Result: nil, Error: diceerrors.ErrWrongTypeOperation},
		},
		"append to key created using SADD": {
			setup: func() {
				key := "setKey"
				// Create a new set object
				initialValues := map[string]struct{}{
					"existingVal": {},
					"anotherVal":  {},
				}
				obj := store.NewObj(initialValues, -1, object.ObjTypeSet, object.ObjEncodingSetStr)
				store.Put(key, obj)
			},
			input:          []string{"setKey", "val"},
			migratedOutput: EvalResponse{Result: nil, Error: diceerrors.ErrWrongTypeOperation},
		},
		"append to key created using HSET": {
			setup: func() {
				key := "hashKey"
				// Create a new hash map object
				initialValues := HashMap{
					"field1": "value1",
					"field2": "value2",
				}
				obj := store.NewObj(initialValues, -1, object.ObjTypeHashMap, object.ObjEncodingHashMap)
				store.Put(key, obj)
			},
			input:          []string{"hashKey", "val"},
			migratedOutput: EvalResponse{Result: nil, Error: diceerrors.ErrWrongTypeOperation},
		},
		"append to key created using SETBIT": {
			setup: func() {
				key := "bitKey"
				// Create a new byte array object
				initialByteArray := NewByteArray(1) // Initialize with 1 byte
				initialByteArray.SetBit(0, true)    // Set the first bit to 1
				obj := store.NewObj(initialByteArray, -1, object.ObjTypeByteArray, object.ObjEncodingByteArray)
				store.Put(key, obj)
			},
			input:          []string{"bitKey", "val"},
			migratedOutput: EvalResponse{Result: nil, Error: diceerrors.ErrWrongTypeOperation},
		},
		"append value with leading zeros": {
			setup: func() {
				store.Del("key_with_leading_zeros")
			},
			input:          []string{"key_with_leading_zeros", "0043"},
			migratedOutput: EvalResponse{Result: 4, Error: nil}, // The length of "0043" is 4
		},
	}

	runMigratedEvalTests(t, tests, evalAPPEND, store)
}

func BenchmarkEvalAPPEND(b *testing.B) {
	store := dstore.NewStore(nil, nil)
	for i := 0; i < b.N; i++ {
		evalAPPEND([]string{"key", fmt.Sprintf("val_%d", i)}, store)
	}
}

func testEvalJSONRESP(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"wrong number of args passed": {
			setup:  func() {},
			input:  nil,
			output: []byte("-ERR wrong number of arguments for 'json.resp' command\r\n"),
		},
		"key does not exist": {
			setup:  func() {},
			input:  []string{"NOTEXISTANT_KEY"},
			output: []byte("$-1\r\n"),
		},
		"string json": {
			setup: func() {
				key := "MOCK_KEY"
				value := "\"Roll the Dice\""
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"MOCK_KEY"},
			output: []byte("*1\r\n$13\r\nRoll the Dice\r\n"),
		},
		"integer json": {
			setup: func() {
				key := "MOCK_KEY"
				value := "10"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"MOCK_KEY"},
			output: []byte("*1\r\n:10\r\n"),
		},
		"bool json": {
			setup: func() {
				key := "MOCK_KEY"
				value := "true"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"MOCK_KEY"},
			output: []byte("*1\r\n+true\r\n"),
		},
		"nil json": {
			setup: func() {
				key := "MOCK_KEY"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(nil), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"MOCK_KEY"},
			output: []byte("*1\r\n$-1\r\n"),
		},
		"empty array": {
			setup: func() {
				key := "MOCK_KEY"
				value := "[]"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"MOCK_KEY"},
			output: []byte("*1\r\n+[\r\n"),
		},
		"empty object": {
			setup: func() {
				key := "MOCK_KEY"
				value := "{}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"MOCK_KEY"},
			output: []byte("*1\r\n+{\r\n"),
		},
		"array with mixed types": {
			setup: func() {
				key := "MOCK_KEY"
				value := "[\"dice\", 10, 10.5, true, null]"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"MOCK_KEY"},
			output: []byte("*6\r\n+[\r\n$4\r\ndice\r\n:10\r\n$4\r\n10.5\r\n+true\r\n$-1\r\n"),
		},
		"one layer of nesting no path": {
			setup: func() {
				key := "MOCK_KEY"
				value := "{\"b\": [\"dice\", 10, 10.5, true, null]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"MOCK_KEY"},
			output: []byte("*3\r\n+{\r\n$1\r\nb\r\n*6\r\n+[\r\n$4\r\ndice\r\n:10\r\n$4\r\n10.5\r\n+true\r\n$-1\r\n"),
		},
		"one layer of nesting with path": {
			setup: func() {
				key := "MOCK_KEY"
				value := "{\"b\": [\"dice\", 10, 10.5, true, null]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"MOCK_KEY", "$.b"},
			output: []byte("*1\r\n*6\r\n+[\r\n$4\r\ndice\r\n:10\r\n$4\r\n10.5\r\n+true\r\n$-1\r\n"),
		},
	}

	runEvalTests(t, tests, evalJSONRESP, store)
}

func testEvalZADD(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"ZADD with wrong number of arguments": {
			input: []string{"myzset", "1"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("ZADD"),
			},
		},
		"ZADD with non-numeric score": {
			input: []string{"myzset", "score", "member1"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrInvalidNumberFormat,
			},
		},
		"ZADD new member to non-existing key": {
			input: []string{"myzset", "1", "member1"},
			migratedOutput: EvalResponse{
				Result: 1,
				Error:  nil,
			},
		},
		"ZADD existing member with updated score": {
			setup: func() {
				evalZADD([]string{"myzset", "1", "member1"}, store)
			},
			input: []string{"myzset", "2", "member1"},
			migratedOutput: EvalResponse{
				Result: 0,
				Error:  nil,
			},
		},
		"ZADD multiple members": {
			setup: func() {
				evalZADD([]string{"myzset", "1", "member1"}, store)
			},
			input: []string{"myzset", "2", "member2", "3", "member3"},
			migratedOutput: EvalResponse{
				Result: 2,
				Error:  nil,
			},
		},
		"ZADD with negative score": {
			input: []string{"myzset", "-1", "member_neg"},
			migratedOutput: EvalResponse{
				Result: 1,
				Error:  nil,
			},
		},
		"ZADD with duplicate members": {
			setup: func() {
				evalZADD([]string{"myzset", "1", "member1"}, store)
			},
			input: []string{"myzset", "2", "member1", "2", "member1"},
			migratedOutput: EvalResponse{
				Result: 0,
				Error:  nil,
			},
		},
		"ZADD with extreme float value": {
			input: []string{"myzset", "1e308", "member_large"},
			migratedOutput: EvalResponse{
				Result: 1,
				Error:  nil,
			},
		},
		"ZADD with NaN score": {
			input: []string{"myzset", "NaN", "member_nan"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrInvalidNumberFormat,
			},
		},
		"ZADD with INF score": {
			input: []string{"myzset", "INF", "member_inf"},
			migratedOutput: EvalResponse{
				Result: 1,
				Error:  nil,
			},
		},
		"ZADD to a key of wrong type": {
			setup: func() {
				store.Put("mywrongtypekey", store.NewObj("string_value", -1, object.ObjTypeString, object.ObjEncodingRaw))
			},
			input: []string{"mywrongtypekey", "1", "member1"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongTypeOperation,
			},
		},
	}

	runMigratedEvalTests(t, tests, evalZADD, store)
}

func testEvalZRANGE(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"ZRANGE on non-existing key": {
			input: []string{"non_existing_key", "0", "-1"},
			migratedOutput: EvalResponse{
				Result: []string{},
				Error:  nil,
			},
		},
		"ZRANGE with wrong type key": {
			setup: func() {
				store.Put("mystring", store.NewObj("string_value", -1, object.ObjTypeString, object.ObjEncodingRaw))
			},
			input: []string{"mystring", "0", "-1"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongTypeOperation,
			},
		},
		"ZRANGE with normal indices": {
			setup: func() {
				evalZADD([]string{"myzset", "1", "member1", "2", "member2", "3", "member3"}, store)
			},
			input: []string{"myzset", "0", "1"},
			migratedOutput: EvalResponse{
				Result: []string{"member1", "member2"},
				Error:  nil,
			},
		},
		"ZRANGE with negative indices": {
			setup: func() {
				evalZADD([]string{"myzset", "1", "member1", "2", "member2", "3", "member3"}, store)
			},
			input: []string{"myzset", "-2", "-1"},
			migratedOutput: EvalResponse{
				Result: []string{"member2", "member3"},
				Error:  nil,
			},
		},
		"ZRANGE with start > stop": {
			setup: func() {
				evalZADD([]string{"myzset", "1", "member1", "2", "member2", "3", "member3"}, store)
			},
			input: []string{"myzset", "2", "1"},
			migratedOutput: EvalResponse{
				Result: []string{},
				Error:  nil,
			},
		},
		"ZRANGE with indices out of bounds": {
			setup: func() {
				evalZADD([]string{"myzset", "1", "member1"}, store)
			},
			input: []string{"myzset", "0", "5"},
			migratedOutput: EvalResponse{
				Result: []string{"member1"},
				Error:  nil,
			},
		},
		"ZRANGE WITHSCORES option": {
			setup: func() {
				evalZADD([]string{"myzset", "1", "member1", "2", "member2"}, store)
			},
			input: []string{"myzset", "0", "-1", "WITHSCORES"},
			migratedOutput: EvalResponse{
				Result: []string{"member1", "1", "member2", "2"},
				Error:  nil,
			},
		},
		"ZRANGE with invalid option": {
			setup: func() {
				evalZADD([]string{"myzset", "1", "member1"}, store)
			},
			input: []string{"myzset", "0", "-1", "INVALIDOPTION"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrSyntax,
			},
		},
		"ZRANGE with REV option": {
			setup: func() {
				evalZADD([]string{"myzset", "1", "member1", "2", "member2", "3", "member3"}, store)
			},
			input: []string{"myzset", "0", "-1", "REV"},
			migratedOutput: EvalResponse{
				Result: []string{"member3", "member2", "member1"},
				Error:  nil,
			},
		},
		"ZRANGE with REV and WITHSCORES options": {
			setup: func() {
				evalZADD([]string{"myzset", "1", "member1", "2", "member2", "3", "member3"}, store)
			},
			input: []string{"myzset", "0", "-1", "REV", "WITHSCORES"},
			migratedOutput: EvalResponse{
				Result: []string{"member3", "3", "member2", "2", "member1", "1"},
				Error:  nil,
			},
		},
		"ZRANGE with start index greater than length": {
			setup: func() {
				evalZADD([]string{"myzset", "1", "member1"}, store)
			},
			input: []string{"myzset", "5", "10"},
			migratedOutput: EvalResponse{
				Result: []string{},
				Error:  nil,
			},
		},
		"ZRANGE with negative start index greater than length": {
			setup: func() {
				evalZADD([]string{"myzset", "1", "member1"}, store)
			},
			input: []string{"myzset", "-10", "-5"},
			migratedOutput: EvalResponse{
				Result: []string{},
				Error:  nil,
			},
		},
	}

	runMigratedEvalTests(t, tests, evalZRANGE, store)
}

func testEvalZPOPMIN(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"ZPOPMIN on non-existing key with/without count argument": {
			input: []string{"NON_EXISTING_KEY"},
			migratedOutput: EvalResponse{
				Result: []string{},
				Error:  nil,
			},
		},
		"ZPOPMIN with wrong type of key with/without count argument": {
			setup: func() {
				store.Put("mystring", store.NewObj("string_value", -1, object.ObjTypeString, object.ObjEncodingRaw))
			},
			input: []string{"mystring", "1"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongTypeOperation,
			},
		},
		"ZPOPMIN on existing key (without count argument)": {
			setup: func() {
				evalZADD([]string{"myzset", "1", "member1", "2", "member2"}, store)
			},
			input: []string{"myzset"},
			migratedOutput: EvalResponse{
				Result: []string{"1", "member1"},
				Error:  nil,
			},
		},
		"ZPOPMIN with normal count argument": {
			setup: func() {
				evalZADD([]string{"myzset", "1", "member1", "2", "member2", "3", "member3"}, store)
			},
			input: []string{"myzset", "2"},
			migratedOutput: EvalResponse{
				Result: []string{"1", "member1", "2", "member2"},
				Error:  nil,
			},
		},
		"ZPOPMIN with count argument but multiple members have the same score": {
			setup: func() {
				evalZADD([]string{"myzset", "1", "member1", "1", "member2", "1", "member3"}, store)
			},
			input: []string{"myzset", "2"},
			migratedOutput: EvalResponse{
				Result: []string{"1", "member1", "1", "member2"},
				Error:  nil,
			},
		},
		"ZPOPMIN with negative count argument": {
			setup: func() {
				evalZADD([]string{"myzset", "1", "member1", "2", "member2", "3", "member3"}, store)
			},
			input: []string{"myzset", "-1"},
			migratedOutput: EvalResponse{
				Result: []string{},
				Error:  nil,
			},
		},
		"ZPOPMIN with invalid count argument": {
			setup: func() {
				evalZADD([]string{"myzset", "1", "member1"}, store)
			},
			input: []string{"myzset", "INCORRECT_COUNT_ARGUMENT"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrIntegerOutOfRange,
			},
		},
		"ZPOPMIN with count argument greater than length of sorted set": {
			setup: func() {
				evalZADD([]string{"myzset", "1", "member1", "2", "member2"}, store)
			},
			input: []string{"myzset", "10"},
			migratedOutput: EvalResponse{
				Result: []string{"1", "member1", "2", "member2"},
				Error:  nil,
			},
		},
		"ZPOPMIN on empty sorted set": {
			setup: func() {
				store.Put("myzset", store.NewObj(sortedset.New(), -1, object.ObjTypeSortedSet, object.ObjEncodingBTree)) // Ensure the set exists but is empty
			},
			input: []string{"myzset"},
			migratedOutput: EvalResponse{
				Result: []string{},
				Error:  nil,
			},
		},
		"ZPOPMIN with floating-point scores": {
			setup: func() {
				evalZADD([]string{"myzset", "1.5", "member1", "2.7", "member2"}, store)
			},
			input: []string{"myzset"},
			migratedOutput: EvalResponse{
				Result: []string{"1.5", "member1"},
				Error:  nil,
			},
		},
	}

	runMigratedEvalTests(t, tests, evalZPOPMIN, store)
}

func BenchmarkEvalZPOPMIN(b *testing.B) {
	// Define benchmark cases with varying sizes of sorted sets
	benchmarks := []struct {
		name  string
		setup func(store *dstore.Store)
		input []string
	}{
		{
			name: "ZPOPMIN on small sorted set (10 members)",
			setup: func(store *dstore.Store) {
				evalZADD([]string{"myzset", "1", "member1", "2", "member2", "3", "member3", "4", "member4", "5", "member5", "6", "member6", "7", "member7", "8", "member8", "9", "member9", "10", "member10"}, store)
			},
			input: []string{"myzset", "3"},
		},
		{
			name: "ZPOPMIN on large sorted set (10000 members)",
			setup: func(store *dstore.Store) {
				args := []string{"myzset"}
				for i := 1; i <= 10000; i++ {
					args = append(args, fmt.Sprintf("%d", i), fmt.Sprintf("member%d", i))
				}
				evalZADD(args, store)
			},
			input: []string{"myzset", "10"},
		},
		{
			name: "ZPOPMIN with duplicate scores",
			setup: func(store *dstore.Store) {
				evalZADD([]string{"myzset", "1", "member1", "1", "member2", "1", "member3"}, store)
			},
			input: []string{"myzset", "2"},
		},
	}

	store := dstore.NewStore(nil, nil)

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			bm.setup(store)

			for i := 0; i < b.N; i++ {
				// Reset the store before each run to avoid contamination
				dstore.ResetStore(store)
				bm.setup(store)
				evalZPOPMIN(bm.input, store)
			}
		})
	}
}

func testEvalZRANK(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"ZRANK with non-existing key": {
			input: []string{"non_existing_key", "member"},
			migratedOutput: EvalResponse{
				Result: clientio.NIL,
				Error:  nil,
			},
		},
		"ZRANK with existing member": {
			setup: func() {
				evalZADD([]string{"myzset", "1", "member1", "2", "member2", "3", "member3"}, store)
			},
			input: []string{"myzset", "member2"},
			migratedOutput: EvalResponse{
				Result: int64(1),
				Error:  nil,
			},
		},
		"ZRANK with non-existing member": {
			setup: func() {
				evalZADD([]string{"myzset", "1", "member1", "2", "member2", "3", "member3"}, store)
			},
			input: []string{"myzset", "non_existing_member"},
			migratedOutput: EvalResponse{
				Result: clientio.NIL,
				Error:  nil,
			},
		},
		"ZRANK with WITHSCORE option": {
			setup: func() {
				evalZADD([]string{"myzset", "1", "member1", "2", "member2", "3", "member3"}, store)
			},
			input: []string{"myzset", "member2", "WITHSCORE"},
			migratedOutput: EvalResponse{
				Result: []interface{}{int64(1), "2"},
				Error:  nil,
			},
		},
		"ZRANK with invalid option": {
			setup: func() {
				evalZADD([]string{"myzset", "1", "member1", "2", "member2", "3", "member3"}, store)
			},
			input: []string{"myzset", "member2", "INVALID_OPTION"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrSyntax,
			},
		},
		"ZRANK with multiple members having same score": {
			setup: func() {
				evalZADD([]string{"myzset", "1", "member1", "1", "member2", "1", "member3"}, store)
			},
			input: []string{"myzset", "member3"},
			migratedOutput: EvalResponse{
				Result: int64(2),
				Error:  nil,
			},
		},
		"ZRANK with non-integer scores": {
			setup: func() {
				evalZADD([]string{"myzset", "1.5", "member1", "2.5", "member2"}, store)
			},
			input: []string{"myzset", "member2"},
			migratedOutput: EvalResponse{
				Result: int64(1),
				Error:  nil,
			},
		},
		"ZRANK with too many arguments": {
			input: []string{"myzset", "member", "WITHSCORES", "extra"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("ZRANK"),
			},
		},
	}

	runMigratedEvalTests(t, tests, evalZRANK, store)
}

func BenchmarkEvalZRANK(b *testing.B) {
	store := dstore.NewStore(nil, nil)

	// Set up initial sorted set
	evalZADD([]string{"myzset", "1", "member1", "2", "member2", "3", "member3"}, store)

	benchmarks := []struct {
		name      string
		input     []string
		withScore bool
	}{
		{"ZRANK existing member", []string{"myzset", "member3"}, false},
		{"ZRANK non-existing member", []string{"myzset", "nonexistent"}, false},
		{"ZRANK with WITHSCORE", []string{"myzset", "member2", "WITHSCORE"}, true},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				evalZRANK(bm.input, store)
			}
		})
	}
}

func testEvalBitField(t *testing.T, store *dstore.Store) {
	testCases := map[string]evalTestCase{
		"BITFIELD signed SET": {
			input:  []string{"bits", "set", "i8", "0", "-100"},
			output: clientio.Encode([]int64{0}, false),
		},
		"BITFIELD GET": {
			setup: func() {
				args := []string{"bits", "set", "u8", "0", "255"}
				evalBITFIELD(args, store)
			},
			input:  []string{"bits", "get", "u8", "0"},
			output: clientio.Encode([]int64{255}, false),
		},
		"BITFIELD INCRBY": {
			setup: func() {
				args := []string{"bits", "set", "u8", "0", "255"}
				evalBITFIELD(args, store)
			},
			input:  []string{"bits", "incrby", "u8", "0", "100"},
			output: clientio.Encode([]int64{99}, false),
		},
		"BITFIELD Arity": {
			input:  []string{},
			output: diceerrors.NewErrArity("BITFIELD"),
		},
		"BITFIELD invalid combination of commands in a single operation": {
			input:  []string{"bits", "SET", "u8", "0", "255", "INCRBY", "u8", "0", "100", "GET", "u8"},
			output: []byte("-ERR syntax error\r\n"),
		},
		"BITFIELD invalid bitfield type": {
			input:  []string{"bits", "SET", "a8", "0", "255", "INCRBY", "u8", "0", "100", "GET", "u8"},
			output: []byte("-ERR Invalid bitfield type. Use something like i16 u8. Note that u64 is not supported but i64 is.\r\n"),
		},
		"BITFIELD invalid bit offset": {
			input:  []string{"bits", "SET", "u8", "a", "255", "INCRBY", "u8", "0", "100", "GET", "u8"},
			output: []byte("-ERR bit offset is not an integer or out of range\r\n"),
		},
		"BITFIELD invalid overflow type": {
			input:  []string{"bits", "SET", "u8", "0", "255", "INCRBY", "u8", "0", "100", "OVERFLOW", "wraap"},
			output: []byte("-ERR Invalid OVERFLOW type specified\r\n"),
		},
		"BITFIELD missing arguments in SET": {
			input:  []string{"bits", "SET", "u8", "0", "INCRBY", "u8", "0", "100", "GET", "u8", "288"},
			output: []byte("-ERR value is not an integer or out of range\r\n"),
		},
	}
	runEvalTests(t, testCases, evalBITFIELD, store)
}

func testEvalHINCRBYFLOAT(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"HINCRBYFLOAT on a non-existing key and field": {
			setup:          func() {},
			input:          []string{"key", "field", "0.1"},
			migratedOutput: EvalResponse{Result: "0.1", Error: nil},
		},
		"HINCRBYFLOAT on an existing key and non-existing field": {
			setup: func() {
				key := "key"
				h := make(HashMap)
				obj := &object.Obj{
					TypeEncoding:   object.ObjTypeHashMap | object.ObjEncodingHashMap,
					Value:          h,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)
			},
			input:          []string{"key", "field", "0.1"},
			migratedOutput: EvalResponse{Result: "0.1", Error: nil},
		},
		"HINCRBYFLOAT on an existing key and field with a float value": {
			setup: func() {
				key := "key"
				field := "field"
				h := make(HashMap)
				h[field] = "2.1"
				obj := &object.Obj{
					TypeEncoding:   object.ObjTypeHashMap | object.ObjEncodingHashMap,
					Value:          h,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)
			},
			input:          []string{"key", "field", "0.1"},
			migratedOutput: EvalResponse{Result: "2.2", Error: nil},
		},
		"HINCRBYFLOAT on an existing key and field with an integer value": {
			setup: func() {
				key := "key"
				field := "field"
				h := make(HashMap)
				h[field] = "2"
				obj := &object.Obj{
					TypeEncoding:   object.ObjTypeHashMap | object.ObjEncodingHashMap,
					Value:          h,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)
			},
			input:          []string{"key", "field", "0.1"},
			migratedOutput: EvalResponse{Result: "2.1", Error: nil},
		},
		"HINCRBYFLOAT with a negative increment": {
			setup: func() {
				key := "key"
				field := "field"
				h := make(HashMap)
				h[field] = "2.0"
				obj := &object.Obj{
					TypeEncoding:   object.ObjTypeHashMap | object.ObjEncodingHashMap,
					Value:          h,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)
			},
			input:          []string{"key", "field", "-0.1"},
			migratedOutput: EvalResponse{Result: "1.9", Error: nil},
		},
		"HINCRBYFLOAT by a non-numeric increment": {
			setup: func() {
				key := "key"
				field := "field"
				h := make(HashMap)
				h[field] = "2.0"
				obj := &object.Obj{
					TypeEncoding:   object.ObjTypeHashMap | object.ObjEncodingHashMap,
					Value:          h,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)
			},
			input:          []string{"key", "field", "a"},
			output:         []byte("-ERR value is not an integer or a float\r\n"),
			migratedOutput: EvalResponse{Result: nil, Error: diceerrors.ErrInvalidNumberFormat},
		},
		"HINCRBYFLOAT on a field with non-numeric value": {
			setup: func() {
				key := "key"
				field := "field"
				h := make(HashMap)
				h[field] = "non_numeric"
				obj := &object.Obj{
					TypeEncoding:   object.ObjTypeHashMap | object.ObjEncodingHashMap,
					Value:          h,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)
			},
			input:          []string{"key", "field", "0.1"},
			migratedOutput: EvalResponse{Result: nil, Error: diceerrors.ErrInvalidNumberFormat},
		},
		"HINCRBYFLOAT by a value that would turn float64 to Inf": {
			setup: func() {
				key := "key"
				field := "field"
				h := make(HashMap)
				h[field] = "1e308"
				obj := &object.Obj{
					TypeEncoding:   object.ObjTypeHashMap | object.ObjEncodingHashMap,
					Value:          h,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)
			},
			input:          []string{"key", "field", "1e308"},
			migratedOutput: EvalResponse{Result: nil, Error: diceerrors.ErrOverflow},
		},
		"HINCRBYFLOAT with scientific notation": {
			setup: func() {
				key := "key"
				field := "field"
				h := make(HashMap)
				h[field] = "1e2"
				obj := &object.Obj{
					TypeEncoding:   object.ObjTypeHashMap | object.ObjEncodingHashMap,
					Value:          h,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)
			},
			input:          []string{"key", "field", "1e-1"},
			migratedOutput: EvalResponse{Result: "100.1", Error: nil},
		},
	}

	runMigratedEvalTests(t, tests, evalHINCRBYFLOAT, store)
}

func BenchmarkEvalHINCRBYFLOAT(b *testing.B) {
	store := dstore.NewStore(nil, nil)

	// Setting initial fields with some values
	store.Put("key1", store.NewObj(HashMap{"field1": "1.0", "field2": "1.2"}, maxExDuration, object.ObjTypeHashMap, object.ObjEncodingHashMap))
	store.Put("key2", store.NewObj(HashMap{"field1": "0.1"}, maxExDuration, object.ObjTypeHashMap, object.ObjEncodingHashMap))

	inputs := []struct {
		key   string
		field string
		incr  string
	}{
		{"key1", "field1", "0.1"},
		{"key1", "field1", "-0.1"},
		{"key1", "field2", "1000000.1"},
		{"key1", "field2", "-1000000.1"},
		{"key2", "field1", "-10.1234"},
		{"key3", "field1", "1.5"},  // testing with non-existing key
		{"key2", "field2", "2.75"}, // testing with non-existing field in existing key
	}

	for _, input := range inputs {
		b.Run(fmt.Sprintf("HINCRBYFLOAT %s %s %s", input.key, input.field, input.incr), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = evalHINCRBYFLOAT([]string{"HINCRBYFLOAT", input.key, input.field, input.incr}, store)
			}
		})
	}
}

func testEvalDUMP(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"nil value": {
			setup:  func() {},
			input:  nil,
			output: []byte("-ERR wrong number of arguments for 'dump' command\r\n"),
		},
		"empty array": {
			setup:  func() {},
			input:  []string{},
			output: []byte("-ERR wrong number of arguments for 'dump' command\r\n"),
		},
		"key does not exist": {
			setup:  func() {},
			input:  []string{"NONEXISTENT_KEY"},
			output: []byte("-ERR nil\r\n"),
		}, "dump string value": {
			setup: func() {
				key := "user"
				value := "hello"
				obj := store.NewObj(value, -1, object.ObjTypeString, object.ObjEncodingRaw)
				store.Put(key, obj)
			},
			input: []string{"user"},
			output: clientio.Encode(
				base64.StdEncoding.EncodeToString([]byte{
					0x09, 0x00, 0x00, 0x00, 0x00, 0x05, 0x68, 0x65, 0x6c, 0x6c, 0x6f,
					0xFF, // End marker
					// CRC64 checksum here:
					0x00, 0x47, 0x97, 0x93, 0xBE, 0x36, 0x45, 0xC7,
				}), false),
		},
		"dump integer value": {
			setup: func() {
				key := "INTEGER_KEY"
				value := int64(10)
				obj := store.NewObj(value, -1, object.ObjTypeInt, object.ObjEncodingInt)
				store.Put(key, obj)
			},
			input: []string{"INTEGER_KEY"},
			output: clientio.Encode(base64.StdEncoding.EncodeToString([]byte{
				0x09,
				0xC0,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0A,
				0xFF,
				0x12, 0x77, 0xDE, 0x29, 0x53, 0xDB, 0x44, 0xC2,
			}), false),
		},
		"dump expired key": {
			setup: func() {
				key := "EXPIRED_KEY"
				value := "This will expire"
				obj := store.NewObj(value, -1, object.ObjTypeString, object.ObjEncodingRaw)
				store.Put(key, obj)
				var exDurationMs int64 = -1
				store.SetExpiry(obj, exDurationMs)
			},
			input:  []string{"EXPIRED_KEY"},
			output: []byte("-ERR nil\r\n"),
		},
	}

	runEvalTests(t, tests, evalDUMP, store)
}

func testEvalBitFieldRO(t *testing.T, store *dstore.Store) {
	testCases := map[string]evalTestCase{
		"BITFIELD_RO Arity": {
			input:  []string{},
			output: diceerrors.NewErrArity("BITFIELD_RO"),
		},
		"BITFIELD_RO syntax error": {
			input:  []string{"bits", "GET", "u8"},
			output: []byte("-ERR syntax error\r\n"),
		},
		"BITFIELD_RO invalid bitfield type": {
			input:  []string{"bits", "GET", "a8", "0", "255"},
			output: []byte("-ERR Invalid bitfield type. Use something like i16 u8. Note that u64 is not supported but i64 is.\r\n"),
		},
		"BITFIELD_RO unsupported commands": {
			input:  []string{"bits", "set", "u8", "0", "255"},
			output: []byte("-ERR BITFIELD_RO only supports the GET subcommand\r\n"),
		},
	}
	runEvalTests(t, testCases, evalBITFIELDRO, store)
}

func testEvalGEOADD(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"GEOADD with wrong number of arguments": {
			input:  []string{"mygeo", "1", "2"},
			output: diceerrors.NewErrArity("GEOADD"),
		},
		"GEOADD with non-numeric longitude": {
			input:  []string{"mygeo", "long", "40.7128", "NewYork"},
			output: diceerrors.NewErrWithMessage("ERR invalid longitude"),
		},
		"GEOADD with non-numeric latitude": {
			input:  []string{"mygeo", "-74.0060", "lat", "NewYork"},
			output: diceerrors.NewErrWithMessage("ERR invalid latitude"),
		},
		"GEOADD new member to non-existing key": {
			setup:  func() {},
			input:  []string{"mygeo", "-74.0060", "40.7128", "NewYork"},
			output: clientio.Encode(int64(1), false),
		},
		"GEOADD existing member with updated coordinates": {
			setup: func() {
				evalGEOADD([]string{"mygeo", "-74.0060", "40.7128", "NewYork"}, store)
			},
			input:  []string{"mygeo", "-73.9352", "40.7304", "NewYork"},
			output: clientio.Encode(int64(0), false),
		},
		"GEOADD multiple members": {
			setup: func() {
				evalGEOADD([]string{"mygeo", "-74.0060", "40.7128", "NewYork"}, store)
			},
			input:  []string{"mygeo", "-118.2437", "34.0522", "LosAngeles", "-87.6298", "41.8781", "Chicago"},
			output: clientio.Encode(int64(2), false),
		},
		"GEOADD with NX option (new member)": {
			input:  []string{"mygeo", "NX", "-122.4194", "37.7749", "SanFrancisco"},
			output: clientio.Encode(int64(1), false),
		},
		"GEOADD with NX option (existing member)": {
			setup: func() {
				evalGEOADD([]string{"mygeo", "-74.0060", "40.7128", "NewYork"}, store)
			},
			input:  []string{"mygeo", "NX", "-73.9352", "40.7304", "NewYork"},
			output: clientio.Encode(int64(0), false),
		},
		"GEOADD with XX option (new member)": {
			input:  []string{"mygeo", "XX", "-71.0589", "42.3601", "Boston"},
			output: clientio.Encode(int64(0), false),
		},
		"GEOADD with XX option (existing member)": {
			setup: func() {
				evalGEOADD([]string{"mygeo", "-74.0060", "40.7128", "NewYork"}, store)
			},
			input:  []string{"mygeo", "XX", "-73.9352", "40.7304", "NewYork"},
			output: clientio.Encode(int64(0), false),
		},
		"GEOADD with both NX and XX options": {
			input:  []string{"mygeo", "NX", "XX", "-74.0060", "40.7128", "NewYork"},
			output: diceerrors.NewErrWithMessage("ERR XX and NX options at the same time are not compatible"),
		},
		"GEOADD with invalid option": {
			input:  []string{"mygeo", "INVALID", "-74.0060", "40.7128", "NewYork"},
			output: diceerrors.NewErrArity("GEOADD"),
		},
		"GEOADD to a key of wrong type": {
			setup: func() {
				store.Put("mygeo", store.NewObj("string_value", -1, object.ObjTypeString, object.ObjEncodingRaw))
			},
			input:  []string{"mygeo", "-74.0060", "40.7128", "NewYork"},
			output: []byte("-ERR Existing key has wrong Dice type\r\n"),
		},
		"GEOADD with longitude out of range": {
			input:  []string{"mygeo", "181.0", "40.7128", "Invalid"},
			output: diceerrors.NewErrWithMessage("ERR invalid longitude"),
		},
		"GEOADD with latitude out of range": {
			input:  []string{"mygeo", "-74.0060", "91.0", "Invalid"},
			output: diceerrors.NewErrWithMessage("ERR invalid latitude"),
		},
	}

	runEvalTests(t, tests, evalGEOADD, store)
}

func testEvalGEODIST(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"GEODIST between existing points": {
			setup: func() {
				evalGEOADD([]string{"points", "13.361389", "38.115556", "Palermo"}, store)
				evalGEOADD([]string{"points", "15.087269", "37.502669", "Catania"}, store)
			},
			input:  []string{"points", "Palermo", "Catania"},
			output: clientio.Encode(float64(166274.1440), false), // Example value
		},
		"GEODIST with units (km)": {
			setup: func() {
				evalGEOADD([]string{"points", "13.361389", "38.115556", "Palermo"}, store)
				evalGEOADD([]string{"points", "15.087269", "37.502669", "Catania"}, store)
			},
			input:  []string{"points", "Palermo", "Catania", "km"},
			output: clientio.Encode(float64(166.2741), false), // Example value
		},
		"GEODIST to same point": {
			setup: func() {
				evalGEOADD([]string{"points", "13.361389", "38.115556", "Palermo"}, store)
			},
			input:  []string{"points", "Palermo", "Palermo"},
			output: clientio.Encode(float64(0.0000), false), // Expecting distance 0 formatted to 4 decimals
		},
		// Add other test cases here...
	}

	runEvalTests(t, tests, evalGEODIST, store)
}

func testEvalSINTER(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"intersection of two sets": {
			setup: func() {
				evalSADD([]string{"set1", "a", "b", "c"}, store)
				evalSADD([]string{"set2", "c", "d", "e"}, store)
			},
			input:  []string{"set1", "set2"},
			output: clientio.Encode([]string{"c"}, false),
		},
		"intersection of three sets": {
			setup: func() {
				evalSADD([]string{"set1", "a", "b", "c"}, store)
				evalSADD([]string{"set2", "b", "c", "d"}, store)
				evalSADD([]string{"set3", "c", "d", "e"}, store)
			},
			input:  []string{"set1", "set2", "set3"},
			output: clientio.Encode([]string{"c"}, false),
		},
		"intersection with single set": {
			setup: func() {
				evalSADD([]string{"set1", "a"}, store)
			},
			input:  []string{"set1"},
			output: clientio.Encode([]string{"a"}, false),
		},
		"intersection with a non-existent key": {
			setup: func() {
				evalSADD([]string{"set1", "a", "b", "c"}, store)
			},
			input:  []string{"set1", "nonexistent"},
			output: clientio.Encode([]string{}, false),
		},
		"intersection with wrong type": {
			setup: func() {
				evalSADD([]string{"set1", "a", "b", "c"}, store)
				store.Put("string", &object.Obj{Value: "string", TypeEncoding: object.ObjTypeString})
			},
			input:  []string{"set1", "string"},
			output: []byte("-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"),
		},
		"no arguments": {
			input:  []string{},
			output: diceerrors.NewErrArity("SINTER"),
		},
	}

	runEvalTests(t, tests, evalSINTER, store)
}

func testEvalOBJECTENCODING(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"nil value": {
			setup:  func() {},
			input:  nil,
			output: []byte("-ERR wrong number of arguments for 'object' command\r\n"),
		},
		"empty array": {
			setup:  func() {},
			input:  []string{},
			output: []byte("-ERR wrong number of arguments for 'object' command\r\n"),
		},
		"object with invalid subcommand": {
			setup:  func() {},
			input:  []string{"TESTSUBCOMMAND", "key"},
			output: []byte("-ERR syntax error\r\n"),
		},
		"key does not exist": {
			setup:  func() {},
			input:  []string{"ENCODING", "NONEXISTENT_KEY"},
			output: clientio.RespNIL,
		},
		"key exists": {
			setup: func() {
				evalLPUSH([]string{"EXISTING_KEY", "mock_value"}, store)
			},
			input:  []string{"ENCODING", "EXISTING_KEY"},
			output: []byte("$5\r\ndeque\r\n"),
		},
	}

	runEvalTests(t, tests, evalOBJECT, store)
}

func testEvalJSONSTRAPPEND(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"append to single field": {
			setup: func() {
				key := "doc1"
				value := "{\"a\":\"foo\", \"nested1\": {\"a\": \"hello\"}, \"nested2\": {\"a\": 31}}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"doc1", "$.nested1.a", "\"baz\""},
			output: []byte("*1\r\n:8\r\n"), // Expected length after append
		},
		"append to non-existing key": {
			setup: func() {
				// No setup needed as we are testing a non-existing document.
			},
			input:  []string{"non_existing_doc", "$..a", "\"err\""},
			output: []byte("-ERR Could not perform this operation on a key that doesn't exist\r\n"),
		},
		"append to root node": {
			setup: func() {
				key := "doc1"
				value := "\"abcd\""
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
				store.Put(key, obj)
			},
			input:  []string{"doc1", "$", "\"piu\""},
			output: []byte("*1\r\n:7\r\n"), // Expected length after appending to "abcd"
		},
	}

	// Run the tests
	runEvalTests(t, tests, evalJSONSTRAPPEND, store)
}

func BenchmarkEvalJSONSTRAPPEND(b *testing.B) {
	store := dstore.NewStore(nil, nil)

	// Setup a sample JSON document
	key := "doc1"
	value := "{\"a\":\"foo\", \"nested1\": {\"a\": \"hello\"}, \"nested2\": {\"a\": 31}}"
	var rootData interface{}
	_ = sonic.Unmarshal([]byte(value), &rootData)
	obj := store.NewObj(rootData, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
	store.Put(key, obj)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Benchmark appending to multiple fields
		evalJSONSTRAPPEND([]string{"doc1", "$..a", "\"bar\""}, store)
	}
}

func testEvalINCR(t *testing.T, store *dstore.Store) {
	tests := []evalTestCase{
		{
			name:           "INCR key does not exist",
			input:          []string{"KEY1"},
			migratedOutput: EvalResponse{Result: int64(1), Error: nil},
		},
		{
			name: "INCR key exists",
			setup: func() {
				key := "KEY2"
				obj := store.NewObj(int64(1), -1, object.ObjTypeInt, object.ObjEncodingInt)
				store.Put(key, obj)
			},
			input:          []string{"KEY2"},
			migratedOutput: EvalResponse{Result: int64(2), Error: nil},
		},
		{
			name: "INCR key holding string value",
			setup: func() {
				key := "KEY3"
				obj := store.NewObj("VAL1", -1, object.ObjTypeString, object.ObjEncodingEmbStr)
				store.Put(key, obj)
			},
			input:          []string{"KEY3"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR value is not an integer or out of range")},
		},
		{
			name: "INCR key holding SET type",
			setup: func() {
				evalSADD([]string{"SET1", "1", "2", "3"}, store)
			},
			input:          []string{"SET1"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")},
		},
		{
			name: "INCR key holding MAP type",
			setup: func() {
				evalHSET([]string{"MAP1", "a", "1", "b", "2", "c", "3"}, store)
			},
			input:          []string{"MAP1"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")},
		},
		{
			name:           "INCR More than one args passed",
			input:          []string{"KEY4", "ARG2"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR wrong number of arguments for 'incr' command")},
		},
		{
			name: "INCR Max Overflow",
			setup: func() {
				key := "KEY5"
				obj := store.NewObj(int64(math.MaxInt64), -1, object.ObjTypeInt, object.ObjEncodingInt)
				store.Put(key, obj)
			},
			input:          []string{"KEY5"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR increment or decrement would overflow")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store = setupTest(store)
			if tt.setup != nil {
				tt.setup()
			}

			response := evalINCR(tt.input, store)

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

func testEvalINCRBY(t *testing.T, store *dstore.Store) {
	tests := []evalTestCase{
		{
			name:           "INCRBY key does not exist",
			input:          []string{"KEY1", "2"},
			migratedOutput: EvalResponse{Result: int64(2), Error: nil},
		},
		{
			name: "INCRBY key exists",
			setup: func() {
				key := "KEY2"
				obj := store.NewObj(int64(1), -1, object.ObjTypeInt, object.ObjEncodingInt)
				store.Put(key, obj)
			},
			input:          []string{"KEY2", "3"},
			migratedOutput: EvalResponse{Result: int64(4), Error: nil},
		},
		{
			name: "INCRBY key holding string value",
			setup: func() {
				key := "KEY3"
				obj := store.NewObj("VAL1", -1, object.ObjTypeString, object.ObjEncodingEmbStr)
				store.Put(key, obj)
			},
			input:          []string{"KEY3", "2"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR value is not an integer or out of range")},
		},
		{
			name: "INCRBY key holding SET type",
			setup: func() {
				evalSADD([]string{"SET1", "1", "2", "3"}, store)
			},
			input:          []string{"SET1", "2"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")},
		},
		{
			name: "INCRBY key holding MAP type",
			setup: func() {
				evalHSET([]string{"MAP1", "a", "1", "b", "2", "c", "3"}, store)
			},
			input:          []string{"MAP1", "2"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")},
		},
		{
			name:           "INCRBY Wrong number of args passed",
			input:          []string{"KEY4"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR wrong number of arguments for 'incrby' command")},
		},
		{
			name: "INCRBY Max Overflow",
			setup: func() {
				key := "KEY5"
				obj := store.NewObj(int64(math.MaxInt64-3), -1, object.ObjTypeInt, object.ObjEncodingInt)
				store.Put(key, obj)
			},
			input:          []string{"KEY5", "4"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR increment or decrement would overflow")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store = setupTest(store)
			if tt.setup != nil {
				tt.setup()
			}

			response := evalINCRBY(tt.input, store)

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

func testEvalDECR(t *testing.T, store *dstore.Store) {
	tests := []evalTestCase{
		{
			name:           "DECR key does not exist",
			input:          []string{"KEY1"},
			migratedOutput: EvalResponse{Result: int64(-1), Error: nil},
		},
		{
			name: "DECR key exists",
			setup: func() {
				key := "KEY2"
				obj := store.NewObj(int64(1), -1, object.ObjTypeInt, object.ObjEncodingInt)
				store.Put(key, obj)
			},
			input:          []string{"KEY2"},
			migratedOutput: EvalResponse{Result: int64(0), Error: nil},
		},
		{
			name: "DECR key holding string value",
			setup: func() {
				key := "KEY3"
				obj := store.NewObj("VAL1", -1, object.ObjTypeString, object.ObjEncodingEmbStr)
				store.Put(key, obj)
			},
			input:          []string{"KEY3"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR value is not an integer or out of range")},
		},
		{
			name: "DECR key holding SET type",
			setup: func() {
				evalSADD([]string{"SET1", "1", "2", "3"}, store)
			},
			input:          []string{"SET1"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")},
		},
		{
			name: "DECR key holding MAP type",
			setup: func() {
				evalHSET([]string{"MAP1", "a", "1", "b", "2", "c", "3"}, store)
			},
			input:          []string{"MAP1"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")},
		},
		{
			name:           "DECR More than one args passed",
			input:          []string{"KEY4", "ARG2"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR wrong number of arguments for 'decr' command")},
		},
		{
			name: "DECR Min Overflow",
			setup: func() {
				key := "KEY5"
				obj := store.NewObj(int64(math.MinInt64), -1, object.ObjTypeInt, object.ObjEncodingInt)
				store.Put(key, obj)
			},
			input:          []string{"KEY5"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR increment or decrement would overflow")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store = setupTest(store)
			if tt.setup != nil {
				tt.setup()
			}

			response := evalDECR(tt.input, store)

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

func testEvalDECRBY(t *testing.T, store *dstore.Store) {
	tests := []evalTestCase{
		{
			name:           "DECRBY key does not exist",
			input:          []string{"KEY1", "2"},
			migratedOutput: EvalResponse{Result: int64(-2), Error: nil},
		},
		{
			name: "DECRBY key exists",
			setup: func() {
				key := "KEY2"
				obj := store.NewObj(int64(1), -1, object.ObjTypeInt, object.ObjEncodingInt)
				store.Put(key, obj)
			},
			input:          []string{"KEY2", "3"},
			migratedOutput: EvalResponse{Result: int64(-2), Error: nil},
		},
		{
			name: "DECRBY key holding string value",
			setup: func() {
				key := "KEY3"
				obj := store.NewObj("VAL1", -1, object.ObjTypeString, object.ObjEncodingEmbStr)
				store.Put(key, obj)
			},
			input:          []string{"KEY3", "2"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR value is not an integer or out of range")},
		},
		{
			name: "DECRBY key holding SET type",
			setup: func() {
				evalSADD([]string{"SET1", "1", "2", "3"}, store)
			},
			input:          []string{"SET1", "2"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")},
		},
		{
			name: "DECRBY key holding MAP type",
			setup: func() {
				evalHSET([]string{"MAP1", "a", "1", "b", "2", "c", "3"}, store)
			},
			input:          []string{"MAP1", "2"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")},
		},
		{
			name:           "DECRBY Wrong number of args passed",
			input:          []string{"KEY4"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR wrong number of arguments for 'decrby' command")},
		},
		{
			name: "DECRBY Min Overflow",
			setup: func() {
				key := "KEY5"
				obj := store.NewObj(int64(math.MinInt64+3), -1, object.ObjTypeInt, object.ObjEncodingInt)
				store.Put(key, obj)
			},
			input:          []string{"KEY5", "4"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR increment or decrement would overflow")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store = setupTest(store)
			if tt.setup != nil {
				tt.setup()
			}

			response := evalDECRBY(tt.input, store)

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

func testEvalBFRESERVE(t *testing.T, store *dstore.Store) {
	tests := []evalTestCase{
		{
			name:           "BF.RESERVE with nil value",
			input:          nil,
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR wrong number of arguments for 'bf.reserve' command")},
		},
		{
			name:           "BF.RESERVE with empty array",
			input:          []string{},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR wrong number of arguments for 'bf.reserve' command")},
		},
		{
			name:           "BF.RESERVE with invalid error rate",
			input:          []string{"myBloomFilter", "invalid_rate", "1000"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR bad error rate")},
		},
		{
			name:           "BF.RESERVE successful reserve",
			input:          []string{"myBloomFilter", "0.01", "1000"},
			migratedOutput: EvalResponse{Result: clientio.OK, Error: nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store = setupTest(store)
			if tt.setup != nil {
				tt.setup()
			}

			response := evalBFRESERVE(tt.input, store)

			assert.Equal(t, tt.migratedOutput.Result, response.Result)
			if tt.migratedOutput.Error != nil {
				assert.Error(t, response.Error, tt.migratedOutput.Error.Error())
			}
		})
	}
}

func testEvalBFINFO(t *testing.T, store *dstore.Store) {
	tests := []evalTestCase{
		{
			name:           "BF.INFO with nil value",
			input:          nil,
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR wrong number of arguments for 'bf.info' command")},
		},
		{
			name:           "BF.INFO with empty array",
			input:          []string{},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR wrong number of arguments for 'bf.info' command")},
		},
		{
			name:           "BF.INFO on non-existent filter",
			input:          []string{"nonExistentFilter"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR not found")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store = setupTest(store)
			if tt.setup != nil {
				tt.setup()
			}

			response := evalBFINFO(tt.input, store)
			assert.Equal(t, tt.migratedOutput.Result, response.Result)

			if tt.migratedOutput.Error != nil {
				assert.Error(t, response.Error, tt.migratedOutput.Error.Error())
			}
		})
	}
}

func testEvalBFADD(t *testing.T, store *dstore.Store) {
	tests := []evalTestCase{
		{
			name:           "BF.ADD with nil value",
			input:          nil,
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR wrong number of arguments for 'bf.add' command")},
		},
		{
			name:           "BF.ADD with empty array",
			input:          []string{},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR wrong number of arguments for 'bf.add' command")},
		},
		{
			name:           "BF.ADD to non-existent filter",
			input:          []string{"nonExistentFilter", "element"},
			migratedOutput: EvalResponse{Result: clientio.IntegerOne, Error: nil},
		},
		{
			name: "BF.ADD to existing filter",
			setup: func() {
				evalBFRESERVE([]string{"myBloomFilter", "0.01", "1000"}, store)
			},
			input:          []string{"myBloomFilter", "element"},
			migratedOutput: EvalResponse{Result: clientio.IntegerOne, Error: nil}, // 1 for new addition, 0 if already exists
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store = setupTest(store)
			if tt.setup != nil {
				tt.setup()
			}

			response := evalBFADD(tt.input, store)

			assert.Equal(t, tt.migratedOutput.Result, response.Result)
			if tt.migratedOutput.Error != nil {
				assert.Error(t, response.Error, tt.migratedOutput.Error.Error())
			}

		})
	}
}

func testEvalBFEXISTS(t *testing.T, store *dstore.Store) {
	tests := []evalTestCase{
		{
			name:           "BF.EXISTS with nil value",
			input:          nil,
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR wrong number of arguments for 'bf.exists' command")},
		},
		{
			name:           "BF.EXISTS with empty array",
			input:          []string{},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR wrong number of arguments for 'bf.exists' command")},
		},
		{
			name:           "BF.EXISTS on non-existent filter",
			input:          []string{"nonExistentFilter", "element"},
			migratedOutput: EvalResponse{Result: clientio.IntegerZero, Error: nil}, // Item does not exist
		},
		{
			name: "BF.EXISTS element not in filter",
			setup: func() {
				evalBFRESERVE([]string{"myBloomFilter", "0.01", "1000"}, store)
			},
			input:          []string{"myBloomFilter", "element"},
			migratedOutput: EvalResponse{Result: clientio.IntegerZero, Error: nil},
		},
		{
			name: "BF.EXISTS element in filter",
			setup: func() {
				evalBFRESERVE([]string{"myBloomFilter", "0.01", "1000"}, store)
				evalBFADD([]string{"myBloomFilter", "element"}, store)
			},
			input:          []string{"myBloomFilter", "element"},
			migratedOutput: EvalResponse{Result: clientio.IntegerOne, Error: nil}, // 1 indicates the element exists
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store = setupTest(store)
			if tt.setup != nil {
				tt.setup()
			}

			response := evalBFEXISTS(tt.input, store)

			assert.Equal(t, tt.migratedOutput.Result, response.Result)
			if tt.migratedOutput.Error != nil {
				assert.Error(t, response.Error, tt.migratedOutput.Error.Error())
			}
		})
	}
}

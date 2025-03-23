// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

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

	"github.com/axiomhq/hyperloglog"
	"github.com/bytedance/sonic"
	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/internal/cmd"
	diceerrors "github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/eval/sortedset"
	"github.com/dicedb/dice/internal/object"
	"github.com/dicedb/dice/internal/server/utils"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/ohler55/ojg/jp"
	"github.com/stretchr/testify/assert"
)

type evalTestCase struct {
	name           string
	setup          func()
	input          []string
	output         []byte
	newValidator   func(output interface{})
	migratedOutput EvalResponse
}

type evalMultiShardTestCase struct {
	name      string
	setup     func()
	input     *cmd.DiceDBCmd
	validator func(output interface{})
	output    EvalResponse
}

func setupTest(store *dstore.Store) *dstore.Store {
	dstore.Reset(store)
	return store
}

func TestEval(t *testing.T) {
	store := dstore.NewStore(nil, nil)

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
	testEvalJSONNUMINCRBY(t, store)
	testEvalJSONNUMMULTBY(t, store)
	testEvalJSONTOGGLE(t, store)
	testEvalJSONARRAPPEND(t, store)
	testEvalJSONRESP(t, store)
	testEvalTTL(t, store)
	testEvalPTTL(t, store)
	testEvalDel(t, store)
	testEvalPersist(t, store)
	testEvalEXPIRE(t, store)
	testEvalEXPIRETIME(t, store)
	testEvalEXPIREAT(t, store)
	testEvalGETSET(t, store)
	testEvalHSET(t, store)
	testEvalHMSET(t, store)
	testEvalHKEYS(t, store)
	testEvalPFADD(t, store)
	testEvalPFCOUNT(t, store)
	testEvalPFMERGE(t, store)
	testEvalHGET(t, store)
	testEvalHGETALL(t, store)
	testEvalHMGET(t, store)
	testEvalHSTRLEN(t, store)
	testEvalHEXISTS(t, store)
	testEvalHDEL(t, store)
	testEvalHSCAN(t, store)
	testEvalJSONSTRLEN(t, store)
	testEvalJSONOBJLEN(t, store)
	testEvalHLEN(t, store)
	testEvalLPUSH(t, store)
	testEvalRPUSH(t, store)
	testEvalLPOP(t, store)
	testEvalRPOP(t, store)
	testEvalLLEN(t, store)
	testEvalLINSERT(t, store)
	testEvalLRANGE(t, store)
	testEvalGETDEL(t, store)
	testEvalGETEX(t, store)
	testEvalDUMP(t, store)
	testEvalTYPE(t, store)
	testEvalCOMMAND(t, store)
	testEvalHINCRBY(t, store)
	testEvalJSONOBJKEYS(t, store)
	testEvalGETRANGE(t, store)
	testEvalHSETNX(t, store)
	testEvalPING(t, store)
	testEvalSETEX(t, store)
	testEvalINCRBYFLOAT(t, store)
	testEvalAPPEND(t, store)
	testEvalHRANDFIELD(t, store)
	testEvalSADD(t, store)
	testEvalSREM(t, store)
	testEvalSCARD(t, store)
	testEvalSMEMBERS(t, store)
	testEvalZADD(t, store)
	testEvalZRANGE(t, store)
	testEvalZPOPMAX(t, store)
	testEvalZPOPMIN(t, store)
	testEvalZRANK(t, store)
	testEvalZCARD(t, store)
	testEvalZREM(t, store)
	testEvalZADD(t, store)
	testEvalZRANGE(t, store)
	testEvalHVALS(t, store)
	testEvalBitField(t, store)
	testEvalHINCRBYFLOAT(t, store)
	testEvalBitFieldRO(t, store)
	testEvalGEOADD(t, store)
	testEvalGEODIST(t, store)
	testEvalGEOPOS(t, store)
	testEvalGEOHASH(t, store)
	testEvalJSONSTRAPPEND(t, store)
	testEvalINCR(t, store)
	testEvalINCRBY(t, store)
	testEvalDECR(t, store)
	testEvalDECRBY(t, store)
	testEvalBFRESERVE(t, store)
	testEvalBFINFO(t, store)
	testEvalBFEXISTS(t, store)
	testEvalBFADD(t, store)
	testEvalJSONARRINDEX(t, store)
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
	serverID = fmt.Sprintf("%s:%d", config.Config.Host, config.Config.Port)
	resp := []interface{}{
		"proto", 2,
		"id", serverID,
		"mode", "standalone",
		"role", "master",
		"modules",
		[]interface{}{},
	}

	tests := map[string]evalTestCase{
		"nil value":            {input: nil, output: Encode(resp, false)},
		"empty args":           {input: []string{}, output: Encode(resp, false)},
		"one value":            {input: []string{"HEY"}, output: Encode(resp, false)},
		"more than one values": {input: []string{"HEY", "HELLO"}, output: []byte("-ERR wrong number of arguments for 'hello' command\r\n")},
	}

	runEvalTests(t, tests, evalHELLO, store)
}

func testEvalSET(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"nil value": {
			input:          nil,
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR wrong number of arguments for 'set' command")},
		},
		"empty array": {
			input:          []string{},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR wrong number of arguments for 'set' command")},
		},
		"one value": {
			input:          []string{"KEY"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR wrong number of arguments for 'set' command")},
		},
		"key val pair": {
			input:          []string{"KEY", "VAL"},
			migratedOutput: EvalResponse{Result: OK, Error: nil},
		},
		"key val pair with int val": {
			input:          []string{"KEY", "123456"},
			migratedOutput: EvalResponse{Result: OK, Error: nil},
		},
		"key val pair and expiry key": {
			input:          []string{"KEY", "VAL", Px},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR syntax error")},
		},
		"key val pair and EX no val": {
			input:          []string{"KEY", "VAL", Ex},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR syntax error")},
		},
		"key val pair and valid EX": {
			input:          []string{"KEY", "VAL", Ex, "2"},
			migratedOutput: EvalResponse{Result: OK, Error: nil},
		},
		"key val pair and invalid negative EX": {
			input:          []string{"KEY", "VAL", Ex, "-2"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR invalid expire time in 'set' command")},
		},
		"key val pair and invalid float EX": {
			input:          []string{"KEY", "VAL", Ex, "2.0"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR value is not an integer or out of range")},
		},
		"key val pair and invalid out of range int EX": {
			input:          []string{"KEY", "VAL", Ex, "9223372036854775807"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR invalid expire time in 'set' command")},
		},
		"key val pair and invalid greater than max duration EX": {
			input:          []string{"KEY", "VAL", Ex, "9223372036854775"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR invalid expire time in 'set' command")},
		},
		"key val pair and invalid EX": {
			input:          []string{"KEY", "VAL", Ex, "invalid_expiry_val"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR value is not an integer or out of range")},
		},
		"key val pair and PX no val": {
			input:          []string{"KEY", "VAL", Px},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR syntax error")},
		},
		"key val pair and valid PX": {
			input:          []string{"KEY", "VAL", Px, "2000"},
			migratedOutput: EvalResponse{Result: OK, Error: nil},
		},
		"key val pair and invalid PX": {
			input:          []string{"KEY", "VAL", Px, "invalid_expiry_val"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR value is not an integer or out of range")},
		},
		"key val pair and invalid negative PX": {
			input:          []string{"KEY", "VAL", Px, "-2"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR invalid expire time in 'set' command")},
		},
		"key val pair and invalid float PX": {
			input:          []string{"KEY", "VAL", Px, "2.0"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR value is not an integer or out of range")},
		},

		"key val pair and invalid out of range int PX": {
			input:          []string{"KEY", "VAL", Px, "9223372036854775807"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR invalid expire time in 'set' command")},
		},
		"key val pair and invalid greater than max duration PX": {
			input:          []string{"KEY", "VAL", Px, "9223372036854775"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR invalid expire time in 'set' command")},
		},
		"key val pair and both EX and PX": {
			input:          []string{"KEY", "VAL", Ex, "2", Px, "2000"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR syntax error")},
		},
		"key val pair and PXAT no val": {
			input:          []string{"KEY", "VAL", Pxat},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR syntax error")},
		},
		"key val pair and invalid PXAT": {
			input:          []string{"KEY", "VAL", Pxat, "invalid_expiry_val"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR value is not an integer or out of range")},
		},
		"key val with get": {
			input: []string{"key", "bazz", "GET"},
			setup: func() {
				key := "key"
				value := "bar"
				obj := store.NewObj(value, -1, object.ObjTypeString)
				store.Put(key, obj)
			},
			migratedOutput: EvalResponse{Result: "bar", Error: nil},
		},
		"key val with get and nil get": {
			input:          []string{"key", "bar", "GET"},
			migratedOutput: EvalResponse{Result: NIL, Error: nil},
		},
		"key val with get and but value is json": {
			input: []string{"key", "bar", "GET"},
			setup: func() {
				key := "key"
				value := "{\"a\":2}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			migratedOutput: EvalResponse{Result: nil, Error: diceerrors.ErrWrongTypeOperation},
		},
	}

	runMigratedEvalTests(t, tests, evalSET, store)
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
			input: []string{"foo", Ex, "10"},
			migratedOutput: EvalResponse{
				Result: "bar",
				Error:  nil,
			},
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
			input: []string{"foo", Ex, "10000000000000000"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrInvalidExpireTime("GETEX"),
			},
		},
		"key val pair and EX and string expire time": {
			setup: func() {
				key := "foo"
				value := "bar"
				obj := &object.Obj{
					Value: value,
				}
				store.Put(key, obj)
			},
			input: []string{"foo", Ex, "string"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrIntegerOutOfRange,
			},
		},
		"key val pair and both EX and PERSIST": {
			setup: func() {
				key := "foo"
				value := "bar"
				obj := &object.Obj{
					Value: value,
				}
				store.Put(key, obj)
			},
			input: []string{"foo", Ex, Persist},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrIntegerOutOfRange,
			},
		},
		"key holding json type": {
			setup: func() {
				evalJSONSET([]string{"JSONKEY", "$", "1"}, store)
			},
			input: []string{"JSONKEY"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongTypeOperation,
			},
		},
		"key holding set type": {
			setup: func() {
				evalSADD([]string{"SETKEY", "FRUITS", "APPLE", "MANGO", "BANANA"}, store)
			},
			input: []string{"SETKEY"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongTypeOperation,
			},
		},
	}

	runMigratedEvalTests(t, tests, evalGETEX, store)
}

func testEvalGETDEL(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"nil value": {
			input:          nil,
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR wrong number of arguments for 'getdel' command")},
		},
		"empty array": {
			input:          []string{},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR wrong number of arguments for 'getdel' command")},
		},
		"key does not exist": {
			input:          []string{"NONEXISTENT_KEY"},
			migratedOutput: EvalResponse{Result: NIL, Error: nil},
		},
		"multiple arguments": {
			input:          []string{"KEY1", "KEY2"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR wrong number of arguments for 'getdel' command")},
		},
		"key exists": {
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
			input:          []string{"EXISTING_KEY"},
			migratedOutput: EvalResponse{Result: NIL, Error: nil},
		},
		"key deleted by previous call of GETDEL": {
			setup: func() {
				key := "DELETED_KEY"
				value := "mock_value"
				obj := &object.Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)
				evalGETDEL([]string{key}, store)
			},
			input:          []string{"DELETED_KEY"},
			migratedOutput: EvalResponse{Result: NIL, Error: nil},
		},
	}

	runMigratedEvalTests(t, tests, evalGETDEL, store)
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
			migratedOutput: EvalResponse{Result: NIL, Error: nil},
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
			migratedOutput: EvalResponse{Result: NIL, Error: nil},
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
					assert.True(t, bytes.Equal(b, expectedBytes), "expected and actual byte slices should be equal")
				}
			} else {
				assert.Equal(t, tt.migratedOutput.Result, response.Result)
			}

			if tt.migratedOutput.Error != nil {
				assert.EqualError(t, response.Error, tt.migratedOutput.Error.Error())
			} else {
				assert.NoError(t, response.Error)
			}
		})
	}
}

func testEvalEXPIRE(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"nil value": {
			input: nil,
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("EXPIRE"),
			},
		},
		"empty args": {
			input: []string{},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("EXPIRE"),
			},
		},
		"wrong number of args": {
			input: []string{"KEY1"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("EXPIRE"),
			},
		},
		"key does not exist": {
			input: []string{"NONEXISTENT_KEY", strconv.FormatInt(1, 10)},
			migratedOutput: EvalResponse{
				Result: IntegerZero,
				Error:  nil,
			},
		},
		"invalid expiry time - 0": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "mock_value"
				obj := &object.Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)
			},
			input: []string{"EXISTING_KEY", "0"},
			migratedOutput: EvalResponse{
				Result: IntegerOne,
				Error:  nil,
			},
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
			input: []string{"EXISTING_KEY", strconv.FormatInt(1, 10)},
			migratedOutput: EvalResponse{
				Result: IntegerOne,
				Error:  nil,
			},
		},
		"invalid expiry time - very large integer": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "mock_value"
				obj := &object.Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)
			},
			input: []string{"EXISTING_KEY", strconv.FormatInt(9223372036854776, 10)},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrInvalidExpireTime("EXPIRE"),
			},
		},

		"invalid expiry time - negative integer": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "mock_value"
				obj := &object.Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)
			},
			input: []string{"EXISTING_KEY", strconv.FormatInt(-1, 10)},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrInvalidExpireTime("EXPIRE"),
			},
		},
		"invalid expiry time - empty string": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "mock_value"
				obj := &object.Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)
			},
			input: []string{"EXISTING_KEY", ""},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrIntegerOutOfRange,
			},
		},
		"invalid expiry time - with float number": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "mock_value"
				obj := &object.Obj{
					Value:          value,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(key, obj)
			},
			input: []string{"EXISTING_KEY", "0.456"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrIntegerOutOfRange,
			},
		},
	}

	runMigratedEvalTests(t, tests, evalEXPIRE, store)
}

func testEvalEXPIRETIME(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"wrong number of args": {
			input: []string{"KEY1", "KEY2"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("EXPIRETIME"),
			},
		},
		"key does not exist": {
			input: []string{"NONEXISTENT_KEY"},
			migratedOutput: EvalResponse{
				Result: IntegerNegativeTwo,
				Error:  nil,
			},
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
			input: []string{"EXISTING_KEY"},
			migratedOutput: EvalResponse{
				Result: IntegerNegativeOne,
				Error:  nil,
			},
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
			input: []string{"EXISTING_KEY"},
			migratedOutput: EvalResponse{
				Result: uint64(2724123456123),
				Error:  nil,
			},
		},
	}

	runMigratedEvalTests(t, tests, evalEXPIRETIME, store)
}

func testEvalEXPIREAT(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"nil value": {
			input: nil,
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("EXPIREAT"),
			},
		},
		"empty args": {
			input: []string{},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("EXPIREAT"),
			},
		},
		"wrong number of args": {
			input: []string{"KEY1"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("EXPIREAT"),
			},
		},
		"key does not exist": {
			input: []string{"NONEXISTENT_KEY", strconv.FormatInt(time.Now().Add(2*time.Minute).Unix(), 10)},
			migratedOutput: EvalResponse{
				Result: IntegerZero,
				Error:  nil,
			},
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
			input: []string{"EXISTING_KEY", strconv.FormatInt(time.Now().Add(2*time.Minute).Unix(), 10)},
			migratedOutput: EvalResponse{
				Result: IntegerOne,
				Error:  nil,
			},
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
			input: []string{"EXISTING_KEY", strconv.FormatInt(9223372036854776, 10)},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrInvalidExpireTime("EXPIREAT"),
			},
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
			input: []string{"EXISTING_KEY", strconv.FormatInt(-1, 10)},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrInvalidExpireTime("EXPIREAT"),
			},
		},
	}

	runMigratedEvalTests(t, tests, evalEXPIREAT, store)
}

func testEvalJSONARRTRIM(t *testing.T, store *dstore.Store) {
	tests := []evalTestCase{
		{
			name:  "nil value",
			setup: func() {},
			input: nil,
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("JSON.ARRTRIM"),
			},
		},
		{
			name:  "key does not exist",
			setup: func() {},
			input: []string{"NONEXISTENT_KEY", "$.a", "0", "1"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral("key does not exist"),
			},
		},
		{
			name: "index is not integer",
			setup: func() {
				key := "EXISTING_KEY"
				value := `{"a":2}`
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input: []string{"EXISTING_KEY", "$", "a", "1"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrIntegerOutOfRange,
			},
		},
		{
			name: "index out of bounds",
			setup: func() {
				key := "EXISTING_KEY"
				value := `[1,2,3]`
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input: []string{"EXISTING_KEY", "$", "0", "10"},
			migratedOutput: EvalResponse{
				Result: []interface{}{3},
				Error:  nil,
			},
		},
		{
			name: "root path is not array",
			setup: func() {
				key := "EXISTING_KEY"
				value := `{"a":2}`
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input: []string{"EXISTING_KEY", "$.a", "0", "6"},
			migratedOutput: EvalResponse{
				Result: []interface{}{nil},
				Error:  nil,
			},
		},
		{
			name: "root path is array",
			setup: func() {
				key := "EXISTING_KEY"
				value := `[1,2,3,4,5]`
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input: []string{"EXISTING_KEY", "$", "0", "2"},
			migratedOutput: EvalResponse{
				Result: []interface{}{3},
				Error:  nil,
			},
		},
		{
			name: "subpath array",
			setup: func() {
				key := "EXISTING_KEY"
				value := `{"connection":{"wireless":true,"names":[0,1,2,3,4]},"names":[0,1,2,3,4]}`
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input: []string{"EXISTING_KEY", "$.names", "1", "3"},
			migratedOutput: EvalResponse{
				Result: []interface{}{3},
				Error:  nil,
			},
		},
		{
			name: "subpath two arrays",
			setup: func() {
				key := "EXISTING_KEY"
				value := `{
					"connection": {
						"wireless": true,
						"names": [0,1,2,3,4]
					},
					"names": [0,1,2,3,4]
				}`
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input: []string{"EXISTING_KEY", "$..names", "1", "3"},
			migratedOutput: EvalResponse{
				Result: []interface{}{3, 3},
				Error:  nil,
			},
		},
		{
			name: "subpath not array",
			setup: func() {
				key := "EXISTING_KEY"
				value := `{"connection":{"wireless":true,"names":[0,1,2,3,4]},"names":[0,1,2,3,4]}`
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input: []string{"EXISTING_KEY", "$.connection", "1", "2"},
			migratedOutput: EvalResponse{
				Result: []interface{}{nil},
				Error:  nil,
			},
		},
		{
			name: "subpath array index negative",
			setup: func() {
				key := "EXISTING_KEY"
				value := `{"connection":{"wireless":true,"names":[0,1,2,3,4]},"names":[0,1,2,3,4]}`
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input: []string{"EXISTING_KEY", "$.names", "-3", "-1"},
			migratedOutput: EvalResponse{
				Result: []interface{}{3},
				Error:  nil,
			},
		},
		{
			name: "index negative start larger than stop",
			setup: func() {
				key := "EXISTING_KEY"
				value := `{"connection":{"wireless":true,"names":[0,1,2,3,4]},"names":[0,1,2,3,4]}`
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input: []string{"EXISTING_KEY", "$.names", "-1", "-3"},
			migratedOutput: EvalResponse{
				Result: []interface{}{0},
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

			response := evalJSONARRTRIM(tt.input, store)

			// Handle comparison for byte slices
			if b, ok := response.Result.([]byte); ok && tt.migratedOutput.Result != nil {
				if expectedBytes, ok := tt.migratedOutput.Result.([]byte); ok {
					assert.True(t, bytes.Equal(b, expectedBytes), "expected and actual byte slices should be equal")
				}
			} else {
				assert.Equal(t, tt.migratedOutput.Result, response.Result)
			}

			if tt.migratedOutput.Error != nil {
				assert.EqualError(t, response.Error, tt.migratedOutput.Error.Error())
			} else {
				assert.NoError(t, response.Error)
			}
		})
	}
}

func testEvalJSONARRINSERT(t *testing.T, store *dstore.Store) {
	tests := []evalTestCase{
		{
			name:  "nil value",
			setup: func() {},
			input: nil,
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("JSON.ARRINSERT"),
			},
		},
		{
			name: "key does not exist",
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"a\":2}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input: []string{"NONEXISTENT_KEY", "$.a", "0", "1"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral("key does not exist"),
			},
		},
		{
			name: "index is not integer",
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"a\":2}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input: []string{"EXISTING_KEY", "$.a", "a", "1"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrIntegerOutOfRange,
			},
		},
		{
			name: "index out of bounds",
			setup: func() {
				key := "EXISTING_KEY"
				value := "[1,2,3]"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input: []string{"EXISTING_KEY", "$", "4", "\"a\"", "1"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral("index out of bounds"),
			},
		},
		{
			name: "root path is not array",
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"a\":2}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input: []string{"EXISTING_KEY", "$.a", "0", "6"},
			migratedOutput: EvalResponse{
				Result: []interface{}{nil},
				Error:  nil,
			},
		},
		{
			name: "root path is array",
			setup: func() {
				key := "EXISTING_KEY"
				value := "[1,2]"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input: []string{"EXISTING_KEY", "$", "0", "6", "\"a\"", "3.14"},
			migratedOutput: EvalResponse{
				Result: []interface{}{5},
				Error:  nil,
			},
		},
		{
			name: "subpath array insert positive index",
			setup: func() {
				key := "EXISTING_KEY"
				value := `{"connection":{"wireless":true,"names":["1","2"]},"price":99.98,"names":[3,4]}`
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input: []string{"EXISTING_KEY", "$..names", "2", "7", "8"},
			migratedOutput: EvalResponse{
				Result: []interface{}{4, 4},
				Error:  nil,
			},
		},
		{
			name: "subpath array insert negative index",
			setup: func() {
				key := "EXISTING_KEY"
				value := `{"connection":{"wireless":true,"names":["1","2"]},"price":99.98,"names":[3,4]}`
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input: []string{"EXISTING_KEY", "$..names", "-1", "7", "8"},
			migratedOutput: EvalResponse{
				Result: []interface{}{4, 4},
				Error:  nil,
			},
		},
		{
			name: "array insert with multitype value",
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"a\":[1,2,3]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input: []string{"EXISTING_KEY", "$.a", "0", "1", "null", "3.14", "true", "{\"a\":123}"},
			migratedOutput: EvalResponse{
				Result: []interface{}{8},
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

			response := evalJSONARRINSERT(tt.input, store)

			if b, ok := response.Result.([]byte); ok && tt.migratedOutput.Result != nil {
				if expectedBytes, ok := tt.migratedOutput.Result.([]byte); ok {
					assert.True(t, bytes.Equal(b, expectedBytes), "expected and actual byte slices should be equal")
				}
			} else {
				assert.Equal(t, tt.migratedOutput.Result, response.Result)
			}

			if tt.migratedOutput.Error != nil {
				assert.EqualError(t, response.Error, tt.migratedOutput.Error.Error())
			} else {
				assert.NoError(t, response.Error)
			}
		})
	}
}

func testEvalJSONARRLEN(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"nil value": {
			setup:  func() {},
			input:  nil,
			output: []byte("-ERR wrong number of arguments for 'json.arrlen' command\r\n"),
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("JSON.ARRLEN"),
			},
		},
		"key does not exist": {
			setup:  func() {},
			input:  []string{"NONEXISTENT_KEY"},
			output: []byte("$-1\r\n"),
			migratedOutput: EvalResponse{
				Result: NIL,
				Error:  nil,
			},
		},
		"root not array arrlen": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"age\":13,\"name\":\"a\"}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input:  []string{"EXISTING_KEY"},
			output: []byte("-ERR Path '$' does not exist or not an array\r\n"),
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongTypeOperation,
			},
		},
		"root array arrlen": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "[1,2,3]"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input:  []string{"EXISTING_KEY"},
			output: []byte(":3\r\n"),
			migratedOutput: EvalResponse{
				Result: 3,
				Error:  nil,
			},
		},
		"wildcase no array arrlen": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"age\":13,\"high\":1.60,\"pet\":null,\"flag\":false, \"partner\":{\"name\":\"tom\"}}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},

			input:  []string{"EXISTING_KEY", "$.*"},
			output: []byte("*5\r\n$-1\r\n$-1\r\n$-1\r\n$-1\r\n$-1\r\n"),
			migratedOutput: EvalResponse{
				Result: []interface{}{NIL, NIL, NIL, NIL, NIL},
				Error:  nil,
			},
		},
		"subpath array arrlen": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"age\":13,\"high\":1.60,\"pet\":null,\"language\":[\"python\",\"golang\"], " +
					"\"flag\":false, \"partner\":{\"name\":\"tom\"}}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},

			input:  []string{"EXISTING_KEY", "$.language"},
			output: []byte(":2\r\n"),
			migratedOutput: EvalResponse{
				Result: 2,
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
			response := evalJSONARRLEN(tt.input, store)

			if tt.migratedOutput.Result != nil {
				if slice, ok := tt.migratedOutput.Result.([]interface{}); ok {
					assert.Equal(t, slice, response.Result)
				} else {
					assert.Equal(t, tt.migratedOutput.Result, response.Result)
				}
			}

			if tt.migratedOutput.Error != nil {
				assert.EqualError(t, response.Error, tt.migratedOutput.Error.Error())
			} else {
				assert.NoError(t, response.Error)
			}
		})
	}
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
				Error:  diceerrors.ErrKeyDoesNotExist,
			},
		},
		"jsonobjlen root not object": {
			name: "jsonobjlen root not object",
			setup: func() {
				key := "EXISTING_KEY"
				value := "[1,2,3]"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
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
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
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
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
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
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
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
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
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
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
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
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
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
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
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
					assert.Equal(t, slice, response.Result)
				} else {
					assert.Equal(t, tt.migratedOutput.Result, response.Result)
				}
			}

			if tt.migratedOutput.Error != nil {
				assert.EqualError(t, response.Error, tt.migratedOutput.Error.Error())
			} else {
				assert.NoError(t, response.Error)
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
		"JSON.DEL : nil value": {
			name:  "JSON.DEL : nil value",
			setup: func() {},
			input: nil,
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("JSON.DEL"),
			},
		},
		"JSON.DEL : key does not exist": {
			name:  "JSON.DEL : key does not exist",
			setup: func() {},
			input: []string{"NONEXISTENT_KEY"},
			migratedOutput: EvalResponse{
				Result: IntegerZero,
				Error:  nil,
			},
		},
		"JSON.DEL : root path del": {
			name: "JSON.DEL : root path del",
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"age\":13,\"high\":1.60,\"pet\":null,\"language\":[\"python\",\"golang\"], " +
					"\"flag\":false, \"partner\":{\"name\":\"tom\",\"language\":[\"rust\"]}}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input: []string{"EXISTING_KEY"},
			migratedOutput: EvalResponse{
				Result: 1,
				Error:  nil,
			},
		},
		"JSON.DEL : partial path del": {
			name: "JSON.DEL : partial path del",
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"age\":13,\"high\":1.60,\"pet\":null,\"language\":[\"python\",\"golang\"], " +
					"\"flag\":false, \"partner\":{\"name\":\"tom\",\"language\":[\"rust\"]}}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},

			input: []string{"EXISTING_KEY", "$..language"},
			migratedOutput: EvalResponse{
				Result: 2,
				Error:  nil,
			},
		},
		"JSON.DEL : wildcard path del": {
			name: "JSON.DEL : wildcard path del",
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"age\":13,\"high\":1.60,\"pet\":null,\"language\":[\"python\",\"golang\"], " +
					"\"flag\":false, \"partner\":{\"name\":\"tom\",\"language\":[\"rust\"]}}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},

			input: []string{"EXISTING_KEY", "$.*"},
			migratedOutput: EvalResponse{
				Result: 6,
				Error:  nil,
			},
		},
	}
	runMigratedEvalTests(t, tests, evalJSONDEL, store)
}

func testEvalJSONFORGET(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"JSON.FORGET : nil value": {
			name:  "JSON.FORGET : nil value",
			setup: func() {},
			input: nil,
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("JSON.FORGET"),
			},
		},
		"JSON.FORGET : key does not exist": {
			name:  "JSON.FORGET : key does not exist",
			setup: func() {},
			input: []string{"NONEXISTENT_KEY"},
			migratedOutput: EvalResponse{
				Result: IntegerZero,
				Error:  nil,
			},
		},
		"JSON.FORGET : root path forget": {
			name: "JSON.FORGET : root path forget",
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"age\":13,\"high\":1.60,\"pet\":null,\"language\":[\"python\",\"golang\"], " +
					"\"flag\":false, \"partner\":{\"name\":\"tom\",\"language\":[\"rust\"]}}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input: []string{"EXISTING_KEY"},
			migratedOutput: EvalResponse{
				Result: 1,
				Error:  nil,
			},
		},
		"JSON.FORGET : partial path forget": {
			name: "JSON.FORGET : partial path forget",
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"age\":13,\"high\":1.60,\"pet\":null,\"language\":[\"python\",\"golang\"], " +
					"\"flag\":false, \"partner\":{\"name\":\"tom\",\"language\":[\"rust\"]}}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},

			input: []string{"EXISTING_KEY", "$..language"},
			migratedOutput: EvalResponse{
				Result: 2,
				Error:  nil,
			},
		},
		"JSON.FORGET : wildcard path forget": {
			name: "JSON.FORGET : wildcard path forget",
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"age\":13,\"high\":1.60,\"pet\":null,\"language\":[\"python\",\"golang\"], " +
					"\"flag\":false, \"partner\":{\"name\":\"tom\",\"language\":[\"rust\"]}}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},

			input: []string{"EXISTING_KEY", "$.*"},
			migratedOutput: EvalResponse{
				Result: 6,
				Error:  nil,
			},
		},
	}
	runMigratedEvalTests(t, tests, evalJSONFORGET, store)
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
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
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
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
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
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
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
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
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
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
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
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
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
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
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
					assert.Equal(t, slice, response.Result)
				} else {
					assert.Equal(t, tt.migratedOutput.Result, response.Result)
				}
			}

			if tt.migratedOutput.Error != nil {
				assert.EqualError(t, response.Error, tt.migratedOutput.Error.Error())
			} else {
				assert.NoError(t, response.Error)
			}
		})
	}
}

func testEvalJSONTYPE(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"nil value": {
			setup: func() {},
			input: nil,
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("JSON.TYPE"),
			},
		},
		"empty array": {
			setup: func() {},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("JSON.TYPE"),
			},
		},
		"key does not exist": {
			setup: func() {},
			input: []string{"NONEXISTENT_KEY"},
			migratedOutput: EvalResponse{
				Result: NIL,
				Error:  nil,
			},
		},
		"object type value": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"language\":[\"java\",\"go\",\"python\"]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},

			input: []string{"EXISTING_KEY"},
			migratedOutput: EvalResponse{
				Result: "object",
				Error:  nil,
			},
		},
		"array type value": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"language\":[\"java\",\"go\",\"python\"]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},

			input: []string{"EXISTING_KEY", "$.language"},
			migratedOutput: EvalResponse{
				Result: "array",
				Error:  nil,
			},
		},
		"string type value": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"a\":\"test\"}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},

			input: []string{"EXISTING_KEY", "$.a"},
			migratedOutput: EvalResponse{
				Result: "string",
				Error:  nil,
			},
		},
		"boolean type value": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"flag\":true}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},

			input: []string{"EXISTING_KEY", "$.flag"},
			migratedOutput: EvalResponse{
				Result: "boolean",
				Error:  nil,
			},
		},
		"number type value": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"price\":3}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},

			input: []string{"EXISTING_KEY", "$.price"},
			migratedOutput: EvalResponse{
				Result: "number",
				Error:  nil,
			},
		},
		"null type value": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"price\":3.14}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},

			input: []string{"EXISTING_KEY", "$.language"},
			migratedOutput: EvalResponse{
				Result: EmptyArray,
				Error:  nil,
			},
		},
		"multi type value": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"name\":\"tom\",\"partner\":{\"name\":\"jerry\"}}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},

			input: []string{"EXISTING_KEY", "$..name"},
			migratedOutput: EvalResponse{
				Result: "string",
				Error:  nil,
			},
		},
	}

	runMigratedEvalTests(t, tests, evalJSONTYPE, store)
}

func testEvalJSONGET(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"nil value": {
			setup: func() {},
			input: nil,
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("JSON.GET"),
			},
		},
		"empty array": {
			setup: func() {},
			input: []string{},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("JSON.GET"),
			},
		},
		"key does not exist": {
			setup: func() {},
			input: []string{"NONEXISTENT_KEY"},
			migratedOutput: EvalResponse{
				Result: NIL,
				Error:  nil,
			},
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
			input: []string{"EXISTING_KEY"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongTypeOperation,
			},
		},
		"key exists value": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"a\":2}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},

			input: []string{"EXISTING_KEY"},
			migratedOutput: EvalResponse{
				Result: "{\"a\":2}",
				Error:  nil,
			},
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
			input: []string{"EXISTING_KEY"},
			migratedOutput: EvalResponse{
				Result: NIL,
				Error:  nil,
			},
		},
	}

	runMigratedEvalTests(t, tests, evalJSONGET, store)
}

func testEvalJSONSET(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"nil value": {
			setup: func() {},
			input: nil,
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("JSON.SET"),
			},
		},
		"empty array": {
			setup: func() {},
			input: []string{},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("JSON.SET"),
			},
		},
		"insufficient args": {
			setup: func() {},
			input: []string{},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("JSON.SET"),
			},
		},
		"invalid json path": {
			setup: func() {},
			input: []string{"doc", "$", "{\"a\":}"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral("invalid JSON"),
			},
		},
		"valid json path": {
			setup: func() {
			},
			input: []string{"doc", "$", "{\"a\":2}"},
			migratedOutput: EvalResponse{
				Result: OK,
				Error:  nil,
			},
		},
	}

	runMigratedEvalTests(t, tests, evalJSONSET, store)
}

func testEvalJSONNUMMULTBY(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"JSON.NUMMULTBY : nil value": {
			name:  "JSON.NUMMULTBY : nil value",
			setup: func() {},
			input: nil,
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("JSON.NUMMULTBY"),
			},
		},
		"JSON.NUMMULTBY : empty array": {
			name:  "JSON.NUMMULTBY : empty array",
			setup: func() {},
			input: []string{},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("JSON.NUMMULTBY"),
			},
		},
		"JSON.NUMMULTBY : insufficient args": {
			name:  "JSON.NUMMULTBY : insufficient args",
			setup: func() {},
			input: []string{"doc"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("JSON.NUMMULTBY"),
			},
		},
		"JSON.NUMMULTBY : non-numeric multiplier on existing key": {
			name: "JSON.NUMMULTBY : non-numeric multiplier on existing key",
			setup: func() {
				key := "doc"
				value := "{\"a\":10,\"b\":[{\"a\":2}, {\"a\":5}, {\"a\":\"c\"}]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input: []string{"doc", "$.a", "qwe"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral("expected value at line 1 column 1"),
			},
		},
		"JSON.NUMMULTBY : nummultby on non integer root fields": {
			name: "JSON.NUMMULTBY : nummultby on non integer root fields",
			setup: func() {
				key := "doc"
				value := "{\"a\": \"b\",\"b\":[{\"a\":2}, {\"a\":5}, {\"a\":\"c\"}]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input: []string{"doc", "$.a", "2"},
			migratedOutput: EvalResponse{
				Result: "[null]",
				Error:  nil,
			},
		},
		"JSON.NUMMULTBY : nummultby on recursive fields": {
			name: "JSON.NUMMULTBY : nummultby on recursive fields",
			setup: func() {
				key := "doc"
				value := "{\"a\": \"b\",\"b\":[{\"a\":2}, {\"a\":5}, {\"a\":\"c\"}]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input: []string{"doc", "$..a", "2"},
			migratedOutput: EvalResponse{
				Result: "[4,10,null,null]",
				Error:  nil,
			},
		},
		"JSON.NUMMULTBY : nummultby on integer root fields": {
			name: "JSON.NUMMULTBY : nummultby on integer root fields",
			setup: func() {
				key := "doc"
				value := "{\"a\":10,\"b\":[{\"a\":2}, {\"a\":5}, {\"a\":\"c\"}]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input: []string{"doc", "$.a", "2"},
			migratedOutput: EvalResponse{
				Result: "[20]",
				Error:  nil,
			},
		},
		"JSON.NUMMULTBY : nummultby on non-existent key": {
			name: "JSON.NUMMULTBY : nummultby on non-existent key",
			setup: func() {
				key := "doc"
				value := "{\"a\":10,\"b\":[{\"a\":2}, {\"a\":5}, {\"a\":\"c\"}]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input: []string{"doc", "$..fe", "2"},
			migratedOutput: EvalResponse{
				Result: "[]",
				Error:  nil,
			},
		},
	}
	runMigratedEvalTests(t, tests, evalJSONNUMMULTBY, store)
}

func testEvalJSONARRAPPEND(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"arr append to non array fields": {
			setup: func() {
				key := "array"
				value := "{\"a\":2}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input:  []string{"array", "$.a", "6"},
			output: []byte("*1\r\n$-1\r\n"),
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrJSONPathNotFound("$.a"),
			},
		},
		"arr append single element to an array field": {
			setup: func() {
				key := "array"
				value := "{\"a\":[1,2]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input:  []string{"array", "$.a", "6"},
			output: []byte("*1\r\n:3\r\n"),
			migratedOutput: EvalResponse{
				Result: []int64{3},
				Error:  nil,
			},
		},
		"arr append multiple elements to an array field": {
			setup: func() {
				key := "array"
				value := "{\"a\":[1,2]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input:  []string{"array", "$.a", "6", "7", "8"},
			output: []byte("*1\r\n:5\r\n"),
			migratedOutput: EvalResponse{
				Result: []int64{5},
				Error:  nil,
			},
		},
		"arr append string value": {
			setup: func() {
				key := "array"
				value := "{\"b\":[\"b\",\"c\"]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input:  []string{"array", "$.b", `"d"`},
			output: []byte("*1\r\n:3\r\n"),
			migratedOutput: EvalResponse{
				Result: []int64{3},
				Error:  nil,
			},
		},
		"arr append nested array value": {
			setup: func() {
				key := "array"
				value := "{\"a\":[[1,2]]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input:  []string{"array", "$.a", "[1,2,3]"},
			output: []byte("*1\r\n:2\r\n"),
			migratedOutput: EvalResponse{
				Result: []int64{2},
				Error:  nil,
			},
		},
		"arr append with json value": {
			setup: func() {
				key := "array"
				value := "{\"a\":[{\"b\": 1}]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input:  []string{"array", "$.a", "{\"c\": 3}"},
			output: []byte("*1\r\n:2\r\n"),
			migratedOutput: EvalResponse{
				Result: []int64{2},
				Error:  nil,
			},
		},
		"arr append to append on multiple fields": {
			setup: func() {
				key := "array"
				value := "{\"a\":[1,2],\"b\":{\"a\":[10]}}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input:  []string{"array", "$..a", "6"},
			output: []byte("*2\r\n:2\r\n:3\r\n"),
			migratedOutput: EvalResponse{
				Result: []int64{2, 3},
				Error:  nil,
			},
		},
		"arr append to append on root node": {
			setup: func() {
				key := "array"
				value := "[1,2,3]"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input:  []string{"array", "$", "6"},
			output: []byte("*1\r\n:4\r\n"),
			migratedOutput: EvalResponse{
				Result: []int64{4},
				Error:  nil,
			},
		},
		"arr append to an array with different type": {
			setup: func() {
				key := "array"
				value := "{\"a\":[1,2]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input:  []string{"array", "$.a", `"blue"`},
			output: []byte("*1\r\n:3\r\n"),
			migratedOutput: EvalResponse{
				Result: []int64{3},
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
			response := evalJSONARRAPPEND(tt.input, store)

			if tt.migratedOutput.Result != nil {
				actual, ok := response.Result.([]int64)
				if ok {
					assert.Equal(t, tt.migratedOutput.Result, actual)
				}
			}

			if tt.migratedOutput.Error != nil {
				assert.EqualError(t, response.Error, tt.migratedOutput.Error.Error())
			} else {
				assert.NoError(t, response.Error)
			}
		})
	}
}

func testEvalJSONTOGGLE(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"JSON.TOGGLE : nil value": {
			name:  "JSON.TOGGLE : nil value",
			setup: func() {},
			input: nil,
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("JSON.TOGGLE"),
			},
		},
		"JSON.TOGGLE : no arguments supplied": {
			name:  "JSON.TOGGLE : no arguments supplied",
			setup: func() {},
			input: []string{},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("JSON.TOGGLE"),
			},
		},
		"JSON.TOGGLE : key does not exist": {
			name:  "JSON.TOGGLE : key does not exist",
			setup: func() {},
			input: []string{"NONEXISTENT_KEY", ".active"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrKeyDoesNotExist,
			},
		},
		"JSON.TOGGLE : key exists, toggling boolean true to false": {
			name: "JSON.TOGGLE : key exists, toggling boolean true to false",
			setup: func() {
				key := "EXISTING_KEY"
				value := `{"active":true}`
				var rootData interface{}
				err := sonic.Unmarshal([]byte(value), &rootData)
				if err != nil {
					fmt.Printf("Debug: Error unmarshaling JSON: %v\n", err)
				}
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input: []string{"EXISTING_KEY", ".active"},
			migratedOutput: EvalResponse{
				Result: []interface{}{0},
				Error:  nil,
			},
		},
		"JSON.TOGGLE : key exists, toggling boolean false to true": {
			name: "JSON.TOGGLE : key exists, toggling boolean false to true",
			setup: func() {
				key := "EXISTING_KEY"
				value := `{"active":false}`
				var rootData interface{}
				err := sonic.Unmarshal([]byte(value), &rootData)
				if err != nil {
					fmt.Printf("Debug: Error unmarshaling JSON: %v\n", err)
				}
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input: []string{"EXISTING_KEY", ".active"},
			migratedOutput: EvalResponse{
				Result: []interface{}{1},
				Error:  nil,
			},
		},
		"JSON.TOGGLE : key exists but expired": {
			name: "JSON.TOGGLE : key exists but expired",
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
			input: []string{"EXISTING_KEY", ".active"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrKeyDoesNotExist,
			},
		},
		"JSON.TOGGLE : nested JSON structure with multiple booleans": {
			name: "JSON.TOGGLE : nested JSON structure with multiple booleans",
			setup: func() {
				key := "NESTED_KEY"
				value := `{"isSimple":true,"nested":{"isSimple":false}}`
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input: []string{"NESTED_KEY", "$..isSimple"},
			migratedOutput: EvalResponse{
				Result: []interface{}{0, 1},
				Error:  nil,
			},
		},
		"JSON.TOGGLE : deeply nested JSON structure with multiple matching fields": {
			name: "JSON.TOGGLE : deeply nested JSON structure with multiple matching fields",
			setup: func() {
				key := "DEEP_NESTED_KEY"
				value := `{"field": true, "nested": {"field": false, "nested": {"field": true}}}`
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input: []string{"DEEP_NESTED_KEY", "$..field"},
			migratedOutput: EvalResponse{
				Result: []interface{}{0, 1, 0},
				Error:  nil,
			},
		},
	}
	runMigratedEvalTests(t, tests, evalJSONTOGGLE, store)
}

func testEvalPTTL(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"nil value": {
			setup: func() {},
			input: nil,
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("PTTL"),
			},
		},
		"empty array": {
			setup: func() {},
			input: []string{},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("PTTL"),
			},
		},
		"key does not exist": {
			setup: func() {},
			input: []string{"NONEXISTENT_KEY"},
			migratedOutput: EvalResponse{
				Result: IntegerNegativeTwo,
				Error:  nil,
			},
		},
		"multiple arguments": {
			setup: func() {},
			input: []string{"KEY1", "KEY2"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("PTTL"),
			},
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
			input: []string{"EXISTING_KEY"},
			migratedOutput: EvalResponse{
				Result: IntegerNegativeOne,
				Error:  nil,
			},
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
			newValidator: func(output interface{}) {
				assert.True(t, output != nil)
				assert.True(t, output != IntegerNegativeOne)
				assert.True(t, output != IntegerNegativeTwo)
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
			input: []string{"EXISTING_KEY"},
			migratedOutput: EvalResponse{
				Result: IntegerNegativeTwo,
				Error:  nil,
			},
		},
	}

	runMigratedEvalTests(t, tests, evalPTTL, store)
}

func testEvalTTL(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"nil value": {
			setup: func() {},
			input: nil,
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("TTL"),
			},
		},
		"empty array": {
			setup: func() {},
			input: []string{},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("TTL"),
			},
		},
		"key does not exist": {
			setup: func() {},
			input: []string{"NONEXISTENT_KEY"},
			migratedOutput: EvalResponse{
				Result: IntegerNegativeTwo,
				Error:  nil,
			},
		},
		"multiple arguments": {
			setup: func() {},
			input: []string{"KEY1", "KEY2"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("TTL"),
			},
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
			input: []string{"EXISTING_KEY"},
			migratedOutput: EvalResponse{
				Result: IntegerNegativeOne,
				Error:  nil,
			},
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
			newValidator: func(output interface{}) {
				assert.True(t, output != nil)
				assert.True(t, output != IntegerNegativeOne)
				assert.True(t, output != IntegerNegativeTwo)
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
			input: []string{"EXISTING_KEY"},
			migratedOutput: EvalResponse{
				Result: IntegerNegativeTwo,
				Error:  nil,
			},
		},
	}

	runMigratedEvalTests(t, tests, evalTTL, store)
}

func testEvalDel(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"DEL nil value": {
			name:  "DEL nil value",
			setup: func() {},
			input: nil,
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("DEL"),
			},
		},
		"DEL empty array": {
			name:  "DEL empty array",
			setup: func() {},
			input: []string{},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("DEL"),
			},
		},
		"DEL key does not exist": {
			name:  "DEL key does not exist",
			setup: func() {},
			input: []string{"NONEXISTENT_KEY"},
			migratedOutput: EvalResponse{
				Result: int64(0),
				Error:  nil,
			},
		},
		"DEL key exists": {
			name: "DEL key exists",
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
			input: []string{"EXISTING_KEY"},
			migratedOutput: EvalResponse{
				Result: int64(1),
				Error:  nil,
			},
		},
	}

	runMigratedEvalTests(t, tests, evalDEL, store)
}

// TestEvalPersist tests the evalPersist function using table-driven tests.
func testEvalPersist(t *testing.T, store *dstore.Store) {
	// Define test cases
	tests := map[string]evalTestCase{
		"PERSIST wrong number of arguments": {
			name:  "PERSIST wrong number of arguments",
			input: []string{},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("PERSIST"),
			},
		},
		"PERSIST key does not exist": {
			name:  "PERSIST key does not exist",
			input: []string{"nonexistent"},
			migratedOutput: EvalResponse{
				Result: IntegerZero,
				Error:  nil,
			},
		},
		"PERSIST key exists but no expiration set": {
			name:  "PERSIST key exists but no expiration set",
			input: []string{"existent_no_expiry"},
			setup: func() {
				evalSET([]string{"existent_no_expiry", "value"}, store)
			},
			migratedOutput: EvalResponse{
				Result: IntegerZero,
				Error:  nil,
			},
		},
		"PERSIST key exists and expiration removed": {
			name:  "PERSIST key exists and expiration removed",
			input: []string{"existent_with_expiry"},
			setup: func() {
				evalSET([]string{"existent_with_expiry", "value", Ex, "1"}, store)
			},
			migratedOutput: EvalResponse{
				Result: IntegerOne,
				Error:  nil,
			},
		},
		"PERSIST key exists with expiration set and not expired": {
			name:  "PERSIST key exists with expiration set and not expired",
			input: []string{"existent_with_expiry_not_expired"},
			setup: func() {
				// Simulate setting a key with an expiration time that has not yet passed
				evalSET([]string{"existent_with_expiry_not_expired", "value", Ex, "10000"}, store) // 10000 seconds in the future
			},
			migratedOutput: EvalResponse{
				Result: IntegerOne,
				Error:  nil,
			},
		},
	}

	runMigratedEvalTests(t, tests, evalPERSIST, store)
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
				_, oType := getRawStringOrInt(value)
				var exDurationMs int64 = -1
				keepttl := false

				store.Put(key, store.NewObj(value, exDurationMs, oType), dstore.WithKeepTTL(keepttl))
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
	tests := map[string]evalMultiShardTestCase{
		"PFMERGE nil value": {
			name:  "PFMERGE nil value",
			setup: func() {},
			input: &cmd.DiceDBCmd{
				Cmd: "PFMERGE",
			},
			output: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("PFMERGE"),
			},
		},
		"PFMERGE empty array": {
			name:  "PFMERGE empty array",
			setup: func() {},
			input: &cmd.DiceDBCmd{
				Cmd:  "PFMERGE",
				Args: []string{},
			},
			output: EvalResponse{
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
			input: &cmd.DiceDBCmd{
				Cmd:  "PFMERGE",
				Args: []string{"INVALID_OBJ_DEST_KEY"},
			},
			output: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrInvalidHyperLogLogKey,
			},
		},
		"PFMERGE destKey doesn't exist": {
			name:  "PFMERGE destKey doesn't exist",
			setup: func() {},
			input: &cmd.DiceDBCmd{
				Cmd:  "PFMERGE",
				Args: []string{"NON_EXISTING_DEST_KEY"},
			},
			output: EvalResponse{
				Result: OK,
				Error:  nil,
			},
		},
		"PFMERGE destKey exist": {
			name: "PFMERGE destKey exist",
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
			input: &cmd.DiceDBCmd{
				Cmd:  "PFMERGE",
				Args: []string{"EXISTING_DEST_KEY"},
			},
			output: EvalResponse{
				Result: OK,
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
			input: &cmd.DiceDBCmd{
				Cmd:  "PFMERGE",
				Args: []string{"EXISTING_DEST_KEY", "NON_EXISTING_SRC_KEY"},
			},
			output: EvalResponse{
				Result: OK,
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
			input: &cmd.DiceDBCmd{
				Cmd:  "PFMERGE",
				Args: []string{"EXISTING_DEST_KEY", "EXISTING_SRC_KEY"},
				InternalObjs: []*object.InternalObj{
					{
						Obj: &object.Obj{
							Value: hyperloglog.New(),
							Type:  object.ObjTypeHLL,
						},
					},
				},
			},
			output: EvalResponse{
				Result: OK,
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
				srcKey := "EXISTING_SRC_KEY1"
				srcValue := hyperloglog.New()
				value.Insert([]byte("SRC_VALUE"))
				srcKeyObj := &object.Obj{
					Value:          srcValue,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(srcKey, srcKeyObj)
				srcKey2 := "EXISTING_SRC_KEY2"
				srcValue2 := hyperloglog.New()
				value.Insert([]byte("SRC_VALUE"))
				srcKeyObj2 := &object.Obj{
					Value:          srcValue2,
					LastAccessedAt: uint32(time.Now().Unix()),
				}
				store.Put(srcKey2, srcKeyObj2)
			},
			input: &cmd.DiceDBCmd{
				Cmd:  "PFMERGE",
				Args: []string{"EXISTING_DEST_KEY", "EXISTING_SRC_KEY1", "EXISTING_SRC_KEY2"},
				InternalObjs: []*object.InternalObj{
					{
						Obj: &object.Obj{
							Value: hyperloglog.New(),
							Type:  object.ObjTypeHLL,
						},
					},
					{
						Obj: &object.Obj{
							Value: hyperloglog.New(),
							Type:  object.ObjTypeHLL,
						},
					},
				},
			},
			output: EvalResponse{
				Result: OK,
				Error:  nil,
			},
		},
	}

	runEvalTestsMultiShard(t, tests, evalPFMERGE, store)
}

func testEvalHGET(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"wrong number of args passed": {
			setup: func() {},
			input: nil,
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("HGET"),
			},
		},
		"only key passed": {
			setup: func() {},
			input: []string{"KEY"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("HGET"),
			},
		},
		"key doesn't exist": {
			setup: func() {},
			input: []string{"KEY", "field_name"},
			migratedOutput: EvalResponse{
				Result: NIL,
				Error:  nil,
			},
		},
		"key exists but field_name doesn't exist": {
			setup: func() {
				key := "KEY_MOCK"
				field := "mock_field_name"
				newMap := make(HashMap)
				newMap[field] = "mock_field_value"

				obj := &object.Obj{
					Type:           object.ObjTypeSSMap,
					Value:          newMap,
					LastAccessedAt: uint32(time.Now().Unix()),
				}

				store.Put(key, obj)
			},
			input: []string{"KEY_MOCK", "non_existent_key"},
			migratedOutput: EvalResponse{
				Result: NIL,
				Error:  nil,
			},
		},
		"both key and field_name exist": {
			setup: func() {
				key := "KEY_MOCK"
				field := "mock_field_name"
				newMap := make(HashMap)
				newMap[field] = "mock_field_value"

				obj := &object.Obj{
					Type:           object.ObjTypeSSMap,
					Value:          newMap,
					LastAccessedAt: uint32(time.Now().Unix()),
				}

				store.Put(key, obj)
			},
			input: []string{"KEY_MOCK", "mock_field_name"},
			migratedOutput: EvalResponse{
				Result: "mock_field_value",
				Error:  nil,
			},
		},
	}

	runMigratedEvalTests(t, tests, evalHGET, store)
}

func testEvalHGETALL(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"wrong number of args passed": {
			setup: func() {},
			input: []string{},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("HGETALL"),
			},
		},
		"key doesn't exist": {
			setup: func() {},
			input: []string{"NON_EXISTENT_KEY"},
			migratedOutput: EvalResponse{
				Result: []string{}, // Expect empty result as the key doesn't exist
				Error:  nil,
			},
		},
		"key exists but is empty": {
			setup: func() {
				key := "EMPTY_HASH"
				newMap := make(HashMap) // Empty hash map

				obj := &object.Obj{
					Type:           object.ObjTypeSSMap,
					Value:          newMap,
					LastAccessedAt: uint32(time.Now().Unix()),
				}

				store.Put(key, obj)
			},
			input: []string{"EMPTY_HASH"},
			migratedOutput: EvalResponse{
				Result: []string{}, // Expect empty result as the hash is empty
				Error:  nil,
			},
		},
		"key exists with multiple fields": {
			setup: func() {
				key := "HASH_WITH_FIELDS"
				newMap := make(HashMap)
				newMap["field1"] = "value1"
				newMap["field2"] = "value2"
				newMap["field3"] = "value3"

				obj := &object.Obj{
					Type:           object.ObjTypeSSMap,
					Value:          newMap,
					LastAccessedAt: uint32(time.Now().Unix()),
				}

				store.Put(key, obj)
			},
			input: []string{"HASH_WITH_FIELDS"},
			migratedOutput: EvalResponse{
				Result: []string{"field1", "value1", "field2", "value2", "field3", "value3"},
				Error:  nil,
			},
		},
	}

	runMigratedEvalTests(t, tests, evalHGETALL, store)
}

func testEvalHMGET(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"wrong number of args passed": {
			setup: func() {},
			input: nil,
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("HMGET"),
			},
		},
		"only key passed": {
			setup: func() {},
			input: []string{"KEY"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("HMGET"),
			},
		},
		"key doesn't exist": {
			setup: func() {},
			input: []string{"KEY", "field_name"},
			migratedOutput: EvalResponse{
				Result: []interface{}{nil},
				Error:  nil,
			},
		},
		"key exists but field_name doesn't exist": {
			setup: func() {
				key := "KEY_MOCK"
				field := "mock_field_name"
				newMap := make(HashMap)
				newMap[field] = "mock_field_value"

				obj := &object.Obj{
					Type:           object.ObjTypeSSMap,
					Value:          newMap,
					LastAccessedAt: uint32(time.Now().Unix()),
				}

				store.Put(key, obj)
			},
			input: []string{"KEY_MOCK", "non_existent_key"},
			migratedOutput: EvalResponse{
				Result: []interface{}{nil},
				Error:  nil,
			},
		},
		"both key and field_name exist": {
			setup: func() {
				key := "KEY_MOCK"
				field := "mock_field_name"
				newMap := make(HashMap)
				newMap[field] = "mock_field_value"

				obj := &object.Obj{
					Type:           object.ObjTypeSSMap,
					Value:          newMap,
					LastAccessedAt: uint32(time.Now().Unix()),
				}

				store.Put(key, obj)
			},
			input: []string{"KEY_MOCK", "mock_field_name"},
			migratedOutput: EvalResponse{
				Result: []interface{}{"mock_field_value"},
				Error:  nil,
			},
		},
		"some fields exist some do not": {
			setup: func() {
				key := "KEY_MOCK"
				newMap := HashMap{
					"field1": "value1",
					"field2": "value2",
				}
				obj := &object.Obj{
					Type:           object.ObjTypeSSMap,
					Value:          newMap,
					LastAccessedAt: uint32(time.Now().Unix()),
				}

				store.Put(key, obj)
			},
			input: []string{"KEY_MOCK", "field1", "field2", "field3", "field4"},
			migratedOutput: EvalResponse{
				Result: []interface{}{"value1", "value2", nil, nil}, // Use nil for non-existent fields
				Error:  nil,
			},
		},
	}

	runMigratedEvalTests(t, tests, evalHMGET, store)
}

func testEvalHVALS(t *testing.T, store *dstore.Store) {
	tests := []evalTestCase{
		{
			name:           "HVALS wrong number of args passed",
			setup:          nil,
			input:          nil,
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR wrong number of arguments for 'hvals' command")},
		},
		{
			name:           "HVALS key doesn't exists",
			setup:          nil,
			input:          []string{"NONEXISTENTHVALSKEY"},
			migratedOutput: EvalResponse{Result: EmptyArray, Error: nil},
		},
		{
			name: "HVALS key exists",
			setup: func() {
				key := "MOCK_KEY"
				field := "mock_field"
				newMap := make(HashMap)
				newMap[field] = "mock_value"

				obj := &object.Obj{
					Type:           object.ObjTypeSSMap,
					Value:          newMap,
					LastAccessedAt: uint32(time.Now().Unix()),
				}

				store.Put(key, obj)
			},
			input:          []string{"MOCK_KEY"},
			migratedOutput: EvalResponse{Result: []string{"mock_value"}, Error: nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.setup != nil {
				tt.setup()
			}

			response := evalHVALS(tt.input, store)

			// Handle comparison for byte slices
			if responseBytes, ok := response.Result.([]byte); ok && tt.migratedOutput.Result != nil {
				if expectedBytes, ok := tt.migratedOutput.Result.([]byte); ok {
					assert.True(t, bytes.Equal(responseBytes, expectedBytes), "expected and actual byte slices should be equal")
				}
			} else {
				switch e := tt.migratedOutput.Result.(type) {
				case []interface{}, []string:
					assert.ElementsMatch(t, e, response.Result)
				default:
					assert.Equal(t, tt.migratedOutput.Result, response.Result)
				}
			}

			if tt.migratedOutput.Error != nil {
				assert.EqualError(t, response.Error, tt.migratedOutput.Error.Error())
			} else {
				assert.NoError(t, response.Error)
			}
		})
	}
}

func testEvalHSTRLEN(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"wrong number of args passed": {
			setup:          func() {},
			input:          nil,
			migratedOutput: EvalResponse{Result: nil, Error: diceerrors.ErrWrongArgumentCount("HSTRLEN")},
		},
		"only key passed": {
			setup:          func() {},
			input:          []string{"KEY"},
			migratedOutput: EvalResponse{Result: nil, Error: diceerrors.ErrWrongArgumentCount("HSTRLEN")},
		},
		"key doesn't exist": {
			setup:          func() {},
			input:          []string{"KEY", "field_name"},
			migratedOutput: EvalResponse{Result: IntegerZero, Error: nil},
		},
		"key exists but field_name doesn't exists": {
			setup: func() {
				key := "KEY_MOCK"
				field := "mock_field_name"
				newMap := make(HashMap)
				newMap[field] = "mock_field_value"

				obj := &object.Obj{
					Type:           object.ObjTypeSSMap,
					Value:          newMap,
					LastAccessedAt: uint32(time.Now().Unix()),
				}

				store.Put(key, obj)
			},
			input:          []string{"KEY_MOCK", "non_existent_key"},
			migratedOutput: EvalResponse{Result: IntegerZero, Error: nil},
		},
		"both key and field_name exists": {
			setup: func() {
				key := "KEY_MOCK"
				field := "mock_field_name"
				newMap := make(HashMap)
				newMap[field] = "HelloWorld"

				obj := &object.Obj{
					Type:           object.ObjTypeSSMap,
					Value:          newMap,
					LastAccessedAt: uint32(time.Now().Unix()),
				}

				store.Put(key, obj)
			},
			input:          []string{"KEY_MOCK", "mock_field_name"},
			migratedOutput: EvalResponse{Result: 10, Error: nil},
		},
	}

	runMigratedEvalTests(t, tests, evalHSTRLEN, store)
}

func testEvalHEXISTS(t *testing.T, store *dstore.Store) {
	tests := []evalTestCase{
		{
			name:           "HEXISTS wrong number of args passed",
			setup:          nil,
			input:          nil,
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR wrong number of arguments for 'hexists' command")},
		},
		{
			name:           "HEXISTS only key passed",
			setup:          nil,
			input:          []string{"KEY"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR wrong number of arguments for 'hexists' command")},
		},
		{
			name:           "HEXISTS key doesn't exist",
			setup:          nil,
			input:          []string{"KEY", "field_name"},
			migratedOutput: EvalResponse{Result: IntegerZero, Error: nil},
		},
		{
			name: "HEXISTS key exists but field_name doesn't exists",
			setup: func() {
				key := "KEY_MOCK"
				field := "mock_field_name"
				newMap := make(HashMap)
				newMap[field] = "mock_field_value"

				obj := &object.Obj{
					Type:           object.ObjTypeSSMap,
					Value:          newMap,
					LastAccessedAt: uint32(time.Now().Unix()),
				}

				store.Put(key, obj)
			},
			input:          []string{"KEY_MOCK", "non_existent_key"},
			migratedOutput: EvalResponse{Result: IntegerZero, Error: nil},
		},
		{
			name: "HEXISTS both key and field_name exists",
			setup: func() {
				key := "KEY_MOCK"
				field := "mock_field_name"
				newMap := make(HashMap)
				newMap[field] = "HelloWorld"

				obj := &object.Obj{
					Type:           object.ObjTypeSSMap,
					Value:          newMap,
					LastAccessedAt: uint32(time.Now().Unix()),
				}

				store.Put(key, obj)
			},
			input:          []string{"KEY_MOCK", "mock_field_name"},
			migratedOutput: EvalResponse{Result: IntegerOne, Error: nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.setup != nil {
				tt.setup()
			}

			response := evalHEXISTS(tt.input, store)

			// Handle comparison for byte slices
			if responseBytes, ok := response.Result.([]byte); ok && tt.migratedOutput.Result != nil {
				// If has result
				if expectedBytes, ok := tt.migratedOutput.Result.([]byte); ok {
					// fmt.Printf("%v | %v\n", responseBytes, expectedBytes)
					assert.True(t, bytes.Equal(responseBytes, expectedBytes), "expected and actual byte slices should be equal")
				}
			} else {
				// If has error
				assert.Equal(t, tt.migratedOutput.Result, response.Result)
			}

			if tt.migratedOutput.Error != nil {
				assert.EqualError(t, response.Error, tt.migratedOutput.Error.Error())
			} else {
				assert.NoError(t, response.Error)
			}
		})
	}
}

func testEvalHDEL(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"HDEL with wrong number of args": {
			input: []string{"key"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("HDEL"),
			},
		},
		"HDEL with key does not exist": {
			input: []string{"nonexistent", "field"},
			migratedOutput: EvalResponse{
				Result: int64(0),
				Error:  nil,
			},
		},
		"HDEL with key exists but not a hash": {
			setup: func() {
				evalSET([]string{"string_key", "string_value"}, store)
			},
			input: []string{"string_key", "field"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongTypeOperation,
			},
		},
		"HDEL with delete existing fields": {
			setup: func() {
				evalHSET([]string{"hash_key", "field1", "value1", "field2", "value2"}, store)
			},
			input: []string{"hash_key", "field1", "field2", "nonexistent"},
			migratedOutput: EvalResponse{
				Result: int64(2),
				Error:  nil,
			},
		},
		"HDEL with delete non-existing fields": {
			setup: func() {
				evalHSET([]string{"hash_key", "field1", "value1"}, store)
			},
			input: []string{"hash_key", "nonexistent1", "nonexistent2"},
			migratedOutput: EvalResponse{
				Result: int64(0),
				Error:  nil,
			},
		},
	}
	runMigratedEvalTests(t, tests, evalHDEL, store)
}

func testEvalHSCAN(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"HSCAN with wrong number of args": {
			input:          []string{"key"},
			migratedOutput: EvalResponse{Result: nil, Error: diceerrors.ErrWrongArgumentCount("HSCAN")},
		},
		"HSCAN with key does not exist": {
			input:          []string{"NONEXISTENT_KEY", "0"},
			migratedOutput: EvalResponse{Result: []interface{}{"0", []string{}}, Error: nil},
		},
		"HSCAN with key exists but not a hash": {
			setup: func() {
				evalSET([]string{"string_key", "string_value"}, store)
			},
			input:          []string{"string_key", "0"},
			migratedOutput: EvalResponse{Result: nil, Error: diceerrors.ErrWrongTypeOperation},
		},
		"HSCAN with valid key and cursor": {
			setup: func() {
				evalHSET([]string{"hash_key", "field1", "value1", "field2", "value2"}, store)
			},
			input:          []string{"hash_key", "0"},
			migratedOutput: EvalResponse{Result: []interface{}{"0", []string{"field1", "value1", "field2", "value2"}}, Error: nil},
		},
		"HSCAN with cursor at the end": {
			setup: func() {
				evalHSET([]string{"hash_key", "field1", "value1", "field2", "value2"}, store)
			},
			input:          []string{"hash_key", "2"},
			migratedOutput: EvalResponse{Result: []interface{}{"0", []string{}}, Error: nil},
		},
		"HSCAN with cursor at the beginning": {
			setup: func() {
				evalHSET([]string{"hash_key", "field1", "value1", "field2", "value2"}, store)
			},
			input:          []string{"hash_key", "0"},
			migratedOutput: EvalResponse{Result: []interface{}{"0", []string{"field1", "value1", "field2", "value2"}}, Error: nil},
		},
		"HSCAN with cursor in the middle": {
			setup: func() {
				evalHSET([]string{"hash_key", "field1", "value1", "field2", "value2"}, store)
			},
			input:          []string{"hash_key", "1"},
			migratedOutput: EvalResponse{Result: []interface{}{"0", []string{"field2", "value2"}}, Error: nil},
		},
		"HSCAN with MATCH argument": {
			setup: func() {
				evalHSET([]string{"hash_key", "field1", "value1", "field2", "value2", "field3", "value3"}, store)
			},
			input:          []string{"hash_key", "0", "MATCH", "field[12]*"},
			migratedOutput: EvalResponse{Result: []interface{}{"0", []string{"field1", "value1", "field2", "value2"}}, Error: nil},
		},
		"HSCAN with COUNT argument": {
			setup: func() {
				evalHSET([]string{"hash_key", "field1", "value1", "field2", "value2", "field3", "value3"}, store)
			},
			input:          []string{"hash_key", "0", "COUNT", "2"},
			migratedOutput: EvalResponse{Result: []interface{}{"2", []string{"field1", "value1", "field2", "value2"}}, Error: nil},
		},
		"HSCAN with MATCH and COUNT arguments": {
			setup: func() {
				evalHSET([]string{"hash_key", "field1", "value1", "field2", "value2", "field3", "value3", "field4", "value4"}, store)
			},
			input:          []string{"hash_key", "0", "MATCH", "field[13]*", "COUNT", "1"},
			migratedOutput: EvalResponse{Result: []interface{}{"1", []string{"field1", "value1"}}, Error: nil},
		},
		"HSCAN with invalid MATCH pattern": {
			setup: func() {
				evalHSET([]string{"hash_key", "field1", "value1", "field2", "value2"}, store)
			},
			input:          []string{"hash_key", "0", "MATCH", "[invalid"},
			migratedOutput: EvalResponse{Result: nil, Error: diceerrors.ErrGeneral("Invalid glob pattern: unexpected end of input")},
		},
		"HSCAN with invalid COUNT value": {
			setup: func() {
				evalHSET([]string{"hash_key", "field1", "value1", "field2", "value2"}, store)
			},
			input:          []string{"hash_key", "0", "COUNT", "invalid"},
			migratedOutput: EvalResponse{Result: nil, Error: diceerrors.ErrIntegerOutOfRange},
		},
	}

	runMigratedEvalTests(t, tests, evalHSCAN, store)
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
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
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
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
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
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
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
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
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
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
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
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
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
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
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
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
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
					assert.Equal(t, slice, response.Result)
				} else {
					assert.Equal(t, tt.migratedOutput.Result, response.Result)
				}
			}

			if tt.migratedOutput.Error != nil {
				assert.EqualError(t, response.Error, tt.migratedOutput.Error.Error())
			} else {
				assert.NoError(t, response.Error)
			}
		})
	}
}

func testEvalLPUSH(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"nil value": {
			input:          nil,
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR wrong number of arguments for 'lpush' command")},
		},
		"empty args": {
			input:          []string{},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR wrong number of arguments for 'lpush' command")},
		},
		"wrong number of args": {
			input:          []string{"KEY1"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR wrong number of arguments for 'lpush' command")},
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
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR wrong number of arguments for 'rpush' command")},
		},
		"empty args": {
			input:          []string{},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR wrong number of arguments for 'rpush' command")},
		},
		"wrong number of args": {
			input:          []string{"KEY1"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR wrong number of arguments for 'rpush' command")},
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
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR wrong number of arguments for 'lpop' command")},
		},
		"empty args": {
			input:          []string{},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR wrong number of arguments for 'lpop' command")},
		},
		"more than 2 args": {
			input:          []string{"k", "2", "3"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR wrong number of arguments for 'lpop' command")},
		},
		"non-integer count": {
			input:          []string{"k", "abc"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR value is not an integer or a float")},
		},
		"negative count": {
			input:          []string{"k", "-1"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR value is not an integer or out of range")},
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
			migratedOutput: EvalResponse{Result: NIL, Error: nil},
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
		"pop one element": {
			setup: func() {
				evalRPUSH([]string{"k", "v1", "v2", "v3", "v4"}, store)
			},
			input:          []string{"k"},
			migratedOutput: EvalResponse{Result: "v1", Error: nil},
		},
		"pop two elements": {
			setup: func() {
				evalRPUSH([]string{"k", "v1", "v2", "v3", "v4"}, store)
			},
			input:          []string{"k", "2"},
			migratedOutput: EvalResponse{Result: []string{"v1", "v2"}, Error: nil},
		},
		"pop more elements than available": {
			setup: func() {
				evalRPUSH([]string{"k", "v1", "v2"}, store)
			},
			input:          []string{"k", "5"},
			migratedOutput: EvalResponse{Result: []string{"v1", "v2"}, Error: nil},
		},
		"pop 0 elements": {
			setup: func() {
				evalRPUSH([]string{"k", "v1", "v2"}, store)
			},
			input:          []string{"k", "0"},
			migratedOutput: EvalResponse{Result: NIL, Error: nil},
		},
	}
	runMigratedEvalTests(t, tests, evalLPOP, store)
}

func testEvalRPOP(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"nil value": {
			input:          nil,
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR wrong number of arguments for 'rpop' command")},
		},
		"empty args": {
			input:          []string{},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR wrong number of arguments for 'rpop' command")},
		},
		"wrong number of args": {
			input:          []string{"KEY1", "KEY2"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR wrong number of arguments for 'rpop' command")},
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
			migratedOutput: EvalResponse{Result: NIL, Error: nil},
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
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR wrong number of arguments for 'llen' command")},
		},
		"empty args": {
			input:          []string{},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR wrong number of arguments for 'llen' command")},
		},
		"wrong number of args": {
			input:          []string{"KEY1", "KEY2"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR wrong number of arguments for 'llen' command")},
		},
		"key does not exist": {
			input:          []string{"NONEXISTENT_KEY"},
			migratedOutput: EvalResponse{Result: IntegerZero, Error: nil},
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
		"JSON.NUMINCRBY : incr on numeric field": {
			name: "JSON.NUMINCRBY : incr on numeric field",
			setup: func() {
				key := "number"
				value := "{\"a\": 2}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input: []string{"number", "$.a", "3"},
			migratedOutput: EvalResponse{
				Result: "[5]",
				Error:  nil,
			},
		},

		"JSON.NUMINCRBY : incr on float field": {
			name: "JSON.NUMINCRBY : incr on float field",
			setup: func() {
				key := "number"
				value := "{\"a\": 2.5}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input: []string{"number", "$.a", "1.5"},
			migratedOutput: EvalResponse{
				Result: "[4.0]",
				Error:  nil,
			},
		},

		"JSON.NUMINCRBY : incr on multiple fields": {
			name: "JSON.NUMINCRBY : incr on multiple fields",
			setup: func() {
				key := "number"
				value := "{\"a\": 2, \"b\": 10, \"c\": [15, {\"d\": 20}]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input: []string{"number", "$..*", "5"},
			migratedOutput: EvalResponse{
				Result: "[25,20,null,7,15,null]",
				Error:  nil,
			},
			newValidator: func(output interface{}) {
				outPutString, ok := output.(string)
				if !ok {
					t.Errorf("expected output to be of type string, got %T", output)
					return
				}

				startIndex := strings.Index(outPutString, "[")
				endIndex := strings.Index(outPutString, "]")
				arrayString := outPutString[startIndex+1 : endIndex]
				arr := strings.Split(arrayString, ",")
				assert.ElementsMatch(t, arr, []string{"25", "20", "7", "15", "null", "null"})
			},
		},

		"JSON.NUMINCRBY : incr on array element": {
			name: "JSON.NUMINCRBY : incr on array element",
			setup: func() {
				key := "number"
				value := "{\"a\": [1, 2, 3]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input: []string{"number", "$.a[1]", "5"},
			migratedOutput: EvalResponse{
				Result: "[7]",
				Error:  nil,
			},
		},
		"JSON.NUMINCRBY : incr on non-existent field": {
			name: "JSON.NUMINCRBY : incr on non-existent field",
			setup: func() {
				key := "number"
				value := "{\"a\": 2}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input: []string{"number", "$.b", "3"},
			migratedOutput: EvalResponse{
				Result: "[]",
				Error:  nil,
			},
		},
		"JSON.NUMINCRBY : incr with mixed fields": {
			name: "JSON.NUMINCRBY : incr with mixed fields",
			setup: func() {
				key := "number"
				value := "{\"a\": 5, \"b\": \"not a number\", \"c\": [1, 2]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input: []string{"number", "$..*", "2"},
			migratedOutput: EvalResponse{
				Result: "[3,4,7,null,null]",
				Error:  nil,
			},
			newValidator: func(output interface{}) {
				// Ensure that the output is a string before proceeding
				outPutString, ok := output.(string)
				if !ok {
					t.Errorf("expected output to be of type string, got %T", output)
					return
				}

				// Find the positions of the first "[" and the last "]"
				startIndex := strings.Index(outPutString, "[")
				endIndex := strings.LastIndex(outPutString, "]")

				// Check if both brackets are found
				if startIndex == -1 || endIndex == -1 || startIndex >= endIndex {
					t.Errorf("invalid array format in output string: %s", outPutString)
					return
				}

				// Extract the array substring between the brackets
				arrayString := outPutString[startIndex+1 : endIndex]

				// Split the array string by commas and trim spaces around elements
				arr := strings.Split(arrayString, ",")
				for i := range arr {
					arr[i] = strings.TrimSpace(arr[i])
				}

				// Validate that the array contains the expected elements
				assert.ElementsMatch(t, arr, []string{"3", "4", "7", "null", "null"})
			},
		},

		"JSON.NUMINCRBY : incr on nested fields": {
			name: "JSON.NUMINCRBY : incr on nested fields",
			setup: func() {
				key := "number"
				value := "{\"a\": {\"b\": {\"c\": 10}}}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input: []string{"number", "$..c", "5"},
			migratedOutput: EvalResponse{
				Result: "[15]",
				Error:  nil,
			},
		},
	}
	runMigratedEvalTests(t, tests, evalJSONNUMINCRBY, store)
}

func runEvalTests(t *testing.T, tests map[string]evalTestCase, evalFunc func([]string, *dstore.Store) []byte, store *dstore.Store) {
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			store = setupTest(store)

			if tc.setup != nil {
				tc.setup()
			}

			output := evalFunc(tc.input, store)

			assert.Equal(t, string(tc.output), string(output))
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
				assert.EqualError(t, output.Error, tc.migratedOutput.Error.Error())
				return
			}

			// Handle comparison for byte slices and string slices
			// TODO: Make this generic so that all kind of slices can be handled
			if b, ok := output.Result.([]byte); ok && tc.migratedOutput.Result != nil {
				if expectedBytes, ok := tc.migratedOutput.Result.([]byte); ok {
					// fmt.Println(string(b))
					// fmt.Println(string(expectedBytes))
					assert.True(t, bytes.Equal(b, expectedBytes), "expected and actual byte slices should be equal")
				}
			} else if a, ok := output.Result.([]string); ok && tc.migratedOutput.Result != nil {
				if expectedStringSlice, ok := tc.migratedOutput.Result.([]string); ok {
					assert.ElementsMatch(t, a, expectedStringSlice)
				}
			} else {
				assert.Equal(t, tc.migratedOutput.Result, output.Result)
			}

			assert.NoError(t, output.Error)
		})
	}
}

func runEvalTestsMultiShard(t *testing.T, tests map[string]evalMultiShardTestCase, evalFunc func(*cmd.DiceDBCmd, *dstore.Store) *EvalResponse, store *dstore.Store) {
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			store = setupTest(store)
			if tc.setup != nil {
				tc.setup()
			}

			output := evalFunc(tc.input, store)
			if tc.output.Error != nil {
				assert.Equal(t, tc.output.Error, output.Error)
			}

			if tc.output.Result != nil {
				assert.Equal(t, tc.output.Result, output.Result)
			}
		})
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
			setup: func() {},
			input: nil,
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("HSET"),
			},
		},
		"only key passed": {
			setup: func() {},
			input: []string{"key"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("HSET"),
			},
		},
		"only key and field_name passed": {
			setup: func() {},
			input: []string{"KEY", "field_name"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("HSET"),
			},
		},
		"key, field and value passed": {
			setup: func() {},
			input: []string{"KEY1", "field_name", "value"},
			migratedOutput: EvalResponse{
				Result: int64(1),
				Error:  nil,
			},
		},
		"key, field and value updated": {
			setup: func() {},
			input: []string{"KEY1", "field_name", "value_new"},
			migratedOutput: EvalResponse{
				Result: int64(1),
				Error:  nil,
			},
		},
		"new set of key, field and value added": {
			setup: func() {},
			input: []string{"KEY2", "field_name_new", "value_new_new"},
			migratedOutput: EvalResponse{
				Result: int64(1),
				Error:  nil,
			},
		},
		"apply with duplicate key, field and value names": {
			setup: func() {
				key := "KEY_MOCK"
				field := "mock_field_name"
				newMap := make(HashMap)
				newMap[field] = "mock_field_value"

				obj := &object.Obj{
					Type:           object.ObjTypeSSMap,
					Value:          newMap,
					LastAccessedAt: uint32(time.Now().Unix()),
				}

				store.Put(key, obj)
			},
			input: []string{"KEY_MOCK", "mock_field_name", "mock_field_value"},
			migratedOutput: EvalResponse{
				Result: int64(0),
				Error:  nil,
			},
		},
		"same key -> update value, add new field and value": {
			setup: func() {
				key := "KEY_MOCK"
				field := "mock_field_name"
				mockValue := "mock_field_value"
				newMap := make(HashMap)
				newMap[field] = mockValue

				obj := &object.Obj{
					Type:           object.ObjTypeSSMap,
					Value:          newMap,
					LastAccessedAt: uint32(time.Now().Unix()),
				}

				store.Put(key, obj)
			},
			input: []string{
				"KEY_MOCK",
				"mock_field_name",
				"mock_field_value_new",
				"mock_field_name_new",
				"mock_value_new",
			},
			migratedOutput: EvalResponse{
				Result: int64(1),
				Error:  nil,
			},
		},
	}
	runMigratedEvalTests(t, tests, evalHSET, store)
}

func testEvalHMSET(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"wrong number of args passed": {
			setup: func() {},
			input: nil,
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("HMSET"),
			},
		},
		"only key passed": {
			setup: func() {},
			input: []string{"key"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("HMSET"),
			},
		},
		"only key and field_name passed": {
			setup: func() {},
			input: []string{"KEY", "field_name"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("HMSET"),
			},
		},
		"key, field and value passed": {
			setup: func() {},
			input: []string{"KEY1", "field_name", "value"},
			migratedOutput: EvalResponse{
				Result: OK,
				Error:  nil,
			},
		},
		"key, field and value updated": {
			setup: func() {},
			input: []string{"KEY1", "field_name", "value_new"},
			migratedOutput: EvalResponse{
				Result: OK,
				Error:  nil,
			},
		},
		"new set of key, field and value added": {
			setup: func() {},
			input: []string{"KEY2", "field_name_new", "value_new_new"},
			migratedOutput: EvalResponse{
				Result: OK,
				Error:  nil,
			},
		},
		"apply with duplicate key, field and value names": {
			setup: func() {
				key := "KEY_MOCK"
				field := "mock_field_name"
				newMap := make(HashMap)
				newMap[field] = "mock_field_value"

				obj := &object.Obj{
					Type:           object.ObjTypeSSMap,
					Value:          newMap,
					LastAccessedAt: uint32(time.Now().Unix()),
				}

				store.Put(key, obj)
			},
			input: []string{"KEY_MOCK", "mock_field_name", "mock_field_value"},
			migratedOutput: EvalResponse{
				Result: OK,
				Error:  nil,
			},
		},
		"same key -> update value, add new field and value": {
			setup: func() {
				key := "KEY_MOCK"
				field := "mock_field_name"
				mockValue := "mock_field_value"
				newMap := make(HashMap)
				newMap[field] = mockValue

				obj := &object.Obj{
					Type:           object.ObjTypeSSMap,
					Value:          newMap,
					LastAccessedAt: uint32(time.Now().Unix()),
				}

				store.Put(key, obj)
			},
			input: []string{
				"KEY_MOCK",
				"mock_field_name",
				"mock_field_value_new",
				"mock_field_name_new",
				"mock_value_new",
			},
			migratedOutput: EvalResponse{
				Result: OK,
				Error:  nil,
			},
		},
	}
	runMigratedEvalTests(t, tests, evalHMSET, store)
}

func testEvalHKEYS(t *testing.T, store *dstore.Store) {
	tests := []evalTestCase{
		{
			name:           "HKEYS wrong number of args passed",
			setup:          nil,
			input:          nil,
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR wrong number of arguments for 'hkeys' command")},
		},
		{
			name:           "HKEYS key doesn't exist",
			setup:          nil,
			input:          []string{"KEY"},
			migratedOutput: EvalResponse{Result: EmptyArray, Error: nil},
		},
		{
			name: "HKEYS key exists but not a hash",
			setup: func() {
				evalSET([]string{"string_key", "string_value"}, store)
			},
			input:          []string{"string_key"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR -WRONGTYPE Operation against a key holding the wrong kind of value")},
		},
		{
			name: "HKEYS key exists and is a hash",
			setup: func() {
				key := "KEY_MOCK"
				field1 := "mock_field_name"
				newMap := make(HashMap)
				newMap[field1] = "HelloWorld"

				obj := &object.Obj{
					Type:           object.ObjTypeSSMap,
					Value:          newMap,
					LastAccessedAt: uint32(time.Now().Unix()),
				}

				store.Put(key, obj)
			},
			input:          []string{"KEY_MOCK"},
			migratedOutput: EvalResponse{Result: []string{"mock_field_name"}, Error: nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.setup != nil {
				tt.setup()
			}

			response := evalHKEYS(tt.input, store)

			// fmt.Printf("EvalReponse: %v\n", response)

			// Handle comparison for byte slices
			if responseBytes, ok := response.Result.([]byte); ok && tt.migratedOutput.Result != nil {
				if expectedBytes, ok := tt.migratedOutput.Result.([]byte); ok {
					// fmt.Printf("G: %v | %v\n", responseBytes, expectedBytes)
					assert.True(t, bytes.Equal(responseBytes, expectedBytes), "expected and actual byte slices should be equal")
				}
			} else {
				// fmt.Printf("G1: %v | %v\n", response.Result, tt.migratedOutput.Result)
				switch e := tt.migratedOutput.Result.(type) {
				case []interface{}, []string:
					assert.ElementsMatch(t, e, response.Result)
				default:
					assert.Equal(t, tt.migratedOutput.Result, response.Result)
				}
			}

			if tt.migratedOutput.Error != nil {
				// fmt.Printf("E: %v | %v\n", response.Error, tt.migratedOutput.Error.Error())
				assert.EqualError(t, response.Error, tt.migratedOutput.Error.Error())
			} else {
				assert.NoError(t, response.Error)
			}
		})
	}
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
			setup:          func() {},
			migratedOutput: EvalResponse{Result: nil, Error: diceerrors.ErrWrongArgumentCount("JSON.DEBUG")},
		},

		"wrong subcommand passed": {
			setup:          func() {},
			input:          []string{"WRONG_SUBCOMMAND"},
			migratedOutput: EvalResponse{Result: nil, Error: diceerrors.ErrGeneral("unknown subcommand - try `JSON.DEBUG HELP`")},
		},

		// help subcommand tests
		"help no args": {
			setup:          func() {},
			input:          []string{"HELP"},
			migratedOutput: EvalResponse{Result: []string{"MEMORY <key> [path] - reports memory usage", "HELP                - this message"}, Error: nil},
		},

		"help with args": {
			setup:          func() {},
			input:          []string{"HELP", "EXTRA_ARG"},
			migratedOutput: EvalResponse{Result: []string{"MEMORY <key> [path] - reports memory usage", "HELP                - this message"}, Error: nil},
		},

		// memory subcommand tests
		"memory without args": {
			setup:          func() {},
			input:          []string{"MEMORY"},
			migratedOutput: EvalResponse{Result: nil, Error: diceerrors.ErrWrongArgumentCount("JSON.DEBUG")},
		},

		"memory nonexistent key": {
			setup:          func() {},
			input:          []string{"MEMORY", "NONEXISTENT_KEY"},
			migratedOutput: EvalResponse{Result: IntegerZero, Error: nil},
		},

		// memory subcommand tests for existing key
		"no path": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"a\": 1}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input:          []string{"MEMORY", "EXISTING_KEY"},
			migratedOutput: EvalResponse{Result: 72, Error: nil},
		},

		"root path": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"a\": 1}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input:          []string{"MEMORY", "EXISTING_KEY", "$"},
			migratedOutput: EvalResponse{Result: 72, Error: nil},
		},

		"invalid path": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"a\": 1}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input:          []string{"MEMORY", "EXISTING_KEY", "INVALID_PATH"},
			migratedOutput: EvalResponse{Result: nil, Error: diceerrors.ErrFormatted("Path '$.%v' does not exist", "INVALID_PATH")},
		},

		"valid path": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"a\": 1, \"b\": 2}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input:          []string{"MEMORY", "EXISTING_KEY", "$.a"},
			migratedOutput: EvalResponse{Result: []interface{}{16}, Error: nil},
		},

		// only the first path is picked whether it's valid or not for an object json
		// memory can be fetched only for one path in a command for an object json
		"multiple paths for object json": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "{\"a\": 1, \"b\": \"dice\"}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input:          []string{"MEMORY", "EXISTING_KEY", "$.a", "$.b"},
			migratedOutput: EvalResponse{Result: []interface{}{16}, Error: nil},
		},

		"single index path for array json": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "[\"roll\", \"the\", \"dices\"]"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input:          []string{"MEMORY", "EXISTING_KEY", "$[1]"},
			migratedOutput: EvalResponse{Result: []interface{}{19}, Error: nil},
		},

		"multiple index paths for array json": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "[\"roll\", \"the\", \"dices\"]"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input:          []string{"MEMORY", "EXISTING_KEY", "$[1,2]"},
			migratedOutput: EvalResponse{Result: []interface{}{19, 21}, Error: nil},
		},

		"index path out of range for array json": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "[\"roll\", \"the\", \"dices\"]"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input:          []string{"MEMORY", "EXISTING_KEY", "$[4]"},
			migratedOutput: EvalResponse{Result: EmptyArray, Error: nil},
		},

		"multiple valid and invalid index paths": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "[\"roll\", \"the\", \"dices\"]"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input:          []string{"MEMORY", "EXISTING_KEY", "$[1,2,4]"},
			migratedOutput: EvalResponse{Result: []interface{}{19, 21}, Error: nil},
		},

		"negative index path": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "[\"roll\", \"the\", \"dices\"]"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input:          []string{"MEMORY", "EXISTING_KEY", "$[-1]"},
			migratedOutput: EvalResponse{Result: []interface{}{21}, Error: nil},
		},

		"multiple negative index paths": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "[\"roll\", \"the\", \"dices\"]"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input:          []string{"MEMORY", "EXISTING_KEY", "$[-1,-2]"},
			migratedOutput: EvalResponse{Result: []interface{}{21, 19}, Error: nil},
		},

		"negative index path out of bound": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "[\"roll\", \"the\", \"dices\"]"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input:          []string{"MEMORY", "EXISTING_KEY", "$[-4]"},
			migratedOutput: EvalResponse{Result: nil, Error: diceerrors.ErrFormatted("Path '$.%v' does not exist", "$[-4]")},
		},

		"all paths with asterix for array json": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "[\"roll\", \"the\", \"dices\"]"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input:          []string{"MEMORY", "EXISTING_KEY", "$[*]"},
			migratedOutput: EvalResponse{Result: []interface{}{20, 19, 21}, Error: nil},
		},

		"all paths with semicolon for array json": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "[\"roll\", \"the\", \"dices\"]"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input:          []string{"MEMORY", "EXISTING_KEY", "$[:]"},
			migratedOutput: EvalResponse{Result: []interface{}{20, 19, 21}, Error: nil},
		},

		"array json with mixed types": {
			setup: func() {
				key := "EXISTING_KEY"
				value := "[2, 3.5, true, null, \"dice\", {}, [], {\"a\": 1, \"b\": 2}, [7, 8, 0]]"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input:          []string{"MEMORY", "EXISTING_KEY", "$[:]"},
			migratedOutput: EvalResponse{Result: []interface{}{16, 16, 16, 16, 20, 16, 16, 82, 64}, Error: nil},
		},
	}

	runMigratedEvalTests(t, tests, evalJSONDebug, store)
}

func testEvalHLEN(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"HLEN wrong number of args": {
			input:          []string{},
			migratedOutput: EvalResponse{Result: nil, Error: diceerrors.ErrWrongArgumentCount("HLEN")},
		},
		"HLEN non-existent key": {
			input:          []string{"nonexistent_key"},
			migratedOutput: EvalResponse{Result: IntegerZero, Error: nil},
		},
		"HLEN key exists but not a hash": {
			setup: func() {
				evalSET([]string{"string_key", "string_value"}, store)
			},
			input:          []string{"string_key"},
			migratedOutput: EvalResponse{Result: nil, Error: diceerrors.ErrWrongTypeOperation},
		},
		"HLEN empty hash": {
			setup:          func() {},
			input:          []string{"empty_hash"},
			migratedOutput: EvalResponse{Result: IntegerZero, Error: nil},
		},
		"HLEN hash with elements": {
			setup: func() {
				evalHSET([]string{"hash_key", "field1", "value1", "field2", "value2", "field3", "value3"}, store)
			},
			input:          []string{"hash_key"},
			migratedOutput: EvalResponse{Result: 3, Error: nil},
		},
	}

	runMigratedEvalTests(t, tests, evalHLEN, store)
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

func testEvalJSONARRPOP(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"wrong number of args passed": {
			setup:  func() {},
			input:  nil,
			output: []byte("-ERR wrong number of arguments for 'json.arrpop' command\r\n"),
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("JSON.ARRPOP"),
			},
		},
		"key does not exist": {
			setup:  func() {},
			input:  []string{"NOTEXISTANT_KEY"},
			output: []byte("-ERR could not perform this operation on a key that doesn't exist\r\n"),
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrKeyNotFound,
			},
		},
		"empty array at root path": {
			setup: func() {
				key := "MOCK_KEY"
				value := "[]"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input:  []string{"MOCK_KEY"},
			output: []byte("-ERR Path '$' does not exist or not an array\r\n"),
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongTypeOperation,
			},
		},
		"empty array at nested path": {
			setup: func() {
				key := "MOCK_KEY"
				value := "{\"a\": 1, \"b\": []}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input:  []string{"MOCK_KEY", "$.b"},
			output: []byte("*1\r\n$-1\r\n"),
			migratedOutput: EvalResponse{
				Result: []interface{}{NIL},
				Error:  nil,
			},
		},
		"all paths with asterix": {
			setup: func() {
				key := "MOCK_KEY"
				value := "{\"a\": 1, \"b\": []}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input:  []string{"MOCK_KEY", "$.*"},
			output: []byte("*2\r\n$-1\r\n$-1\r\n"),
			migratedOutput: EvalResponse{
				Result: []interface{}{NIL, NIL},
				Error:  nil,
			},
		},
		"array root path no index": {
			setup: func() {
				key := "MOCK_KEY"
				value := "[0, 1, 2, 3, 4, 5]"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input:  []string{"MOCK_KEY"},
			output: []byte(":5\r\n"),
			migratedOutput: EvalResponse{
				Result: float64(5),
				Error:  nil,
			},
		},
		"array root path valid positive index": {
			setup: func() {
				key := "MOCK_KEY"
				value := "[0, 1, 2, 3, 4, 5]"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input:  []string{"MOCK_KEY", "$", "2"},
			output: []byte(":2\r\n"),
			migratedOutput: EvalResponse{
				Result: float64(2),
				Error:  nil,
			},
		},
		"array root path out of bound positive index": {
			setup: func() {
				key := "MOCK_KEY"
				value := "[0, 1, 2, 3, 4, 5]"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input:  []string{"MOCK_KEY", "$", "10"},
			output: []byte(":5\r\n"),
			migratedOutput: EvalResponse{
				Result: float64(5),
				Error:  nil,
			},
		},
		"array root path valid negative index": {
			setup: func() {
				key := "MOCK_KEY"
				value := "[0, 1, 2, 3, 4, 5]"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input:  []string{"MOCK_KEY", "$", "-2"},
			output: []byte(":4\r\n"),
			migratedOutput: EvalResponse{
				Result: float64(4),
				Error:  nil,
			},
		},
		"array root path out of bound negative index": {
			setup: func() {
				key := "MOCK_KEY"
				value := "[0, 1, 2, 3, 4, 5]"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input:  []string{"MOCK_KEY", "$", "-10"},
			output: []byte(":0\r\n"),
			migratedOutput: EvalResponse{
				Result: float64(0),
				Error:  nil,
			},
		},
		"array at root path updated correctly": {
			setup: func() {
				key := "MOCK_KEY"
				value := "[0, 1, 2, 3, 4, 5]"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input:  []string{"MOCK_KEY", "$", "2"},
			output: []byte(":0\r\n"),
			newValidator: func(output interface{}) {
				key := "MOCK_KEY"
				obj := store.Get(key)
				want := []interface{}{float64(0), float64(1), float64(3), float64(4), float64(5)}
				equal := reflect.DeepEqual(obj.Value, want)
				assert.Equal(t, equal, true)
			},
			migratedOutput: EvalResponse{
				Result: float64(2),
				Error:  nil,
			},
		},
		"nested array updated correctly": {
			setup: func() {
				key := "MOCK_KEY"
				value := "{\"a\": 2, \"b\": [0, 1, 2, 3, 4, 5]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input:  []string{"MOCK_KEY", "$.b", "2"},
			output: []byte("*1\r\n:2\r\n"),
			newValidator: func(output interface{}) {
				key := "MOCK_KEY"
				path := "$.b"
				obj := store.Get(key)

				expr, err := jp.ParseString(path)
				assert.Nil(t, err, "error parsing path")

				results := expr.Get(obj.Value)
				assert.Equal(t, len(results), 1)

				want := []interface{}{float64(0), float64(1), float64(3), float64(4), float64(5)}

				equal := reflect.DeepEqual(results[0], want)
				assert.Equal(t, equal, true)
			},
			migratedOutput: EvalResponse{
				Result: []interface{}{float64(2)},
				Error:  nil,
			},
		},
	}
	runMigratedEvalTests(t, tests, evalJSONARRPOP, store)
}

func testEvalTYPE(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"TYPE incorrect number of arguments": {
			name:  "TYPE incorrect number of arguments",
			setup: func() {},
			input: []string{},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("TYPE"),
			},
		},
		"TYPE key does not exist": {
			name:  "TYPE key does not exist",
			setup: func() {},
			input: []string{"nonexistent_key"},
			migratedOutput: EvalResponse{
				Result: "none",
				Error:  nil,
			},
		},
		"TYPE key exists and is of type String": {
			name: "TYPE key exists and is of type String",
			setup: func() {
				store.Put("string_key", store.NewObj("value", -1, object.ObjTypeString))
			},
			input: []string{"string_key"},
			migratedOutput: EvalResponse{
				Result: "string",
				Error:  nil,
			},
		},
		"TYPE key exists and is of type List": {
			name: "TYPE key exists and is of type List",
			setup: func() {
				evalLPUSH([]string{"list_key", "value"}, store)
			},
			input: []string{"list_key"},
			migratedOutput: EvalResponse{
				Result: "list",
				Error:  nil,
			},
		},
		"TYPE key exists and is of type Set": {
			name: "TYPE key exists and is of type Set",
			setup: func() {
				store.Put("set_key", store.NewObj([]byte("value"), -1, object.ObjTypeSet))
			},
			input: []string{"set_key"},
			migratedOutput: EvalResponse{
				Result: "set",
				Error:  nil,
			},
		},
		"TYPE key exists and is of type Hash": {
			name: "TYPE key exists and is of type Hash",
			setup: func() {
				store.Put("hash_key", store.NewObj([]byte("value"), -1, object.ObjTypeSSMap))
			},
			input: []string{"hash_key"},
			migratedOutput: EvalResponse{
				Result: "hash",
				Error:  nil,
			},
		},
	}
	runMigratedEvalTests(t, tests, evalTYPE, store)
}

func BenchmarkEvalTYPE(b *testing.B) {
	store := dstore.NewStore(nil, nil)

	// Define different types of objects to benchmark
	objectTypes := map[string]func(){
		"String": func() {
			store.Put("string_key", store.NewObj("value", -1, object.ObjTypeString))
		},
		"List": func() {
			store.Put("list_key", store.NewObj([]byte("value"), -1, object.ObjTypeDequeue))
		},
		"Set": func() {
			store.Put("set_key", store.NewObj([]byte("value"), -1, object.ObjTypeSet))
		},
		"Hash": func() {
			store.Put("hash_key", store.NewObj([]byte("value"), -1, object.ObjTypeSSMap))
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
			migratedOutput: EvalResponse{
				Result: []string([]string{"COMMAND <subcommand> [<arg> [value] [opt] ...]. Subcommands are:", "(no subcommand)", "     Return details about all DiceDB commands.", "COUNT", "     Return the total number of commands in this DiceDB server.", "LIST", "     Return a list of all commands in this DiceDB server.", "INFO [<command-name> ...]", "     Return details about the specified DiceDB commands. If no command names are given, documentation details for all commands are returned.", "DOCS [<command-name> ...]", "\tReturn documentation details about multiple diceDB commands.\n\tIf no command names are given, documentation details for all\n\tcommands are returned.", "GETKEYS <full-command>", "     Return the keys from a full DiceDB command.", "HELP", "     Print this help."}),
				Error:  nil,
			},
		},
		"command help with wrong number of arguments": {
			input: []string{"HELP", "EXTRA-ARGS"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("COMMAND|HELP"),
			},
		},
		"command info valid command SET": {
			input: []string{"INFO", "SET"},
			migratedOutput: EvalResponse{
				Result: []interface{}([]interface{}{[]interface{}{"set", -3, 1, 0, 0, []interface{}(nil)}}),
				Error:  nil,
			},
		},
		"command info valid command GET": {
			input: []string{"INFO", "GET"},
			migratedOutput: EvalResponse{
				Result: []interface{}([]interface{}{[]interface{}{"get", 2, 1, 0, 0, []interface{}(nil)}}),
				Error:  nil,
			},
		},
		"command info valid command PING": {
			input: []string{"INFO", "PING"},
			migratedOutput: EvalResponse{
				Result: []interface{}([]interface{}{[]interface{}{"ping", -1, 0, 0, 0, []interface{}(nil)}}),
				Error:  nil,
			},
		},
		"command info multiple valid commands": {
			input: []string{"INFO", "SET", "GET"},
			migratedOutput: EvalResponse{
				Result: []interface{}([]interface{}{[]interface{}{"set", -3, 1, 0, 0, []interface{}(nil)}, []interface{}{"get", 2, 1, 0, 0, []interface{}(nil)}}),
				Error:  nil,
			},
		},
		"command info invalid command": {
			input: []string{"INFO", "INVALID_CMD"},
			migratedOutput: EvalResponse{
				Result: []interface{}([]interface{}{[]uint8{0x24, 0x2d, 0x31, 0xd, 0xa}}),
				Error:  nil,
			},
		},
		"command info mixture of valid and invalid commands": {
			input: []string{"INFO", "SET", "INVALID_CMD"},
			migratedOutput: EvalResponse{
				Result: []interface{}([]interface{}{[]interface{}{"set", -3, 1, 0, 0, []interface{}(nil)}, []uint8{0x24, 0x2d, 0x31, 0xd, 0xa}}),
				Error:  nil,
			},
		},
		"command count with wrong number of arguments": {
			input: []string{"COUNT", "EXTRA-ARGS"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("COMMAND|COUNT"),
			},
		},
		"command list with wrong number of arguments": {
			input: []string{"LIST", "EXTRA-ARGS"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("COMMAND|LIST"),
			},
		},
		"command unknown": {
			input: []string{"UNKNOWN"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral("unknown subcommand 'UNKNOWN'. Try COMMAND HELP."),
			},
		},
		"command getkeys with incorrect number of arguments": {
			input: []string{"GETKEYS"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("COMMAND|GETKEYS"),
			},
		},
		"command getkeys with unknown command": {
			input: []string{"GETKEYS", "UNKNOWN"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral("invalid command specified"),
			},
		},
		"command getkeys with a command that accepts no key arguments": {
			input: []string{"GETKEYS", "FLUSHDB"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral("the command has no key arguments"),
			},
		},
		"command getkeys with an invalid number of arguments for a command": {
			input: []string{"GETKEYS", "SET", "key1"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral("invalid number of arguments specified for command"),
			},
		},
		"command docs valid command SET": {
			input: []string{"DOCS", "SET"},
			migratedOutput: EvalResponse{
				Result: []interface{}([]interface{}{[]interface{}{"set", []interface{}{"summary", "SET puts a new <key, value> pair in db as in the args\n\t\targs must contain key and value.\n\t\targs can also contain multiple options -\n\t\tEX or ex which will set the expiry time(in secs) for the key\n\t\tReturns encoded error response if at least a <key, value> pair is not part of args\n\t\tReturns encoded error response if expiry tme value in not integer\n\t\tReturns encoded OK RESP once new entry is added\n\t\tIf the key already exists then the value will be overwritten and expiry will be discarded", "arity", -3, "beginIndex", 1, "lastIndex", 0, "step", 0}}}),
				Error:  nil,
			},
		},
		"command docs valid command GET": {
			input: []string{"DOCS", "GET"},
			migratedOutput: EvalResponse{
				Result: []interface{}([]interface{}{[]interface{}{"get", []interface{}{"summary", "GET returns the value for the queried key in args\n\t\tThe key should be the only param in args\n\t\tThe RESP value of the key is encoded and then returned\n\t\tGET returns RespNIL if key is expired or it does not exist", "arity", 2, "beginIndex", 1, "lastIndex", 0, "step", 0}}}),
				Error:  nil,
			},
		},
		"command docs valid command PING": {
			input: []string{"DOCS", "PING"},
			migratedOutput: EvalResponse{
				Result: []interface{}([]interface{}{[]interface{}{"ping", []interface{}{"summary", "PING returns with an encoded \"PONG\" If any message is added with the ping command,the message will be returned.", "arity", -1, "beginIndex", 0, "lastIndex", 0, "step", 0}}}),
				Error:  nil,
			},
		},
		"command docs multiple valid commands": {
			input: []string{"DOCS", "SET", "GET"},
			migratedOutput: EvalResponse{
				Result: []interface{}([]interface{}{[]interface{}{"set", []interface{}{"summary", "SET puts a new <key, value> pair in db as in the args\n\t\targs must contain key and value.\n\t\targs can also contain multiple options -\n\t\tEX or ex which will set the expiry time(in secs) for the key\n\t\tReturns encoded error response if at least a <key, value> pair is not part of args\n\t\tReturns encoded error response if expiry tme value in not integer\n\t\tReturns encoded OK RESP once new entry is added\n\t\tIf the key already exists then the value will be overwritten and expiry will be discarded", "arity", -3, "beginIndex", 1, "lastIndex", 0, "step", 0}}, []interface{}{"get", []interface{}{"summary", "GET returns the value for the queried key in args\n\t\tThe key should be the only param in args\n\t\tThe RESP value of the key is encoded and then returned\n\t\tGET returns RespNIL if key is expired or it does not exist", "arity", 2, "beginIndex", 1, "lastIndex", 0, "step", 0}}}),
				Error:  nil,
			},
		},
		"command docs invalid command": {
			input: []string{"DOCS", "INVALID_CMD"},
			migratedOutput: EvalResponse{
				Result: []interface{}([]interface{}(nil)),
				Error:  nil,
			},
		},
		"command docs mixture of valid and invalid commands": {
			input: []string{"DOCS", "SET", "INVALID_CMD"},
			migratedOutput: EvalResponse{
				Result: []interface{}([]interface{}{[]interface{}{"set", []interface{}{"summary", "SET puts a new <key, value> pair in db as in the args\n\t\targs must contain key and value.\n\t\targs can also contain multiple options -\n\t\tEX or ex which will set the expiry time(in secs) for the key\n\t\tReturns encoded error response if at least a <key, value> pair is not part of args\n\t\tReturns encoded error response if expiry tme value in not integer\n\t\tReturns encoded OK RESP once new entry is added\n\t\tIf the key already exists then the value will be overwritten and expiry will be discarded", "arity", -3, "beginIndex", 1, "lastIndex", 0, "step", 0}}}),
				Error:  nil,
			},
		},
		"command docs unknown command": {
			input: []string{"UNKNOWN"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral("unknown subcommand 'UNKNOWN'. Try COMMAND HELP."),
			},
		},
	}

	runMigratedEvalTests(t, tests, evalCommand, store)
}

func testEvalJSONOBJKEYS(t *testing.T, store *dstore.Store) {
	tests := []evalTestCase{
		{
			name:  "nil value",
			setup: func() {},
			input: nil,
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("JSON.OBJKEYS"),
			},
		},
		{
			name:  "empty args",
			setup: func() {},
			input: []string{},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("JSON.OBJKEYS"),
			},
		},
		{
			name:  "key does not exist",
			setup: func() {},
			input: []string{"NONEXISTENT_KEY"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral("could not perform this operation on a key that doesn't exist"),
			},
		},
		{
			name: "root not object",
			setup: func() {
				key := "EXISTING_KEY"
				value := "[1]"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input: []string{"EXISTING_KEY"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongTypeOperation,
			},
		},
		{
			name: "wildcard no object objkeys",
			setup: func() {
				key := "EXISTING_KEY"
				value := `{"name":"John","age":30,"pets":null,"languages":["python","golang"],"flag":false}`
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input: []string{"EXISTING_KEY", "$.*"},
			migratedOutput: EvalResponse{
				Result: []interface{}{nil, nil, nil, nil, nil},
				Error:  nil,
			},
		},
		{
			name: "incompatible type (int)",
			setup: func() {
				key := "EXISTING_KEY"
				value := `{"person":{"name":"John","age":30},"languages":["python","golang"]}`
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input: []string{"EXISTING_KEY", "$.person.age"},
			migratedOutput: EvalResponse{
				Result: []interface{}{nil},
				Error:  nil,
			},
		},
		{
			name: "incompatible type (string)",
			setup: func() {
				key := "EXISTING_KEY"
				value := `{"person":{"name":"John","age":30},"languages":["python","golang"]}`
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input: []string{"EXISTING_KEY", "$.person.name"},
			migratedOutput: EvalResponse{
				Result: []interface{}{nil},
				Error:  nil,
			},
		},
		{
			name: "incompatible type (array)",
			setup: func() {
				key := "EXISTING_KEY"
				value := `{"person":{"name":"John","age":30},"languages":["python","golang"]}`
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
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

			response := evalJSONOBJKEYS(tt.input, store)

			if b, ok := response.Result.([]byte); ok && tt.migratedOutput.Result != nil {
				if expectedBytes, ok := tt.migratedOutput.Result.([]byte); ok {
					assert.True(t, bytes.Equal(b, expectedBytes), "expected and actual byte slices should be equal")
				}
			} else {
				assert.Equal(t, tt.migratedOutput.Result, response.Result)
			}

			if tt.migratedOutput.Error != nil {
				assert.EqualError(t, response.Error, tt.migratedOutput.Error.Error())
			} else {
				assert.NoError(t, response.Error)
			}
		})
	}
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
		store.Put("STRING_KEY", store.NewObj("Hello World", maxExDuration, object.ObjTypeString))
	}
	setupForIntegerValue := func() {
		store.Put("INTEGER_KEY", store.NewObj("1234", maxExDuration, object.ObjTypeString))
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
		"GETRANGE against byte array with valid range: 0 4": {
			setup: func() {
				key := "BYTEARRAY_KEY"
				store.Put(key, store.NewObj(&ByteArray{data: []byte{0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x20, 0x77, 0x6f, 0x72, 0x6c, 0x64}}, maxExDuration, object.ObjTypeByteArray))
			},
			input:          []string{"BYTEARRAY_KEY", "0", "4"},
			migratedOutput: EvalResponse{Result: "hello", Error: nil},
		},
		"GETRANGE against byte array with valid range: 6 -1": {
			setup: func() {
				key := "BYTEARRAY_KEY"
				store.Put(key, store.NewObj(&ByteArray{data: []byte{0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x20, 0x77, 0x6f, 0x72, 0x6c, 0x64}}, maxExDuration, object.ObjTypeByteArray))
			},
			input:          []string{"BYTEARRAY_KEY", "6", "-1"},
			migratedOutput: EvalResponse{Result: "world", Error: nil},
		},
		"GETRANGE against byte array with invalid range: 20 30": {
			setup: func() {
				key := "BYTEARRAY_KEY"
				store.Put(key, store.NewObj(&ByteArray{data: []byte{0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x20, 0x77, 0x6f, 0x72, 0x6c, 0x64}}, maxExDuration, object.ObjTypeByteArray))
			},
			input:          []string{"BYTEARRAY_KEY", "20", "30"},
			migratedOutput: EvalResponse{Result: "", Error: nil},
		},
	}

	runMigratedEvalTests(t, tests, evalGETRANGE, store)
}

func BenchmarkEvalGETRANGE(b *testing.B) {
	store := dstore.NewStore(nil, nil)
	store.Put("BENCHMARK_KEY", store.NewObj("Hello World", maxExDuration, object.ObjTypeString))

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
			setup: func() {},
			input: nil,
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("HSETNX"),
			},
		},
		"only key passed": {
			setup: func() {},
			input: []string{"key"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("HSETNX"),
			},
		},
		"only key and field_name passed": {
			setup: func() {},
			input: []string{"KEY", "field_name"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("HSETNX"),
			},
		},
		"more than one field and value passed": {
			setup: func() {},
			input: []string{"KEY", "field1", "value1", "field2", "value2"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("HSETNX"),
			},
		},
		"key, field and value passed": {
			setup: func() {},
			input: []string{"KEY1", "field_name", "value"},
			migratedOutput: EvalResponse{
				Result: int64(1),
				Error:  nil,
			},
		},
		"new set of key, field and value added": {
			setup: func() {},
			input: []string{"KEY2", "field_name_new", "value_new"},
			migratedOutput: EvalResponse{
				Result: int64(1),
				Error:  nil,
			},
		},
		"apply with duplicate key, field and value names": {
			setup: func() {
				key := "KEY_MOCK"
				field := "mock_field_name"
				newMap := make(HashMap)
				newMap[field] = "mock_field_value"

				obj := &object.Obj{
					Type:           object.ObjTypeSSMap,
					Value:          newMap,
					LastAccessedAt: uint32(time.Now().Unix()),
				}

				store.Put(key, obj)
			},
			input: []string{"KEY_MOCK", "mock_field_name", "mock_field_value_2"},
			migratedOutput: EvalResponse{
				Result: int64(0),
				Error:  nil,
			},
		},
		"key exists, field exists but value is nil": {
			setup: func() {
				key := "KEY_EXISTING"
				field := "existing_field"
				newMap := make(HashMap)
				newMap[field] = "existing_value"

				obj := &object.Obj{
					Type:           object.ObjTypeSSMap,
					Value:          newMap,
					LastAccessedAt: uint32(time.Now().Unix()),
				}

				store.Put(key, obj)
			},
			input: []string{"KEY_EXISTING", "existing_field", "new_value"},
			migratedOutput: EvalResponse{
				Result: int64(0),
				Error:  nil,
			},
		},
	}

	runMigratedEvalTests(t, tests, evalHSETNX, store)
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
					Type:           object.ObjTypeSSMap,
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
					Type:           object.ObjTypeSSMap,
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
					Type:           object.ObjTypeSSMap,
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
					Type:           object.ObjTypeSSMap,
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
					Type:           object.ObjTypeSSMap,
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
					Type:           object.ObjTypeSSMap,
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
		"key exp value pair":                     {input: []string{"KEY", "123", "VAL"}, migratedOutput: EvalResponse{Result: OK, Error: nil}},
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
				assert.Equal(t, OK, output)

				// Check if the key was set correctly
				getValue := evalGET([]string{"TEST_KEY"}, store)
				assert.Equal(t, "TEST_VALUE", getValue.Result)

				// Check if the TTL is set correctly (should be 5 seconds or less)
				ttlResponse := evalTTL([]string{"TEST_KEY"}, store)
				ttl, ok := ttlResponse.Result.(uint64)
				assert.True(t, ok, "TTL result should be an uint64")
				assert.True(t, ttl > 0 && ttl <= 5, "TTL should be between 0 and 5 seconds")
				assert.Nil(t, ttlResponse.Error, "TTL command should not return an error")

				// Wait for the key to expire
				mockTime.SetTime(mockTime.CurrTime.Add(6 * time.Second))

				// Check if the key has been deleted after expiry
				expiredValue := evalGET([]string{"TEST_KEY"}, store)
				assert.Equal(t, NIL, expiredValue.Result)
			},
		},
		"update existing key": {
			setup: func() {
				evalSET([]string{"EXISTING_KEY", "OLD_VALUE"}, store)
			},
			input: []string{"EXISTING_KEY", "10", "NEW_VALUE"},
			newValidator: func(output interface{}) {
				assert.Equal(t, OK, output)

				// Check if the key was updated correctly
				getValue := evalGET([]string{"EXISTING_KEY"}, store)
				assert.Equal(t, "NEW_VALUE", getValue.Result)

				// Check if the TTL is set correctly
				ttlResponse := evalTTL([]string{"EXISTING_KEY"}, store)
				ttl, ok := ttlResponse.Result.(uint64)
				assert.True(t, ok, "TTL result should be an uint64")
				assert.True(t, ttl > 0 && ttl <= 10, "TTL should be between 0 and 10 seconds")
				assert.Nil(t, ttlResponse.Error, "TTL command should not return an error")
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
						assert.True(t, bytes.Equal(b, expectedBytes), "expected and actual byte slices should be equal")
					}
				} else {
					assert.Equal(t, tt.migratedOutput.Result, response.Result)
				}

				if tt.migratedOutput.Error != nil {
					assert.EqualError(t, response.Error, tt.migratedOutput.Error.Error())
				} else {
					assert.NoError(t, response.Error)
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
				obj := store.NewObj(value, -1, object.ObjTypeString)
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
				obj := store.NewObj(value, -1, object.ObjTypeInt)
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
				obj := store.NewObj(value, -1, object.ObjTypeInt)
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
				obj := store.NewObj(value, -1, object.ObjTypeInt)
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
				obj := store.NewObj(value, -1, object.ObjTypeString)
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
				obj := store.NewObj(value, -1, object.ObjTypeString)
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
				obj := store.NewObj(value, -1, object.ObjTypeString)
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
				obj := store.NewObj(value, -1, object.ObjTypeString)
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
				obj := store.NewObj(value, -1, object.ObjTypeString)
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
				obj := store.NewObj(value, -1, object.ObjTypeString)
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
					assert.True(t, bytes.Equal(b, expectedBytes), "expected and actual byte slices should be equal")
				}
			} else {
				assert.Equal(t, tt.migratedOutput.Result, response.Result)
			}

			if tt.migratedOutput.Error != nil {
				assert.EqualError(t, response.Error, tt.migratedOutput.Error.Error())
			} else {
				assert.NoError(t, response.Error)
			}
		})
	}
}

func BenchmarkEvalINCRBYFLOAT(b *testing.B) {
	store := dstore.NewStore(nil, nil)
	store.Put("key1", store.NewObj("1", maxExDuration, object.ObjTypeString))
	store.Put("key2", store.NewObj("1.2", maxExDuration, object.ObjTypeString))

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
				Result: RespType(0),
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
					Type:           object.ObjTypeSSMap,
					Value:          newMap,
					LastAccessedAt: uint32(time.Now().Unix()),
				}

				store.Put(key, obj)
			},
			input: []string{"KEY_MOCK"},
			newValidator: func(output interface{}) {
				assert.True(t, output != nil)
				stringSlice, ok := output.([]string)
				if !ok {
					assert.Error(t, diceerrors.ErrUnexpectedType("[]string", reflect.TypeOf(output)))
				}
				resultString := strings.Join(stringSlice, " ")
				assert.True(t,
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
					Type:           object.ObjTypeSSMap,
					Value:          newMap,
					LastAccessedAt: uint32(time.Now().Unix()),
				}

				store.Put(key, obj)
			},
			input: []string{"KEY_MOCK", "2"},
			newValidator: func(output interface{}) {
				assert.True(t, output != nil)
				stringSlice, ok := output.([]string)
				if !ok {
					assert.Error(t, diceerrors.ErrUnexpectedType("[]string", reflect.TypeOf(output)))
				}
				decodedResult := strings.Join(stringSlice, " ")
				fields := []string{"field1", "field2", "field3"}
				count := 0

				for _, field := range fields {
					if strings.Contains(decodedResult, field) {
						count++
					}
				}

				assert.True(t, count == 2)
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
					Type:           object.ObjTypeSSMap,
					Value:          newMap,
					LastAccessedAt: uint32(time.Now().Unix()),
				}

				store.Put(key, obj)
			},
			input: []string{"KEY_MOCK", "2", WithValues},
			newValidator: func(output interface{}) {
				assert.True(t, output != nil)
				stringSlice, ok := output.([]string)
				if !ok {
					assert.Error(t, diceerrors.ErrUnexpectedType("[]string", reflect.TypeOf(output)))
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
				obj := store.NewObj(value, -1, object.ObjTypeString)
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
			newValidator: func(output interface{}) {
				obj := store.Get("key")
				oType := obj.Type
				if oType != object.ObjTypeInt {
					t.Errorf("unexpected encoding")
					return
				}
			},
		},
		"append string value to existing key having integer value": {
			setup: func() {
				key := "key"
				value := "123"
				storedValue, _ := strconv.ParseInt(value, 10, 64)
				obj := store.NewObj(storedValue, -1, object.ObjTypeInt)
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
				obj := store.NewObj(value, -1, object.ObjTypeString)
				store.Put(key, obj)
			},
			input:          []string{"key", ""},
			migratedOutput: EvalResponse{Result: 0, Error: nil},
		},
		"append empty string to existing key": {
			setup: func() {
				key := "key"
				value := "val"
				obj := store.NewObj(value, -1, object.ObjTypeString)
				store.Put(key, obj)
			},
			input:          []string{"key", ""},
			migratedOutput: EvalResponse{Result: 3, Error: nil},
		},
		"append modifies the encoding from int to raw": {
			setup: func() {
				store.Del("key")
				storedValue, _ := strconv.ParseInt("1", 10, 64)
				obj := store.NewObj(storedValue, -1, object.ObjTypeInt)
				store.Put("key", obj)
			},
			input:          []string{"key", "2"},
			migratedOutput: EvalResponse{Result: 2, Error: nil},
			newValidator: func(output interface{}) {
				obj := store.Get("key")
				oType := obj.Type
				if oType != object.ObjTypeString {
					t.Errorf("unexpected encoding")
					return
				}
			},
		},
		"append to key created using LPUSH": {
			setup: func() {
				key := "listKey"
				value := "val"
				// Create a new list object
				obj := store.NewObj(NewDeque(), -1, object.ObjTypeDequeue)
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
				obj := store.NewObj(initialValues, -1, object.ObjTypeSet)
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
				obj := store.NewObj(initialValues, -1, object.ObjTypeSSMap)
				store.Put(key, obj)
			},
			input:          []string{"hashKey", "val"},
			migratedOutput: EvalResponse{Result: nil, Error: diceerrors.ErrWrongTypeOperation},
		},
		"append to key containing byte array": {
			setup: func() {
				key := "bitKey"
				// Create a new byte array object
				initialByteArray := NewByteArray(2) // Initialize with 2 byte
				initialByteArray.SetBit(2, true)    // Set the third bit to 1
				initialByteArray.SetBit(3, true)    // Set the fourth bit to 1
				initialByteArray.SetBit(5, true)    // Set the sixth bit to 1
				initialByteArray.SetBit(10, true)   // Set the eleventh bit to 1
				initialByteArray.SetBit(11, true)   // Set the twelfth bit to 1
				initialByteArray.SetBit(14, true)   // Set the fifteenth bit to 1
				obj := store.NewObj(initialByteArray, -1, object.ObjTypeByteArray)
				store.Put(key, obj)
			},
			input:          []string{"bitKey", "1"},
			migratedOutput: EvalResponse{Result: 3, Error: nil},
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
			setup:          func() {},
			input:          nil,
			migratedOutput: EvalResponse{Result: nil, Error: diceerrors.ErrWrongArgumentCount("JSON.RESP")},
		},
		"key does not exist": {
			setup:          func() {},
			input:          []string{"NOTEXISTANT_KEY"},
			migratedOutput: EvalResponse{Result: NIL, Error: nil},
		},
		"string json": {
			setup: func() {
				key := "MOCK_KEY"
				value := "\"Roll the Dice\""
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input:          []string{"MOCK_KEY"},
			migratedOutput: EvalResponse{Result: []interface{}{"Roll the Dice"}, Error: nil},
		},
		"integer json": {
			setup: func() {
				key := "MOCK_KEY"
				value := "10"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input:          []string{"MOCK_KEY"},
			migratedOutput: EvalResponse{Result: []interface{}{float64(10)}, Error: nil},
		},
		"bool json": {
			setup: func() {
				key := "MOCK_KEY"
				value := "true"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input:          []string{"MOCK_KEY"},
			migratedOutput: EvalResponse{Result: []interface{}{true}, Error: nil},
		},
		"nil json": {
			setup: func() {
				key := "MOCK_KEY"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(nil), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input:          []string{"MOCK_KEY"},
			migratedOutput: EvalResponse{Result: []interface{}{nil}, Error: nil},
		},
		"empty array": {
			setup: func() {
				key := "MOCK_KEY"
				value := "[]"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input:          []string{"MOCK_KEY"},
			migratedOutput: EvalResponse{Result: []interface{}{"["}, Error: nil},
		},
		"empty object": {
			setup: func() {
				key := "MOCK_KEY"
				value := "{}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input:          []string{"MOCK_KEY"},
			migratedOutput: EvalResponse{Result: []interface{}{"{"}, Error: nil},
		},
		"array with mixed types": {
			setup: func() {
				key := "MOCK_KEY"
				value := "[\"dice\", 10, 10.5, true, null]"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input:          []string{"MOCK_KEY"},
			migratedOutput: EvalResponse{Result: []interface{}{"[", "dice", float64(10), float64(10.5), true, interface{}(nil)}, Error: nil},
		},
		"one layer of nesting no path": {
			setup: func() {
				key := "MOCK_KEY"
				value := "{\"b\": [\"dice\", 10, 10.5, true, null]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input:          []string{"MOCK_KEY"},
			migratedOutput: EvalResponse{Result: []interface{}{"{", "b", []interface{}{"[", "dice", float64(10), float64(10.5), true, interface{}(nil)}}, Error: nil},
		},
		"one layer of nesting with path": {
			setup: func() {
				key := "MOCK_KEY"
				value := "{\"b\": [\"dice\", 10, 10.5, true, null]}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input:          []string{"MOCK_KEY", "$.b"},
			migratedOutput: EvalResponse{Result: []interface{}{[]interface{}{"[", "dice", float64(10), float64(10.5), true, interface{}(nil)}}, Error: nil},
		},
	}

	runMigratedEvalTests(t, tests, evalJSONRESP, store)
}

func testEvalSADD(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"SADD with wrong number of arguments": {
			input: []string{},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("SADD"),
			},
		},
		"SADD new member to non existing key": {
			input: []string{"myset", "member"},
			migratedOutput: EvalResponse{
				Result: 1,
				Error:  nil,
			},
		},
		"SADD new members to non existing key": {
			input: []string{"myset", "member1", "member2"},
			migratedOutput: EvalResponse{
				Result: 2,
				Error:  nil,
			},
		},
		"SADD new member to existing key": {
			setup: func() {
				evalSADD([]string{"myset", "member1"}, store)
			},
			input: []string{"myset", "member2"},
			migratedOutput: EvalResponse{
				Result: 1,
				Error:  nil,
			},
		},
		"SADD existing member to existing key": {
			setup: func() {
				evalSADD([]string{"myset", "member1"}, store)
			},
			input: []string{"myset", "member1"},
			migratedOutput: EvalResponse{
				Result: 0,
				Error:  nil,
			},
		},
		"SADD new and existing member to existing key": {
			setup: func() {
				evalSADD([]string{"myset", "member1"}, store)
			},
			input: []string{"myset", "member1", "member2", "member3"},
			migratedOutput: EvalResponse{
				Result: 2,
				Error:  nil,
			},
		},
		"SADD member to key with invalid type": {
			setup: func() {
				evalSET([]string{"key", "value"}, store)
			},
			input: []string{"key", "member"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongTypeOperation,
			},
		},
	}

	runMigratedEvalTests(t, tests, evalSADD, store)
}

func testEvalSREM(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"SREM with wrong number of arguments": {
			input: []string{},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("SREM"),
			},
		},
		"SREM on key with invalid type": {
			setup: func() {
				evalSET([]string{"key", "value"}, store)
			},
			input: []string{"key", "member"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongTypeOperation,
			},
		},
		"SREM with non existing key": {
			input: []string{"myset", "member"},
			migratedOutput: EvalResponse{
				Result: 0,
				Error:  nil,
			},
		},
		"SREM on existing key with existing member": {
			setup: func() {
				evalSADD([]string{"myset", "a"}, store)
			},
			input: []string{"myset", "a"},
			migratedOutput: EvalResponse{
				Result: 1,
				Error:  nil,
			},
		},
		"SREM on existing key with not existing member": {
			setup: func() {
				evalSADD([]string{"myset", "a"}, store)
			},
			input: []string{"myset", "b"},
			migratedOutput: EvalResponse{
				Result: 0,
				Error:  nil,
			},
		},
		"SREM on existing key with existing and not existing members": {
			setup: func() {
				evalSADD([]string{"myset", "a"}, store)
			},
			input: []string{"myset", "a", "b"},
			migratedOutput: EvalResponse{
				Result: 1,
				Error:  nil,
			},
		},
		"SREM on existing key with repeated existing members": {
			setup: func() {
				evalSADD([]string{"myset", "a"}, store)
			},
			input: []string{"myset", "a", "b", "a"},
			migratedOutput: EvalResponse{
				Result: 1,
				Error:  nil,
			},
		},
	}

	runMigratedEvalTests(t, tests, evalSREM, store)
}

func testEvalSCARD(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"SCARD with wrong number of arguments": {
			input: []string{"mykey", "value"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("SCARD"),
			},
		},
		"SCARD on key with invalid type": {
			setup: func() {
				evalSET([]string{"mykey", "value"}, store)
			},
			input: []string{"mykey"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongTypeOperation,
			},
		},
		"SCARD with non existing key": {
			input: []string{"mykey"},
			migratedOutput: EvalResponse{
				Result: 0,
				Error:  nil,
			},
		},
		"SCARD with existing key and no member": {
			setup: func() {
				evalSADD([]string{"mykey"}, store)
			},
			input: []string{"mykey"},
			migratedOutput: EvalResponse{
				Result: 0,
				Error:  nil,
			},
		},
		"SCARD with existing key": {
			setup: func() {
				evalSADD([]string{"mykey", "a", "b"}, store)
			},
			input: []string{"mykey"},
			migratedOutput: EvalResponse{
				Result: 2,
				Error:  nil,
			},
		},
	}

	runMigratedEvalTests(t, tests, evalSCARD, store)
}

func testEvalSMEMBERS(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"SMEMBERS with wrong number of arguments": {
			input: []string{"mykey", "mykey"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("SMEMBERS"),
			},
		},
		"SMEMBERS on key with invalid type": {
			setup: func() {
				evalSET([]string{"mykey", "value"}, store)
			},
			input: []string{"mykey"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongTypeOperation,
			},
		},
		"SMEMBERS with non existing key": {
			input: []string{"mykey"},
			migratedOutput: EvalResponse{
				Result: []string{},
				Error:  nil,
			},
		},
		"SMEMBERS with  existing key": {
			setup: func() {
				evalSADD([]string{"mykey", "a", "b"}, store)
			},
			input: []string{"mykey"},
			migratedOutput: EvalResponse{
				Result: []string{"a", "b"},
				Error:  nil,
			},
		},
		"SMEMBERS with  existing key and no members": {
			setup: func() {
				evalSADD([]string{"mykey"}, store)
			},
			input: []string{"mykey"},
			migratedOutput: EvalResponse{
				Result: []string{},
				Error:  nil,
			},
		},
	}
	runMigratedEvalTests(t, tests, evalSMEMBERS, store)
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
				store.Put("mywrongtypekey", store.NewObj("string_value", -1, object.ObjTypeString))
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
				store.Put("mystring", store.NewObj("string_value", -1, object.ObjTypeString))
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
				Error:  diceerrors.ErrInvalidSyntax("ZRANGE"),
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
				store.Put("mystring", store.NewObj("string_value", -1, object.ObjTypeString))
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
				store.Put("myzset", store.NewObj(sortedset.New(), -1, object.ObjTypeSortedSet)) // Ensure the set exists but is empty
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
				dstore.Reset(store)
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
				Result: NIL,
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
				Result: NIL,
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
				Error:  diceerrors.ErrInvalidSyntax("ZRANK"),
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

func testEvalZREM(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"ZREM with wrong number of arguments": {
			input: []string{"myzset"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("ZREM"),
			},
		},
		"ZREM with missing key": {
			input: []string{},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("ZREM"),
			},
		},
		"ZREM with wrong type key": {
			setup: func() {
				store.Put("string_key", store.NewObj("string_value", -1, object.ObjTypeString))
			},
			input: []string{"string_key", "field"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongTypeOperation,
			},
		},
		"ZREM with non-existent key": {
			input: []string{"non_existent_key", "field"},
			migratedOutput: EvalResponse{
				Result: IntegerZero,
				Error:  nil,
			},
		},
		"ZREM with non-existent element": {
			setup: func() {
				evalZADD([]string{"myzset", "1", "one"}, store)
			},
			input: []string{"myzset", "two"},
			migratedOutput: EvalResponse{
				Result: int64(0),
				Error:  nil,
			},
		},
		"ZREM with sorted set holding single element": {
			setup: func() {
				evalZADD([]string{"myzset", "1", "one"}, store)
			},
			input: []string{"myzset", "one"},
			migratedOutput: EvalResponse{
				Result: int64(1),
				Error:  nil,
			},
		},
		"ZREM with sorted set holding multiple elements": {
			setup: func() {
				evalZADD([]string{"myzset", "1", "one", "2", "two", "3", "three"}, store)
			},
			input: []string{"myzset", "one", "two"},
			migratedOutput: EvalResponse{
				Result: int64(2),
				Error:  nil,
			},
		},
	}

	runMigratedEvalTests(t, tests, evalZREM, store)
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

func testEvalZCARD(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"ZCARD with wrong number of arguments": {
			input: []string{"myzset", "field"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("ZCARD"),
			},
		},
		"ZCARD with missing key": {
			input: []string{},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("ZCARD"),
			},
		},
		"ZCARD with wrong type key": {
			setup: func() {
				store.Put("string_key", store.NewObj("string_value", -1, object.ObjTypeString))
			},
			input: []string{"string_key"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongTypeOperation,
			},
		},
		"ZCARD with non-existent key": {
			input: []string{"non_existent_key"},
			migratedOutput: EvalResponse{
				Result: IntegerZero,
				Error:  nil,
			},
		},
		"ZCARD with sorted set holding single element": {
			setup: func() {
				evalZADD([]string{"myzset", "1", "one"}, store)
			},
			input: []string{"myzset"},
			migratedOutput: EvalResponse{
				Result: int64(1),
				Error:  nil,
			},
		},
		"ZCARD with sorted set holding multiple elements": {
			setup: func() {
				evalZADD([]string{"myzset", "1", "one", "2", "two", "3", "three"}, store)
			},
			input: []string{"myzset"},
			migratedOutput: EvalResponse{
				Result: int64(3),
				Error:  nil,
			},
		},
	}

	runMigratedEvalTests(t, tests, evalZCARD, store)
}

func testEvalBitField(t *testing.T, store *dstore.Store) {
	testCases := map[string]evalTestCase{
		"BITFIELD signed SET": {
			input: []string{"bits", "set", "i8", "0", "-100"},
			migratedOutput: EvalResponse{
				Result: []interface{}{int64(0)},
				Error:  nil,
			},
		},
		"BITFIELD GET": {
			setup: func() {
				args := []string{"bits", "set", "u8", "0", "255"}
				evalBITFIELD(args, store)
			},
			input: []string{"bits", "get", "u8", "0"},
			migratedOutput: EvalResponse{
				Result: []interface{}{int64(255)},
				Error:  nil,
			},
		},
		"BITFIELD INCRBY": {
			setup: func() {
				args := []string{"bits", "set", "u8", "0", "255"}
				evalBITFIELD(args, store)
			},
			input: []string{"bits", "incrby", "u8", "0", "100"},
			migratedOutput: EvalResponse{
				Result: []interface{}{int64(99)},
				Error:  nil,
			},
		},
		"BITFIELD Arity": {
			input:          []string{},
			migratedOutput: EvalResponse{Result: nil, Error: diceerrors.ErrWrongArgumentCount("BITFIELD")},
		},
		"BITFIELD invalid combination of commands in a single operation": {
			input:          []string{"bits", "SET", "u8", "0", "255", "INCRBY", "u8", "0", "100", "GET", "u8"},
			migratedOutput: EvalResponse{Result: nil, Error: diceerrors.ErrInvalidSyntax("BITFIELD")},
		},
		"BITFIELD invalid bitfield type": {
			input:          []string{"bits", "SET", "a8", "0", "255", "INCRBY", "u8", "0", "100", "GET", "u8"},
			migratedOutput: EvalResponse{Result: nil, Error: diceerrors.ErrGeneral("Invalid bitfield type. Use something like i16 u8. Note that u64 is not supported but i64 is")},
		},
		"BITFIELD invalid bit offset": {
			input:          []string{"bits", "SET", "u8", "a", "255", "INCRBY", "u8", "0", "100", "GET", "u8"},
			migratedOutput: EvalResponse{Result: nil, Error: diceerrors.ErrGeneral("bit offset is not an integer or out of range")},
		},
		"BITFIELD invalid overflow type": {
			input:          []string{"bits", "SET", "u8", "0", "255", "INCRBY", "u8", "0", "100", "OVERFLOW", "wraap"},
			migratedOutput: EvalResponse{Result: nil, Error: diceerrors.ErrGeneral("Invalid OVERFLOW type specified")},
		},
		"BITFIELD missing arguments in SET": {
			input:          []string{"bits", "SET", "u8", "0", "INCRBY", "u8", "0", "100", "GET", "u8", "288"},
			migratedOutput: EvalResponse{Result: nil, Error: diceerrors.ErrIntegerOutOfRange},
		},
	}

	runMigratedEvalTests(t, testCases, evalBITFIELD, store)
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
					Type:           object.ObjTypeSSMap,
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
					Type:           object.ObjTypeSSMap,
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
					Type:           object.ObjTypeSSMap,
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
					Type:           object.ObjTypeSSMap,
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
					Type:           object.ObjTypeSSMap,
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
					Type:           object.ObjTypeSSMap,
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
					Type:           object.ObjTypeSSMap,
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
					Type:           object.ObjTypeSSMap,
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
	store.Put("key1", store.NewObj(HashMap{"field1": "1.0", "field2": "1.2"}, maxExDuration, object.ObjTypeSSMap))
	store.Put("key2", store.NewObj(HashMap{"field1": "0.1"}, maxExDuration, object.ObjTypeSSMap))

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
			setup: func() {},
			input: nil,
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("DUMP"),
			},
		},
		"empty array": {
			setup: func() {},
			input: []string{},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("DUMP"),
			},
		},
		"key does not exist": {
			setup: func() {},
			input: []string{"NONEXISTENT_KEY"},
			migratedOutput: EvalResponse{
				Result: NIL,
				Error:  nil,
			},
		}, "dump string value": {
			setup: func() {
				key := "user"
				value := "hello"
				obj := store.NewObj(value, -1, object.ObjTypeString)
				store.Put(key, obj)
			},
			input: []string{"user"},
			migratedOutput: EvalResponse{
				Result: base64.StdEncoding.EncodeToString([]byte{
					0x09, 0x00, 0x00, 0x00, 0x00, 0x05, 0x68, 0x65, 0x6c, 0x6c, 0x6f,
					0xFF, // End marker
					// CRC64 checksum here:
					0x00, 0x47, 0x97, 0x93, 0xBE, 0x36, 0x45, 0xC7,
				}),
				Error: nil,
			},
		},
		"dump integer value": {
			setup: func() {
				key := "INTEGER_KEY"
				value := int64(10)
				obj := store.NewObj(value, -1, object.ObjTypeInt)
				store.Put(key, obj)
			},
			input: []string{"INTEGER_KEY"},
			migratedOutput: EvalResponse{
				Result: "CQUAAAAAAAAACv9+l81XgsShqw==",
				Error:  nil,
			},
		},
		"dump expired key": {
			setup: func() {
				key := "EXPIRED_KEY"
				value := "This will expire"
				obj := store.NewObj(value, -1, object.ObjTypeString)
				store.Put(key, obj)
				var exDurationMs int64 = -1
				store.SetExpiry(obj, exDurationMs)
			},
			input: []string{"EXPIRED_KEY"},
			migratedOutput: EvalResponse{
				Result: NIL,
				Error:  nil,
			},
		},
	}

	runMigratedEvalTests(t, tests, evalDUMP, store)
}

func testEvalBitFieldRO(t *testing.T, store *dstore.Store) {
	testCases := map[string]evalTestCase{
		"BITFIELD_RO Arity": {
			input: []string{},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("BITFIELD_RO"),
			},
		},
		"BITFIELD_RO syntax error": {
			input:          []string{"bits", "GET", "u8"},
			migratedOutput: EvalResponse{Result: nil, Error: diceerrors.ErrInvalidSyntax("BITFIELD_RO")},
		},
		"BITFIELD_RO invalid bitfield type": {
			input:          []string{"bits", "GET", "a8", "0", "255"},
			migratedOutput: EvalResponse{Result: nil, Error: diceerrors.ErrGeneral("Invalid bitfield type. Use something like i16 u8. Note that u64 is not supported but i64 is")},
		},
		"BITFIELD_RO unsupported commands": {
			input:          []string{"bits", "set", "u8", "0", "255"},
			migratedOutput: EvalResponse{Result: nil, Error: diceerrors.ErrGeneral("BITFIELD_RO only supports the GET subcommand")},
		},
	}

	runMigratedEvalTests(t, testCases, evalBITFIELDRO, store)
}

func testEvalGEOADD(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"GEOADD with wrong number of arguments": {
			input: []string{"mygeo", "1", "2"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("GEOADD"),
			},
		},
		"GEOADD with non-numeric longitude": {
			input: []string{"mygeo", "long", "40.7128", "NewYork"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral("invalid longitude"),
			},
		},
		"GEOADD with non-numeric latitude": {
			input: []string{"mygeo", "-74.0060", "lat", "NewYork"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral("invalid latitude"),
			},
		},
		"GEOADD new member to non-existing key": {
			setup: func() {},
			input: []string{"mygeo", "-74.0060", "40.7128", "NewYork"},
			migratedOutput: EvalResponse{
				Result: 1,
				Error:  nil,
			},
		},
		"GEOADD existing member with updated coordinates": {
			setup: func() {
				evalGEOADD([]string{"mygeo", "-74.0060", "40.7128", "NewYork"}, store)
			},
			input: []string{"mygeo", "-73.9352", "40.7304", "NewYork"},
			migratedOutput: EvalResponse{
				Result: 0,
				Error:  nil,
			},
		},
		"GEOADD multiple members": {
			setup: func() {
				evalGEOADD([]string{"mygeo", "-74.0060", "40.7128", "NewYork"}, store)
			},
			input: []string{"mygeo", "-118.2437", "34.0522", "LosAngeles", "-87.6298", "41.8781", "Chicago"},
			migratedOutput: EvalResponse{
				Result: 2,
				Error:  nil,
			},
		},
		"GEOADD with NX option (new member)": {
			input: []string{"mygeo", "NX", "-122.4194", "37.7749", "SanFrancisco"},
			migratedOutput: EvalResponse{
				Result: 1,
				Error:  nil,
			},
		},
		"GEOADD with NX option (existing member)": {
			setup: func() {
				evalGEOADD([]string{"mygeo", "-74.0060", "40.7128", "NewYork"}, store)
			},
			input: []string{"mygeo", "NX", "-73.9352", "40.7304", "NewYork"},
			migratedOutput: EvalResponse{
				Result: 0,
				Error:  nil,
			},
		},
		"GEOADD with XX option (new member)": {
			input: []string{"mygeo", "XX", "-71.0589", "42.3601", "Boston"},
			migratedOutput: EvalResponse{
				Result: 0,
				Error:  nil,
			},
		},
		"GEOADD with XX option (existing member)": {
			setup: func() {
				evalGEOADD([]string{"mygeo", "-74.0060", "40.7128", "NewYork"}, store)
			},
			input: []string{"mygeo", "XX", "-73.9352", "40.7304", "NewYork"},
			migratedOutput: EvalResponse{
				Result: 0,
				Error:  nil,
			},
		},
		"GEOADD with both NX and XX options": {
			input:  []string{"mygeo", "NX", "XX", "-74.0060", "40.7128", "NewYork"},
			output: diceerrors.NewErrWithMessage("ERR XX and NX options at the same time are not compatible"),
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral("XX and NX options at the same time are not compatible"),
			},
		},
		"GEOADD with invalid option": {
			input: []string{"mygeo", "INVALID", "-74.0060", "40.7128", "NewYork"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("GEOADD"),
			},
		},
		"GEOADD to a key of wrong type": {
			setup: func() {
				store.Put("mygeo", store.NewObj("string_value", -1, object.ObjTypeString))
			},
			input: []string{"mygeo", "-74.0060", "40.7128", "NewYork"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongTypeOperation,
			},
		},
		"GEOADD with longitude out of range": {
			input: []string{"mygeo", "181.0", "40.7128", "Invalid"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral("invalid longitude"),
			},
		},
		"GEOADD with latitude out of range": {
			input: []string{"mygeo", "-74.0060", "91.0", "Invalid"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral("invalid latitude"),
			},
		},
	}

	runMigratedEvalTests(t, tests, evalGEOADD, store)
}

func testEvalGEODIST(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"GEODIST between existing points": {
			setup: func() {
				evalGEOADD([]string{"points", "13.361389", "38.115556", "Palermo"}, store)
				evalGEOADD([]string{"points", "15.087269", "37.502669", "Catania"}, store)
			},
			input: []string{"points", "Palermo", "Catania"},
			migratedOutput: EvalResponse{
				Result: float64(166274.1440),
				Error:  nil,
			},
		},
		"GEODIST with units (km)": {
			setup: func() {
				evalGEOADD([]string{"points", "13.361389", "38.115556", "Palermo"}, store)
				evalGEOADD([]string{"points", "15.087269", "37.502669", "Catania"}, store)
			},
			input: []string{"points", "Palermo", "Catania", "km"},
			migratedOutput: EvalResponse{
				Result: float64(166.2741),
				Error:  nil,
			},
		},
		"GEODIST to same point": {
			setup: func() {
				evalGEOADD([]string{"points", "13.361389", "38.115556", "Palermo"}, store)
			},
			input: []string{"points", "Palermo", "Palermo"},
			migratedOutput: EvalResponse{
				Result: float64(0.0000),
				Error:  nil,
			},
		},
		// Add other test cases here...
	}

	runMigratedEvalTests(t, tests, evalGEODIST, store)
}

func testEvalGEOPOS(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"GEOPOS for existing single point": {
			setup: func() {
				evalGEOADD([]string{"index", "13.361387", "38.115556", "Palermo"}, store)
			},
			input: []string{"index", "Palermo"},
			migratedOutput: EvalResponse{
				Result: []interface{}{[]interface{}{float64(13.361387), float64(38.115556)}},
				Error:  nil,
			},
		},
		"GEOPOS for multiple existing points": {
			setup: func() {
				evalGEOADD([]string{"points", "13.361387", "38.115556", "Palermo"}, store)
				evalGEOADD([]string{"points", "15.087265", "37.502668", "Catania"}, store)
			},
			input: []string{"points", "Palermo", "Catania"},
			migratedOutput: EvalResponse{
				Result: []interface{}{
					[]interface{}{float64(13.361387), float64(38.115556)},
					[]interface{}{float64(15.087265), float64(37.502668)},
				},
				Error: nil,
			},
		},
		"GEOPOS for a point that does not exist": {
			setup: func() {
				evalGEOADD([]string{"index", "13.361387", "38.115556", "Palermo"}, store)
			},
			input: []string{"index", "NonExisting"},
			migratedOutput: EvalResponse{
				Result: []interface{}{nil},
				Error:  nil,
			},
		},
		"GEOPOS for multiple points, one existing and one non-existing": {
			setup: func() {
				evalGEOADD([]string{"index", "13.361387", "38.115556", "Palermo"}, store)
			},
			input: []string{"index", "Palermo", "NonExisting"},
			migratedOutput: EvalResponse{
				Result: []interface{}{
					[]interface{}{float64(13.361387), float64(38.115556)},
					nil,
				},
				Error: nil,
			},
		},
		"GEOPOS for empty index": {
			setup: func() {
				evalGEOADD([]string{"", "13.361387", "38.115556", "Palermo"}, store)
			},
			input: []string{"", "Palermo"},
			migratedOutput: EvalResponse{
				Result: []interface{}{
					[]interface{}{float64(13.361387), float64(38.115556)},
				},
				Error: nil,
			},
		},
		"GEOPOS with no members in key": {
			input: []string{"index", "Palermo"},
			migratedOutput: EvalResponse{
				Result: NIL,
				Error:  nil,
			},
		},
		"GEOPOS with invalid number of arguments": {
			input: []string{"index"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("GEOPOS"),
			},
		},
		"GEOPOS for a key not used for setting geospatial values": {
			setup: func() {
				evalSET([]string{"k", "v"}, store)
			},
			input: []string{"k", "v"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  errors.New("WRONGTYPE Operation against a key holding the wrong kind of value"),
			},
		},
	}

	runMigratedEvalTests(t, tests, evalGEOPOS, store)
}

func testEvalGEOHASH(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"GEOHASH with wrong number of arguments": {
			input: []string{"mygeo"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("GEOHASH"),
			},
		},
		"GEOHASH with non-existent key": {
			input: []string{"nonexistent", "member1"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrKeyNotFound,
			},
		},
		"GEOHASH with existing key but missing member": {
			setup: func() {
				evalGEOADD([]string{"mygeo", "-74.0060", "40.7128", "NewYork"}, store)
			},
			input: []string{"mygeo", "missingMember"},
			migratedOutput: EvalResponse{
				Result: []interface{}{(nil)},
				Error:  nil,
			},
		},
		"GEOHASH for single member": {
			setup: func() {
				evalGEOADD([]string{"mygeo", "-74.0060", "40.7128", "NewYork"}, store)
			},
			input: []string{"mygeo", "NewYork"},
			migratedOutput: EvalResponse{
				Result: []interface{}{"dr5regw3pp"},
				Error:  nil,
			},
		},
		"GEOHASH for multiple members": {
			setup: func() {
				evalGEOADD([]string{"mygeo", "-74.0060", "40.7128", "NewYork"}, store)
				evalGEOADD([]string{"mygeo", "-118.2437", "34.0522", "LosAngeles"}, store)
			},
			input: []string{"mygeo", "NewYork", "LosAngeles"},
			migratedOutput: EvalResponse{
				Result: []interface{}{"dr5regw3pp", "9q5ctr186n"},
				Error:  nil,
			},
		},
		"GEOHASH with a key of wrong type": {
			setup: func() {
				store.Put("mygeo", store.NewObj("string_value", -1, object.ObjTypeString))
			},
			input: []string{"mygeo", "member1"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongTypeOperation,
			},
		},
	}

	runMigratedEvalTests(t, tests, evalGEOHASH, store)
}

func testEvalJSONSTRAPPEND(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"append to single field": {
			setup: func() {
				key := "doc1"
				value := "{\"a\":\"foo\", \"nested1\": {\"a\": \"hello\"}, \"nested2\": {\"a\": 31}}"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input: []string{"doc1", "$.nested1.a", "\"baz\""},
			migratedOutput: EvalResponse{
				Result: 8,
				Error:  nil,
			},
		},
		"append to non-existing key": {
			setup: func() {
				// No setup needed as we are testing a non-existing document.
			},
			input: []string{"non_existing_doc", "$..a", "\"err\""},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrKeyDoesNotExist,
			},
		},
		"append to root node": {
			setup: func() {
				key := "doc1"
				value := "\"abcd\""
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(value), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input: []string{"doc1", "$", "\"piu\""},
			migratedOutput: EvalResponse{
				Result: 7,
				Error:  nil,
			},
		},
	}

	// Run the tests
	runMigratedEvalTests(t, tests, evalJSONSTRAPPEND, store)
}

func BenchmarkEvalJSONSTRAPPEND(b *testing.B) {
	store := dstore.NewStore(nil, nil)

	// Setup a sample JSON document
	key := "doc1"
	value := "{\"a\":\"foo\", \"nested1\": {\"a\": \"hello\"}, \"nested2\": {\"a\": 31}}"
	var rootData interface{}
	_ = sonic.Unmarshal([]byte(value), &rootData)
	obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
	store.Put(key, obj)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Benchmark appending to multiple fields
		evalJSONSTRAPPEND([]string{"doc1", "$..a", "\"bar\""}, store)
	}
}

func testEvalZPOPMAX(t *testing.T, store *dstore.Store) {
	setup := func() {
		evalZADD([]string{"myzset", "1", "member1", "2", "member2", "3", "member3"}, store)
	}

	tests := map[string]evalTestCase{
		"ZPOPMAX without key": {
			setup: func() {},
			input: []string{"KEY_INVALID"},
			migratedOutput: EvalResponse{
				Result: []string{},
				Error:  nil,
			},
		},
		"ZPOPMAX on wrongtype of key": {
			setup: func() {
				evalSET([]string{"mystring", "shankar"}, store)
			},
			input: []string{"mystring", "1"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongTypeOperation,
			},
		},
		"ZPOPMAX without count argument": {
			setup: setup,
			input: []string{"myzset"},
			migratedOutput: EvalResponse{
				Result: []string{"3", "member3"},
				Error:  nil,
			},
		},
		"ZPOPMAX with count argument": {
			setup: setup,
			input: []string{"myzset", "2"},
			migratedOutput: EvalResponse{
				Result: []string{"3", "member3", "2", "member2"},
				Error:  nil,
			},
		},
		"ZPOPMAX with count more than the elements in sorted set": {
			setup: setup,
			input: []string{"myzset", "4"},
			migratedOutput: EvalResponse{
				Result: []string{"3", "member3", "2", "member2", "1", "member1"},
			},
		},
		"ZPOPMAX with count as zero": {
			setup: setup,
			input: []string{"myzsert", "0"},
			migratedOutput: EvalResponse{
				Result: []string{},
				Error:  nil,
			},
		},
		"ZPOPMAX on an empty sorted set": {
			setup: func() {
				store.Put("myzset", store.NewObj(sortedset.New(), -1, object.ObjTypeSortedSet))
			},
			input: []string{"myzset"},
			migratedOutput: EvalResponse{
				Result: []string{},
				Error:  nil,
			},
		},
	}
	runMigratedEvalTests(t, tests, evalZPOPMAX, store)
}

func BenchmarkEvalZPOPMAX(b *testing.B) {
	// Define benchmark cases with varying sizes of sorted sets
	benchmarks := []struct {
		name  string
		setup func(store *dstore.Store)
		input []string
	}{
		{
			name: "ZPOPMAX on small sorted set (10 members)",
			setup: func(store *dstore.Store) {
				evalZADD([]string{"sortedSet", "1", "member1", "2", "member2", "3", "member3", "4", "member4", "5", "member5", "6", "member6", "7", "member7", "8", "member8", "9", "member9", "10", "member10"}, store)
			},
			input: []string{"sortedSet", "3"},
		},
		{
			name: "ZPOPMAX on large sorted set (10000 members)",
			setup: func(store *dstore.Store) {
				args := []string{"sortedSet"}
				for i := 1; i <= 10000; i++ {
					args = append(args, fmt.Sprintf("%d", i), fmt.Sprintf("member%d", i))
				}
				evalZADD(args, store)
			},
			input: []string{"sortedSet", "10"},
		},
		{
			name: "ZPOPMAX with large sorted set with duplicate scores",
			setup: func(store *dstore.Store) {
				args := []string{"sortedSet"}
				for i := 1; i <= 10000; i++ {
					args = append(args, "1", fmt.Sprintf("member%d", i))
				}
				evalZADD(args, store)
			},
			input: []string{"sortedSet", "2"},
		},
	}

	store := dstore.NewStore(nil, nil)

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			bm.setup(store)

			for i := 0; i < b.N; i++ {
				// Reset the store before each run to avoid contamination
				dstore.Reset(store)
				bm.setup(store)
				evalZPOPMAX(bm.input, store)
			}
		})
	}
}
func BenchmarkZCOUNT(b *testing.B) {
	store := dstore.NewStore(nil, nil)

	// Populate the sorted set with some members for basic benchmarks
	evalZADD([]string{"key", "10", "member1", "20", "member2", "30", "member3"}, store)

	// Benchmark for basic ZCOUNT
	b.Run("Basic ZCOUNT", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			evalZCOUNT([]string{"key", "10", "30"}, store) // Count members with scores between 10 and 30
		}
	})

	// Benchmark for large ZCOUNT
	b.Run("Large ZCOUNT", func(b *testing.B) {
		// Setup a large sorted set
		for i := 0; i < 10000; i++ {
			evalZADD([]string{"key", fmt.Sprintf("%d", i), fmt.Sprintf("member%d", i)}, store)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			evalZCOUNT([]string{"key", "0", "10000"}, store) // Count all members
		}
	})

	// Benchmark for edge cases
	b.Run("Edge Case ZCOUNT", func(b *testing.B) {
		// Reset the store and set up members
		store = dstore.NewStore(nil, nil)
		evalZADD([]string{"key", "5", "member1", "15", "member2", "25", "member3"}, store)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			evalZCOUNT([]string{"key", "-inf", "+inf"}, store) // Count all members
			evalZCOUNT([]string{"key", "10", "10"}, store)     // Count boundary member
			evalZCOUNT([]string{"key", "100", "200"}, store)   // Count out-of-range
		}
	})

	// Benchmark for concurrent ZCOUNT
	b.Run("Concurrent ZCOUNT", func(b *testing.B) {
		// Populate the sorted set with some members
		evalZADD([]string{"key", "10", "member1", "20", "member2", "30", "member3"}, store)

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				evalZCOUNT([]string{"key", "0", "100"}, store) // Perform concurrent ZCOUNT
			}
		})
	})
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
				obj := store.NewObj(int64(1), -1, object.ObjTypeInt)
				store.Put(key, obj)
			},
			input:          []string{"KEY2"},
			migratedOutput: EvalResponse{Result: int64(2), Error: nil},
		},
		{
			name: "INCR key holding string value",
			setup: func() {
				key := "KEY3"
				obj := store.NewObj("VAL1", -1, object.ObjTypeString)
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
				obj := store.NewObj(int64(math.MaxInt64), -1, object.ObjTypeInt)
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
					assert.True(t, bytes.Equal(b, expectedBytes), "expected and actual byte slices should be equal")
				}
			} else {
				assert.Equal(t, tt.migratedOutput.Result, response.Result)
			}

			if tt.migratedOutput.Error != nil {
				assert.EqualError(t, response.Error, tt.migratedOutput.Error.Error())
			} else {
				assert.NoError(t, response.Error)
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
				obj := store.NewObj(int64(1), -1, object.ObjTypeInt)
				store.Put(key, obj)
			},
			input:          []string{"KEY2", "3"},
			migratedOutput: EvalResponse{Result: int64(4), Error: nil},
		},
		{
			name: "INCRBY key holding string value",
			setup: func() {
				key := "KEY3"
				obj := store.NewObj("VAL1", -1, object.ObjTypeString)
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
				obj := store.NewObj(int64(math.MaxInt64-3), -1, object.ObjTypeInt)
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
					assert.True(t, bytes.Equal(b, expectedBytes), "expected and actual byte slices should be equal")
				}
			} else {
				assert.Equal(t, tt.migratedOutput.Result, response.Result)
			}

			if tt.migratedOutput.Error != nil {
				assert.EqualError(t, response.Error, tt.migratedOutput.Error.Error())
			} else {
				assert.NoError(t, response.Error)
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
				obj := store.NewObj(int64(1), -1, object.ObjTypeInt)
				store.Put(key, obj)
			},
			input:          []string{"KEY2"},
			migratedOutput: EvalResponse{Result: int64(0), Error: nil},
		},
		{
			name: "DECR key holding string value",
			setup: func() {
				key := "KEY3"
				obj := store.NewObj("VAL1", -1, object.ObjTypeString)
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
				obj := store.NewObj(int64(math.MinInt64), -1, object.ObjTypeInt)
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
					assert.True(t, bytes.Equal(b, expectedBytes), "expected and actual byte slices should be equal")
				}
			} else {
				assert.Equal(t, tt.migratedOutput.Result, response.Result)
			}

			if tt.migratedOutput.Error != nil {
				assert.EqualError(t, response.Error, tt.migratedOutput.Error.Error())
			} else {
				assert.NoError(t, response.Error)
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
				obj := store.NewObj(int64(1), -1, object.ObjTypeInt)
				store.Put(key, obj)
			},
			input:          []string{"KEY2", "3"},
			migratedOutput: EvalResponse{Result: int64(-2), Error: nil},
		},
		{
			name: "DECRBY key holding string value",
			setup: func() {
				key := "KEY3"
				obj := store.NewObj("VAL1", -1, object.ObjTypeString)
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
				obj := store.NewObj(int64(math.MinInt64+3), -1, object.ObjTypeInt)
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
					assert.True(t, bytes.Equal(b, expectedBytes), "expected and actual byte slices should be equal")
				}
			} else {
				assert.Equal(t, tt.migratedOutput.Result, response.Result)
			}

			if tt.migratedOutput.Error != nil {
				assert.EqualError(t, response.Error, tt.migratedOutput.Error.Error())
			} else {
				assert.NoError(t, response.Error)
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
			migratedOutput: EvalResponse{Result: OK, Error: nil},
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
			migratedOutput: EvalResponse{Result: nil, Error: diceerrors.ErrKeyNotFound},
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
			migratedOutput: EvalResponse{Result: IntegerOne, Error: nil},
		},
		{
			name: "BF.ADD to existing filter",
			setup: func() {
				evalBFRESERVE([]string{"myBloomFilter", "0.01", "1000"}, store)
			},
			input:          []string{"myBloomFilter", "element"},
			migratedOutput: EvalResponse{Result: IntegerOne, Error: nil}, // 1 for new addition, 0 if already exists
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
			migratedOutput: EvalResponse{Result: IntegerZero, Error: nil}, // Item does not exist
		},
		{
			name: "BF.EXISTS element not in filter",
			setup: func() {
				evalBFRESERVE([]string{"myBloomFilter", "0.01", "1000"}, store)
			},
			input:          []string{"myBloomFilter", "element"},
			migratedOutput: EvalResponse{Result: IntegerZero, Error: nil},
		},
		{
			name: "BF.EXISTS element in filter",
			setup: func() {
				evalBFRESERVE([]string{"myBloomFilter", "0.01", "1000"}, store)
				evalBFADD([]string{"myBloomFilter", "element"}, store)
			},
			input:          []string{"myBloomFilter", "element"},
			migratedOutput: EvalResponse{Result: IntegerOne, Error: nil}, // 1 indicates the element exists
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

func testEvalLINSERT(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"nil value": {
			input:          nil,
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("-wrong number of arguments for LINSERT")},
		},
		"empty args": {
			input:          []string{},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("-wrong number of arguments for LINSERT")},
		},
		"wrong number of args": {
			input:          []string{"KEY1", "KEY2"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("-wrong number of arguments for LINSERT")},
		},
		"key does not exist": {
			input:          []string{"NONEXISTENT_KEY", "before", "pivot", "element"},
			migratedOutput: EvalResponse{Result: 0, Error: nil},
		},
		"key exists": {
			setup: func() {
				evalLPUSH([]string{"EXISTING_KEY", "mock_value"}, store)
			},
			input:          []string{"EXISTING_KEY", "before", "mock_value", "element"},
			migratedOutput: EvalResponse{Result: int64(2), Error: nil},
		},
		"key with different type": {
			setup: func() {
				evalSET([]string{"EXISTING_KEY", "mock_value"}, store)
			},
			input:          []string{"EXISTING_KEY", "before", "mock_value", "element"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("-WRONGTYPE Operation against a key holding the wrong kind of value")},
		},
	}
	runMigratedEvalTests(t, tests, evalLINSERT, store)
}

func testEvalLRANGE(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"nil value": {
			input:          nil,
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("-wrong number of arguments for LRANGE")},
		},
		"empty args": {
			input:          []string{},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("-wrong number of arguments for LRANGE")},
		},
		"wrong number of args": {
			input:          []string{"KEY1"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("-wrong number of arguments for LRANGE")},
		},
		"invalid start": {
			input:          []string{"KEY1", "014f", "2"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("-ERR value is not an integer or out of range")},
		},
		"invalid stop": {
			input:          []string{"KEY1", "2", "f2"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("-ERR value is not an integer or out of range")},
		},
		"key does not exist": {
			input:          []string{"NONEXISTENT_KEY", "2", "4"},
			migratedOutput: EvalResponse{Result: []string{}, Error: nil},
		},
		"key exists": {
			setup: func() {
				evalLPUSH([]string{"EXISTING_KEY", "pivot_value"}, store)
				evalLINSERT([]string{"EXISTING_KEY", "before", "pivot_value", "before_value"}, store)
				evalLINSERT([]string{"EXISTING_KEY", "after", "pivot_value", "after_value"}, store)
			},
			input:          []string{"EXISTING_KEY", "0", "5"},
			migratedOutput: EvalResponse{Result: []string{"before_value", "pivot_value", "after_value"}, Error: nil},
		},
		"key with different type": {
			setup: func() {
				evalSET([]string{"EXISTING_KEY", "mock_value"}, store)
			},
			input:          []string{"EXISTING_KEY", "0", "4"},
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("-WRONGTYPE Operation against a key holding the wrong kind of value")},
		},
	}
	runMigratedEvalTests(t, tests, evalLRANGE, store)
}

func testEvalJSONARRINDEX(t *testing.T, store *dstore.Store) {
	normalArray := `[0,1,2,3,4,3]`
	tests := []evalTestCase{
		{
			name:  "nil value",
			setup: func() {},
			input: nil,
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("JSON.ARRINDEX"),
			},
		},
		{
			name:  "empty args",
			setup: func() {},
			input: []string{},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("JSON.ARRINDEX"),
			},
		},
		{
			name: "start index is invalid",
			setup: func() {
				key := "EXISTING_KEY"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(normalArray), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input: []string{"EXISTING_KEY", "$", "3", "abc"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  errors.New("ERR Couldn't parse as integer"),
			},
		},
		{
			name: "stop index is invalid",
			setup: func() {
				key := "EXISTING_KEY"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(normalArray), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input: []string{"EXISTING_KEY", "$", "3", "4", "abc"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  errors.New("ERR Couldn't parse as integer"),
			},
		},
		{
			name: "start and stop optional param valid",
			setup: func() {
				key := "EXISTING_KEY"
				var rootData interface{}
				_ = sonic.Unmarshal([]byte(normalArray), &rootData)
				obj := store.NewObj(rootData, -1, object.ObjTypeJSON)
				store.Put(key, obj)
			},
			input: []string{"EXISTING_KEY", "$", "4", "4", "5"},
			migratedOutput: EvalResponse{
				Result: []interface{}{4},
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

			response := evalJSONARRINDEX(tt.input, store)
			assert.Equal(t, tt.migratedOutput.Result, response.Result)
			if tt.migratedOutput.Error != nil {
				assert.EqualError(t, response.Error, tt.migratedOutput.Error.Error())
			} else {
				assert.NoError(t, response.Error)
			}
		})
	}
}

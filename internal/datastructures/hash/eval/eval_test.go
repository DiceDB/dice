package eval

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"testing"

	"github.com/dicedb/dice/internal/clientio"
	ds "github.com/dicedb/dice/internal/datastructures"
	"github.com/dicedb/dice/internal/datastructures/hash"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/stretchr/testify/assert"
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
					fmt.Println(string(b))
					fmt.Println(string(expectedBytes))
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

func setupTest(store *dstore.Store) *dstore.Store {
	dstore.ResetStore(store)
	return store
}

func TestEval(t *testing.T) {
	store := dstore.NewStore(nil, nil, nil)
	testEvalHSET(t, store)
	testEvalHKEYS(t, store)
	testEvalHVALS(t, store)
	testEvalHGET(t, store)
	testEvalHGETALL(t, store)
	testEvalHMGET(t, store)
	testEvalHSTRLEN(t, store)
	testEvalHLEN(t, store)
	testEvalHEXISTS(t, store)
	testEvalHDEL(t, store)
	testEvalHINCRBY(t, store)
	testEvalHINCRBYFLOAT(t, store)
}

func testEvalHSET(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"wrong number of args passed": {
			setup: func() {},
			input: nil,
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  ds.ErrWrongArgumentCount("HSET"),
			},
		},
		"only key passed": {
			setup: func() {},
			input: []string{"key"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  ds.ErrWrongArgumentCount("HSET"),
			},
		},
		"only key and field_name passed": {
			setup: func() {},
			input: []string{"KEY", "field_name"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  ds.ErrWrongArgumentCount("HSET"),
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
				newMap := hash.NewHash()
				newMap.(*hash.Hash).Add(field, "mock_field_value", -1)
				store.Put(key, newMap)
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
				newMap := hash.NewHash()
				newMap.(*hash.Hash).Add(field, mockValue, -1)
				store.Put(key, newMap)
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

func testEvalHKEYS(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{

		"HKEYS wrong number of args passed": {
			setup:          nil,
			input:          nil,
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR wrong number of arguments for 'hkeys' command")},
		},
		"HKEYS key doesn't exist": {
			setup:          nil,
			input:          []string{"KEY"},
			migratedOutput: EvalResponse{Result: clientio.EmptyArray, Error: nil},
		},
		"HKEYS key exists and is a hash": {
			setup: func() {
				key := "KEY_MOCK"
				field1 := "mock_field_name"
				obj := hash.NewHash()
				obj.(*hash.Hash).Add(field1, "mock_field_value", -1)
				store.Put(key, obj)
			},
			input:          []string{"KEY_MOCK"},
			migratedOutput: EvalResponse{Result: []string{"mock_field_name"}, Error: nil},
		},
	}

	runMigratedEvalTests(t, tests, evalHKEYS, store)
}

func testEvalHVALS(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"HVALS wrong number of args passed": {
			setup:          nil,
			input:          nil,
			migratedOutput: EvalResponse{Result: nil, Error: errors.New("ERR wrong number of arguments for 'hvals' command")},
		},
		"HVALS key doesn't exist": {
			setup:          nil,
			input:          []string{"KEY"},
			migratedOutput: EvalResponse{Result: clientio.EmptyArray, Error: nil},
		},
		"HVALS key exists and is a hash": {
			setup: func() {
				key := "KEY_MOCK"
				field1 := "mock_field_name"
				obj := hash.NewHash()
				obj.(*hash.Hash).Add(field1, "mock_field_value", -1)
				store.Put(key, obj)
			},
			input:          []string{"KEY_MOCK"},
			migratedOutput: EvalResponse{Result: []string{"mock_field_value"}, Error: nil},
		},
	}

	runMigratedEvalTests(t, tests, evalHVALS, store)
}

func testEvalHGET(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"wrong number of args passed": {
			setup: func() {},
			input: nil,
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  ds.ErrWrongArgumentCount("HGET"),
			},
		},
		"only key passed": {
			setup: func() {},
			input: []string{"KEY"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  ds.ErrWrongArgumentCount("HGET"),
			},
		},
		"key doesn't exist": {
			setup: func() {},
			input: []string{"KEY", "field_name"},
			migratedOutput: EvalResponse{
				Result: clientio.NIL,
				Error:  nil,
			},
		},
		"key exists but field_name doesn't exist": {
			setup: func() {
				key := "KEY_MOCK"
				field := "mock_field_name"
				obj := hash.NewHash()
				obj.(*hash.Hash).Add(field, "mock_field_value", -1)
				store.Put(key, obj)
			},
			input: []string{"KEY_MOCK", "non_existent_key"},
			migratedOutput: EvalResponse{
				Result: clientio.NIL,
				Error:  nil,
			},
		},
		"both key and field_name exist": {
			setup: func() {
				key := "KEY_MOCK"
				field := "mock_field_name"
				obj := hash.NewHash()
				obj.(*hash.Hash).Add(field, "mock_field_value", -1)
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
				Error:  ds.ErrWrongArgumentCount("HGETALL"),
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
				obj := hash.NewHash()
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
				obj := hash.NewHash()
				obj.(*hash.Hash).Add("field1", "value1", -1)
				obj.(*hash.Hash).Add("field2", "value2", -1)
				obj.(*hash.Hash).Add("field3", "value3", -1)
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
				Error:  ds.ErrWrongArgumentCount("HMGET"),
			},
		},
		"only key passed": {
			setup: func() {},
			input: []string{"KEY"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  ds.ErrWrongArgumentCount("HMGET"),
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
				obj := hash.NewHash()
				obj.(*hash.Hash).Add(field, "mock_field_value", -1)
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
				obj := hash.NewHash()
				obj.(*hash.Hash).Add(field, "mock_field_value", -1)
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
				field1 := "field1"
				field2 := "field2"
				obj := hash.NewHash()
				obj.(*hash.Hash).Add(field1, "value1", -1)
				obj.(*hash.Hash).Add(field2, "value2", -1)
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

func testEvalHSTRLEN(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"wrong number of args passed": {
			setup:          func() {},
			input:          nil,
			migratedOutput: EvalResponse{Result: nil, Error: ds.ErrWrongArgumentCount("HSTRLEN")},
		},
		"only key passed": {
			setup:          func() {},
			input:          []string{"KEY"},
			migratedOutput: EvalResponse{Result: nil, Error: ds.ErrWrongArgumentCount("HSTRLEN")},
		},
		"key doesn't exist": {
			setup:          func() {},
			input:          []string{"KEY", "field_name"},
			migratedOutput: EvalResponse{Result: clientio.IntegerZero, Error: nil},
		},
		"key exists but field_name doesn't exists": {
			setup: func() {
				key := "KEY_MOCK"
				field := "mock_field_name"
				obj := hash.NewHash()
				obj.(*hash.Hash).Add(field, "mock_field_value", -1)
				store.Put(key, obj)
			},
			input:          []string{"KEY_MOCK", "non_existent_key"},
			migratedOutput: EvalResponse{Result: clientio.IntegerZero, Error: nil},
		},
		"both key and field_name exists": {
			setup: func() {
				key := "KEY_MOCK"
				field := "mock_field_name"
				obj := hash.NewHash()
				obj.(*hash.Hash).Add(field, "mock_field_value", -1)
				store.Put(key, obj)
			},
			input:          []string{"KEY_MOCK", "mock_field_name"},
			migratedOutput: EvalResponse{Result: 16, Error: nil},
		},
	}

	runMigratedEvalTests(t, tests, evalHSTRLEN, store)
}

func testEvalHEXISTS(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"HEXISTS wrong number of args passed": {
			setup: nil,
			input: nil,
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  ds.ErrWrongArgumentCount("HEXISTS"),
			},
		},
		"HEXISTS only key passed": {
			setup: nil,
			input: []string{"KEY"},
			migratedOutput: EvalResponse{Result: nil,
				Error: ds.ErrWrongArgumentCount("HEXISTS")},
		},
		"HEXISTS key doesn't exist": {
			setup:          nil,
			input:          []string{"KEY", "field_name"},
			migratedOutput: EvalResponse{Result: clientio.IntegerZero, Error: nil},
		},
		"HEXISTS key exists but field_name doesn't exists": {
			setup: func() {
				key := "KEY_MOCK"
				field := "mock_field_name"
				obj := hash.NewHash()
				obj.(*hash.Hash).Add(field, "mock_field_value", -1)
				store.Put(key, obj)
			},
			input:          []string{"KEY_MOCK", "non_existent_key"},
			migratedOutput: EvalResponse{Result: clientio.IntegerZero, Error: nil},
		},
		"HEXISTS both key and field_name exists": {
			setup: func() {
				key := "KEY_MOCK"
				field := "mock_field_name"
				obj := hash.NewHash()
				obj.(*hash.Hash).Add(field, "mock_field_value", -1)
				store.Put(key, obj)
			},
			input:          []string{"KEY_MOCK", "mock_field_name"},
			migratedOutput: EvalResponse{Result: clientio.IntegerOne, Error: nil},
		},
	}

	runMigratedEvalTests(t, tests, evalHEXISTS, store)
}

func testEvalHDEL(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"HDEL with wrong number of args": {
			input: []string{"key"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  ds.ErrWrongArgumentCount("HDEL"),
			},
		},
		"HDEL with key does not exist": {
			input: []string{"nonexistent", "field"},
			migratedOutput: EvalResponse{
				Result: int64(0),
				Error:  nil,
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

func testEvalHLEN(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"HLEN wrong number of args": {
			input:          []string{},
			migratedOutput: EvalResponse{Result: nil, Error: ds.ErrWrongArgumentCount("HLEN")},
		},
		"HLEN non-existent key": {
			input:          []string{"nonexistent_key"},
			migratedOutput: EvalResponse{Result: clientio.IntegerZero, Error: nil},
		},
		"HLEN empty hash": {
			setup:          func() {},
			input:          []string{"empty_hash"},
			migratedOutput: EvalResponse{Result: clientio.IntegerZero, Error: nil},
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

func testEvalHINCRBY(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"invalid number of args passed": {
			setup:          func() {},
			input:          nil,
			migratedOutput: EvalResponse{Result: nil, Error: ds.ErrWrongArgumentCount("HINCRBY")},
		},
		"only key is passed in args": {
			setup:          func() {},
			input:          []string{"key"},
			migratedOutput: EvalResponse{Result: nil, Error: ds.ErrWrongArgumentCount("HINCRBY")},
		},
		"only key and field is passed in args": {
			setup:          func() {},
			input:          []string{"key field"},
			migratedOutput: EvalResponse{Result: nil, Error: ds.ErrWrongArgumentCount("HINCRBY")},
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
				obj := hash.NewHash()
				obj.(*hash.Hash).Add(field, "10", -1)
				store.Put(key, obj)
			},
			input:          []string{"key", "field", "10"},
			migratedOutput: EvalResponse{Result: int64(20), Error: nil},
		},
		"increment value is not int64": {
			setup:          func() {},
			input:          []string{"key", "field", "hello"},
			migratedOutput: EvalResponse{Result: nil, Error: ds.ErrIntegerOutOfRange},
		},
		"increment value is greater than the bound of int64": {
			setup:          func() {},
			input:          []string{"key", "field", "99999999999999999999999999999999999999999999999999999"},
			migratedOutput: EvalResponse{Result: nil, Error: ds.ErrIntegerOutOfRange},
		},
		"update the existing field which has spaces": {
			setup: func() {
				key := "key"
				field := "field"
				obj := hash.NewHash()
				obj.(*hash.Hash).Add(field, "   10   ", -1)
				store.Put(key, obj)
			},
			input:          []string{"key", "field", "10"},
			migratedOutput: EvalResponse{Result: nil, Error: ds.ErrHashValueNotInteger},
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
				obj := hash.NewHash()
				obj.(*hash.Hash).Add(field, "-10", -1)
				store.Put(key, obj)
			},
			input:          []string{"key", "field", "-10"},
			migratedOutput: EvalResponse{Result: int64(-20), Error: nil},
		},
		"updating the existing field which would lead to positive overflow": {
			setup: func() {
				key := "key"
				field := "field"
				obj := hash.NewHash()
				obj.(*hash.Hash).Add(field, fmt.Sprintf("%v", math.MaxInt64), -1)
				fmt.Sprintf("%v", math.MaxInt64)
				store.Put(key, obj)
			},
			input:          []string{"key", "field", "10"},
			migratedOutput: EvalResponse{Result: nil, Error: ds.ErrOverflow},
		},
		"updating the existing field which would lead to negative overflow": {
			setup: func() {
				key := "key"
				field := "field"
				obj := hash.NewHash()
				obj.(*hash.Hash).Add(field, fmt.Sprintf("%v", math.MinInt64), -1)
				store.Put(key, obj)
			},
			input:          []string{"key", "field", "-10"},
			migratedOutput: EvalResponse{Result: nil, Error: ds.ErrOverflow},
		},
	}

	runMigratedEvalTests(t, tests, evalHINCRBY, store)
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
				obj := hash.NewHash()
				store.Put(key, obj)
			},
			input:          []string{"key", "field", "0.1"},
			migratedOutput: EvalResponse{Result: "0.1", Error: nil},
		},
		"HINCRBYFLOAT on an existing key and field with a float value": {
			setup: func() {
				key := "key"
				field := "field"
				obj := hash.NewHash()
				obj.(*hash.Hash).Add(field, "2.1", -1)
				store.Put(key, obj)
			},
			input:          []string{"key", "field", "0.1"},
			migratedOutput: EvalResponse{Result: "2.2", Error: nil},
		},
		"HINCRBYFLOAT on an existing key and field with an integer value": {
			setup: func() {
				key := "key"
				field := "field"
				obj := hash.NewHash()
				obj.(*hash.Hash).Add(field, "2", -1)
				store.Put(key, obj)
			},
			input:          []string{"key", "field", "0.1"},
			migratedOutput: EvalResponse{Result: "2.1", Error: nil},
		},
		"HINCRBYFLOAT with a negative increment": {
			setup: func() {
				key := "key"
				field := "field"
				obj := hash.NewHash()
				obj.(*hash.Hash).Add(field, "2", -1)
				store.Put(key, obj)
			},
			input:          []string{"key", "field", "-0.1"},
			migratedOutput: EvalResponse{Result: "1.9", Error: nil},
		},
		"HINCRBYFLOAT by a non-numeric increment": {
			setup: func() {
				key := "key"
				field := "field"
				obj := hash.NewHash()
				obj.(*hash.Hash).Add(field, "2", -1)
				store.Put(key, obj)
			},
			input:          []string{"key", "field", "a"},
			output:         []byte("-ERR value is not an integer or a float\r\n"),
			migratedOutput: EvalResponse{Result: nil, Error: ds.ErrInvalidNumberFormat},
		},
		"HINCRBYFLOAT on a field with non-numeric value": {
			setup: func() {
				key := "key"
				field := "field"
				obj := hash.NewHash()
				obj.(*hash.Hash).Add(field, "non_numeric", -1)
				store.Put(key, obj)
			},
			input:          []string{"key", "field", "0.1"},
			migratedOutput: EvalResponse{Result: nil, Error: ds.ErrInvalidNumberFormat},
		},
		"HINCRBYFLOAT by a value that would turn float64 to Inf": {
			setup: func() {
				key := "key"
				field := "field"
				obj := hash.NewHash()
				obj.(*hash.Hash).Add(field, fmt.Sprintf("%v", math.MaxFloat64), -1)
				store.Put(key, obj)
			},
			input:          []string{"key", "field", "1e308"},
			migratedOutput: EvalResponse{Result: nil, Error: ds.ErrOverflow},
		},
		"HINCRBYFLOAT with scientific notation": {
			setup: func() {
				key := "key"
				field := "field"
				obj := hash.NewHash()
				obj.(*hash.Hash).Add(field, "1e2", -1)
				store.Put(key, obj)
			},
			input:          []string{"key", "field", "1e-1"},
			migratedOutput: EvalResponse{Result: "100.1", Error: nil},
		},
	}

	runMigratedEvalTests(t, tests, evalHINCRBYFLOAT, store)
}

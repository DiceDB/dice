package async

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestHINCRBY(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	defer FireCommand(conn, "FLUSHDB")

	testcases := []TestCase{
		{
			name:     "HINCRBY Wrong number of arguments provided",
			commands: []string{"HINCRBY", "HINCRBY key", "HINCRBY key field"},
			expected: []interface{}{"ERR wrong number of arguments for 'hincrby' command", "ERR wrong number of arguments for 'hincrby' command", "ERR wrong number of arguments for 'hincrby' command"},
		},
		{
			name:     "HINCRBY should increment when key doesn't exist",
			commands: []string{"HINCRBY key field-1 10"},
			expected: []interface{}{int64(10)},
		},
		{
			name:     "HINCRBY should increment when key exists and a field doesn't exist",
			commands: []string{"HSET new-key field-1 10", "HINCRBY new-key field-2 10"},
			expected: []interface{}{int64(1), int64(10)},
		},
		{
			name:     "HINCRBY should increment on existing key and field",
			commands: []string{"HINCRBY key field-1 10"},
			expected: []interface{}{int64(20)},
		},
		{
			name:     "HINCRBY should give error when trying to increment a value which is not integer",
			commands: []string{"HSET key field value", "HINCRBY key field 20"},
			expected: []interface{}{int64(1), "ERR hash value is not an integer"},
		},
		{
			name:     "HINCRBY should give error when increment value is greater than max int64 value",
			commands: []string{"HINCRBY key field 9999999999999999999999999999999999999"},
			expected: []interface{}{"ERR value is not an integer or out of range"},
		},
		{
			name:     "HINCRBY should give error when increment value is not an integer",
			commands: []string{"HINCRBY key field value"},
			expected: []interface{}{"ERR value is not an integer or out of range"},
		},
		{
			name:     "HINCRBY should give error when trying to increment a key which is not a hash value",
			commands: []string{"SET key value", "HINCRBY key value 10"},
			expected: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
		{
			name:     "HINCRBY should give error when the increment value adds up to overflow",
			commands: []string{"HSET new-key value 9000000000000000000", "HINCRBY new-key value 1000000000000000000"},
			expected: []interface{}{int64(1), "ERR increment or decrement would overflow"},
		},
		{
			name:     "HINCRBY should give error when the value decrements to overflow",
			commands: []string{"HSET new-key value -9000000000000000000", "HINCRBY new-key value -1000000000000000000"},
			expected: []interface{}{int64(0), "ERR increment or decrement would overflow"},
		},
		{
			name:     "HINCRBY should decrement the value",
			commands: []string{"HSET new-key value 10", "HINCRBY new-key value -1"},
			expected: []interface{}{int64(0), int64(9)},
		},
		{
			name:     "HINCRBY should give integer error when trying to increment a key which is not a hash value with a value which is not integer",
			commands: []string{"SET key value", "HINCRBY key value ten"},
			expected: []interface{}{"OK", "ERR value is not an integer or out of range"},
		},
	}

	for _, tc := range testcases {
		for i, cmd := range tc.commands {
			result := FireCommand(conn, cmd)
			assert.DeepEqual(t, tc.expected[i], result)
		}
	}
}

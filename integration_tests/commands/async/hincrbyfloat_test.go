package async

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestHINCRBYFLOAT(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	defer FireCommand(conn, "FLUSHDB")
	defer FireCommand(conn, "DEL key new-key")

	testcases := []TestCase{
		{
			name:     "HINCRBYFLOAT Wrong number of arguments provided",
			commands: []string{"HINCRBYFLOAT", "HINCRBYFLOAT key", "HINCRBYFLOAT key field"},
			expected: []interface{}{"ERR wrong number of arguments for 'hincrbyfloat' command", "ERR wrong number of arguments for 'hincrbyfloat' command", "ERR wrong number of arguments for 'hincrbyfloat' command"},
		},
		{
			name:     "HINCRBYFLOAT should increment when key doesn't exist",
			commands: []string{"HINCRBYFLOAT key field-1 10.2"},
			expected: []interface{}{"10.2"},
		},
		{
			name:     "HINCRBYFLOAT should increment when key exists and a field doesn't exist",
			commands: []string{"HSET new-key field-1 10", "HINCRBYFLOAT new-key field-2 10.2"},
			expected: []interface{}{int64(1), "10.2"},
		},
		{
			name:     "HINCRBYFLOAT should increment on existing key and field",
			commands: []string{"HINCRBYFLOAT key field-1 10.2"},
			expected: []interface{}{"20.4"},
		},
		{
			name:     "HINCRBYFLOAT should give error when trying to increment a value which is not integer or float",
			commands: []string{"HSET key field value", "HINCRBYFLOAT key field 20.2"},
			expected: []interface{}{int64(1), "ERR value is not an integer or a float"},
		},
		{
			name:     "HINCRBYFLOAT should give error when increment value is not an integer",
			commands: []string{"HINCRBYFLOAT key field value"},
			expected: []interface{}{"ERR value is not an integer or a float"},
		},
		{
			name:     "HINCRBYFLOAT should give error when trying to increment a key which is not a hash value",
			commands: []string{"SET key value", "HINCRBYFLOAT key value 10"},
			expected: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
		{
			name:     "HINCRBYFLOAT should decrement the value",
			commands: []string{"HSET new-key value 10.2", "HINCRBYFLOAT new-key value -1.5"},
			expected: []interface{}{int64(1), "8.7"},
		},
		{
			name:     "HINCRBYFLOAT should give integer error when trying to increment a key which is not a hash value with a value which is not integer or a float",
			commands: []string{"SET key value", "HINCRBYFLOAT key value ten"},
			expected: []interface{}{"OK", "ERR value is not an integer or a float"},
		},
	}

	for _, tc := range testcases {
		for i, cmd := range tc.commands {
			result := FireCommand(conn, cmd)
			assert.DeepEqual(t, tc.expected[i], result)
		}
	}
}

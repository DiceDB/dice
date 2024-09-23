package async

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestHSTRLEN(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()
	defer FireCommand(conn, "DEL key_hStrLen1 key_hStrLen2 key_hStrLen3 key")

	testCases := []TestCase{
		{
			commands: []string{"HSTRLEN", "HSTRLEN KEY", "HSTRLEN KEY FIELD ANOTHER_FIELD"},
			expected: []interface{}{"ERR wrong number of arguments for 'hstrlen' command",
				"ERR wrong number of arguments for 'hstrlen' command",
				"ERR wrong number of arguments for 'hstrlen' command"},
		},
		{
			commands: []string{"HSET key_hStrLen1 field value", "HSTRLEN wrong_key_hStrLen field"},
			expected: []interface{}{int64(1), int64(0)},
		},
		{
			commands: []string{"HSET key_hStrLen2 field value", "HSTRLEN key_hStrLen2 wrong_field"},
			expected: []interface{}{int64(1), int64(0)},
		},
		{
			commands: []string{"HSET key_hStrLen3 field HelloWorld", "HSTRLEN key_hStrLen3 field"},
			expected: []interface{}{int64(1), int64(10)},
		},
		{
			commands: []string{"SET key value", "HSTRLEN key field"},
			expected: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
	}

	for _, tc := range testCases {
		for i, cmd := range tc.commands {
			result := FireCommand(conn, cmd)
			assert.DeepEqual(t, tc.expected[i], result)
		}
	}
}

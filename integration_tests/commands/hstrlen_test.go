package commands

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestHSTRLEN(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()
	defer FireCommand(conn, "DEL key_hStrLen key")

	testCases := []TestCase{
		{
			commands: []string{"HSTRLEN", "HSTRLEN KEY", "HSTRLEN KEY FIELD ANOTHER_FIELD"},
			expected: []interface{}{"ERR wrong number of arguments for 'hstrlen' command",
				"ERR wrong number of arguments for 'hstrlen' command",
				"ERR wrong number of arguments for 'hstrlen' command"},
		},
		{
			commands: []string{"HSET key_hStrLen field value", "HSET key_hStrLen field HelloWorld"},
			expected: []interface{}{ONE, ZERO},
		},
		{
			commands: []string{"HSTRLEN wrong_key_hStrLen field"},
			expected: []interface{}{ZERO},
		},
		{
			commands: []string{"HSTRLEN key_hStrLen wrong_field"},
			expected: []interface{}{ZERO},
		},
		{
			commands: []string{"HSTRLEN key_hStrLen field"},
			expected: []interface{}{int64(10)},
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

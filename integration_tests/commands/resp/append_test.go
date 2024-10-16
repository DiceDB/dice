package resp

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestAPPEND(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	testCases := []TestCase{
		{
			name:     "APPEND and GET a new Val",
			commands: []string{"APPEND k newVal", "GET k"},
			expected: []interface{}{int64(6), "newVal"},
		},
		{
			name:     "APPEND to an existing key and GET",
			commands: []string{"SET k Bhima", "APPEND k Shankar", "GET k"},
			expected: []interface{}{"OK", int64(12), "BhimaShankar"},
		},
		{
			name:     "APPEND without input value",
			commands: []string{"APPEND k"},
			expected: []interface{}{"ERR wrong number of arguments for 'append' command"},
		},
		{
			name:     "APPEND empty string to an existing key with empty string",
			commands: []string{"SET k \"\"", "APPEND k \"\""},
			expected: []interface{}{"OK", int64(0)},
		},
		{
			name:     "APPEND to key created using LPUSH",
			commands: []string{"LPUSH m bhima", "APPEND m shankar"},
			expected: []interface{}{int64(1), "WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
		{
			name:     "APPEND value with leading zeros",
			commands: []string{"APPEND z 0043"},
			expected: []interface{}{int64(4)},
		},
		{
			name:     "APPEND to key created using SADD",
			commands: []string{"SADD key apple", "APPEND key banana"},
			expected: []interface{}{int64(1), "WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// deleteTestKeys([]string{"k"}, store)
			FireCommand(conn, "DEL k")

			for i, cmd := range tc.commands {
				result := FireCommand(conn, cmd)
				assert.DeepEqual(t, tc.expected[i], result)
			}
		})
	}
}

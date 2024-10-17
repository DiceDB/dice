package websocket

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestAppend(t *testing.T) {
	exec := NewWebsocketCommandExecutor()
	conn := exec.ConnectToServer()

	testCases := []TestCase{
		{
			name:     "APPEND and GET a new Val",
			commands: []string{"APPEND k newVal", "GET k"},
			expected: []interface{}{float64(6), "newVal"},
		},
		{
			name:     "APPEND to an existing key and GET",
			commands: []string{"SET k Bhima", "APPEND k Shankar", "GET k"},
			expected: []interface{}{"OK", float64(12), "BhimaShankar"},
		},
		{
			name:     "APPEND without input value",
			commands: []string{"APPEND k"},
			expected: []interface{}{"ERR wrong number of arguments for 'append' command"},
		},
		{
			name:     "APPEND to key created using LPUSH",
			commands: []string{"LPUSH z bhima", "APPEND z shankar"},
			expected: []interface{}{float64(1), "WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
		{
			name:     "APPEND value with leading zeros",
			commands: []string{"APPEND key1 0043"},
			expected: []interface{}{float64(4)},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result := exec.FireCommand(conn, cmd)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}
}

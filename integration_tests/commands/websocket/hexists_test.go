package websocket

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestHExists(t *testing.T) {
	exec := NewWebsocketCommandExecutor()

	testCases := []TestCase{
		{
			name:     "WS Check if field exists when k f and v are set",
			commands: []string{"HSET key field value", "HEXISTS key field"},
			expected: []interface{}{float64(1), "1"},
		},
		{
			name:     "WS Check if field exists when k exists but not f and v",
			commands: []string{"HSET key field1 value", "HEXISTS key field"},
			expected: []interface{}{float64(1), "0"},
		},
		{
			name:     "WS Check if field exists when no k,f and v exist",
			commands: []string{"HEXISTS key field"},
			expected: []interface{}{"0"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			conn := exec.ConnectToServer()
			exec.FireCommand(conn, "HDEL key field")

			for i, cmd := range tc.commands {
				result := exec.FireCommand(conn, cmd)
				assert.DeepEqual(t, tc.expected[i], result)
			}
		})
	}
}

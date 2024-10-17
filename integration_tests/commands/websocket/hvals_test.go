package websocket

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestHVals(t *testing.T) {
	exec := NewWebsocketCommandExecutor()

	testCases := []TestCase{
		{
			name:     "WS One or more vals exist",
			commands: []string{"HSET key field value", "HSET key field1 value1", "HVALS key"},
			expected: []interface{}{float64(1), float64(1), []interface{}{"value", "value1"}},
		},
		{
			name:     "WS No values exist",
			commands: []string{"HVALS key"},
			expected: []interface{}{[]interface{}{}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			conn := exec.ConnectToServer()
			exec.FireCommand(conn, "HDEL key field")
			exec.FireCommand(conn, "HDEL key field1")

			for i, cmd := range tc.commands {
				result := exec.FireCommand(conn, cmd)
				assert.DeepEqual(t, tc.expected[i], result)
			}
		})
	}
}

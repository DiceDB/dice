package websocket

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestExists(t *testing.T) {
	exec := NewWebsocketCommandExecutor()

	testCases := []TestCase{
		{
			name:     "Test EXISTS command",
			commands: []string{"SET key value", "EXISTS key", "EXISTS key2"},
			expected: []interface{}{"OK", int64(1), int64(0)},
		},
		{
			name:     "Test EXISTS command with multiple keys",
			commands: []string{"SET key value", "SET key2 value2", "EXISTS key key2 key3", "EXISTS key key2 key3 key4", "DEL key", "EXISTS key key2 key3 key4"},
			expected: []interface{}{"OK", "OK", int64(2), int64(2), int64(1), int64(1)},
		},
		{
			name:     "Test EXISTS an expired key",
			commands: []string{"SET key value ex 1", "EXISTS key", "EXISTS key"},
			expected: []interface{}{"OK", int64(1), int64(0)},
		},
		{
			name:     "Test EXISTS with multiple keys and expired key",
			commands: []string{"SET key value ex 2", "SET key2 value2", "SET key3 value3", "EXISTS key key2 key3", "EXISTS key key2 key3"},
			expected: []interface{}{"OK", "OK", "OK", int64(3), int64(2)},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			conn := exec.ConnectToServer()
			for i, cmd := range tc.commands {
				result := exec.FireCommand(conn, cmd)
				assert.DeepEqual(t, tc.expected[i], result)
			}
		})
	}
}

package websocket

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestGETRANGE(t *testing.T) {
	exec := NewWebsocketCommandExecutor()
	conn := exec.ConnectToServer()

	testCases := []TestCase{
		{
			name:     "Get range on a string",
			commands: []string{"SET k1 shankar", "GETRANGE k1 0 7"},
			expected: []interface{}{"OK", "shankar"},
		},
		{
			name:     "Get range on a non existent key",
			commands: []string{"GETRANGE k2 0 7"},
			expected: []interface{}{""},
		},
		{
			name:     "Get range on wrong key type",
			commands: []string{"LPUSH k3 shankar", "GETRANGE k3 0 7"},
			expected: []interface{}{float64(1), "WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
		{
			name:     "GETRANGE against string value: 0, -1",
			commands: []string{"GETRANGE k1 0 -1"},
			expected: []interface{}{"shankar"},
		},
		{
			name:     "GETRANGE against string value: 5, 3",
			commands: []string{"GETRANGE k1 5 3"},
			expected: []interface{}{""},
		},
		{
			name:     "GETRANGE against integer value: -1, -100",
			commands: []string{"GETRANGE k1 -1 -100"},
			expected: []interface{}{""},
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

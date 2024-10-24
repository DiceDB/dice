package websocket

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDel(t *testing.T) {
	exec := NewWebsocketCommandExecutor()

	testCases := []TestCase{
		{
			name:     "DEL with set key",
			commands: []string{"SET k1 v1", "DEL k1", "GET k1"},
			expected: []interface{}{"OK", float64(1), "(nil)"},
		},
		{
			name:     "DEL with multiple keys",
			commands: []string{"SET k1 v1", "SET k2 v2", "DEL k1 k2", "GET k1", "GET k2"},
			expected: []interface{}{"OK", "OK", float64(2), "(nil)", "(nil)"},
		},
		{
			name:     "DEL with key not set",
			commands: []string{"GET k3", "DEL k3"},
			expected: []interface{}{"(nil)", float64(0)},
		},
		{
			name:     "DEL with no keys or arguments",
			commands: []string{"DEL"},
			expected: []interface{}{"ERR wrong number of arguments for 'del' command"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			conn := exec.ConnectToServer()
			for i, cmd := range tc.commands {
				result, err := exec.FireCommandAndReadResponse(conn, cmd)
				assert.Nil(t, err)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}
}

package websocket

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppend(t *testing.T) {
	exec := NewWebsocketCommandExecutor()

	testCases := []struct {
		name       string
		commands   []string
		expected   []interface{}
		cleanupKey string
	}{
		{
			name:       "APPEND and GET a new Val",
			commands:   []string{"APPEND k newVal", "GET k"},
			expected:   []interface{}{float64(6), "newVal"},
			cleanupKey: "k",
		},
		{
			name:       "APPEND to an existing key and GET",
			commands:   []string{"SET k Bhima", "APPEND k Shankar", "GET k"},
			expected:   []interface{}{"OK", float64(12), "BhimaShankar"},
			cleanupKey: "k",
		},
		{
			name:       "APPEND without input value",
			commands:   []string{"APPEND k"},
			expected:   []interface{}{"ERR wrong number of arguments for 'append' command"},
			cleanupKey: "k",
		},
		{
			name:       "APPEND to key created using LPUSH",
			commands:   []string{"LPUSH z bhima", "APPEND z shankar"},
			expected:   []interface{}{float64(1), "WRONGTYPE Operation against a key holding the wrong kind of value"},
			cleanupKey: "z",
		},
		{
			name:       "APPEND value with leading zeros",
			commands:   []string{"APPEND key1 0043"},
			expected:   []interface{}{float64(4)},
			cleanupKey: "key1",
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
			DeleteKey(t, conn, exec, tc.cleanupKey)
		})
	}
}

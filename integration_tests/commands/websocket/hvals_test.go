package websocket

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHVals(t *testing.T) {
	exec := NewWebsocketCommandExecutor()

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
		delays   []time.Duration
	}{
		{
			name:     "WS No values exist",
			commands: []string{"HVALS key"},
			expected: []interface{}{"*0"},
			delays:   []time.Duration{3 * time.Second},
		},
		{
			name:     "WS One or more vals exist",
			commands: []string{"HSET key field value", "HVALS key"},
			expected: []interface{}{float64(1), []interface{}{"value"}},
			delays:   []time.Duration{0, 0, 3 * time.Second},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			conn := exec.ConnectToServer()
			exec.FireCommandAndReadResponse(conn, "HDEL key field")
			exec.FireCommandAndReadResponse(conn, "HDEL key field1")
			exec.FireCommandAndReadResponse(conn, "DEL key")

			for i, cmd := range tc.commands {
				t.Logf("Executing command: %s", cmd)
				if tc.delays[i] > 0 {
					time.Sleep(tc.delays[i])
				}

				result, err := exec.FireCommandAndReadResponse(conn, cmd)
				t.Logf("Received result: %v for command: %s", result, cmd)
				assert.Nil(t, err, "Error encountered while executing command %s: %v", cmd, err)
				assert.Equal(t, tc.expected[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}
}

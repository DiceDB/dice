package websocket

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHExists(t *testing.T) {
	exec := NewWebsocketCommandExecutor()

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
		delays   []time.Duration
	}{
		// {
		// 	name:     "WS Check if field exists when k f and v are set",
		// 	commands: []string{"HSET key field value", "HEXISTS key field"},
		// 	expected: []interface{}{float64(1), "1"},
		// 	delays:   []time.Duration{0, 0},
		// },
		// {
		// 	name:     "WS Check if field exists when k exists but not f and v",
		// 	commands: []string{"HSET key field1 value", "HEXISTS key field"},
		// 	expected: []interface{}{float64(1), "0"},
		// 	delays:   []time.Duration{0, 0},
		// },
		// {
		// 	name:     "WS Check if field exists when no k,f and v exist",
		// 	commands: []string{"HEXISTS key field"},
		// 	expected: []interface{}{"0"},
		// 	delays:   []time.Duration{0},
		// },
		{
			name:     "WS Check if field exists when k f and v are set",
			commands: []string{"HSET key field value", "HEXISTS key field"},
			// Adjusted expected values to match the actual float64 type
			expected: []interface{}{float64(1), float64(1)},
			delays:   []time.Duration{0, 0},
		},
		{
			name:     "WS Check if field exists when k exists but not f and v",
			commands: []string{"HSET key field1 value", "HEXISTS key field"},
			expected: []interface{}{float64(1), float64(0)},
			delays:   []time.Duration{0, 0},
		},
		{
			name:     "WS Check if field exists when no k,f and v exist",
			commands: []string{"HEXISTS key field"},
			expected: []interface{}{float64(0)},
			delays:   []time.Duration{0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			conn := exec.ConnectToServer()

			t.Log("Clearing keys before test execution")
			exec.FireCommandAndReadResponse(conn, "HDEL key field")
			exec.FireCommandAndReadResponse(conn, "HDEL key field1")

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

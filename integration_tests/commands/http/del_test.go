package http

import (
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func TestDel(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []struct {
		name     string
		commands []HTTPCommand
		expected []interface{}
		delays   []time.Duration
	}{
		{
			name: "DEL with set key",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "k1", "value": "v1"}},
				{Command: "DEL", Body: map[string]interface{}{"key": "k1"}},
				{Command: "GET", Body: map[string]interface{}{"key": "k1"}},
			},
			expected: []interface{}{"OK", int64(1), "(nil)"},
			delays:   []time.Duration{0, 0},
		},
		{
			name: "DEL with multiple keys",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "k1", "value": "v1"}},
				{Command: "SET", Body: map[string]interface{}{"key": "k2", "value": "v2"}},
				{Command: "DEL", Body: map[string]interface{}{"key": "k1"}},
				{Command: "DEL", Body: map[string]interface{}{"key": "k2"}},
				{Command: "GET", Body: map[string]interface{}{"key": "k1"}},
				{Command: "GET", Body: map[string]interface{}{"key": "k2"}},
			},
			expected: []interface{}{"OK", "OK", int64(1), int64(1), "(nil)", "(nil)"},
		},
		{
			name: "DEL with key not set",
			commands: []HTTPCommand{
				{Command: "GET", Body: map[string]interface{}{"key": "k3"}},
				{Command: "DEL", Body: map[string]interface{}{"key": "k3"}},
			},
			expected: []interface{}{"(nil)", int64(0)},
		},
		{
			name: "DEL with no keys or arguments",
			commands: []HTTPCommand{
				{Command: "DEL", Body: map[string]interface{}{}},
			},
			expected: []interface{}{"ERR wrong number of arguments for 'del' command"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				if tc.delays[i] > 0 {
					time.Sleep(tc.delays[i])
				}
				result, err := exec.FireCommand(cmd)
				if err != nil {
					// Check if the error message matches the expected result
					assert.Equal(t, tc.expected[i], err.Error(), "Error message mismatch for cmd %s", cmd)
				} else {
					assert.Equal(t, tc.expected[i], result, "Value mismatch for cmd %s, expected %v, got %v", cmd, tc.expected[i], result)
				}
			}
		})
	}
}

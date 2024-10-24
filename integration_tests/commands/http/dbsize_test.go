package http

import (
	"gotest.tools/v3/assert"
	"testing"
	"time"
)

func TestDBSize(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"keys": [...]string{"key1", "key2", "key3"}}})

	testCases := []TestCase{
		{
			name: "DBSIZE with 3 keys",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "k1", "value": "v1"}},
				{Command: "SET", Body: map[string]interface{}{"key": "k2", "value": "v2"}},
				{Command: "SET", Body: map[string]interface{}{"key": "k3", "value": "v3"}},
				{Command: "DBSIZE", Body: nil},
			},
			expected: []interface{}{"OK", "OK", "OK", int64(3)},
		},
		{
			name: "DBSIZE with repetitive keys",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "k1", "value": "v1"}},
				{Command: "SET", Body: map[string]interface{}{"key": "k2", "value": "v2"}},
				{Command: "SET", Body: map[string]interface{}{"key": "k3", "value": "v3"}},
				{Command: "SET", Body: map[string]interface{}{"key": "k1", "value": "v3"}},
				{Command: "SET", Body: map[string]interface{}{"key": "k2", "value": "v22"}},
				{Command: "DBSIZE", Body: nil},
			},
			expected: []interface{}{"OK", "OK", "OK", "OK", "OK", int64(3)},
		},
		{
			name: "DBSIZE with expired keys",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "k1", "value": "v1"}},
				{Command: "SET", Body: map[string]interface{}{"key": "k2", "value": "v2"}},
				{Command: "SET", Body: map[string]interface{}{"key": "k3", "value": "v3"}},
				{Command: "DBSIZE", Body: nil},
				{Command: "EXPIRE", Body: map[string]interface{}{"k3": "v3", "seconds": 1}},
				{Command: "SLEEP", Body: map[string]interface{}{"key": 2}},
				{Command: "DBSIZE", Body: nil},
			},
			expected: []interface{}{"OK", "OK", "OK", int64(3), int64(1), int64(2)},
		},
		{
			name: "DBSIZE after deleting a key",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "k1", "value": "v1"}},
				{Command: "SET", Body: map[string]interface{}{"key": "k2", "value": "v2"}},
				{Command: "SET", Body: map[string]interface{}{"key": "k3", "value": "v3"}},
				{Command: "DBSIZE", Body: nil},
				{Command: "DEL", Body: map[string]interface{}{"key": "k1"}},
				{Command: "DBSIZE", Body: nil},
			},
			expected: []interface{}{"OK", "OK", "OK", int64(3), int64(1), int64(2)},
			delays:   []time.Duration{0, 0, 0, 0, 0, 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"keys": [...]string{"key1", "key2", "key3"}}})

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

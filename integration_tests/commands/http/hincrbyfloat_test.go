package http

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHINCRBYFLOAT(t *testing.T) {
	exec := NewHTTPCommandExecutor()
	exec.FireCommand(HTTPCommand{
		Command: "DEL",
		Body: map[string]interface{}{
			"keys": []interface{}{"key", "new-key"},
		},
	})
	testCases := []struct {
		name          string
		commands      []HTTPCommand
		expected      []interface{}
		delay         []time.Duration
		errorExpected bool
	}{
		{
			name: "HINCRBYFLOAT Wrong number of arguments provided",
			commands: []HTTPCommand{
				{Command: "HINCRBYFLOAT", Body: map[string]interface{}{"key": ""}},
				{Command: "HINCRBYFLOAT", Body: map[string]interface{}{"key": "key"}},
				{Command: "HINCRBYFLOAT", Body: map[string]interface{}{"key": "key", "field": "field"}},
			},
			expected:      []interface{}{"ERR wrong number of arguments for 'hincrbyfloat' command", "ERR wrong number of arguments for 'hincrbyfloat' command", "ERR wrong number of arguments for 'hincrbyfloat' command"},
			delay:         []time.Duration{0, 0, 0},
			errorExpected: true,
		},
		{
			name: "HINCRBYFLOAT should increment when key doesn't exist",
			commands: []HTTPCommand{
				{Command: "HINCRBYFLOAT", Body: map[string]interface{}{"key": "key", "field": "field-1", "value": 10.2}},
			},
			expected:      []interface{}{"10.2"},
			delay:         []time.Duration{0},
			errorExpected: false,
		},
		{
			name: "HINCRBYFLOAT should increment when key exists and a field doesn't exist",
			commands: []HTTPCommand{
				{Command: "HSET", Body: map[string]interface{}{"key": "new-key", "field": "field-1", "value": 10.2}},
				{Command: "HINCRBYFLOAT", Body: map[string]interface{}{"key": "new-key", "field": "field-2", "value": 10.2}},
			},
			expected:      []interface{}{float64(1), "10.2"},
			delay:         []time.Duration{0, 0},
			errorExpected: false,
		},
		{
			name: "HINCRBYFLOAT should increment on existing key and field",
			commands: []HTTPCommand{
				{Command: "HINCRBYFLOAT", Body: map[string]interface{}{"key": "key", "field": "field-1", "value": 10.2}},
			},
			expected:      []interface{}{"20.4"},
			delay:         []time.Duration{0},
			errorExpected: false,
		},
		{
			name: "HINCRBYFLOAT should decrement on existing key and field",
			commands: []HTTPCommand{
				{Command: "HINCRBYFLOAT", Body: map[string]interface{}{"key": "key", "field": "field-1", "value": -10.2}},
			},
			expected:      []interface{}{"10.2"},
			delay:         []time.Duration{0},
			errorExpected: false,
		},
		{
			name: "HINCRBYFLOAT should give error when trying to increment a key which is not a hash value with a value which is not integer or a float",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "key", "value": "value"}},
				{Command: "HINCRBYFLOAT", Body: map[string]interface{}{"key": "key", "field": "field-1", "value": "ten"}},
			},
			expected:      []interface{}{"OK", "ERR value is not an integer or a float"},
			delay:         []time.Duration{0, 0},
			errorExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				if tc.delay[i] > 0 {
					time.Sleep(tc.delay[i])
				}
				result, err := exec.FireCommand(cmd)
				if err != nil {
					// Check if the error message matches the expected result
					assert.Equal(t, tc.expected[i], err.Error(), "Error message mismatch for cmd %s", cmd)
				} else {
					assert.Equal(t, tc.expected[i], result, "field mismatch for cmd %s, expected %v, got %v", cmd, tc.expected[i], result)
				}
			}
		})
	}
}

package http

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHINCRBY(t *testing.T) {
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
			name: "HINCRBY Wrong number of arguments provided",
			commands: []HTTPCommand{
				{Command: "HINCRBY", Body: map[string]interface{}{"key": ""}},
				{Command: "HINCRBY", Body: map[string]interface{}{"key": "key"}},
				{Command: "HINCRBY", Body: map[string]interface{}{"key": "key", "field": "field"}},
			},
			expected:      []interface{}{"ERR wrong number of arguments for 'hincrby' command", "ERR wrong number of arguments for 'hincrby' command", "ERR wrong number of arguments for 'hincrby' command"},
			delay:         []time.Duration{0, 0, 0},
			errorExpected: true,
		},
		{
			name: "HINCRBY should increment when key doesn't exist",
			commands: []HTTPCommand{
				{Command: "HINCRBY", Body: map[string]interface{}{"key": "key", "field": "field-1", "value": 10}},
			},
			expected:      []interface{}{float64(10)},
			delay:         []time.Duration{0},
			errorExpected: false,
		},
		{
			name: "HINCRBY should increment when key exists and a field doesn't exist",
			commands: []HTTPCommand{
				{Command: "HSET", Body: map[string]interface{}{"key": "new-key", "field": "field-1", "value": 10}},
				{Command: "HINCRBY", Body: map[string]interface{}{"key": "new-key", "field": "field-2", "value": 10}},
			},
			expected:      []interface{}{float64(1), float64(10)},
			delay:         []time.Duration{0, 0},
			errorExpected: false,
		},
		{
			name: "HINCRBY should increment on existing key and field",
			commands: []HTTPCommand{
				{Command: "HINCRBY", Body: map[string]interface{}{"key": "key", "field": "field-1", "value": 10}},
			},
			expected:      []interface{}{float64(20)},
			delay:         []time.Duration{0},
			errorExpected: false,
		},
		{
			name: "HINCRBY should decrement on existing key and field",
			commands: []HTTPCommand{
				{Command: "HINCRBY", Body: map[string]interface{}{"key": "key", "field": "field-1", "value": -10}},
			},
			expected:      []interface{}{float64(10)},
			delay:         []time.Duration{0},
			errorExpected: false,
		},
		{
			name: "HINCRBY should give error when increment field is greater than max int64 field",
			commands: []HTTPCommand{
				{Command: "HINCRBY", Body: map[string]interface{}{"key": "key", "field": "field", "value": float64(9999999999999999999999999999999999999)}},
			},
			expected:      []interface{}{"ERR value is not an integer or out of range"},
			delay:         []time.Duration{0},
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

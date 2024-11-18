package http

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestINCRBYFLOAT(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	invalidArgMessage := "ERR wrong number of arguments for 'incrbyfloat' command"
	invalidIncrTypeMessage := "ERR value is not a valid float"
	valueOutOfRangeMessage := "ERR value is out of range"

	testCases := []struct {
		name     string
		commands []HTTPCommand
		expected []interface{}
		delays   []time.Duration
	}{
		{
			name: "Invalid number of arguments",
			commands: []HTTPCommand{
				{Command: "INCRBYFLOAT", Body: map[string]interface{}{"key": nil}},
				{Command: "INCRBYFLOAT", Body: map[string]interface{}{"key": "foo"}},
			},
			expected: []interface{}{invalidArgMessage, invalidArgMessage},
			delays:   []time.Duration{0, 0},
		},
		{
			name: "Increment a non existing key",
			commands: []HTTPCommand{
				{Command: "INCRBYFLOAT", Body: map[string]interface{}{"key": "foo", "value": 0.1}},
				{Command: "GET", Body: map[string]interface{}{"key": "foo"}},
			},
			expected: []interface{}{"0.1", "0.1"},
			delays:   []time.Duration{0, 0},
		},
		{
			name: "Increment a key with an integer value",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "1"}},
				{Command: "INCRBYFLOAT", Body: map[string]interface{}{"key": "foo", "value": 0.1}},
				{Command: "GET", Body: map[string]interface{}{"key": "foo"}},
			},
			expected: []interface{}{"OK", "1.1", "1.1"},
			delays:   []time.Duration{0, 0, 0},
		},
		{
			name: "Increment and then decrement a key with the same value",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "1"}},
				{Command: "INCRBYFLOAT", Body: map[string]interface{}{"key": "foo", "value": 0.1}},
				{Command: "GET", Body: map[string]interface{}{"key": "foo"}},
				{Command: "INCRBYFLOAT", Body: map[string]interface{}{"key": "foo", "value": -0.1}},
				{Command: "GET", Body: map[string]interface{}{"key": "foo"}},
			},
			expected: []interface{}{"OK", "1.1", "1.1", "1", "1"},
			delays:   []time.Duration{0, 0, 0, 0, 0},
		},
		{
			name: "Increment a non numeric value",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "INCRBYFLOAT", Body: map[string]interface{}{"key": "foo", "value": 0.1}},
			},
			expected: []interface{}{"OK", invalidIncrTypeMessage},
			delays:   []time.Duration{0, 0},
		},
		{
			name: "Increment by a non numeric value",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "1"}},
				{Command: "INCRBYFLOAT", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
			},
			expected: []interface{}{"OK", invalidIncrTypeMessage},
			delays:   []time.Duration{0, 0},
		},
		{
			name: "Increment by both integer and float",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "1"}},
				{Command: "INCRBYFLOAT", Body: map[string]interface{}{"key": "foo", "value": 1}},
				{Command: "INCRBYFLOAT", Body: map[string]interface{}{"key": "foo", "value": 0.1}},
			},
			expected: []interface{}{"OK", "2", "2.1"},
			delays:   []time.Duration{0, 0, 0},
		},
		{
			name: "Increment that would make the value Inf",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "1e308"}},
				{Command: "INCRBYFLOAT", Body: map[string]interface{}{"key": "foo", "value": 1e308}},
				{Command: "INCRBYFLOAT", Body: map[string]interface{}{"key": "foo", "value": -1e308}},
			},
			expected: []interface{}{"OK", valueOutOfRangeMessage, "0"},
			delays:   []time.Duration{0, 0, 0},
		},
		{
			name: "Increment that would make the value -Inf",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "-1e308"}},
				{Command: "INCRBYFLOAT", Body: map[string]interface{}{"key": "foo", "value": -1e308}},
				{Command: "INCRBYFLOAT", Body: map[string]interface{}{"key": "foo", "value": 1e308}},
			},
			expected: []interface{}{"OK", valueOutOfRangeMessage, "0"},
			delays:   []time.Duration{0, 0, 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"key": "foo"}})

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

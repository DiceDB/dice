package http

import (
	"math"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestINCR(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"keys": [...]string{"key1", "key2"}}})

	testCases := []struct {
		name     string
		commands []HTTPCommand
		expected []interface{}
		delays   []time.Duration
	}{
		{
			name: "Increment multiple keys",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "key1", "value": 0}},
				{Command: "INCR", Body: map[string]interface{}{"key": "key1"}},
				{Command: "INCR", Body: map[string]interface{}{"key": "key1"}},
				{Command: "INCR", Body: map[string]interface{}{"key": "key2"}},
				{Command: "GET", Body: map[string]interface{}{"key": "key1"}},
				{Command: "GET", Body: map[string]interface{}{"key": "key2"}},
			},
			expected: []interface{}{"OK", float64(1), float64(2), float64(1), float64(2), float64(1)},
			delays:   []time.Duration{0, 0, 0, 0, 0, 0},
		},
		{
			name: "Increment to and from max int64",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "max_int", "value": strconv.Itoa(math.MaxInt64 - 1)}},
				{Command: "INCR", Body: map[string]interface{}{"key": "max_int"}},
				{Command: "INCR", Body: map[string]interface{}{"key": "max_int"}},
				{Command: "SET", Body: map[string]interface{}{"key": "max_int", "value": strconv.Itoa(math.MaxInt64)}},
				{Command: "INCR", Body: map[string]interface{}{"key": "max_int"}},
			},
			expected: []interface{}{"OK", float64(math.MaxInt64), "ERR increment or decrement would overflow", "OK", "ERR increment or decrement would overflow"},
			delays:   []time.Duration{0, 0, 0, 0, 0},
		},
		{
			name: "Increment from min int64",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "min_int", "value": strconv.Itoa(math.MinInt64)}},
				{Command: "INCR", Body: map[string]interface{}{"key": "min_int"}},
				{Command: "INCR", Body: map[string]interface{}{"key": "min_int"}},
			},
			expected: []interface{}{"OK", float64(math.MinInt64 + 1), float64(math.MinInt64 + 2)},
			delays:   []time.Duration{0, 0, 0},
		},
		{
			name: "Increment non-integer values",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "float_key", "value": "3.14"}},
				{Command: "INCR", Body: map[string]interface{}{"key": "float_key"}},
				{Command: "SET", Body: map[string]interface{}{"key": "string_key", "value": "hello"}},
				{Command: "INCR", Body: map[string]interface{}{"key": "string_key"}},
				{Command: "SET", Body: map[string]interface{}{"key": "bool_key", "value": "true"}},
				{Command: "INCR", Body: map[string]interface{}{"key": "bool_key"}},
			},
			expected: []interface{}{"OK", "ERR value is not an integer or out of range", "OK", "ERR value is not an integer or out of range", "OK", "ERR value is not an integer or out of range"},
			delays:   []time.Duration{0, 0, 0, 0, 0, 0},
		},
		{
			name: "Increment non-existent key",
			commands: []HTTPCommand{
				{Command: "INCR", Body: map[string]interface{}{"key": "non_existent"}},
				{Command: "GET", Body: map[string]interface{}{"key": "non_existent"}},
				{Command: "INCR", Body: map[string]interface{}{"key": "non_existent"}},
			},
			expected: []interface{}{float64(1), float64(1), float64(2)},
			delays:   []time.Duration{0, 0, 0},
		},
		{
			name: "Increment string representing integers",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "str_int1", "value": "42"}},
				{Command: "INCR", Body: map[string]interface{}{"key": "str_int1"}},
				{Command: "SET", Body: map[string]interface{}{"key": "str_int2", "value": "-10"}},
				{Command: "INCR", Body: map[string]interface{}{"key": "str_int2"}},
				{Command: "SET", Body: map[string]interface{}{"key": "str_int3", "value": "0"}},
				{Command: "INCR", Body: map[string]interface{}{"key": "str_int3"}},
			},
			expected: []interface{}{"OK", float64(43), "OK", float64(-9), "OK", float64(1)},
			delays:   []time.Duration{0, 0, 0, 0, 0, 0},
		},
		{
			name: "Increment with expiry",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "expiry_key", "value": 0, "ex": 1}},
				{Command: "INCR", Body: map[string]interface{}{"key": "expiry_key"}},
				{Command: "INCR", Body: map[string]interface{}{"key": "expiry_key"}},
				{Command: "INCR", Body: map[string]interface{}{"key": "expiry_key"}},
			},
			expected: []interface{}{"OK", float64(1), float64(2), float64(1)},
			delays:   []time.Duration{0, 0, 0, 1 * time.Second},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"keys": [...]string{"key1", "key2", "expiry_key", "max_int", "min_int", "float_key", "string_key", "bool_key"}}})

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

func TestINCRBY(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []struct {
		name     string
		commands []HTTPCommand
		expected []interface{}
		delays   []time.Duration
	}{
		{
			name: "INCRBY with positive increment",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "key", "value": 3}},
				{Command: "INCRBY", Body: map[string]interface{}{"key": "key", "value": 2}},
				{Command: "INCRBY", Body: map[string]interface{}{"key": "key", "value": 1}},
				{Command: "GET", Body: map[string]interface{}{"key": "key"}},
			},
			expected: []interface{}{"OK", float64(5), float64(6), float64(6)},
			delays:   []time.Duration{0, 0, 0, 0},
		},
		{
			name: "INCRBY with negative increment",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "key", "value": 100}},
				{Command: "INCRBY", Body: map[string]interface{}{"key": "key", "value": -2}},
				{Command: "INCRBY", Body: map[string]interface{}{"key": "key", "value": -10}},
				{Command: "INCRBY", Body: map[string]interface{}{"key": "key", "value": -88}},
				{Command: "INCRBY", Body: map[string]interface{}{"key": "key", "value": -100}},
				{Command: "GET", Body: map[string]interface{}{"key": "key"}},
			},
			expected: []interface{}{"OK", float64(98), float64(88), float64(0), float64(-100), float64(-100)},
			delays:   []time.Duration{0, 0, 0, 0, 0, 0},
		},
		{
			name: "INCRBY with unset key",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "key", "value": 3}},
				{Command: "INCRBY", Body: map[string]interface{}{"key": "unsetKey", "value": 2}},
				{Command: "GET", Body: map[string]interface{}{"key": "key"}},
				{Command: "GET", Body: map[string]interface{}{"key": "unsetKey"}},
			},
			expected: []interface{}{"OK", float64(2), float64(3), float64(2)},
			delays:   []time.Duration{0, 0, 0, 0},
		},
		{
			name: "edge case with maximum int value",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "key", "value": strconv.Itoa(math.MaxInt64 - 1)}},
				{Command: "INCRBY", Body: map[string]interface{}{"key": "key", "value": 1}},
				{Command: "INCRBY", Body: map[string]interface{}{"key": "key", "value": 1}},
				{Command: "GET", Body: map[string]interface{}{"key": "key"}},
			},
			expected: []interface{}{"OK", float64(math.MaxInt64), "ERR increment or decrement would overflow", float64(math.MaxInt64)},
			delays:   []time.Duration{0, 0, 0, 0},
		},
		{
			name: "edge case with minimum int value",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "key", "value": strconv.Itoa(math.MinInt64 + 1)}},
				{Command: "INCRBY", Body: map[string]interface{}{"key": "key", "value": -1}},
				{Command: "INCRBY", Body: map[string]interface{}{"key": "key", "value": -1}},
				{Command: "GET", Body: map[string]interface{}{"key": "key"}},
			},
			expected: []interface{}{"OK", float64(math.MinInt64), "ERR increment or decrement would overflow", float64(math.MinInt64)},
			delays:   []time.Duration{0, 0, 0, 0},
		},
		{
			name: "edge case with string values",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "key", "value": 1}},
				{Command: "INCRBY", Body: map[string]interface{}{"key": "stringKey", "value": "abc"}},
			},
			expected: []interface{}{"OK", "ERR value is not an integer or out of range"},
			delays:   []time.Duration{0, 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"keys": [...]string{"key", "unsetKey", "stringkey"}}})

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

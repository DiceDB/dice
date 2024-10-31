package http

import (
	"math"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDECR(t *testing.T) {
	exec := NewHTTPCommandExecutor()
	exec.FireCommand(HTTPCommand{Command: "FLUSHDB"})
	exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"keys": [...]string{"key1", "key2", "key3"}}})

	testCases := []struct {
		name     string
		commands []HTTPCommand
		expected []interface{}
		delays   []time.Duration
	}{
		{
			name: "Decrement multiple keys",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "key1", "value": 3}},
				{Command: "DECR", Body: map[string]interface{}{"key": "key1"}},
				{Command: "DECR", Body: map[string]interface{}{"key": "key1"}},
				{Command: "DECR", Body: map[string]interface{}{"key": "key2"}},
				{Command: "GET", Body: map[string]interface{}{"key": "key1"}},
				{Command: "GET", Body: map[string]interface{}{"key": "key2"}},
				{Command: "SET", Body: map[string]interface{}{"key": "key3", "value": strconv.Itoa(math.MinInt64 + 1)}},
				{Command: "DECR", Body: map[string]interface{}{"key": "key3"}},
				{Command: "DECR", Body: map[string]interface{}{"key": "key3"}},
			},
			expected: []interface{}{"OK", float64(2), float64(1), float64(-1), float64(1), float64(-1), "OK", float64(math.MinInt64), "ERR increment or decrement would overflow"},
			delays:   []time.Duration{0, 0, 0, 0, 0, 0, 0, 0, 0},
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

func TestDECRBY(t *testing.T) {
	exec := NewHTTPCommandExecutor()
	exec.FireCommand(HTTPCommand{Command: "FLUSHDB"})
	exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"keys": [...]string{"key1", "key2", "key3"}}})

	testCases := []struct {
		name     string
		commands []HTTPCommand
		expected []interface{}
		delays   []time.Duration
	}{
		{
			name: "Decrement multiple keys",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "key1", "value": 3}},
				{Command: "SET", Body: map[string]interface{}{"key": "key3", "value": strconv.Itoa(math.MinInt64 + 1)}},
				{Command: "DECRBY", Body: map[string]interface{}{"key": "key1", "value": 2}},
				{Command: "DECRBY", Body: map[string]interface{}{"key": "key1", "value": 1}},
				{Command: "DECRBY", Body: map[string]interface{}{"key": "key4", "value": 1}},
				{Command: "DECRBY", Body: map[string]interface{}{"key": "key3", "value": 1}},
				{Command: "DECRBY", Body: map[string]interface{}{"key": "key3", "value": strconv.Itoa(math.MinInt64)}},
				{Command: "DECRBY", Body: map[string]interface{}{"key": "key5", "value": "abc"}},
				{Command: "GET", Body: map[string]interface{}{"key": "key1"}},
				{Command: "GET", Body: map[string]interface{}{"key": "key4"}},
			},
			expected: []interface{}{"OK", "OK", float64(1), float64(0), float64(-1), float64(math.MinInt64), "ERR increment or decrement would overflow", "ERR value is not an integer or out of range", float64(0), float64(-1)},
			delays:   []time.Duration{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
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

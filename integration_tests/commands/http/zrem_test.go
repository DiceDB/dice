package http

import (
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestZREM(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []struct {
		name     string
		commands []HTTPCommand
		expected []interface{}
		delays   []time.Duration
	}{
		{
			name: "ZREM with wrong number of arguments",
			commands: []HTTPCommand{
				{Command: "ZREM", Body: nil},
				{Command: "ZREM", Body: map[string]interface{}{"key": "myzset"}},
			},
			expected: []interface{}{
				"ERR wrong number of arguments for 'zrem' command",
				"ERR wrong number of arguments for 'zrem' command"},
			delays: []time.Duration{0, 0},
		},
		{
			name: "ZREM with wrong type of key",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "string_key", "value": "string_value"}},
				{Command: "ZREM", Body: map[string]interface{}{"key": "string_key", "field": "string_value"}},
			},
			expected: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
			delays:   []time.Duration{0, 0},
		},
		{
			name: "ZREM with non-existent key",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset", "values": []string{"1", "one"}}},
				{Command: "ZREM", Body: map[string]interface{}{"key": "wrong_myzset", "field": "one"}},
			},
			expected: []interface{}{float64(1), float64(0)},
			delays:   []time.Duration{0, 0},
		},
		{
			name: "ZREM with non-existent element",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset", "values": []string{"1", "one"}}},
				{Command: "ZREM", Body: map[string]interface{}{"key": "wrong_myzset", "field": "two"}},
			},
			expected: []interface{}{float64(1), float64(0)},
			delays:   []time.Duration{0, 0},
		},
		{
			name: "ZREM with sorted set holding single element",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset", "values": []string{"1", "one"}}},
				{Command: "ZREM", Body: map[string]interface{}{"key": "myzset", "values": []string{"one"}}},
			},
			expected: []interface{}{float64(1), float64(1)},
			delays:   []time.Duration{0, 0},
		},
		{
			name: "ZREM with sorted set holding multiple elements",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset", "values": []string{"1", "one", "2", "two", "3", "three", "4", "four"}}},
				{Command: "ZREM", Body: map[string]interface{}{"key": "myzset", "values": []string{"four", "five"}}},
				{Command: "ZREM", Body: map[string]interface{}{"key": "myzset", "values": []string{"one", "two"}}},
			},
			expected: []interface{}{float64(4), float64(1), float64(2)},
			delays:   []time.Duration{0, 0, 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"keys": []string{"string_key", "myzset"}}})

			for i, cmd := range tc.commands {
				if tc.delays[i] > 0 {
					time.Sleep(tc.delays[i])
				}
				result, err := exec.FireCommand(cmd)
				if err != nil {
					// Check if the error message matches the expected result
					log.Println(tc.expected[i])
					assert.Equal(t, tc.expected[i], err.Error(), "Error message mismatch for cmd %s", cmd)
				} else {
					assert.Equal(t, tc.expected[i], result, "Value mismatch for cmd %s, expected %v, got %v", cmd, tc.expected[i], result)
				}
			}
		})
	}
}

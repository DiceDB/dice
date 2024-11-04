package http

import (
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHLen(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []struct {
		name     string
		commands []HTTPCommand
		expected []interface{}
		delays   []time.Duration
	}{
		{
			name: "HLEN with wrong number of arguments",
			commands: []HTTPCommand{
				{Command: "HLEN", Body: map[string]interface{}{"key": "", "field": ""}},
				{Command: "HLEN", Body: map[string]interface{}{"key": "KEY", "field": "field", "another_field": "another_field"}},
			},
			expected: []interface{}{
				"ERR wrong number of arguments for 'hlen' command",
				"ERR wrong number of arguments for 'hlen' command"},
			delays: []time.Duration{0, 0},
		},
		{
			name: "HLEN with wrong key",
			commands: []HTTPCommand{
				{Command: "HSET", Body: map[string]interface{}{"key": "key_hLen1", "field": "field", "value": "value"}},
				{Command: "HLEN", Body: map[string]interface{}{"key": "wrong_key_hLen1"}},
			},
			expected: []interface{}{float64(1), float64(0)},
			delays:   []time.Duration{0, 0},
		},
		{
			name: "HLEN with single field",
			commands: []HTTPCommand{
				{Command: "HSET", Body: map[string]interface{}{"key": "key_hLen2", "field": "field", "value": "HelloWorld"}},
				{Command: "HLEN", Body: map[string]interface{}{"key": "key_hLen2"}},
			},
			expected: []interface{}{float64(1), float64(1)},
			delays:   []time.Duration{0, 0},
		},
		{
			name: "HLEN with multiple fields",
			commands: []HTTPCommand{
				{Command: "HSET", Body: map[string]interface{}{"key": "key_hLen3", "key_values": map[string]interface{}{"field1": "value1", "field2": "value2", "field3": "value3"}}},
				{Command: "HLEN", Body: map[string]interface{}{"key": "key_hLen3"}},
			},
			expected: []interface{}{float64(3), float64(3)},
			delays:   []time.Duration{0, 0},
		},

		{
			name: "HLEN with wrong type",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "key", "value": "value"}},
				{Command: "HLEN", Body: map[string]interface{}{"key": "key"}},
			},
			expected: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
			delays:   []time.Duration{0, 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"keys": [...]string{"KEY", "key", "key_hLen1", "key_hLen2", "key_hLen3"}}})

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

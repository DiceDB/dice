package http

import (
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHStrLen(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []struct {
		name     string
		commands []HTTPCommand
		expected []interface{}
		delays   []time.Duration
	}{
		{
			name: "HSTRLEN with wrong number of arguments",
			commands: []HTTPCommand{
				{Command: "HSTRLEN", Body: map[string]interface{}{"key": "KEY"}},
				{Command: "HSTRLEN", Body: map[string]interface{}{"key": "KEY", "field": "field", "another_field": "another_field"}},
			},
			expected: []interface{}{
				"ERR wrong number of arguments for 'hstrlen' command",
				"ERR wrong number of arguments for 'hstrlen' command"},
			delays: []time.Duration{0, 0},
		},
		{
			name: "HSTRLEN with wrong key",
			commands: []HTTPCommand{
				{Command: "HSET", Body: map[string]interface{}{"key": "key_hStrLen1", "field": "field", "value": "value"}},
				{Command: "HSTRLEN", Body: map[string]interface{}{"key": "wrong_key_hStrLen", "field": "field"}},
			},
			expected: []interface{}{float64(1), float64(0)},
			delays:   []time.Duration{0, 0},
		},
		{
			name: "HSTRLEN with wrong field",
			commands: []HTTPCommand{
				{Command: "HSET", Body: map[string]interface{}{"key": "key_hStrLen2", "field": "field", "value": "value"}},
				{Command: "HSTRLEN", Body: map[string]interface{}{"key": "key_hStrLen2", "field": "wrong_field"}},
			},
			expected: []interface{}{float64(1), float64(0)},
			delays:   []time.Duration{0, 0},
		},
		{
			name: "HSTRLEN",
			commands: []HTTPCommand{
				{Command: "HSET", Body: map[string]interface{}{"key": "key_hStrLen3", "field": "field", "value": "HelloWorld"}},
				{Command: "HSTRLEN", Body: map[string]interface{}{"key": "key_hStrLen3", "field": "field"}},
			},
			expected: []interface{}{float64(1), float64(10)},
			delays:   []time.Duration{0, 0},
		},
		{
			name: "HSTRLEN with wrong type",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "key", "value": "value"}},
				{Command: "HSTRLEN", Body: map[string]interface{}{"key": "key", "field": "field"}},
			},
			expected: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
			delays:   []time.Duration{0, 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"keys": [...]string{"KEY", "key"}}})

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

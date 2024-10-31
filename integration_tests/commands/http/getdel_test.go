package http

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetDel(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []struct {
		name     string
		commands []HTTPCommand
		expected []interface{}
		delays   []time.Duration
	}{
		{
			name: "GetDel",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "k", "value": "v"}},
				{Command: "GETDEL", Body: map[string]interface{}{"key": "k"}},
				{Command: "GETDEL", Body: map[string]interface{}{"key": "k"}},
				{Command: "GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", "v", nil, nil},
			delays:   []time.Duration{0, 0, 0, 0},
		},
		{
			name: "GetDel with expiration, checking if key exist and is already expired",
			commands: []HTTPCommand{
				{Command: "GETDEL", Body: map[string]interface{}{"key": "k"}},
				{Command: "SET", Body: map[string]interface{}{"key": "k", "value": "v", "ex": 2}},
				{Command: "GETDEL", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{nil, "OK", nil},
			delays:   []time.Duration{0, 0, 3 * time.Second},
		},
		{
			name: "GetDel with expiration, checking if key exist and is not yet expired",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "k", "value": "v", "ex": 40}},
				{Command: "GETDEL", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", "v"},
			delays:   []time.Duration{0, 2 * time.Second},
		},
		{
			name: "GetDel with invalid command",
			commands: []HTTPCommand{
				{Command: "GETDEL", Body: map[string]interface{}{"key": "k", "value": "v"}},
			},
			expected: []interface{}{
				"ERR wrong number of arguments for 'getdel' command",
			},
			delays: []time.Duration{0, 0},
		},
		{
			name: "Getdel with value created from Setbit",
			commands: []HTTPCommand{
				{Command: "SETBIT", Body: map[string]interface{}{"key": "k", "offset": 1, "value": 1}},
				{Command: "GET", Body: map[string]interface{}{"key": "k"}},
				{Command: "GETDEL", Body: map[string]interface{}{"key": "k"}},
				{Command: "GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{float64(0), "@", "@", nil},
			delays:   []time.Duration{0, 0, 0, 0},
		},
		{
			name: "GetDel with Set object should return wrong type error",
			commands: []HTTPCommand{
				{Command: "SADD", Body: map[string]interface{}{"key": "myset", "member": "member1"}},
				{Command: "GETDEL", Body: map[string]interface{}{"key": "myset"}},
			},
			expected: []interface{}{float64(1), "WRONGTYPE Operation against a key holding the wrong kind of value"},
			delays:   []time.Duration{0, 0},
		},
		{
			name: "GetDel with JSON object should return wrong type error",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "value": 1}},
				{Command: "GETDEL", Body: map[string]interface{}{"key": "k"}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value", "1"},
			delays:   []time.Duration{0, 0, 0},
		},
	}

	for _, tc := range testCases {
		exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"key": "k"}})
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

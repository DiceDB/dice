package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAPPEND(t *testing.T) {
	exec := NewHTTPCommandExecutor()
	testCases := []struct {
		name     string
		commands []HTTPCommand
		expected []interface{}
		cleanup  []HTTPCommand
	}{
		{
			name: "APPEND and GET a new Val",
			commands: []HTTPCommand{
				{Command: "APPEND", Body: map[string]interface{}{"key": "k", "value": "newVal"}},
				{Command: "GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{float64(6), "newVal"},
			cleanup: []HTTPCommand{
				{Command: "del", Body: map[string]interface{}{"key": "k"}},
			},
		},
		{
			name: "APPEND to an exisiting key and GET",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "k", "value": "Bhima"}},
				{Command: "APPEND", Body: map[string]interface{}{"key": "k", "value": "Shankar"}},
				{Command: "GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", float64(12), "BhimaShankar"},
			cleanup: []HTTPCommand{
				{Command: "del", Body: map[string]interface{}{"key": "k"}},
			},
		},
		{
			name: "APPEND without input value",
			commands: []HTTPCommand{
				{Command: "APPEND", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"ERR wrong number of arguments for 'append' command"},
			cleanup: []HTTPCommand{
				{Command: "del", Body: map[string]interface{}{"key": "k"}},
			},
		},
		{
			name: "APPEND empty string to an exsisting key with empty string",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "k", "value": ""}},
				{Command: "APPEND", Body: map[string]interface{}{"key": "k", "value": ""}},
			},
			expected: []interface{}{"OK", float64(0)},
			cleanup: []HTTPCommand{
				{Command: "del", Body: map[string]interface{}{"key": "k"}},
			},
		},
		{
			name: "APPEND to key created using LPUSH",
			commands: []HTTPCommand{
				{Command: "LPUSH", Body: map[string]interface{}{"key": "m", "value": "bhima"}},
				{Command: "APPEND", Body: map[string]interface{}{"key": "m", "value": "shankar"}},
			},
			expected: []interface{}{float64(1), "WRONGTYPE Operation against a key holding the wrong kind of value"},
			cleanup: []HTTPCommand{
				{Command: "del", Body: map[string]interface{}{"key": "m"}},
			},
		},
		{
			name: "APPEND value with leading zeros",
			commands: []HTTPCommand{
				{Command: "APPEND", Body: map[string]interface{}{"key": "z", "value": "0043"}},
			},
			expected: []interface{}{float64(4)},
			cleanup: []HTTPCommand{
				{Command: "del", Body: map[string]interface{}{"key": "z"}},
			},
		},
		{
			name: "APPEND to key created using SADD",
			commands: []HTTPCommand{
				{Command: "SADD", Body: map[string]interface{}{"key": "key", "value": "apple"}},
				{Command: "APPEND", Body: map[string]interface{}{"key": "key", "value": "banana"}},
			},
			expected: []interface{}{float64(1), "WRONGTYPE Operation against a key holding the wrong kind of value"},
			cleanup: []HTTPCommand{
				{Command: "del", Body: map[string]interface{}{"key": "key"}},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)
				assert.Equal(t, tc.expected[i], result)
			}
			exec.FireCommand(tc.cleanup[0])
		})
	}
}

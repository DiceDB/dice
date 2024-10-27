package http

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestCommandGetKeys(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []TestCase{
		{
			name: "Set command",
			commands: []HTTPCommand{
				{Command: "COMMAND/GETKEYS", Body: map[string]interface{}{"key": "SET", "keys": []interface{}{"1", "2"}, "values": []interface{}{"2", "3"}}},
			},
			expected: []interface{}{[]interface{}{"1"}},
		},
		{
			name: "Get command",
			commands: []HTTPCommand{
				{Command: "COMMAND/GETKEYS", Body: map[string]interface{}{"key": "GET", "field": "key"}},
			},
			expected: []interface{}{[]interface{}{"key"}},
		},
		{
			name: "TTL command",
			commands: []HTTPCommand{
				{Command: "COMMAND/GETKEYS", Body: map[string]interface{}{"key": "TTL", "field": "key"}},
			},
			expected: []interface{}{[]interface{}{"key"}},
		},
		{
			name: "Del command",
			commands: []HTTPCommand{
				{Command: "COMMAND/GETKEYS", Body: map[string]interface{}{"key": "DEL", "field": "1 2 3 4 5 6 7"}},
			},
			expected: []interface{}{[]interface{}{"1 2 3 4 5 6 7"}},
		},
		{
			name: "MSET command",
			commands: []HTTPCommand{
				{Command: "COMMAND/GETKEYS", Body: map[string]interface{}{"key": "MSET", "keys": []interface{}{"key1 key2"}, "values": []interface{}{" val1 val2"}}},
			},
			expected: []interface{}{[]interface{}{"key1 key2"}},
		},
		{
			name: "Expire command",
			commands: []HTTPCommand{
				{Command: "COMMAND/GETKEYS", Body: map[string]interface{}{"key": "EXPIRE", "field": "key", "values": []interface{}{"time", "extra"}}},
			},
			expected: []interface{}{[]interface{}{"key"}},
		},
		{
			name: "PING command",
			commands: []HTTPCommand{
				{Command: "COMMAND/GETKEYS", Body: map[string]interface{}{"key": "PING"}},
			},
			expected: []interface{}{"ERR the command has no key arguments"},
		},
		{
			name: "Invalid Get command",
			commands: []HTTPCommand{
				{Command: "COMMAND/GETKEYS", Body: map[string]interface{}{"key": "GET"}},
			},
			expected: []interface{}{"ERR invalid number of arguments specified for command"},
		},
		{
			name: "Abort command",
			commands: []HTTPCommand{
				{Command: "COMMAND/GETKEYS", Body: map[string]interface{}{"key": "ABORT"}},
			},
			expected: []interface{}{"ERR the command has no key arguments"},
		},
		{
			name: "Invalid command",
			commands: []HTTPCommand{
				{Command: "COMMAND/GETKEYS", Body: map[string]interface{}{"key": "NotValidCommand"}},
			},
			expected: []interface{}{"ERR invalid command specified"},
		},
		{
			name: "Wrong number of arguments",
			commands: []HTTPCommand{
				{Command: "COMMAND/GETKEYS", Body: map[string]interface{}{"key": ""}},
			},
			expected: []interface{}{"ERR invalid command specified"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)
				assert.DeepEqual(t, tc.expected[i], result)
			}
		})
	}
}

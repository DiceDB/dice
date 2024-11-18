package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandInfo(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []TestCase{
		{
			name: "Set command",
			commands: []HTTPCommand{
				{Command: "COMMAND/INFO", Body: map[string]interface{}{"key": "SET"}},
			},
			expected: []interface{}{[]interface{}{[]interface{}{"set", float64(-3), float64(1), float64(0), float64(0), []any{}}}},
		},
		{
			name: "Get command",
			commands: []HTTPCommand{
				{Command: "COMMAND/INFO", Body: map[string]interface{}{"key": "GET"}},
			},
			expected: []interface{}{[]interface{}{[]interface{}{"get", float64(2), float64(1), float64(0), float64(0), []any{}}}},
		},
		{
			name: "PING command",
			commands: []HTTPCommand{
				{Command: "COMMAND/INFO", Body: map[string]interface{}{"key": "PING"}},
			},
			expected: []interface{}{[]interface{}{[]interface{}{"ping", float64(-1), float64(0), float64(0), float64(0), []any{}}}},
		},
		{
			name: "Invalid command",
			commands: []HTTPCommand{
				{Command: "COMMAND/INFO", Body: map[string]interface{}{"key": "INVALID_CMD"}},
			},
			expected: []interface{}{[]interface{}{nil}},
		},
		{
			name: "Combination of valid and Invalid command",
			commands: []HTTPCommand{
				{Command: "COMMAND/INFO", Body: map[string]interface{}{"keys": []interface{}{"SET", "INVALID_CMD"}}},
			},
			expected: []interface{}{[]interface{}{
				[]interface{}{"set", float64(-3), float64(1), float64(0), float64(0), []any{}},
				nil,
			}},
		},
		{
			name: "Combination of multiple valid commands",
			commands: []HTTPCommand{
				{Command: "COMMAND/INFO", Body: map[string]interface{}{"keys": []interface{}{"SET", "GET"}}},
			},
			expected: []interface{}{[]interface{}{
				[]interface{}{"set", float64(-3), float64(1), float64(0), float64(0), []any{}},
				[]interface{}{"get", float64(2), float64(1), float64(0), float64(0), []any{}},
			}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}
}

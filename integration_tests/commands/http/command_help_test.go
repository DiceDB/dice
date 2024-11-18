package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandHelp(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []TestCase{
		{
			name: "Command help should not support any argument",
			commands: []HTTPCommand{
				{Command: "COMMAND/HELP", Body: map[string]interface{}{"key": ""}},
			},
			expected: []interface{}{"ERR wrong number of arguments for 'command|help' command"},
		},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)
				assert.Equal(t, tc.expected[i], result)

			}

		})
	}
}

package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandList(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []TestCase{
		{
			name: "Command list should not be empty",
			commands: []HTTPCommand{
				{Command: "COMMAND/LIST", Body: map[string]interface{}{"key": ""}},
			},
			expected: []interface{}{"ERR wrong number of arguments for 'command|list' command"},
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

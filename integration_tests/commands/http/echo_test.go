package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEchoHttp(t *testing.T) {

	exec := NewHTTPCommandExecutor()

	testCases := []struct {
		name     string
		commands []HTTPCommand
		expected []interface{}
	}{
		{
			name: "ECHO with invalid number of arguments",
			commands: []HTTPCommand{
				{Command: "ECHO", Body: map[string]interface{}{"key": "", "value": ""}},
			},
			expected: []interface{}{"ERR wrong number of arguments for 'echo' command"},
		},
		{
			name: "ECHO with one argument",
			commands: []HTTPCommand{
				{Command: "ECHO", Body: map[string]interface{}{"value": "hello world"}}, // Providing one argument "hello world"
			},
			expected: []interface{}{"hello world"},
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

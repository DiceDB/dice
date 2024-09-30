package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSELECT(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []struct {
		name     string
		commands []HTTPCommand
		expected []interface{}
	}{
		{
			name: "SELECT command response",
			commands: []HTTPCommand{
				{Command: "SELECT", Body: map[string]interface{}{"value": "1"}},
			},
			expected: []interface{}{"OK"},
		},
		{
			name: "SELECT command error response",
			commands: []HTTPCommand{
				{Command: "SELECT", Body: map[string]interface{}{"value": ""}},
			},
			expected: []interface{}{"OK"},
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

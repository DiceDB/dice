package http

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestGetKeys(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []TestCase{
		{
			name: "Set command",
			commands: []HTTPCommand{
				{Command: "COMMAND/GETKEYS/SET", Body: map[string]interface{}{"key": "1", "value": "2"}},
			},
			expected: []interface{}{[]interface{}{"1"}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result := exec.FireCommand(cmd)
				assert.DeepEqual(t, tc.expected[i], result)
			}
		})
	}
}

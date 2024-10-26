package http

import (
	"gotest.tools/v3/assert"
	"testing"
)

func TestDiscard(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []TestCase{
		{
			name: "Discard commands in a transaction",
			commands: []HTTPCommand{
				{Command: "MULTI", Body: nil},
				{Command: "SET", Body: map[string]interface{}{"key": "key1", "value": "value1"}},
				{Command: "DISCARD", Body: nil},
			},
			expected: []interface{}{"OK", "OK", "OK"},
		},
		{
			name: "Error when DISCARD is used outside a transaction",
			commands: []HTTPCommand{
				{Command: "DISCARD", Body: nil},
			},
			expected: []interface{}{"ERR DISCARD without MULTI"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"keys": [...]string{"key1"}}})

			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)
				assert.Equal(t, tc.expected[i], result, "Value mismatch for command: %v", cmd)
			}
		})
	}
}

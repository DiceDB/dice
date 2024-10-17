package http

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestHExists(t *testing.T) {
	cmdExec := NewHTTPCommandExecutor()

	testCases := []TestCase{
		{
			name: "Check if field exists when k f and v are set",
			commands: []HTTPCommand{
				{Command: "HSET", Body: map[string]interface{}{"key": "k", "field": "f", "value": "v"}},
				{Command: "HEXISTS", Body: map[string]interface{}{"key": "k", "field": "f"}},
			},
			expected: []interface{}{int64(1), int64(1)},
		},
		{
			name: "Check if field exists when k exists but not f and v",
			commands: []HTTPCommand{
				{Command: "HSET", Body: map[string]interface{}{"key": "k", "field": "f1", "value": "v"}},
				{Command: "HEXISTS", Body: map[string]interface{}{"key": "k", "field": "f"}},
			},
			expected: []interface{}{int(1), int64(0)},
		},
		{
			name:     "Check if field exists when no k,f and v exist",
			commands: []HTTPCommand{
				{Command: "HEXISTS", Body: map[string]interface{}{"key": "k", "field": "f"}},
			},
			expected: []interface{}{int64(0)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmdExec.FireCommand(HTTPCommand{
				Command: "HDEL",
				Body:    map[string]interface{}{"key": "k", "field": "f"},
			})

			for i, cmd := range tc.commands {
				result, _ := cmdExec.FireCommand(cmd)
				assert.DeepEqual(t, tc.expected[i], result)
			}
		})
	}
}

package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHKeys(t *testing.T) {
	cmdExec := NewHTTPCommandExecutor()

	testCases := []TestCase{
		{
			name: "HTTP One or more keys exist",
			commands: []HTTPCommand{
				{Command: "HSET", Body: map[string]interface{}{"key": "k", "field": "f", "value": "v"}},
				{Command: "HSET", Body: map[string]interface{}{"key": "k", "field": "f1", "value": "v1"}},
				{Command: "HKEYS", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{float64(1), float64(1), []interface{}{"f", "f1"}},
		},
		{
			name: "HTTP No keys exist",
			commands: []HTTPCommand{
				{Command: "HKEYS", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{nil},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmdExec.FireCommand(HTTPCommand{
				Command: "HDEL",
				Body:    map[string]interface{}{"key": "k", "field": "f"},
			})
			cmdExec.FireCommand(HTTPCommand{
				Command: "HDEL",
				Body:    map[string]interface{}{"key": "k", "field": "f1"},
			})

			for i, cmd := range tc.commands {
				result, _ := cmdExec.FireCommand(cmd)
				switch e := tc.expected[i].(type) {
				case []interface{}:
					assert.ElementsMatch(t, e, tc.expected[i])
				default:
					assert.Equal(t, tc.expected[i], result)
				}
			}
		})
	}

	cmdExec.FireCommand(HTTPCommand{
		Command: "HDEL",
		Body:    map[string]interface{}{"key": "k", "field": "f"},
	})
	cmdExec.FireCommand(HTTPCommand{
		Command: "HDEL",
		Body:    map[string]interface{}{"key": "k", "field": "f1"},
	})
}

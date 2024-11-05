package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDBSize(t *testing.T) {
	exec := NewHTTPCommandExecutor()
	exec.FireCommand(HTTPCommand{Command: "FLUSHDB"})

	testCases := []TestCase{
		{
			name: "DBSIZE with 3 keys",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "k1", "value": "v1"}},
				{Command: "SET", Body: map[string]interface{}{"key": "k2", "value": "v2"}},
				{Command: "SET", Body: map[string]interface{}{"key": "k3", "value": "v3"}},
				{Command: "DBSIZE", Body: nil},
			},
			expected: []interface{}{"OK", "OK", "OK", float64(3)},
		},
		{
			name: "DBSIZE with repetitive keys",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "k1", "value": "v1"}},
				{Command: "SET", Body: map[string]interface{}{"key": "k2", "value": "v2"}},
				{Command: "SET", Body: map[string]interface{}{"key": "k3", "value": "v3"}},
				{Command: "SET", Body: map[string]interface{}{"key": "k1", "value": "v3"}},
				{Command: "SET", Body: map[string]interface{}{"key": "k2", "value": "v22"}},
				{Command: "DBSIZE", Body: nil},
			},
			expected: []interface{}{"OK", "OK", "OK", "OK", "OK", float64(3)},
		},
		{
			name: "DBSIZE with expired keys",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "k1", "value": "v1"}},
				{Command: "SET", Body: map[string]interface{}{"key": "k2", "value": "v2"}},
				{Command: "SET", Body: map[string]interface{}{"key": "k3", "value": "v3"}},
				{Command: "DBSIZE", Body: nil},
				{Command: "EXPIRE", Body: map[string]interface{}{"key": "k3", "seconds": 1}},
				{Command: "SLEEP", Body: map[string]interface{}{"key": 2}},
				{Command: "DBSIZE", Body: nil},
			},
			expected: []interface{}{"OK", "OK", "OK", float64(3), float64(1), "OK", float64(2)},
		},
		{
			name: "DBSIZE after deleting a key",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "k1", "value": "v1"}},
				{Command: "SET", Body: map[string]interface{}{"key": "k2", "value": "v2"}},
				{Command: "SET", Body: map[string]interface{}{"key": "k3", "value": "v3"}},
				{Command: "DBSIZE", Body: nil},
				{Command: "DEL", Body: map[string]interface{}{"key": "k1"}},
				{Command: "DBSIZE", Body: nil},
			},
			expected: []interface{}{"OK", "OK", "OK", float64(3), float64(1), float64(2)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"keys": [...]string{"k1", "k2", "k3"}}})

			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}
}

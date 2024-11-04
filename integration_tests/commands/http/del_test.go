package http

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDel(t *testing.T) {
	exec := NewHTTPCommandExecutor()
	exec.FireCommand(HTTPCommand{Command: "FLUSHDB"})

	testCases := []TestCase{
		{
			name: "DEL with set key",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "k1", "value": "v1"}},
				{Command: "DEL", Body: map[string]interface{}{"keys": [...]string{"k1"}}},
				{Command: "GET", Body: map[string]interface{}{"key": "k1"}},
			},
			expected: []interface{}{"OK", float64(1), nil},
		},
		{
			name: "DEL with multiple keys",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "k1", "value": "v1"}},
				{Command: "SET", Body: map[string]interface{}{"key": "k2", "value": "v2"}},
				{Command: "DEL", Body: map[string]interface{}{"keys": [...]string{"k1", "k2"}}},
				{Command: "GET", Body: map[string]interface{}{"key": "k1"}},
				{Command: "GET", Body: map[string]interface{}{"key": "k2"}},
			},
			expected: []interface{}{"OK", "OK", float64(2), nil, nil},
		},
		{
			name: "DEL with key not set",
			commands: []HTTPCommand{
				{Command: "GET", Body: map[string]interface{}{"key": "k3"}},
				{Command: "DEL", Body: map[string]interface{}{"key": "k3"}},
			},
			expected: []interface{}{nil, float64(0)},
		},
		{
			name: "DEL with no keys or arguments",
			commands: []HTTPCommand{
				{Command: "DEL", Body: map[string]interface{}{}},
			},
			expected: []interface{}{"Invalid HTTP request format"},
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

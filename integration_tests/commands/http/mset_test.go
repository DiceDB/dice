package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMSET(t *testing.T) {
	exec := NewHTTPCommandExecutor()
	testCases := []TestCase{
		{
			name: "MSET with one key-value pair",
			commands: []HTTPCommand{
				{Command: "MSET", Body: map[string]interface{}{"key_values": map[string]interface{}{"k1": "v1"}}},
				{Command: "GET", Body: map[string]interface{}{"key": "k1"}},
			},
			expected: []interface{}{"OK", "v1"},
		},
		{
			name: "MSET with multiple key-value pairs",
			commands: []HTTPCommand{
				{Command: "MSET", Body: map[string]interface{}{"key_values": map[string]interface{}{"k1": "v1", "k2": "v2"}}},
				{Command: "GET", Body: map[string]interface{}{"key": "k1"}},
				{Command: "GET", Body: map[string]interface{}{"key": "k2"}},
			},
			expected: []interface{}{"OK", "v1", "v2"},
		},
		{
			name: "MSET with integers arguments",
			commands: []HTTPCommand{
				{Command: "MSET", Body: map[string]interface{}{"key_values": map[string]interface{}{"k1": 12345, "k2": 12345}}},
				{Command: "GET", Body: map[string]interface{}{"key": "k1"}},
				{Command: "GET", Body: map[string]interface{}{"key": "k2"}},
			},
			expected: []interface{}{"OK", float64(12345), float64(12345)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			exec.FireCommand(HTTPCommand{
				Command: "DEL",
				Body:    map[string]interface{}{"keys": []interface{}{"k1"}},
			})
			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)

				assert.Equal(t, tc.expected[i], result)
			}
		})
	}

	// Deleting the used keys
	exec.FireCommand(HTTPCommand{
		Command: "DEL",
		Body: map[string]interface{}{
			"keys": []interface{}{"k1", "k2"},
		},
	})
}

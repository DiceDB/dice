package http

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestGETRANGE(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []TestCase{
		{
			name: "Get range on a string",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "k1", "value": "shankar"}},
				{Command: "GETRANGE", Body: map[string]interface{}{"key": "k1", "values": []interface{}{0, 7}}},
			},
			expected: []interface{}{"OK", "shankar"},
		},
		{
			name: "Get range on a non existent key",
			commands: []HTTPCommand{
				{Command: "GETRANGE", Body: map[string]interface{}{"key": "k2", "values": []interface{}{0, 7}}},
			},
			expected: []interface{}{""},
		},
		{
			name: "Get range on wrong key type",
			commands: []HTTPCommand{
				{Command: "LPUSH", Body: map[string]interface{}{"key": "k3", "value": "shankar"}},
				{Command: "GETRANGE", Body: map[string]interface{}{"key": "k3", "values": []interface{}{0, 7}}},
			},
			expected: []interface{}{float64(1), "WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
		{
			name: "GETRANGE against string value: 0, -1",
			commands: []HTTPCommand{
				{Command: "GETRANGE", Body: map[string]interface{}{"key": "k1", "values": []interface{}{0, -1}}},
			},
			expected: []interface{}{"shankar"},
		},
		{
			name: "GETRANGE against string value: 5, 3",
			commands: []HTTPCommand{
				{Command: "GETRANGE", Body: map[string]interface{}{"key": "k1", "values": []interface{}{5, 3}}},
			},
			expected: []interface{}{""},
		},
		{
			name: "GETRANGE against integer value: -1, -100",
			commands: []HTTPCommand{
				{Command: "GETRANGE", Body: map[string]interface{}{"key": "k1", "values": []interface{}{-1, -100}}},
			},
			expected: []interface{}{""},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)
				assert.Equal(t, tc.expected[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}
}

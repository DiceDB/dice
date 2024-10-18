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
				{Command: "SET", Body: map[string]interface{}{"key": "test1", "value": "shankar"}},
				{Command: "GETRANGE", Body: map[string]interface{}{"key": "test1", "values": []interface{}{0, 7}}},
			},
			expected: []interface{}{"OK", "shankar"},
		},
		{
			name: "Get range on a non existent key",
			commands: []HTTPCommand{
				{Command: "GETRANGE", Body: map[string]interface{}{"key": "test2", "values": []interface{}{0, 7}}},
			},
			expected: []interface{}{""},
		},
		{
			name: "Get range on wrong key type",
			commands: []HTTPCommand{
				{Command: "LPUSH", Body: map[string]interface{}{"key": "test3", "value": "shankar"}},
				{Command: "GETRANGE", Body: map[string]interface{}{"key": "test3", "values": []interface{}{0, 7}}},
			},
			expected: []interface{}{float64(1), "WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
		{
			name: "GETRANGE against string value: 0, -1",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "test4", "value": "shankar"}},
				{Command: "GETRANGE", Body: map[string]interface{}{"key": "test4", "values": []interface{}{0, -1}}},
			},
			expected: []interface{}{"OK", "shankar"},
		},
		{
			name: "GETRANGE against string value: 5, 3",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "test5", "value": "shankar"}},
				{Command: "GETRANGE", Body: map[string]interface{}{"key": "test5", "values": []interface{}{5, 3}}},
			},
			expected: []interface{}{"OK", ""},
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

package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandRename(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []TestCase{
		{
			name: "Set key and Rename key",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "sourceKey", "value": "hello"}},
				{Command: "GET", Body: map[string]interface{}{"key": "sourceKey"}},
				{Command: "RENAME", Body: map[string]interface{}{"keys": []interface{}{"sourceKey", "destKey"}}},
				{Command: "GET", Body: map[string]interface{}{"key": "destKey"}},
				{Command: "GET", Body: map[string]interface{}{"key": "sourceKey"}},
			},
			expected: []interface{}{"OK", "hello", "OK", "hello", nil},
		},
		{
			name: "same key for source and destination on Rename",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "Key", "value": "hello"}},
				{Command: "GET", Body: map[string]interface{}{"key": "Key"}},
				{Command: "RENAME", Body: map[string]interface{}{"keys": []interface{}{"Key", "Key"}}},
				{Command: "GET", Body: map[string]interface{}{"key": "Key"}},
			},
			expected: []interface{}{"OK", "hello", "OK", "hello"},
		},
		{
			name: "If source key doesn't exists",
			commands: []HTTPCommand{
				{Command: "RENAME", Body: map[string]interface{}{"keys": []interface{}{"unknownKey", "Key"}}},
			},
			expected: []interface{}{"ERR no such key"},
		},
		{
			name: "If source key doesn't exists and renaming the same key to the same key",
			commands: []HTTPCommand{
				{Command: "RENAME", Body: map[string]interface{}{"keys": []interface{}{"unknownKey", "unknownKey"}}},
			},
			expected: []interface{}{"ERR no such key"},
		},
		{
			name: "If destination Key already presents",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "destinationKey", "value": "world"}},
				{Command: "SET", Body: map[string]interface{}{"key": "newKey", "value": "hello"}},
				{Command: "RENAME", Body: map[string]interface{}{"keys": []interface{}{"newKey", "destinationKey"}}},
				{Command: "GET", Body: map[string]interface{}{"key": "newKey"}},
				{Command: "GET", Body: map[string]interface{}{"key": "destinationKey"}},
			},
			expected: []interface{}{"OK", "OK", "OK", nil, "hello"},
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

package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestType(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []TestCase{
		{
			name: "TYPE with invalid number of arguments",
			commands: []HTTPCommand{
				{Command: "TYPE", Body: map[string]interface{}{"keys": []interface{}{}}},
			},
			expected:      []interface{}{"ERR wrong number of arguments for 'type' command"},
			errorExpected: true,
		},
		{
			name: "TYPE for non-existent key",
			commands: []HTTPCommand{
				{Command: "TYPE", Body: map[string]interface{}{"key": "k1"}},
			},
			expected:      []interface{}{"none"},
			errorExpected: false,
		},
		{
			name: "TYPE for key with String value",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "k1", "value": "v1"}},
				{Command: "TYPE", Body: map[string]interface{}{"key": "k1"}},
			},
			expected:      []interface{}{"OK", "string"},
			errorExpected: false,
		},
		{
			name: "TYPE for key with List value",
			commands: []HTTPCommand{
				{Command: "LPUSH", Body: map[string]interface{}{"key": "k1", "value": "v1"}},
				{Command: "TYPE", Body: map[string]interface{}{"key": "k1"}},
			},
			expected:      []interface{}{float64(1), "list"},
			errorExpected: false,
		},
		{
			name: "TYPE for key with Set value",
			commands: []HTTPCommand{
				{Command: "SADD", Body: map[string]interface{}{"key": "k1", "value": "v1"}},
				{Command: "TYPE", Body: map[string]interface{}{"key": "k1"}},
			},
			expected:      []interface{}{float64(1), "set"},
			errorExpected: false,
		},
		{
			name: "TYPE for key with Hash value",
			commands: []HTTPCommand{
				{Command: "HSET", Body: map[string]interface{}{"key": "k1", "field": "field1", "value": "v1"}},
				{Command: "TYPE", Body: map[string]interface{}{"key": "k1"}},
			},
			expected:      []interface{}{float64(1), "hash"},
			errorExpected: false,
		},
		{
			name: "TYPE for key with value created from SETBIT command",
			commands: []HTTPCommand{
				{Command: "SETBIT", Body: map[string]interface{}{"values": []interface{}{"k1", 1, 1}}},
				{Command: "TYPE", Body: map[string]interface{}{"key": "k1"}},
			},
			expected:      []interface{}{float64(0), "string"},
			errorExpected: false,
		},
		{
			name: "TYPE for key with value created from BITOP command",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "key1", "value": "foobar"}},
				{Command: "SET", Body: map[string]interface{}{"key": "key2", "value": "abcdef"}},
				{Command: "BITOP", Body: map[string]interface{}{"values": []interface{}{"AND", "dest", "key1", "key2"}}},
				{Command: "TYPE", Body: map[string]interface{}{"key": "dest"}},
			},
			expected:      []interface{}{"OK", "OK", float64(6), "string"},
			errorExpected: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			keys := []string{"foo", "k1", "key1", "key2"}

			for _, key := range keys {
				exec.FireCommand(HTTPCommand{
					Command: "DEL",
					Body:    map[string]interface{}{"key": key},
				})
			}

			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}

}

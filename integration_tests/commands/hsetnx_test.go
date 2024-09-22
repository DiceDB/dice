package commands

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestHSETNX(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	testCases := []TestCase{
		{
			commands: []string{"HSETNX key field value", "HSET key field value_new"},
			expected: []interface{}{ONE, ZERO},
		},
		{
			commands: []string{"HSETNX key field1 value1"},
			expected: []interface{}{ONE},
		},
		{
			commands: []string{"HSETNX key_new field value", "HSETNX key_new field new_value", "HSETNX key_new"},
			expected: []interface{}{ONE, ZERO, "ERR wrong number of arguments for 'hsetnx' command"},
		},
		{
			commands: []string{"SET k v", "HSETNX k f v"},
			expected: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
	}

	for _, tc := range testCases {
		for i, cmd := range tc.commands {
			result := FireCommand(conn, cmd)
			assert.DeepEqual(t, tc.expected[i], result)
		}
	}
}

package tests

import (
	"testing"

	"gotest.tools/v3/assert"
)

var ZERO int64 = 0
var ONE int64 = 1
var TWO int64 = 2

func TestHSET(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	testCases := []TestCase{
		{
			commands: []string{"HSET key field value", "HSET key field value_new"},
			expected: []interface{}{ONE, ZERO},
		},
		{
			commands: []string{"HSET key field1 value1"},
			expected: []interface{}{ONE},
		},
		{
			commands: []string{"HSET key field2 value2 field3 value3"},
			expected: []interface{}{TWO},
		},
		{
			commands: []string{"HSET key_new field value", "HSET key_new field new_value", "HSET key_new"},
			expected: []interface{}{ONE, ZERO, "ERR wrong number of arguments for command"},
		},
		{
			commands: []string{"SET k v", "HSET k f v"},
			expected: []interface{}{"OK", "ERR key already exists"},
		},
	}

	for _, tc := range testCases {
		for i, cmd := range tc.commands {
			result := fireCommand(conn, cmd)
			assert.DeepEqual(t, tc.expected[i], result)
		}
	}
}

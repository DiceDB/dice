package async

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestHSETNX(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	testCases := []TestCase{
		{
			commands: []string{"HSETNX key_nx_t1 field value", "HSET key_nx_t1 field value_new"},
			expected: []interface{}{ONE, ZERO},
		},
		{
			commands: []string{"HSETNX key_nx_t2 field1 value1"},
			expected: []interface{}{ONE},
		},
		{
			commands: []string{"HSETNX key_nx_t3 field value", "HSETNX key_nx_t3 field new_value", "HSETNX key_nx_t3"},
			expected: []interface{}{ONE, ZERO, "ERR wrong number of arguments for 'hsetnx' command"},
		},
		{
			commands: []string{"SET key_nx_t4 v", "HSETNX key_nx_t4 f v"},
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

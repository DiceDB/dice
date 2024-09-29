package async

import (
	"testing"

	"github.com/dicedb/dice/testutils"
	"gotest.tools/v3/assert"
)

func TestHKEYS(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()
	defer FireCommand(conn, "DEL key_hKeys1 key_hKeys2 key_hKeys3 key")

	testCases := []TestCase{
		{
			commands: []string{"HKEYS", "HKEYS KEY1 KEY2"},
			expected: []interface{}{"ERR wrong number of arguments for 'hkeys' command",
				"ERR wrong number of arguments for 'hkeys' command"},
		},
		{
			commands: []string{"HSET key_hKeys1 field1 value1", "HKEYS wrong_key_hKeys1"},
			expected: []interface{}{int64(1), []interface{}{}},
		},
		{
			commands: []string{"HSET key_hKeys2 field2 value2", "HKEYS key_hKeys2"},
			expected: []interface{}{int64(1), []interface{}{"field2"}},
		},
		{
			commands: []string{"HSET key_hKeys3 field3 value3 field4 value4", "HKEYS key_hKeys3"},
			expected: []interface{}{int64(2), []interface{}{"field3", "field4"}},
		},
		{
			commands: []string{"SET key value", "HKEYS key"},
			expected: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
	}

	for _, tc := range testCases {
		for i, cmd := range tc.commands {
			result := FireCommand(conn, cmd)
			if slice, ok := tc.expected[i].([]interface{}); ok {
				assert.Assert(t, testutils.UnorderedEqual(slice, result))
			} else {
				assert.DeepEqual(t, tc.expected[i], result)
			}
		}
	}
}

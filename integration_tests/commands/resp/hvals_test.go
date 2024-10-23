package resp

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestHVals(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	testCases := []TestCase{
		{
			name:     "RESP HVALS with multiple fields",
			commands: []string{"HSET hvalsKey field value", "HSET hvalsKey field2 value_new", "HVALS hvalsKey"},
			expected: []interface{}{int64(1), int64(1), []any{string("value"), string("value_new")}},
		},
		{
			name:     "RESP HVALS with non-existing key",
			commands: []string{"HVALS hvalsKey01"},
			expected: []interface{}{[]any{}},
		},
		{
			name:     "HVALS on wrong key type",
			commands: []string{"SET hvalsKey02 field", "HVALS hvalsKey02"},
			expected: []interface{}{"OK", "ERR -WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
		{
			name:     "HVALS with wrong number of arguments",
			commands: []string{"HVALS hvalsKey03 x", "HVALS"},
			expected: []interface{}{"ERR wrong number of arguments for 'hvals' command", "ERR wrong number of arguments for 'hvals' command"},
		},
		{
			name:     "RESP One or more vals exist",
			commands: []string{"HSET key field value", "HSET key field1 value1", "HVALS key"},
			expected: []interface{}{int64(1), int64(1), []interface{}{"value", "value1"}},
		},
		{
			name:     "RESP No values exist",
			commands: []string{"HVALS key"},
			expected: []interface{}{[]interface{}{}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			FireCommand(conn, "HDEL key field")
			FireCommand(conn, "HDEL key field1")

			for i, cmd := range tc.commands {
				result := FireCommand(conn, cmd)
				assert.DeepEqual(t, tc.expected[i], result)
			}
		})
	}
}

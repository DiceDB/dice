package async

import (
	"testing"

	testifyAssert "github.com/stretchr/testify/assert"
)

func TestHKEYS(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()
	defer FireCommand(conn, "DEL key_hKeys1 key_hKeys2 key_hKeys3 key")

	testCases := []TestCase{
		{
			name:     "HKEYS with wrong number of arguments",
			commands: []string{"HKEYS", "HKEYS KEY1 KEY2"},
			expected: []interface{}{"ERR wrong number of arguments for 'hkeys' command",
				"ERR wrong number of arguments for 'hkeys' command"},
		},
		{
			name:     "HKEYS with non-existent key",
			commands: []string{"HSET key_hKeys1 field1 value1", "HKEYS wrong_key_hKeys1"},
			expected: []interface{}{int64(1), []interface{}{}},
		},
		{
			name:     "HKEYS with key containing hash with one field",
			commands: []string{"HSET key_hKeys2 field2 value2", "HKEYS key_hKeys2"},
			expected: []interface{}{int64(1), []interface{}{"field2"}},
		},
		{
			name:     "HKEYS with key containing hash with multiple fields",
			commands: []string{"HSET key_hKeys3 field3 value3 field4 value4", "HKEYS key_hKeys3"},
			expected: []interface{}{int64(2), []interface{}{"field3", "field4"}},
		},
		{
			name:     "HKEYS with key containing a non-hash value",
			commands: []string{"SET key value", "HKEYS key"},
			expected: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
	}

	for _, tc := range testCases {
		for i, cmd := range tc.commands {
			result := FireCommand(conn, cmd)
			if slice, ok := tc.expected[i].([]interface{}); ok {
				testifyAssert.ElementsMatch(t, slice, result)
			} else {
				testifyAssert.Equal(t, tc.expected[i], result)
			}
		}
	}
}

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
			name:     "HKEYS with key containing hash with multiple fields",
			commands: []string{"HSET key_hkeys field1 value1", "HSET key_hkeys field2 value2", "HKEYS key_hkeys"},
			expected: []interface{}{int64(1), int64(1), []interface{}{"field1", "field2"}},
		},
		{
			name:     "HKEYS with non-existent key",
			commands: []string{"HKEYS key_hkeys01"},
			expected: []interface{}{[]interface{}{}},
		},
		{
			name:     "HKEYS with key containing a non-hash value",
			commands: []string{"SET key_hkeys02 field1", "HKEYS key_hkeys02"},
			expected: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
		{
			name:     "HKEYS with wrong number of arguments",
			commands: []string{"HKEYS key_hkeys03 x", "HKEYS"},
			expected: []interface{}{"ERR wrong number of arguments for 'hkeys' command",
					"ERR wrong number of arguments for 'hkeys' command"},
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
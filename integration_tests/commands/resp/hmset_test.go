package resp

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHMSET(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()
	defer FireCommand(conn, "DEL key key_new key2")

	testCases := []TestCase{
		{
			commands: []string{"HMSET key field value field2 value2", "HGET key field",
				"HGET key field2", "HMSET key field3 value_new field value4", "HGET key field3", "HGET key field2",
				"HGET key field"},
			expected: []interface{}{"OK", "value", "value2", "OK", "value_new", "value2", "value4"},
		},
		{
			commands: []string{"HMSET key2 field1 value1", "HGET key2 xxxx", "HGET key2 field1"},
			expected: []interface{}{"OK", "(nil)", "value1"},
		},
		{
			commands: []string{"HMSET key field2 value2 field3 value3"},
			expected: []interface{}{"OK"},
		},
		{
			commands: []string{"HMSET key_new field value field2 value2", "HMSET key_new field new_value", "HMSET key_new",
				"HGET key_new field", "HGET key_new field2"},
			expected: []interface{}{"OK", "OK", "ERR wrong number of arguments for 'hmset' command", "new_value", "value2"},
		},
		{
			commands: []string{"SET k v", "HMSET k f v"},
			expected: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
	}

	for _, tc := range testCases {
		for i, cmd := range tc.commands {
			result := FireCommand(conn, cmd)
			assert.Equal(t, tc.expected[i], result)
		}
	}
}

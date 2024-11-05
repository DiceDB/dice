package resp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHMGET(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()
	defer FireCommand(conn, "DEL key_hmGet key_hmGet1")

	testCases := []TestCase{
		{
			name:     "hmget existing keys and fields",
			commands: []string{"HSET key_hmGet field value", "HSET key_hmGet field2 value_new", "HMGET key_hmGet field field2"},
			expected: []interface{}{ONE, ONE, []interface{}{"value", "value_new"}},
		},
		{
			name:     "hmget key does not exist",
			commands: []string{"HMGET doesntexist field"},
			expected: []interface{}{[]interface{}{"(nil)"}},
		},
		{
			name:     "hmget field does not exist",
			commands: []string{"HMGET key_hmGet field3"},
			expected: []interface{}{[]interface{}{"(nil)"}},
		},
		{
			name:     "hmget some fields do not exist",
			commands: []string{"HMGET key_hmGet field field2 field3 field3"},
			expected: []interface{}{[]interface{}{"value", "value_new", "(nil)", "(nil)"}},
		},
		{
			name:     "hmget with wrongtype",
			commands: []string{"SET key_hmGet1 field", "HMGET key_hmGet1 field"},
			expected: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
		{
			name:     "wrong number of arguments",
			commands: []string{"HMGET key_hmGet", "HMGET"},
			expected: []interface{}{"ERR wrong number of arguments for 'hmget' command",
				"ERR wrong number of arguments for 'hmget' command"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				// Fire the command and get the result
				result := FireCommand(conn, cmd)
				assert.Equal(t, result, tc.expected[i])
			}
		})
	}
}

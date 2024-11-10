package resp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHGET(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()
	defer FireCommand(conn, "DEL key_hGet key")

	testCases := []TestCase{
		{
			commands: []string{"HGET", "HGET KEY", "HGET KEY FIELD ANOTHER_FIELD"},
			expected: []interface{}{"ERR wrong number of arguments for 'hget' command",
				"ERR wrong number of arguments for 'hget' command",
				"ERR wrong number of arguments for 'hget' command"},
		},
		{
			commands: []string{"HSET key_hGet field value", "HSET key_hGet field newvalue"},
			expected: []interface{}{ONE, ZERO},
		},
		{
			commands: []string{"HGET wrong_key_hGet field"},
			expected: []interface{}{"(nil)"},
		},
		{
			commands: []string{"HGET key_hGet wrong_field"},
			expected: []interface{}{"(nil)"},
		},
		{
			commands: []string{"HGET key_hGet field"},
			expected: []interface{}{"newvalue"},
		},
		{
			commands: []string{"SET key value", "HGET key field"},
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

package async

import (
	"testing"

	testifyAssert "github.com/stretchr/testify/assert"
)

func TestHEXISTS(t *testing.T) {
	conn := getLocalConnection()
	FireCommand(conn, "DEL key_hExists1 key_hExists2 key_hExists3 key")

	defer conn.Close()
	defer FireCommand(conn, "DEL key_hExists1 key_hExists2 key_hExists3 key")

	testCases := []TestCase{
		{
			commands: []string{"HEXISTS", "HEXISTS KEY", "HEXISTS KEY FIELD ANOTHER_FIELD"},
			expected: []interface{}{"ERR wrong number of arguments for 'hexists' command",
				"ERR wrong number of arguments for 'hexists' command",
				"ERR wrong number of arguments for 'hexists' command"},
		},
		{
			commands: []string{"HSET key_hExists1 field value", "HEXISTS wrong_key_hExists field"},
			expected: []interface{}{int64(1), int64(0)},
		},
		{
			commands: []string{"HSET key_hExists2 field value", "HEXISTS key_hExists2 wrong_field"},
			expected: []interface{}{int64(1), int64(0)},
		},
		{
			commands: []string{"HSET key_hExists3 field HelloWorld", "HEXISTS key_hExists3 field"},
			expected: []interface{}{int64(1), int64(1)},
		},
		{
			commands: []string{"SET key value", "HEXISTS key field"},
			expected: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
	}

	for _, tc := range testCases {
		for i, cmd := range tc.commands {
			result := FireCommand(conn, cmd)
			testifyAssert.Equal(t, tc.expected[i], result)
		}
	}
}

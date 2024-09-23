package async

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestHLEN(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()
	defer FireCommand(conn, "DEL key_hLen hash1 hash2")

	testCases := []TestCase{
		{
			commands: []string{"HLEN", "HLEN key1 key2"},
			expected: []interface{}{
				"ERR wrong number of arguments for 'hlen' command",
				"ERR wrong number of arguments for 'hlen' command",
			},
		},
		{
			commands: []string{"HSET key_hLen field1 value1", "HLEN key_hLen"},
			expected: []interface{}{int64(1), int64(1)},
		},
		{
			commands: []string{"HSET key_hLen field2 value2", "HLEN key_hLen"},
			expected: []interface{}{int64(1), int64(2)},
		},
		{
			commands: []string{"HLEN nonexistent_key"},
			expected: []interface{}{int64(0)},
		},
		{
			commands: []string{"SET hash1 value", "HLEN hash1"},
			expected: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
		{
			commands: []string{
				"HSET hash2 f1 v1 f2 v2 f3 v3",
				"HLEN hash2",
				"DEL hash2",
				"HLEN hash2",
			},
			expected: []interface{}{int64(3), int64(3), int64(1), int64(0)},
		},
	}

	for _, tc := range testCases {
		for i, cmd := range tc.commands {
			result := FireCommand(conn, cmd)
			assert.DeepEqual(t, tc.expected[i], result)
		}
	}
}

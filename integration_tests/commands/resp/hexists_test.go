package resp

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestHExists(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	testCases := []TestCase{
		{
			name:     "RESP Check if field exists when k f and v are set",
			commands: []string{"HSET key field value", "HEXISTS key field"},
			expected: []interface{}{int64(1), int64(1)},
		},
		{
			name:     "RESP Check if field exists when k exists but not f and v",
			commands: []string{"HSET key field1 value", "HEXISTS key field"},
			expected: []interface{}{int(1), int64(0)},
		},
		{
			name:     "RESP Check if field exists when no k,f and v exist",
			commands: []string{"HEXISTS key field"},
			expected: []interface{}{int64(0)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			FireCommand(conn, "HDEL key field")

			for i, cmd := range tc.commands {
				result := FireCommand(conn, cmd)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}
}

package tests

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestBgrewriteaof(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	testCases := []TestCase{
		{
			commands: []string{"SET k1 v1", "SET k2 v2", "BGREWRITEAOF"},
			expected: []interface{}{"OK", "OK", "OK"},
		},
	}

	for _, tc := range testCases {
		for i, cmd := range tc.commands {
			result := fireCommand(conn, cmd)
			assert.DeepEqual(t, tc.expected[i], result)
		}
	}
}

package commands

import (
	"testing"

	dstore "github.com/dicedb/dice/internal/store"
	"gotest.tools/v3/assert"
)

func TestType(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
		setup    func(store *dstore.Store)
	}{
		{
			name:     "TYPE for non-existent key",
			commands: []string{"TYPE k1"},
			expected: []interface{}{"none"},
		},
		{
			name:     "TYPE for key with String value",
			commands: []string{"SET k1 v1", "TYPE k1"},
			expected: []interface{}{"OK", "string"},
		},
		{
			name:     "TYPE for key with List value",
			commands: []string{"LPUSH k1 v1", "TYPE k1"},
			expected: []interface{}{"OK", "list"},
		},
		{
			name:     "TYPE for key with Set value",
			commands: []string{"SADD k1 v1", "TYPE k1"},
			expected: []interface{}{int64(1), "set"},
		},
		{
			name:     "TYPE for key with Hash value",
			commands: []string{"HSET k1 field1 v1", "TYPE k1"},
			expected: []interface{}{int64(1), "hash"},
		},
		// TODO: Adding the support of command ZADD and XADD to run the integration test for ZSet and Stream.
		/*
			{
				name:     "TYPE for key with ZSet value",
				commands: []string{"ZADD k1 1 v1", "TYPE k1"},
				expected: []interface{}{int64(1), "zset"},
			},
			{
				name:     "TYPE for key with Stream value",
				commands: []string{"XADD k1 * field1 v1", "TYPE k1"},
				expected: []interface{}{"OK", "stream"},
			},
		*/
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			FireCommand(conn, "DEL k1")

			for i, cmd := range tc.commands {
				result := FireCommand(conn, cmd)
				assert.Equal(t, tc.expected[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}
}

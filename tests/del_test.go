package tests

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestDel(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name:     "DEL with set key",
			commands: []string{"SET k1 v1", "DEL k1"},
			expected: []interface{}{"OK", int64(1)},
		},
		{
			name:     "DEL with key not set",
			commands: []string{"DEL k2"},
			expected: []interface{}{int64(0)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			deleteTestKeys([]string{"k1", "k2"})
			for i, cmd := range tc.commands {
				result := fireCommand(conn, cmd)
				assert.Equal(t, tc.expected[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}
}


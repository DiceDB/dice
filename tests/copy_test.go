package tests

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestCopy(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name:     "COPY when source key doesn't exist",
			commands: []string{"COPY k3 k2"},
			expected: []interface{}{int64(0)},
		},
		{
			name:     "COPY with no REPLACE",
			commands: []string{"SET k1 v1", "COPY k1 k2", "GET k1", "GET k2"},
			expected: []interface{}{"OK", int64(1), "v1", "v1"},
		},
		{
			name:     "COPY with REPLACE",
            commands: []string{"SET k4 v1", "SET k5 v2", "GET k5", "COPY k4 k5 REPLACE", "GET k5"},
			expected: []interface{}{"OK", "OK", "v2", int64(1), "v1"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			deleteTestKeys([]string{"k1", "k2", "k3", "k4", "k5"})
			for i, cmd := range tc.commands {
				result := fireCommand(conn, cmd)
				assert.Equal(t, tc.expected[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}
}

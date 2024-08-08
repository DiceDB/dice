package tests

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestMset(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name:     "MSET with one key-value pair",
			commands: []string{"MSET k1 v1", "GET k1"},
			expected: []interface{}{"OK", "v1"},
		},
		{
			name:     "MSET with multiple key-value pairs",
			commands: []string{"MSET k1 v1 k2 v2", "GET k1", "GET k2"},
			expected: []interface{}{"OK", "v1", "v2"},
		},
		{
			name:     "MSET with odd number of arguments",
			commands: []string{"MSET k1 v1 k2"},
			expected: []interface{}{"ERR wrong number of arguments for 'mset' command"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			deleteTestKeys([]string{"k1", "k2"})
			for i, cmd := range tc.commands {
				result := fireCommand(conn, cmd)
				assert.DeepEqual(t, tc.expected[i], result)
			}
		})
	}
}

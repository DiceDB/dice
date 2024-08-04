package tests

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestSet(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name:     "Set and Get Simple Value",
			commands: []string{"SET k v", "GET k"},
			expected: []interface{}{"OK", "v"},
		},
		{
			name:     "Overwrite Existing Key",
			commands: []string{"SET k v1", "SET k v2", "GET k"},
			expected: []interface{}{"OK", "OK", "v2"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			deleteTestKeys([]string{"k"})
			for i, cmd := range tc.commands {
				result := fireCommand(conn, cmd)
				assert.DeepEqual(t, tc.expected[i], result)
			}
		})
	}
}

func TestSetWithOptions(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name:     "XX on non-existing key",
			commands: []string{"DEL k", "SET k v XX", "GET k"},
			expected: []interface{}{int64(0), "(nil)", "(nil)"},
		},
		// ... (keep other test cases)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			deleteTestKeys([]string{"k"})
			for i, cmd := range tc.commands {
				result := fireCommand(conn, cmd)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}
}

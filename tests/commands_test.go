package tests

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestCommandRandomKey(t *testing.T) {
	var randomKeyTestCases = []struct {
		name     string
		inCmd    []string
		expected []interface{}
	}{
		{
			name:     "Try to get RandomKey with 0 keys in map",
			inCmd:    []string{"RANDOMKEY"},
			expected: []interface{}{"(nil)"},
		},
		{
			name:     "set a key and get a RandomKey",
			inCmd:    []string{"set Key hello", "RANDOMKEY"},
			expected: []interface{}{"OK", "Key"},
		},
		{
			name:  "Set another two keys and check the RandomKey",
			inCmd: []string{"set Key2 hello2", ""},
		},
	}
	conn := getLocalConnection()
	defer conn.Close()

	for _, tc := range randomKeyTestCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.inCmd {
				result := fireCommand(conn, cmd)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}
}

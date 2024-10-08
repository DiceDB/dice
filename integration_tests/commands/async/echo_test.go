package async

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEcho(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name:     "ECHO with invalid number of arguments",
			commands: []string{"ECHO"},
			expected: []interface{}{"ERR wrong number of arguments for 'echo' command"},
		},
		{
			name:     "ECHO with one argument",
			commands: []string{"ECHO \"hello world\""},
			expected: []interface{}{"hello world"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result := FireCommand(conn, cmd)
				assert.Equal(t, tc.expected[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}
}

package tests

import (
	"fmt"
	"testing"

	"gotest.tools/v3/assert"
)

func TestCommandCOMMAND(t *testing.T) {
	conn := getLocalConnection()

	tests := []struct {
		subcommand string
		expected   string
	}{
		{
			subcommand: "invalid-cmd",
			expected:   "ERR unknown subcommand 'invalid-cmd'. Try COMMAND HELP",
		},
	}

	for _, test := range tests {
		actual := fireCommand(conn, fmt.Sprintf("COMMAND %s", test.subcommand))
		assert.Equal(t, actual, test.expected)
	}
}

package async

import (
	"strings"
	"testing"

	"gotest.tools/v3/assert"
)

func TestClientSetName(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	tests := []struct {
		name     string
		command  string
		expected string
	}{
		{
			name:     "Set valid name without spaces",
			command:  "K",
			expected: "OK",
		},
		{
			name:     "Set valid name with trailing space",
			command:  "K ",
			expected: "OK",
		},
		{
			name:     "Too many arguments for SETNAME",
			command:  "K K",
			expected: "ERR wrong number of arguments for 'client|setname' command",
		},
		{
			name:     "Name with space between characters",
			command:  "\"K K\"",
			expected: "ERR Client names cannot contain spaces, newlines or special characters.",
		},
		{
			name:     "Empty name argument",
			command:  " ",
			expected: "ERR wrong number of arguments for 'client|setname' command",
		},
		{
			name:     "Missing name argument",
			command:  "",
			expected: "ERR wrong number of arguments for 'client|setname' command",
		},
		{
			name:     "Name with newline character",
			command:  "\n",
			expected: "ERR Client names cannot contain spaces, newlines or special characters.",
		},
		{
			name:     "Name with valid character followed by newline",
			command:  "K\n",
			expected: "ERR Client names cannot contain spaces, newlines or special characters.",
		},
		{
			name:     "Name with special character",
			command:  "K%",
			expected: "ERR Client names cannot contain spaces, newlines or special characters.",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := FireCommand(conn, "CLIENT SETNAME "+tc.command)
			assert.DeepEqual(t, tc.expected, result)
		})
	}
}

func TestClientGetName(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	tests := []struct {
		name     string
		command  string
		expected string
	}{
		{
			name:     "GetName with invalid argument",
			command:  "CLIENT GETNAME invalid-arg",
			expected: "ERR wrong number of arguments for 'client|getname' command",
		},
		{
			name:     "GetName with no name set",
			command:  "CLIENT GETNAME",
			expected: "(nil)",
		},
		{
			name:     "SetName with invalid name containing space",
			command:  "CLIENT SETNAME \"K K\"; CLIENT GETNAME",
			expected: "(nil)",
		},
		{
			name:     "SetName with valid name",
			command:  "CLIENT SETNAME K; CLIENT GETNAME",
			expected: "K",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Split multiple commands (like "CLIENT SETNAME" followed by "CLIENT GETNAME") if needed
			commands := strings.Split(tt.command, "; ")
			var result interface{}
			for _, cmd := range commands {
				result = FireCommand(conn, strings.TrimSpace(cmd))
			}
			assert.DeepEqual(t, tt.expected, result)
		})
	}
}

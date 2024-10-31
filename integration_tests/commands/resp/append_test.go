package resp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAPPEND(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
		cleanup  []string
	}{
		{
			name:     "APPEND and GET a new Val",
			commands: []string{"APPEND k newVal", "GET k"},
			expected: []interface{}{int64(6), "newVal"},
			cleanup:  []string{"del k"},
		},
		{
			name:     "APPEND to an existing key and GET",
			commands: []string{"SET k Bhima", "APPEND k Shankar", "GET k"},
			expected: []interface{}{"OK", int64(12), "BhimaShankar"},
			cleanup:  []string{"del k"},
		},
		{
			name:     "APPEND without input value",
			commands: []string{"APPEND k"},
			expected: []interface{}{"ERR wrong number of arguments for 'append' command"},
			cleanup:  []string{"del k"},
		},
		{
			name:     "APPEND empty string to an existing key with empty string",
			commands: []string{"SET k \"\"", "APPEND k \"\""},
			expected: []interface{}{"OK", int64(0)},
			cleanup:  []string{"del k"},
		},
		{
			name:     "APPEND to key created using LPUSH",
			commands: []string{"LPUSH m bhima", "APPEND m shankar"},
			expected: []interface{}{int64(1), "WRONGTYPE Operation against a key holding the wrong kind of value"},
			cleanup:  []string{"del m"},
		},
		{
			name:     "APPEND value with leading zeros",
			commands: []string{"APPEND z 0043"},
			expected: []interface{}{int64(4)},
			cleanup:  []string{"del z"},
		},
		{
			name:     "APPEND to key created using SADD",
			commands: []string{"SADD key apple", "APPEND key banana"},
			expected: []interface{}{int64(1), "WRONGTYPE Operation against a key holding the wrong kind of value"},
			cleanup:  []string{"del key"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			for i := 0; i < len(tc.commands); i++ {
				result := FireCommand(conn, tc.commands[i])
				expected := tc.expected[i]
				assert.Equal(t, expected, result)
			}

			for _, cmd := range tc.cleanup {
				FireCommand(conn, cmd)
			}
		})
	}
}

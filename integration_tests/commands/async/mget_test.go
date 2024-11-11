package async

import (
	"testing"

	"github.com/dicedb/dice/testutils"
	"github.com/stretchr/testify/assert"
)

func TestMGET(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	defer FireCommand(conn, "DEL k1")
	defer FireCommand(conn, "DEL k2")

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name:     "MGET With non-existing keys",
			commands: []string{"MGET k1 k2"},
			expected: []interface{}{[]interface{}{"(nil)", "(nil)"}},
		},
		{
			name:     "MGET With existing keys",
			commands: []string{"MSET k1 v1 k2 v2", "MGET k1 k2"},
			expected: []interface{}{"OK", []interface{}{"v1", "v2"}},
		},
		{
			name:     "MGET with existing and non existing keys",
			commands: []string{"set k1 v1", "MGET k1 k2"},
			expected: []interface{}{"OK", []interface{}{"v1", "(nil)"}},
		},
		{
			name:     "MGET without any keys",
			commands: []string{"MGET"},
			expected: []interface{}{"ERR wrong number of arguments for 'mget' command"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			FireCommand(conn, "DEL k1")
			FireCommand(conn, "DEL k2")

			for i, cmd := range tc.commands {
				result := FireCommand(conn, cmd)
				if slice, ok := tc.expected[i].([]interface{}); ok {
					assert.True(t, testutils.UnorderedEqual(slice, result))
				} else {
					assert.Equal(t, tc.expected[i], result)
				}
			}
		})
	}
}

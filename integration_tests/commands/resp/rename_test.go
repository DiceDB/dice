package resp

import (
	"testing"

	"github.com/dicedb/dice/testutils"
	testifyAssert "github.com/stretchr/testify/assert"
	"gotest.tools/v3/assert"
)

func TestRename(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	simpleJSON := `{"name":"John","age":30}`

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name:     "RENAME when source key doesn't exist",
			commands: []string{"RENAME k1 k2"},
			expected: []interface{}{int64(0)},
		},
		{
			name:     "RENAME with existing source key",
			commands: []string{"SET k1 v1", "RENAME k1 k2", "GET k1", "GET k2"},
			expected: []interface{}{"OK", "OK", "(nil)", "v1"},
		},
		{
			name:     "RENAME with existing destination key",
			commands: []string{"SET k1 v1", "SET k2 v2", "RENAME k1 k2", "GET k1", "GET k2"},
			expected: []interface{}{"OK", "OK", "(nil)", "v1"},
		},
		{
			name:     "RENAME JSON key",
			commands: []string{`JSON.SET k1 $ ` + simpleJSON, "RENAME k1 k2", "JSON.GET k2"},
			expected: []interface{}{"OK", int64(1), simpleJSON},
		},
		{
			name:     "RENAME with expiry",
			commands: []string{"SET k1 v1 EX 5", "RENAME k1 k2", "GET k1", "GET k2", "SLEEP 7", "GET k1", "GET k2"},
			expected: []interface{}{"OK", "OK", "(nil)", "v1", "OK", "(nil)", "(nil)"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Cleanup any existing keys before the test
			FireCommand(conn, "DEL k1")
			FireCommand(conn, "DEL k2")

			for i, cmd := range tc.commands {
				result := FireCommand(conn, cmd)
				resStr, resOk := result.(string)
				expStr, expOk := tc.expected[i].(string)

				// Compare JSON strings if both are valid JSON
				if resOk && expOk && testutils.IsJSONResponse(resStr) && testutils.IsJSONResponse(expStr) {
					testifyAssert.JSONEq(t, expStr, resStr)
				} else {
					assert.Equal(t, tc.expected[i], result, "Value mismatch for cmd %s", cmd)
				}
			}
		})
	}
}

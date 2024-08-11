package tests

import (
	"github.com/dicedb/dice/testutils"
	"gotest.tools/v3/assert"
	"strconv"
	"testing"
	"time"
)

func TestEvalEXPIREAT(t *testing.T) {
	conn := getLocalConnection()
	expireAt := time.Now().Add(10 * time.Second).Unix()

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name:     "Set with EXPIREAT command",
			commands: []string{"SET test_key test_value", "EXPIREAT test_key " + strconv.FormatInt(expireAt, 10)},
			expected: []interface{}{"OK", "OK"},
		},
		{
			name: "Check if key expires after EXPIREAT",
			commands: []string{
				"SET test_key test_value",
				"EXPIREAT test_key " + strconv.FormatInt(expireAt, 10),
			},
			expected: []interface{}{"OK", "OK"},
		},
		{
			name: "Check if key is nil after expiration",
			commands: []string{
				"GET test_key",
			},
			expected: []interface{}{"nil"}, // Expect "nil" or empty after expiration
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				// Special handling for expiration check
				if i == len(tc.commands)-1 {
					time.Sleep(11 * time.Second) // Wait to ensure the key expires
				}

				result := fireCommand(conn, cmd)

				if i == len(tc.commands)-1 {
					assert.Assert(t, result == nil || result == "nil" || result == "", "Expected nil or empty value after expiration, got:", result)
				} else if slice, ok := tc.expected[i].([]interface{}); ok {
					assert.Assert(t, testutils.UnorderedEqual(slice, result))
				} else {
					assert.DeepEqual(t, tc.expected[i], result)
				}
			}
		})
	}
}

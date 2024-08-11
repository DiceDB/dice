package tests

import (
	"gotest.tools/v3/assert"
	"strconv"
	"testing"
	"time"
)

func TestEvalEXPIREAT(t *testing.T) {
	conn := getLocalConnection()

	expireInSeconds := int64(10) // Set expiration time to 10 seconds later

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name:     "Set with EXPIREAT command",
			commands: []string{"SET test_key test_value", "EXPIREAT test_key " + strconv.FormatInt(expireInSeconds, 10)},
			expected: []interface{}{"OK", "OK"}, // Expect "OK" for successful SET and EXPIREAT commands
		},
		{
			name: "Check if key is nil after expiration",
			commands: []string{
				"GET test_key",
			},
			expected: []interface{}{"nil"}, // Expect "nil" or empty value after expiration
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				if i == len(tc.commands)-1 {
					time.Sleep(12 * time.Second) // Wait 12 seconds to ensure the key expires
				}

				result := fireCommand(conn, cmd)

				if i == len(tc.commands)-1 { // Only check the result of the final command
					if result == "nil" || result == "" {
						assert.Assert(t, true) // Expiration was successful
					} else {
						t.Fatalf("Expected nil or empty value after expiration, got: %v", result)
					}
				} else {
					assert.DeepEqual(t, tc.expected[i], result) // Assert successful execution for other commands
				}
			}
		})
	}
}

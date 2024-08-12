package tests

import (
	"gotest.tools/v3/assert"
	"strconv"
	"testing"
	"time"
)

func TestEvalEXPIREAT(t *testing.T) {
	// Establish a connection to the local server
	conn := getLocalConnection()

	// Set the expiration time to 2 seconds
	expireInSeconds := int64(2)

	// Define test cases with their respective commands and expected outcomes
	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			// Test setting a key with the EXPIREAT command
			name:     "Set with EXPIREAT command",
			commands: []string{"SET test_key test_value", "EXPIREAT test_key " + strconv.FormatInt(expireInSeconds, 10)},
			expected: []interface{}{"OK", "OK"}, // Expect "OK" for both SET and EXPIREAT commands
		},
		{
			// Test retrieving the key after it should have expired
			name: "Check if key is nil after expiration",
			commands: []string{
				"GET test_key",
			},
			expected: []interface{}{"nil"}, // Expect "nil" after key expiration
		},
	}

	// Iterate over each test case
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				// If this is the last command, wait 3 seconds to ensure the key has expired
				if i == len(tc.commands)-1 {
					time.Sleep(3 * time.Second)
				}

				// Execute the command
				result := fireCommand(conn, cmd)

				// Validate the result of the last command
				if i == len(tc.commands)-1 {
					if result == "nil" || result == "" {
						assert.Assert(t, true) // Key has expired as expected
					} else {
						t.Fatalf("Expected nil or empty value after expiration, got: %v", result)
					}
				} else {
					// Validate the result of other commands
					assert.DeepEqual(t, tc.expected[i], result)
				}
			}
		})
	}
}

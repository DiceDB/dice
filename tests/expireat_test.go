package tests

import (
	"gotest.tools/v3/assert"
	"strconv"
	"testing"
	"time"
)

func TestEvalEXPIREAT(t *testing.T) {
	conn := getLocalConnection()

	// Set the expiration time to 2 seconds
	expireInSeconds := int64(2)

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			// Test setting a key with the EXPIREAT command
			name:     "Set with EXPIREAT command",
			commands: []string{"SET test_key test_value", "EXPIREAT test_key " + strconv.FormatInt(expireInSeconds, 10)},
			expected: []interface{}{"OK", int64(1)}, // Expect "OK" for SET and 1 (int64) for successful EXPIREAT command
		},
		{
			// Test retrieving the key after it should have expired
			name: "Check if key is nil after expiration",
			commands: []string{
				"GET test_key",
			},
			expected: []interface{}{"(nil)"}, // Expect "(nil)" after key expiration
		},
		{
			// Test EXPIREAT on a non-existent key
			name:     "EXPIREAT non-existent key",
			commands: []string{"EXPIREAT non_existent_key " + strconv.FormatInt(expireInSeconds, 10)},
			expected: []interface{}{int64(0)}, // Expect 0 (int64) indicating the key does not exist
		},
		{
			// Test EXPIREAT with a time in the past (key should be deleted immediately)
			name:     "EXPIREAT with past time",
			commands: []string{"SET test_key test_value", "EXPIREAT test_key 1"},
			expected: []interface{}{"OK", int64(1)}, // Expect "OK" for SET, 1 (int64) indicating the key was deleted
		},
		{
			// Test EXPIREAT with invalid syntax (no timeout provided)
			name:     "EXPIREAT with invalid syntax",
			commands: []string{"SET test_key test_value", "EXPIREAT test_key"},
			expected: []interface{}{"OK", "ERR wrong number of arguments for 'EXPIREAT' command"}, // Expect an error message for EXPIREAT
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				// If this is the last command, wait 3 seconds to ensure the key has expired
				if i == len(tc.commands)-1 && tc.name != "EXPIREAT non-existent key" && tc.name != "EXPIREAT with invalid syntax" {
					time.Sleep(3 * time.Second)
				}

				result := fireCommand(conn, cmd)

				if i == len(tc.commands)-1 {
					if tc.name == "Check if key is nil after expiration" && (result == "(nil)" || result == "") {
						assert.Assert(t, true) // Key has expired as expected
					} else if tc.name == "EXPIREAT with past time" && result == "nil" {
						assert.Assert(t, true) // Key should be deleted immediately for past time
					} else {
						assert.DeepEqual(t, tc.expected[i], result) // Validate the result for other scenarios
					}
				} else {
					// Validate the result of other commands
					assert.DeepEqual(t, tc.expected[i], result)
				}
			}
		})
	}
}

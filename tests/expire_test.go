package tests

import (
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func TestExpire(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	testCases := []struct {
		name     string
		setup    string
		commands []string
		expected []interface{}
		delay    []time.Duration
	}{
		{
			name:  "Set with EXPIRE command",
			setup: "",
			commands: []string{
				"SET test_key test_value",
				"EXPIRE test_key 1",
			},
			expected: []interface{}{"OK", int64(1)},
			delay:    []time.Duration{0, 0},
		},
		{
			name:  "Check if key is nil after expiration",
			setup: "SET test_key test_value",
			commands: []string{
				"EXPIRE test_key 1",
				"GET test_key",
			},
			expected: []interface{}{int64(1), "(nil)"},
			delay:    []time.Duration{0, 1100 * time.Millisecond},
		},
		{
			name:  "EXPIRE non-existent key",
			setup: "",
			commands: []string{
				"EXPIRE non_existent_key 1",
			},
			expected: []interface{}{int64(0)},
			delay:    []time.Duration{0, 0},
		},
		{
			name:  "EXPIRE with past time",
			setup: "SET test_key test_value",
			commands: []string{
				"EXPIRE test_key -1",
				"GET test_key",
			},
			expected: []interface{}{int64(1), "(nil)"},
			delay:    []time.Duration{0, 0},
		},
		{
			name:  "EXPIRE with invalid syntax",
			setup: "SET test_key test_value",
			commands: []string{
				"EXPIRE test_key",
			},
			expected: []interface{}{"ERR wrong number of arguments for 'expire' command"},
			delay:    []time.Duration{0},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			if tc.setup != "" {
				fireCommand(conn, tc.setup)
			}

			// Execute commands
			var results []interface{}
			for i, cmd := range tc.commands {
				// Wait if delay is specified
				if tc.delay[i] > 0 {
					time.Sleep(tc.delay[i])
				}
				result := fireCommand(conn, cmd)
				results = append(results, result)
			}

			// Validate results
			for i, expected := range tc.expected {
				if i >= len(results) {
					t.Fatalf("Not enough results. Expected %d, got %d", len(tc.expected), len(results))
				}

				if expected == "(nil)" {
					assert.Assert(t, results[i] == "(nil)" || results[i] == "",
						"Expected nil or empty result, got %v", results[i])
				} else {
					assert.DeepEqual(t, expected, results[i])
				}
			}
		})
	}
}

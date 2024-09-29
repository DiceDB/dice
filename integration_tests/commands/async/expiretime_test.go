package async

import (
	"strconv"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func TestExpiretime(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	futureUnixTimestamp := time.Now().Unix() + 1

	testCases := []struct {
		name     string
		setup    string
		commands []string
		expected []interface{}
		delay    []time.Duration
	}{
		{
			name:  "EXPIRETIME command",
			setup: "SET test_key test_value",
			commands: []string{
				"EXPIREAT test_key " + strconv.FormatInt(futureUnixTimestamp, 10),
				"EXPIRETIME test_key",
			},
			expected: []interface{}{int64(1), futureUnixTimestamp},
			delay:    []time.Duration{0, 0},
		},
		{
			name:  "EXPIRETIME non-existent key",
			setup: "",
			commands: []string{
				"EXPIRETIME non_existent_key",
			},
			expected: []interface{}{int64(-2)},
			delay:    []time.Duration{0},
		},
		{
			name:  "EXPIRETIME with past time",
			setup: "SET test_key test_value",
			commands: []string{
				"EXPIREAT test_key 1724167183",
				"EXPIRETIME test_key",
			},
			expected: []interface{}{int64(1), int64(-2)},
			delay:    []time.Duration{0, 0},
		},
		{
			name:  "EXPIRETIME with invalid syntax",
			setup: "SET test_key test_value",
			commands: []string{
				"EXPIRETIME",
				"EXPIRETIME key1 key2",
			},
			expected: []interface{}{"ERR wrong number of arguments for 'expiretime' command", "ERR wrong number of arguments for 'expiretime' command"},
			delay:    []time.Duration{0, 0},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			if tc.setup != "" {
				FireCommand(conn, tc.setup)
			}

			// Execute commands
			var results []interface{}
			for i, cmd := range tc.commands {
				// Wait if delay is specified
				if tc.delay[i] > 0 {
					time.Sleep(tc.delay[i])
				}
				result := FireCommand(conn, cmd)
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

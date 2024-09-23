package async

import (
	"strconv"
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
			expected: []interface{}{"ERR invalid expire time in 'expire' command", "test_value"},
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
		{
			name:  "Test(NX): Set the expiration only if the key has no expiration time",
			setup: "",
			commands: []string{
				"SET test_key test_value",
				"EXPIRE test_key " + strconv.FormatInt(1, 10) + " NX",
				"EXPIRE test_key " + strconv.FormatInt(1, 10) + " NX",
			},
			expected: []interface{}{"OK", int64(1), int64(0)},
			delay:    []time.Duration{0, 0, 0},
		},

		{
			name:  "Test(XX): Set the expiration only if the key already has an expiration time",
			setup: "",
			commands: []string{
				"SET test_key test_value",
				"EXPIRE test_key " + strconv.FormatInt(10, 10) + " XX",
				"TTL test_key",
				"EXPIRE test_key " + strconv.FormatInt(10, 10),
				"EXPIRE test_key " + strconv.FormatInt(10, 10) + " XX",
			},
			expected: []interface{}{"OK", int64(0), int64(-1), int64(1), int64(1)},
			delay:    []time.Duration{0, 0, 0, 0, 0},
		},

		{
			name:  "TEST(GT): Set the expiration only if the new expiration time is greater than the current one",
			setup: "",
			commands: []string{
				"SET test_key test_value",
				"EXPIRE test_key " + strconv.FormatInt(10, 10) + " GT",
				"TTL test_key",
				"EXPIRE test_key " + strconv.FormatInt(10, 10),
				"EXPIRE test_key " + strconv.FormatInt(20, 10) + " GT",
			},
			expected: []interface{}{"OK", int64(0), int64(-1), int64(1), int64(1)},
			delay:    []time.Duration{0, 0, 0, 0, 0},
		},

		{
			name:  "TEST(LT): Set the expiration only if the new expiration time is less than the current one",
			setup: "",
			commands: []string{
				"SET test_key test_value",
				"EXPIRE test_key " + strconv.FormatInt(10, 10) + " LT",
				"EXPIRE test_key " + strconv.FormatInt(20, 10) + " LT",
			},
			expected: []interface{}{"OK", int64(1), int64(0)},
			delay:    []time.Duration{0, 0, 0},
		},

		{
			name:  "TEST(LT): Set the expiration only if the new expiration time is less than the current one",
			setup: "",
			commands: []string{
				"SET test_key test_value",
				"EXPIRE test_key " + strconv.FormatInt(10, 10) + " LT",
				"EXPIRE test_key " + strconv.FormatInt(20, 10) + " LT",
			},
			expected: []interface{}{"OK", int64(1), int64(0)},
			delay:    []time.Duration{0, 0, 0},
		},

		{
			name:  "TEST(NX + LT/GT)",
			setup: "",
			commands: []string{
				"SET test_key test_value",
				"EXPIRE test_key " + strconv.FormatInt(20, 10) + " NX",
				"EXPIRE test_key " + strconv.FormatInt(20, 10) + " NX" + " LT",
				"EXPIRE test_key " + strconv.FormatInt(20, 10) + " NX" + " GT",
				"GET test_key",
			},
			expected: []interface{}{"OK", int64(1),
				"ERR NX and XX, GT or LT options at the same time are not compatible",
				"ERR NX and XX, GT or LT options at the same time are not compatible",
				"test_value"},
			delay: []time.Duration{0, 0, 0, 0, 0},
		},
		{
			name:  "TEST(XX + LT/GT)",
			setup: "",
			commands: []string{
				"SET test_key test_value",
				"EXPIRE test_key " + strconv.FormatInt(20, 10),
				"EXPIRE test_key " + strconv.FormatInt(5, 10) + " XX" + " LT",
				"EXPIRE test_key " + strconv.FormatInt(10, 10) + " XX" + " GT",
				"EXPIRE test_key " + strconv.FormatInt(20, 10) + " XX" + " GT",
				"GET test_key",
			},
			expected: []interface{}{"OK", int64(1), int64(1), int64(1), int64(1), "test_value"},
			delay:    []time.Duration{0, 0, 0, 0, 0, 0},
		},
		{
			name:  "Test if value is nil after expiration",
			setup: "",
			commands: []string{
				"SET test_key test_value",
				"EXPIRE test_key " + strconv.FormatInt(20, 10),
				"EXPIRE test_key " + strconv.FormatInt(2, 10) + " XX" + " LT",
				"GET test_key",
			},
			expected: []interface{}{"OK", int64(1), int64(1), "(nil)"},
			delay:    []time.Duration{0, 0, 0, 2 * time.Second},
		},
		{
			name:  "Test if value is nil after expiration",
			setup: "",
			commands: []string{
				"SET test_key test_value",
				"EXPIRE test_key " + strconv.FormatInt(2, 10) + " NX",
				"GET test_key",
			},
			expected: []interface{}{"OK", int64(1), "(nil)"},
			delay:    []time.Duration{0, 0, 2 * time.Second},
		},
		{
			name:  "Invalid Command Test",
			setup: "",
			commands: []string{
				"SET test_key test_value",
				"EXPIRE test_key " + strconv.FormatInt(1, 10) + " XX" + " " + "rr",
				"EXPIRE test_key " + strconv.FormatInt(1, 10) + " XX" + " " + "NX",
				"EXPIRE test_key " + strconv.FormatInt(1, 10) + " GT" + " " + "lt",
				"EXPIRE test_key " + strconv.FormatInt(1, 10) + " GT" + " " + "lt" + " " + "xx",
				"EXPIRE test_key " + strconv.FormatInt(1, 10) + " GT" + " " + "lt" + " " + "nx",
				"EXPIRE test_key " + strconv.FormatInt(1, 10) + " nx" + " " + "xx" + " " + "gt",
				"EXPIRE test_key " + strconv.FormatInt(1, 10) + " nx" + " " + "xx" + " " + "lt",
			},
			expected: []interface{}{"OK", "ERR Unsupported option rr",
				"ERR NX and XX, GT or LT options at the same time are not compatible",
				"ERR GT and LT options at the same time are not compatible",
				"ERR GT and LT options at the same time are not compatible",
				"ERR NX and XX, GT or LT options at the same time are not compatible",
				"ERR NX and XX, GT or LT options at the same time are not compatible",
				"ERR NX and XX, GT or LT options at the same time are not compatible"},
			delay: []time.Duration{0, 0, 0, 0, 0, 0, 0, 0},
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

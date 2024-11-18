package http

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExpireAtHttp(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []struct {
		name          string
		setup         HTTPCommand
		commands      []HTTPCommand
		expected      []interface{}
		delay         []time.Duration
		errorExpected bool
	}{
		{
			name:  "Set with EXPIREAT command",
			setup: HTTPCommand{Command: "SET", Body: map[string]interface{}{"key": "test_key", "value": "test_value"}},
			commands: []HTTPCommand{
				{Command: "EXPIREAT", Body: map[string]interface{}{"key": "test_key", "seconds": strconv.FormatInt(time.Now().Unix()+1, 10)}},
			},
			expected:      []interface{}{float64(1)},
			delay:         []time.Duration{0},
			errorExpected: false,
		},
		{
			name:  "Check if key is nil after expiration",
			setup: HTTPCommand{Command: "SET", Body: map[string]interface{}{"key": "test_key", "value": "test_value"}},
			commands: []HTTPCommand{
				{Command: "EXPIREAT", Body: map[string]interface{}{"key": "test_key", "seconds": strconv.FormatInt(time.Now().Unix()+1, 10)}},
				{Command: "GET", Body: map[string]interface{}{"key": "test_key"}},
			},
			expected:      []interface{}{float64(1), nil},
			delay:         []time.Duration{0, 1100 * time.Millisecond},
			errorExpected: false,
		},
		{
			name:  "EXPIREAT non-existent key",
			setup: HTTPCommand{Command: "", Body: map[string]interface{}{}},
			commands: []HTTPCommand{
				{Command: "EXPIREAT", Body: map[string]interface{}{"key": "non_existent_key", "seconds": strconv.FormatInt(time.Now().Unix()+1, 10)}},
			},
			expected:      []interface{}{float64(0)},
			delay:         []time.Duration{0},
			errorExpected: false,
		},
		{
			name:  "EXPIREAT with past time",
			setup: HTTPCommand{Command: "SET", Body: map[string]interface{}{"key": "test_key", "value": "test_value"}},
			commands: []HTTPCommand{
				{Command: "EXPIREAT", Body: map[string]interface{}{"key": "test_key", "seconds": strconv.FormatInt(-1, 10)}},
				{Command: "GET", Body: map[string]interface{}{"key": "test_key"}},
			},
			expected:      []interface{}{"ERR invalid expire time in 'expireat' command", "test_value"},
			delay:         []time.Duration{0, 0},
			errorExpected: true,
		},
		{
			name:  "EXPIREAT with invalid syntax",
			setup: HTTPCommand{Command: "SET", Body: map[string]interface{}{"key": "test_key", "value": "test_value"}},
			commands: []HTTPCommand{
				{Command: "EXPIREAT", Body: map[string]interface{}{"key": "test_key"}},
			},
			expected:      []interface{}{"ERR wrong number of arguments for 'expireat' command"},
			delay:         []time.Duration{0},
			errorExpected: true,
		},
		{
			name:  "Test(NX): Set the expiration only if the key has no expiration time",
			setup: HTTPCommand{Command: "SET", Body: map[string]interface{}{"key": "test_key", "value": "test_value"}},
			commands: []HTTPCommand{
				{Command: "EXPIREAT", Body: map[string]interface{}{"key": "test_key", "seconds": strconv.FormatInt(time.Now().Unix()+10, 10), "nx": true}},
				{Command: "EXPIREAT", Body: map[string]interface{}{"key": "test_key", "seconds": strconv.FormatInt(time.Now().Unix()+1, 10), "nx": true}},
			},
			expected:      []interface{}{float64(1), float64(0)},
			delay:         []time.Duration{0, 0},
			errorExpected: false,
		},
		{
			name:  "Test(XX): Set the expiration only if the key already has an expiration time",
			setup: HTTPCommand{Command: "SET", Body: map[string]interface{}{"key": "test_key", "value": "test_value"}},
			commands: []HTTPCommand{
				{Command: "EXPIREAT", Body: map[string]interface{}{"key": "test_key", "seconds": strconv.FormatInt(time.Now().Unix()+10, 10), "xx": true}},
				{Command: "TTL", Body: map[string]interface{}{"key": "test_key"}},
				{Command: "EXPIREAT", Body: map[string]interface{}{"key": "test_key", "seconds": strconv.FormatInt(time.Now().Unix()+10, 10)}},
				{Command: "EXPIREAT", Body: map[string]interface{}{"key": "test_key", "seconds": strconv.FormatInt(time.Now().Unix()+10, 10), "xx": true}},
			},
			expected:      []interface{}{float64(0), float64(-1), float64(1), float64(1)},
			delay:         []time.Duration{0, 0, 0, 0},
			errorExpected: false,
		},
		{
			name:  "TEST(GT): Set the expiration only if the new expiration time is greater than the current one",
			setup: HTTPCommand{Command: "SET", Body: map[string]interface{}{"key": "test_key", "value": "test_value"}},
			commands: []HTTPCommand{
				{Command: "EXPIREAT", Body: map[string]interface{}{"key": "test_key", "seconds": strconv.FormatInt(time.Now().Unix()+10, 10), "gt": true}},
				{Command: "TTL", Body: map[string]interface{}{"key": "test_key"}},
				{Command: "EXPIREAT", Body: map[string]interface{}{"key": "test_key", "seconds": strconv.FormatInt(time.Now().Unix()+10, 10)}},
				{Command: "EXPIREAT", Body: map[string]interface{}{"key": "test_key", "seconds": strconv.FormatInt(time.Now().Unix()+20, 10), "gt": true}},
			},
			expected:      []interface{}{float64(0), float64(-1), float64(1), float64(1)},
			delay:         []time.Duration{0, 0, 0, 0},
			errorExpected: false,
		},
		{
			name:  "TEST(LT): Set the expiration only if the new expiration time is less than the current one",
			setup: HTTPCommand{Command: "SET", Body: map[string]interface{}{"key": "test_key", "value": "test_value"}},
			commands: []HTTPCommand{
				{Command: "EXPIREAT", Body: map[string]interface{}{"key": "test_key", "seconds": strconv.FormatInt(time.Now().Unix()+10, 10), "lt": true}},
				{Command: "EXPIREAT", Body: map[string]interface{}{"key": "test_key", "seconds": strconv.FormatInt(time.Now().Unix()+20, 10), "lt": true}},
			},
			expected:      []interface{}{float64(1), float64(0)},
			delay:         []time.Duration{0, 0},
			errorExpected: false,
		},
		{
			name:  "TEST(NX + LT/GT)",
			setup: HTTPCommand{Command: "SET", Body: map[string]interface{}{"key": "test_key", "value": "test_value"}},
			commands: []HTTPCommand{

				{Command: "EXPIREAT", Body: map[string]interface{}{"key": "test_key", "seconds": strconv.FormatInt(time.Now().Unix()+20, 10), "nx": true}},
				{Command: "EXPIREAT", Body: map[string]interface{}{"key": "test_key", "seconds": strconv.FormatInt(time.Now().Unix()+20, 10), "nx": true, "lt": true}},
				{Command: "EXPIREAT", Body: map[string]interface{}{"key": "test_key", "seconds": strconv.FormatInt(time.Now().Unix()+20, 10), "nx": true, "gt": true}},
				{Command: "GET", Body: map[string]interface{}{"key": "test_key"}},
			},
			expected: []interface{}{float64(1),
				"ERR NX and XX, GT or LT options at the same time are not compatible",
				"ERR NX and XX, GT or LT options at the same time are not compatible",
				"test_value"},
			delay:         []time.Duration{0, 0, 0, 0},
			errorExpected: true,
		},
		{
			name:  "TEST(XX + LT/GT)",
			setup: HTTPCommand{Command: "SET", Body: map[string]interface{}{"key": "test_key", "value": "test_value"}},
			commands: []HTTPCommand{

				{Command: "EXPIREAT", Body: map[string]interface{}{"key": "test_key", "seconds": strconv.FormatInt(time.Now().Unix()+20, 10)}},
				{Command: "EXPIREAT", Body: map[string]interface{}{"key": "test_key", "seconds": strconv.FormatInt(time.Now().Unix()+5, 10), "xx": true, "lt": true}},
				{Command: "EXPIREAT", Body: map[string]interface{}{"key": "test_key", "seconds": strconv.FormatInt(time.Now().Unix()+10, 10), "xx": true, "gt": true}},
				{Command: "EXPIREAT", Body: map[string]interface{}{"key": "test_key", "seconds": strconv.FormatInt(time.Now().Unix()+20, 10), "xx": true, "gt": true}},
				{Command: "GET", Body: map[string]interface{}{"key": "test_key"}},
			},
			expected:      []interface{}{float64(1), float64(1), float64(1), float64(1), "test_value"},
			delay:         []time.Duration{0, 0, 0, 0, 0, 0},
			errorExpected: false,
		},
		{
			name:  "Test if value is nil after expiration (XX + LT)",
			setup: HTTPCommand{Command: "SET", Body: map[string]interface{}{"key": "test_key", "value": "test_value"}},
			commands: []HTTPCommand{

				{Command: "EXPIREAT", Body: map[string]interface{}{"key": "test_key", "seconds": strconv.FormatInt(time.Now().Unix()+20, 10)}},
				{Command: "EXPIREAT", Body: map[string]interface{}{"key": "test_key", "seconds": strconv.FormatInt(time.Now().Unix()+2, 10), "xx": true, "lt": true}},
				{Command: "GET", Body: map[string]interface{}{"key": "test_key"}},
			},
			expected:      []interface{}{float64(1), float64(1), nil},
			delay:         []time.Duration{0, 0, 2 * time.Second},
			errorExpected: false,
		},
		{
			name:  "Test if value is nil after expiration (NX)",
			setup: HTTPCommand{Command: "SET", Body: map[string]interface{}{"key": "test_key", "value": "test_value"}},
			commands: []HTTPCommand{

				{Command: "EXPIREAT", Body: map[string]interface{}{"key": "test_key", "seconds": strconv.FormatInt(time.Now().Unix()+2, 10), "nx": true}},
				{Command: "GET", Body: map[string]interface{}{"key": "test_key"}},
			},
			expected:      []interface{}{float64(1), nil},
			delay:         []time.Duration{0, 2 * time.Second},
			errorExpected: false,
		},
		{
			name:  "Invalid Command Test",
			setup: HTTPCommand{Command: "SET", Body: map[string]interface{}{"key": "test_key", "value": "test_value"}},
			commands: []HTTPCommand{

				{Command: "EXPIREAT", Body: map[string]interface{}{"key": "test_key", "seconds": strconv.FormatInt(time.Now().Unix()+1, 10), "xx": true, "rr": true}},
				{Command: "EXPIREAT", Body: map[string]interface{}{"key": "test_key", "seconds": strconv.FormatInt(time.Now().Unix()+1, 10), "xx": true, "nx": true}},
				{Command: "EXPIREAT", Body: map[string]interface{}{"key": "test_key", "seconds": strconv.FormatInt(time.Now().Unix()+1, 10), "gt": true, "lt": true}},
				{Command: "EXPIREAT", Body: map[string]interface{}{"key": "test_key", "seconds": strconv.FormatInt(time.Now().Unix()+1, 10), "gt": true, "lt": true, "xx": true}},
				{Command: "EXPIREAT", Body: map[string]interface{}{"key": "test_key", "seconds": strconv.FormatInt(time.Now().Unix()+1, 10), "gt": true, "lt": true, "nx": true}},
				{Command: "EXPIREAT", Body: map[string]interface{}{"key": "test_key", "seconds": strconv.FormatInt(time.Now().Unix()+1, 10), "nx": true, "xx": true, "gt": true}},
				{Command: "EXPIREAT", Body: map[string]interface{}{"key": "test_key", "seconds": strconv.FormatInt(time.Now().Unix()+1, 10), "nx": true, "xx": true, "lt": true}},
			},
			expected: []interface{}{"ERR Unsupported option rr",
				"ERR NX and XX, GT or LT options at the same time are not compatible",
				"ERR GT and LT options at the same time are not compatible",
				"ERR GT and LT options at the same time are not compatible",
				"ERR NX and XX, GT or LT options at the same time are not compatible",
				"ERR NX and XX, GT or LT options at the same time are not compatible",
				"ERR NX and XX, GT or LT options at the same time are not compatible"},
			delay:         []time.Duration{0, 0, 0, 0, 0, 0, 0},
			errorExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setup.Command != "" {
				_, err := exec.FireCommand(tc.setup)
				if err != nil && !tc.errorExpected {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			var results []interface{}
			for i, cmd := range tc.commands {
				if tc.delay[i] > 0 {
					time.Sleep(tc.delay[i])
				}

				result, err := exec.FireCommand(cmd)
				if err != nil && !tc.errorExpected {
					t.Fatalf("Command failed: %v", err)
				}
				results = append(results, result)
			}

			// Validate results
			for i, expected := range tc.expected {
				if i >= len(results) {
					t.Fatalf("Not enough results. Expected %d, got %d", len(tc.expected), len(results))
				}

				if expected == nil {
					assert.True(t, results[i] == nil || results[i] == "",
						"Expected nil or empty result, got %v", results[i])
				} else {
					assert.Equal(t, expected, results[i])
				}
			}
		})
	}
}

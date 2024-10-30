package http

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExpireTimeHttp(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	futureUnixTimestamp := time.Now().Unix() + 1

	testCases := []struct {
		name          string
		setup         HTTPCommand
		commands      []HTTPCommand
		expected      []interface{}
		delay         []time.Duration
		errorExpected bool
	}{
		{
			name:  "EXPIRETIME command",
			setup: HTTPCommand{Command: "SET", Body: map[string]interface{}{"key": "test_key", "value": "test_value"}},
			commands: []HTTPCommand{
				{Command: "EXPIREAT", Body: map[string]interface{}{"key": "test_key", "seconds": strconv.FormatInt(futureUnixTimestamp, 10)}},
				{Command: "EXPIRETIME", Body: map[string]interface{}{"key": "test_key"}},
			},
			expected:      []interface{}{float64(1), float64(futureUnixTimestamp)},
			delay:         []time.Duration{0, 0},
			errorExpected: false,
		},
		{
			name:  "EXPIRETIME non-existent key",
			setup: HTTPCommand{Command: "", Body: map[string]interface{}{}},
			commands: []HTTPCommand{
				{Command: "EXPIRETIME", Body: map[string]interface{}{"key": "non_existent_key"}},
			},
			expected:      []interface{}{float64(-2)},
			delay:         []time.Duration{0},
			errorExpected: false,
		},
		{
			name:  "EXPIRETIME with past time",
			setup: HTTPCommand{Command: "SET", Body: map[string]interface{}{"key": "test_key", "value": "test_value"}},
			commands: []HTTPCommand{
				{Command: "EXPIREAT", Body: map[string]interface{}{"key": "test_key", "seconds": "1724167183"}},
				{Command: "EXPIRETIME", Body: map[string]interface{}{"key": "test_key"}},
			},
			expected:      []interface{}{float64(1), float64(-2)},
			delay:         []time.Duration{0, 0},
			errorExpected: false,
		},
		{
			name:  "EXPIRETIME with invalid syntax",
			setup: HTTPCommand{Command: "SET", Body: map[string]interface{}{"key": "test_key", "value": "test_value"}},
			commands: []HTTPCommand{
				{Command: "EXPIRETIME", Body: map[string]interface{}{"": ""}},
				{Command: "EXPIRETIME", Body: map[string]interface{}{"keys": []interface{}{"key1", "key2"}}},
			},
			expected:      []interface{}{"ERR wrong number of arguments for 'expiretime' command", "ERR wrong number of arguments for 'expiretime' command"},
			delay:         []time.Duration{0, 0},
			errorExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// setup
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

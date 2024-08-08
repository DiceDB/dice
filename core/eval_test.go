package core

import (
	"errors"
	"strconv"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

type testCase struct {
	input  []string
	output []byte
}

func TestEval(t *testing.T) {
	testCases := map[string]func(*testing.T){
		"SET": testEvalSET,
	}

	for name, testFunc := range testCases {
		t.Run(name, testFunc)
	}
}

func testEvalSET(t *testing.T) {
	tests := map[string]testCase{
		"nil value":                       {input: nil, output: []byte("-ERR wrong number of arguments for 'set' command\r\n")},
		"empty array":                     {input: []string{}, output: []byte("-ERR wrong number of arguments for 'set' command\r\n")},
		"one value":                       {input: []string{"KEY"}, output: []byte("-ERR wrong number of arguments for 'set' command\r\n")},
		"key val pair":                    {input: []string{"KEY", "VAL"}, output: RESP_OK},
		"key val pair and expiry key":     {input: []string{"KEY", "VAL", "PX"}, output: []byte("-ERR syntax error\r\n")},
		"key val pair and EX no val":      {input: []string{"KEY", "VAL", "EX"}, output: []byte("-ERR syntax error\r\n")},
		"key val pair and valid EX":       {input: []string{"KEY", "VAL", "EX", "2"}, output: RESP_OK},
		"key val pair and invalid EX":     {input: []string{"KEY", "VAL", "EX", "invalid_expiry_val"}, output: []byte("-ERR value is not an integer or out of range\r\n")},
		"key val pair and valid PX":       {input: []string{"KEY", "VAL", "PX", "2000"}, output: RESP_OK},
		"key val pair and invalid PX":     {input: []string{"KEY", "VAL", "PX", "invalid_expiry_val"}, output: []byte("-ERR value is not an integer or out of range\r\n")},
		"key val pair and both EX and PX": {input: []string{"KEY", "VAL", "EX", "2", "PX", "2000"}, output: []byte("-ERR syntax error\r\n")},
		"key val pair and PXAT no val":    {input: []string{"KEY", "VAL", "PXAT"}, output: []byte("-ERR syntax error\r\n")},
		"key val pair and invalid PXAT":   {input: []string{"KEY", "VAL", "PXAT", "invalid_expiry_val"}, output: []byte("-ERR value is not an integer or out of range\r\n")},
		"key val pair and expired PXAT":   {input: []string{"KEY", "VAL", "PXAT", "2"}, output: RESP_OK},
		"key val pair and negative PXAT":  {input: []string{"KEY", "VAL", "PXAT", "-123456"}, output: []byte("-ERR invalid expire time in 'set' command\r\n")},
		"key val pair and valid PXAT":     {input: []string{"KEY", "VAL", "PXAT", strconv.FormatInt(time.Now().Add(2*time.Minute).UnixMilli(), 10)}, output: RESP_OK},
	}

	runTests(t, tests, evalSET)
}

func runTests(t *testing.T, tests map[string]testCase, evalFunc func([]string) []byte) {
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			output := evalFunc(tc.input)
			assert.Equal(t, string(tc.output), string(output))
		})
	}
}

// TestEvalPersist tests the evalPersist function using table-driven tests.
func TestEvalPersist(t *testing.T) {
	// Define test cases
	tests := []struct {
		name     string
		args     []string
		setup    func()
		expected []byte
	}{
		{
			name:     "wrong number of arguments",
			args:     []string{"key1", "key2"},
			expected: Encode(errors.New("ERR wrong number of arguments for 'persist' command"), false),
		},
		{
			name:     "key does not exist",
			args:     []string{"nonexistent"},
			expected: RESP_ZERO,
		},
		{
			name: "key exists but no expiration set",
			args: []string{"existent_no_expiry"},
			setup: func() {
				evalSET([]string{"existent_no_expiry", "value"})
			},
			expected: RESP_MINUS_1,
		},
		{
			name: "key exists and expiration removed",
			args: []string{"existent_with_expiry"},
			setup: func() {
				evalSET([]string{"existent_with_expiry", "value", "EX", "1"})
			},
			expected: RESP_ONE,
		},
		{
			name: "key exists with expiration set and not expired",
			args: []string{"existent_with_expiry_not_expired"},
			setup: func() {
				// Simulate setting a key with an expiration time that has not yet passed
				evalSET([]string{"existent_with_expiry_not_expired", "value", "EX", "10000"}) // 10000 seconds in the future
			},
			expected: RESP_ONE,
		},
	}

	// Run test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup test case
			if tc.setup != nil {
				tc.setup()
			}

			// Call function under test
			got := evalPersist(tc.args)

			// Assert expected result
			if string(got) != string(tc.expected) {
				t.Errorf("evalPersist(%v) = %v, want %v", tc.args, string(got), string(tc.expected))
			}
		})
	}
}

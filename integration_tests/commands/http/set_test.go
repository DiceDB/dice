package http

import (
	"strconv"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

type TestCase struct {
	name          string
	commands      []HTTPCommand
	expected      []interface{}
	errorExpected bool
}

func TestSet(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []TestCase{
		{
			name: "Set and Get Simple Value",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "k", "value": "v"}},
				{Command: "GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", "v"},
		},
		{
			name: "Set and Get Integer Value",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "k", "value": 123456789}},
				{Command: "GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", "1.23456789e+08"},
		},
		{
			name: "Overwrite Existing Key",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "k", "value": "v1"}},
				{Command: "SET", Body: map[string]interface{}{"key": "k", "value": 5}},
				{Command: "GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", "OK", float64(5)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			exec.FireCommand(HTTPCommand{
				Command: "DEL",
				Body:    map[string]interface{}{"key": "k"},
			})

			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)
				assert.DeepEqual(t, tc.expected[i], result)
			}
		})
	}
}

func TestSetWithOptions(t *testing.T) {
	exec := NewHTTPCommandExecutor()
	expiryTime := strconv.FormatInt(time.Now().Add(1*time.Minute).UnixMilli(), 10)

	testCases := []TestCase{
		{
			name: "Set with EX option",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "k", "value": "v", "ex": 3}},
				{Command: "GET", Body: map[string]interface{}{"key": "k"}},
				{Command: "SLEEP", Body: map[string]interface{}{"key": 3}},
				{Command: "GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", "v", "OK", "(nil)"},
		},
		{
			name: "Set with PX option",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "k", "value": "v", "px": 2000}},
				{Command: "GET", Body: map[string]interface{}{"key": "k"}},
				{Command: "SLEEP", Body: map[string]interface{}{"key": 3}},
				{Command: "GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", "v", "OK", "(nil)"},
		},
		{
			name: "Set with EX and PX option",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "k", "value": "v", "ex": 2, "px": 2000}},
			},
			expected: []interface{}{"ERR syntax error"},
		},
		{
			name: "XX on non-existing key",
			commands: []HTTPCommand{
				{Command: "DEL", Body: map[string]interface{}{"key": "k"}},
				{Command: "SET", Body: map[string]interface{}{"key": "k", "value": "v", "xx": true}},
				{Command: "GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{float64(0), "(nil)", "(nil)"},
		},
		{
			name: "NX on non-existing key",
			commands: []HTTPCommand{
				{Command: "DEL", Body: map[string]interface{}{"key": "k"}},
				{Command: "SET", Body: map[string]interface{}{"key": "k", "value": "v", "nx": true}},
				{Command: "GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{float64(0), "OK", "v"},
		},
		{
			name: "NX on existing key",
			commands: []HTTPCommand{
				{Command: "DEL", Body: map[string]interface{}{"key": "k"}},
				{Command: "SET", Body: map[string]interface{}{"key": "k", "value": "v", "nx": true}},
				{Command: "GET", Body: map[string]interface{}{"key": "k"}},
				{Command: "SET", Body: map[string]interface{}{"key": "k", "value": "v", "nx": true}},
			},
			expected: []interface{}{float64(0), "OK", "v", "(nil)"},
		},
		{
			name: "PXAT option",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "k", "value": "v", "pxat": expiryTime}},
				{Command: "GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", "v"},
		},
		{
			name: "PXAT option with delete",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "k1", "value": "v1", "pxat": expiryTime}},
				{Command: "GET", Body: map[string]interface{}{"key": "k1"}},
				{Command: "SLEEP", Body: map[string]interface{}{"key": 4}},
				{Command: "DEL", Body: map[string]interface{}{"key": "k1"}},
			},
			expected: []interface{}{"OK", "v1", "OK", float64(1)},
		},
		{
			name: "PXAT option with invalid unix time ms",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "k2", "value": "v2", "pxat": "123123"}},
				{Command: "GET", Body: map[string]interface{}{"key": "k2"}},
			},
			expected: []interface{}{"OK", "(nil)"},
		},
		{
			name: "XX on existing key",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "k", "value": "v1"}},
				{Command: "SET", Body: map[string]interface{}{"key": "k", "value": "v2", "xx": true}},
				{Command: "GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", "OK", "v2"},
		},
		{
			name: "Multiple XX operations",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "k", "value": "v1"}},
				{Command: "SET", Body: map[string]interface{}{"key": "k", "value": "v2", "xx": true}},
				{Command: "SET", Body: map[string]interface{}{"key": "k", "value": "v3", "xx": true}},
				{Command: "GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", "OK", "OK", "v3"},
		},
		{
			name: "EX option",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "k", "value": "v", "ex": 1}},
				{Command: "GET", Body: map[string]interface{}{"key": "k"}},
				{Command: "SLEEP", Body: map[string]interface{}{"key": 2}},
				{Command: "GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", "v", "OK", "(nil)"},
		},
		{
			name: "XX option",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "k", "value": "v", "xx": true, "ex": 1}},
				{Command: "GET", Body: map[string]interface{}{"key": "k"}},
				{Command: "SLEEP", Body: map[string]interface{}{"key": 2}},
				{Command: "GET", Body: map[string]interface{}{"key": "k"}},
				{Command: "SET", Body: map[string]interface{}{"key": "k", "value": "v", "xx": true, "ex": 1}},
				{Command: "GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"(nil)", "(nil)", "OK", "(nil)", "(nil)", "(nil)"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"key": "k"}})
			exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"key": "k1"}})
			exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"key": "k2"}})
			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}
}

func TestSetWithExat(t *testing.T) {
	exec := NewHTTPCommandExecutor()
	Etime := strconv.FormatInt(time.Now().Unix()+5, 10)
	BadTime := "123123"

	testCases := []TestCase{
		{
			name: "SET with EXAT",
			commands: []HTTPCommand{
				{Command: "DEL", Body: map[string]interface{}{"key": "k"}},
				{Command: "SET", Body: map[string]interface{}{"key": "k", "value": "v", "exat": Etime}},
				{Command: "GET", Body: map[string]interface{}{"key": "k"}},
				{Command: "TTL", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{float64(0), "OK", "v", float64(4)},
		},
		{
			name: "SET with invalid EXAT expires key immediately",
			commands: []HTTPCommand{
				{Command: "DEL", Body: map[string]interface{}{"key": "k"}},
				{Command: "SET", Body: map[string]interface{}{"key": "k", "value": "v", "exat": BadTime}},
				{Command: "GET", Body: map[string]interface{}{"key": "k"}},
				{Command: "TTL", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{float64(0), "OK", "(nil)", float64(-2)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Ensure key is deleted before the test
			exec.FireCommand(HTTPCommand{
				Command: "DEL",
				Body:    map[string]interface{}{"key": "k"},
			})

			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)
				if cmd.Command == "TTL" {
					assert.Assert(t, result.(float64) <= tc.expected[i].(float64))
				} else {
					assert.DeepEqual(t, tc.expected[i], result)
				}
			}
		})
	}
}

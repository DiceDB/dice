package tests

import (
	"testing"

	"gotest.tools/v3/assert"
)

var getInfoCommandTestCases = []struct {
	name     string
	inCmd    string
	expected interface{}
}{
	{
		name:  "Single command - GET",
		inCmd: "GET",
		expected: []interface{}{
			[]interface{}{
				"GET",
				int64(2),
				[]interface{}{"readonly", "fast"},
				int64(1),
				int64(1),
				int64(1),
				[]interface{}{"@read", "@string", "@fast"},
				[]interface{}{},
				[]interface{}{}, // Additional command metadata can go here
			},
		},
	},
	{
		name:  "Single command - SET",
		inCmd: "SET",
		expected: []interface{}{
			[]interface{}{
				"SET",
				int64(-3),
				[]interface{}{"write", "denyoom"},
				int64(1),
				int64(1),
				int64(1),
				[]interface{}{"@write", "@string", "@slow"},
				[]interface{}{},
				[]interface{}{}, // Additional command metadata can go here
			},
		},
	},
	{
		name:  "Unknown command",
		inCmd: "FOO",
		expected: []interface{}{
			nil,
		},
	},
	{
		name:  "Multiple commands - GET and SET",
		inCmd: "GET SET",
		expected: []interface{}{
			[]interface{}{
				"GET",
				int64(2),
				[]interface{}{"readonly", "fast"},
				int64(1),
				int64(1),
				int64(1),
				[]interface{}{"@read", "@string", "@fast"},
				[]interface{}{},
				[]interface{}{}, // Additional command metadata can go here
			},
			[]interface{}{
				"SET",
				int64(-3),
				[]interface{}{"write", "denyoom"},
				int64(1),
				int64(1),
				int64(1),
				[]interface{}{"@write", "@string", "@slow"},
				[]interface{}{},
				[]interface{}{}, // Additional command metadata can go here
			},
		},
	},
	{
		name:  "Mixed valid and invalid commands",
		inCmd: "GET FOO SET",
		expected: []interface{}{
			[]interface{}{
				"GET",
				int64(2),
				[]interface{}{"readonly", "fast"},
				int64(1),
				int64(1),
				int64(1),
				[]interface{}{"@read", "@string", "@fast"},
				[]interface{}{},
				[]interface{}{}, // Additional command metadata can go here
			},
			nil,
			[]interface{}{
				"SET",
				int64(-3),
				[]interface{}{"write", "denyoom"},
				int64(1),
				int64(1),
				int64(1),
				[]interface{}{"@write", "@string", "@slow"},
				[]interface{}{},
				[]interface{}{}, // Additional command metadata can go here
			},
		},
	},
}

func TestInfoCommand(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	for _, tc := range getInfoCommandTestCases {
		t.Run(tc.name, func(t *testing.T) {
			result := fireCommand(conn, "COMMAND INFO "+tc.inCmd)
			assert.DeepEqual(t, tc.expected, result)
		})
	}
}

func BenchmarkInfoCommand(b *testing.B) {
	conn := getLocalConnection()
	defer conn.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tc := range getInfoCommandTestCases {
			fireCommand(conn, "COMMAND INFO "+tc.inCmd)
		}
	}
}

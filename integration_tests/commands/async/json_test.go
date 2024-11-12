package async

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func arraysArePermutations[T comparable](a, b []T) bool {
	// If lengths are different, they cannot be permutations
	if len(a) != len(b) {
		return false
	}

	// Count occurrences of each element in array 'a'
	countA := make(map[T]int)
	for _, elem := range a {
		countA[elem]++
	}

	// Subtract occurrences based on array 'b'
	for _, elem := range b {
		countA[elem]--
		if countA[elem] < 0 {
			return false
		}
	}

	// Check if all counts are zero
	for _, count := range countA {
		if count != 0 {
			return false
		}
	}

	return true
}

func TestJsonSTRAPPEND(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	simpleJSON := `{"name":"John","age":30}`

	testCases := []struct {
		name     string
		setCmd   string
		getCmd   string
		expected interface{}
	}{
		{
			name:     "STRAPPEND to nested string",
			setCmd:   `JSON.SET doc $ ` + simpleJSON,
			getCmd:   `JSON.STRAPPEND doc $.name " Doe"`,
			expected: []interface{}{int64(8)},
		},
		{
			name:     "STRAPPEND to multiple paths",
			setCmd:   `JSON.SET doc $ ` + simpleJSON,
			getCmd:   `JSON.STRAPPEND doc $..name "baz"`,
			expected: []interface{}{int64(7)},
		},
		{
			name:     "STRAPPEND to non-string",
			setCmd:   `JSON.SET doc $ ` + simpleJSON,
			getCmd:   `JSON.STRAPPEND doc $.age " years"`,
			expected: []interface{}{"(nil)"},
		},
		{
			name:     "STRAPPEND with empty string",
			setCmd:   `JSON.SET doc $ ` + simpleJSON,
			getCmd:   `JSON.STRAPPEND doc $.name ""`,
			expected: []interface{}{int64(4)},
		},
		{
			name:     "STRAPPEND to non-existent path",
			setCmd:   `JSON.SET doc $ ` + simpleJSON,
			getCmd:   `JSON.STRAPPEND doc $.nonexistent " test"`,
			expected: []interface{}{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Use FLUSHDB to clear all keys before each test
			result := FireCommand(conn, "FLUSHDB")
			assert.Equal(t, "OK", result)

			result = FireCommand(conn, tc.setCmd)
			assert.Equal(t, "OK", result)

			result = FireCommand(conn, tc.getCmd)
			assert.ElementsMatch(t, tc.expected, result)

		})
	}
}

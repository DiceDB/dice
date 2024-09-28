package async

import (
	testifyAssert "github.com/stretchr/testify/assert"
	"testing"

	"gotest.tools/v3/assert"
)

func TestHvals(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()
	defer FireCommand(conn, "DEL hvalsKey hvalsKey01 hvalsKey02")

	testCases := []TestCase{
		{
			name:     "HVALS with multiple fields",
			commands: []string{"HSET hvalsKey field value", "HSET hvalsKey field2 value_new", "HVALS hvalsKey"},
			expected: []interface{}{ONE, ONE, []string{"value", "value_new"}},
		},
		{
			name:     "HVALS with non-existing key",
			commands: []string{"HVALS hvalsKey01"},
			expected: []interface{}{[]string{}},
		},
		{
			name:     "HVALS on wrong key type",
			commands: []string{"SET hvalsKey02 field", "HVALS hvalsKey02"},
			expected: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
		{
			name:     "HVALS with wrong number of arguments",
			commands: []string{"HVALS hvalsKey03 x", "HVALS"},
			expected: []interface{}{"ERR wrong number of arguments for 'hvals' command", "ERR wrong number of arguments for 'hvals' command"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result := FireCommand(conn, cmd)

				// Type check for expected value and result
				expectedList, isExpectedList := tc.expected[i].([]string)
				resultList, isResultList := result.([]interface{})

				// If both are lists, compare them unordered
				if isExpectedList && isResultList && len(resultList) == len(expectedList) {
					testifyAssert.ElementsMatch(t, expectedList, convertToStringSlice(resultList))
				} else {
					// Otherwise, do a deep comparison
					assert.DeepEqual(t, tc.expected[i], result)
				}
			}
		})
	}
}

// Helper function to convert []interface{} to []string for easier comparison
func convertToStringSlice(input []interface{}) []string {
	output := make([]string, len(input))
	for i, v := range input {
		output[i] = v.(string)
	}
	return output
}

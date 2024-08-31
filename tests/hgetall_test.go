package tests

import (
	"gotest.tools/v3/assert"
	"reflect"
	"testing"
)

func TestHGETALL(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	testCases := []TestCase{
		{
			commands: []string{"HSET key field value", "HSET key field2 value_new", "HGETALL key"},
			expected: []interface{}{ONE, ONE, []string{"field", "value", "field2", "value_new"}},
		},
		{
			commands: []string{"HGETALL key3"},
			expected: []interface{}{[]interface{}{}},
		},
		{
			commands: []string{"SET key field", "HGETALL key"},
			expected: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
		{
			commands: []string{"HGETALL key x", "HGETALL"},
			expected: []interface{}{"ERR wrong number of arguments for 'hgetall' command",
				"ERR wrong number of arguments for 'hgetall' command"},
		},
	}

	for _, tc := range testCases {
		for i, cmd := range tc.commands {
			result := fireCommand(conn, cmd)
			expectedResults, ok := tc.expected[i].([]string)
			results, ok2 := result.([]interface{})

			if ok && ok2 && len(results) == len(expectedResults) {
				expectedResultsMap := make(map[string]string)
				resultsMap := make(map[string]string)

				for i := 0; i < len(expectedResultsMap); i += 2 {
					expectedResultsMap[expectedResults[i]] = expectedResults[i+1]
					resultsMap[results[i].(string)] = results[i+1].(string)
				}
				reflect.DeepEqual(resultsMap, expectedResultsMap)

			} else {
				assert.DeepEqual(t, tc.expected[i], result)
			}
		}
	}
}

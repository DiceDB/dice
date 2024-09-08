package tests

import (
	"reflect"
	"testing"

	"gotest.tools/v3/assert"
)

func TestHGET(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()
	defer fireCommand(conn, "DEL key_hGet key_hGet02")

	testCases := []TestCase{
		{
			commands: []string{"HGET", "HGET KEY", "HGET KEY FIELD ANOTHER_FIELD"},
			expected: []interface{}{"ERR wrong number of arguments for 'hget' command",
				"ERR wrong number of arguments for 'hget' command",
				"ERR wrong number of arguments for 'hget' command"},
		},
		{
			commands: []string{"HSET key_hGet field value", "HSET key_hGet field newvalue"},
			expected: []interface{}{ONE, ZERO},
		},
		{
			commands: []string{"HGET wrong_key_hGet field"},
			expected: []interface{}{"(nil)"},
		},
		{
			commands: []string{"HGET key_hGet wrong_field"},
			expected: []interface{}{"(nil)"},
		},
		{
			commands: []string{"HGET key_hGet field"},
			expected: []interface{}{"newvalue"},
		},
		{
			commands: []string{"SET key value", "HGET key field"},
			expected: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result := fireCommand(conn, cmd)
				expectedResults, ok := tc.expected[i].([]string)
				results, ok2 := result.([]interface{})

				if ok && ok2 && len(results) == len(expectedResults) {
					expectedResultsMap := make(map[string]string)
					resultsMap := make(map[string]string)

					for i := 0; i < len(results); i += 2 {
						expectedResultsMap[expectedResults[i]] = expectedResults[i+1]
						resultsMap[results[i].(string)] = results[i+1].(string)
					}
					if !reflect.DeepEqual(resultsMap, expectedResultsMap) {
						t.Fatalf("Assertion failed: expected true, got false")
					}

				} else {
					assert.DeepEqual(t, tc.expected[i], result)
				}
			}
		})
	}
}

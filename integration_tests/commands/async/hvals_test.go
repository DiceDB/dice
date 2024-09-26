package async

import (
	"reflect"
	"testing"

	"gotest.tools/v3/assert"
)

func TestHvals(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()
	defer FireCommand(conn, "DEL key_hVals key_hVals02")

	testCases := []TestCase{
		{
			commands: []string{"HSET key_hVals field value", "HSET key_hVals field2 value_new", "HVALS key_hVals"},
			expected: []interface{}{ONE, ONE, []string{"value", "value_new"}},
		},
		{
			commands: []string{"HVALS key_hVals01"},
			expected: []interface{}{[]interface{}{}},
		},
		{
			commands: []string{"SET key_hVals02 field", "HVALS key_hVals02"},
			expected: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
		{
			commands: []string{"HVALS key_hVals03 x", "HVALS"},
			expected: []interface{}{"ERR wrong number of arguments for 'hvals' command",
				"ERR wrong number of arguments for 'hvals' command"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result := FireCommand(conn, cmd)
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

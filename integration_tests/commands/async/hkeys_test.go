package async

import (
	"reflect"
	"testing"

	"gotest.tools/v3/assert"
)

func TestHKEYS(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()
	defer FireCommand(conn, "DEL key_hkeys key_hkeys02")

	testCases := []TestCase{
		{
			commands: []string{"HSET key_hkeys field1 value1", "HSET key_hkeys field2 value2", "HKEYS key_hkeys"},
			expected: []interface{}{ONE, ONE, []string{"field1", "field2",}},
		},
		{
			commands: []string{"HKEYS key_hkeys01"},
			expected: []interface{}{[]interface{}{}},
		},
		{
			commands: []string{"SET key_hkeys02 field1", "HKEYS key_hkeys02"},
			expected: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
		{
			commands: []string{"HKEYS key_hkeys03 x", "HKEYS"},
			expected: []interface{}{"ERR wrong number of arguments for 'hkeys' command",
				"ERR wrong number of arguments for 'hkeys' command"},
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

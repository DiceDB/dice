package async

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

var ZERO int64 = 0
var ONE int64 = 1
var TWO int64 = 2

func TestHGETALL(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()
	defer FireCommand(conn, "DEL key_hGetAll key_hGetAll02")

	testCases := []TestCase{
		{
			commands: []string{"HSET key_hGetAll field value", "HSET key_hGetAll field2 value_new", "HGETALL key_hGetAll"},
			expected: []interface{}{ONE, ONE, []string{"field", "value", "field2", "value_new"}},
		},
		{
			commands: []string{"HGETALL key_hGetAll01"},
			expected: []interface{}{[]interface{}{}},
		},
		{
			commands: []string{"SET key_hGetAll02 field", "HGETALL key_hGetAll02"},
			expected: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
		{
			commands: []string{"HGETALL key_hGetAll03 x", "HGETALL"},
			expected: []interface{}{"ERR wrong number of arguments for 'hgetall' command",
				"ERR wrong number of arguments for 'hgetall' command"},
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
					assert.Equal(t, tc.expected[i], result)
				}
			}
		})
	}
}

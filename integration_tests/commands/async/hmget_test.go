package async

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"gotest.tools/v3/assert"
)

func TestHMGET(t *testing.T) {
	conn := getLocalConnection()
	var smallHashKeys []string
	var smallHashValues []string

	var bigHashKeys []string
	var bigHashValues []string

	// Setup smallHash
	for i := 0; i < 8; i++ {
		key := fmt.Sprintf("field%d", i)
		value := fmt.Sprintf("value%d", i)
		smallHashKeys = append(smallHashKeys, key)
		smallHashValues = append(smallHashValues, value)
		FireCommand(conn, "HSET smallHash "+key+" "+value)
	}

	// Setup bigHash
	for i := 0; i < 1024; i++ {
		key := fmt.Sprintf("field%d", i)
		value := fmt.Sprintf("value%d", i)
		bigHashKeys = append(bigHashKeys, key)
		bigHashValues = append(bigHashValues, value)
		FireCommand(conn, "HSET bigHash "+key+" "+value)
	}

	defer conn.Close()
	defer FireCommand(conn, "DEL key_hmGet key_hmGet1 smallHash bigHash")

	testCases := []TestCase{
		{
			commands: []string{"HSET key_hmGet field value", "HSET key_hmGet field2 value_new", "HMGET key_hmGet field field2"},
			expected: []interface{}{ONE, ONE, []string{"value", "value_new"}},
		},
		{
			commands: []string{"HMGET doesntexist field"},
			expected: []interface{}{[]interface{}{"(nil)"}},
		},
		{
			commands: []string{"HMGET smallHash field"},
			expected: []interface{}{[]interface{}{"(nil)"}},
		},
		{
			commands: []string{"HMGET bigHash field"},
			expected: []interface{}{[]interface{}{"(nil)"}},
		},
		{
			commands: []string{"HMGET bigHash " + strings.Join(bigHashKeys, " ")},
			expected: []interface{}{bigHashValues},
		},
		{
			commands: []string{"HMGET smallHash " + strings.Join(smallHashKeys, " ")},
			expected: []interface{}{smallHashValues},
		},
		{
			commands: []string{"SET key_hmGet1 field", "HMGET key_hmGet1 field"},
			expected: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
		{
			commands: []string{"HMGET key_hmGet", "HMGET"},
			expected: []interface{}{"ERR wrong number of arguments for 'hmget' command",
				"ERR wrong number of arguments for 'hmget' command"},
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

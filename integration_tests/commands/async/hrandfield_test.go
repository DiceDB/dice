package async

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestHRANDFIELD(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()
	defer FireCommand(conn, "DEL key_hrandfield key_hrandfield02 key_hrandfield03")

	testCases := []TestCase{
		{
			commands: []string{"HSET key_hrandfield field value", "HSET key_hrandfield field2 value2", "HRANDFIELD key_hrandfield"},
			expected: []interface{}{ONE, ONE, []string{"field", "field2"}},
		},
		{
			commands: []string{"HSET key_hrandfield field3 value3", "HRANDFIELD key_hrandfield 2"},
			expected: []interface{}{ONE, []string{"field", "field2", "field3"}},
		},
		{
			commands: []string{"HRANDFIELD key_hrandfield 2 WITHVALUES"},
			expected: []interface{}{[]string{"field", "value", "field2", "value2", "field3", "value3"}},
		},
		{
			commands: []string{"HRANDFIELD key_hrandfield_nonexistent"},
			expected: []interface{}{"(nil)"},
		},
		{
			commands: []string{"HRANDFIELD"},
			expected: []interface{}{"ERR wrong number of arguments for 'hrandfield' command"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result := FireCommand(conn, cmd)
				expectedResults, ok := tc.expected[i].([]string)
				results, ok2 := result.([]interface{})

				if ok && ok2 {
					resultsList := make([]string, len(results))
					for i, r := range results {
						resultsList[i] = r.(string)
					}

					expectedMap := make(map[string]struct{})
					for _, expected := range expectedResults {
						expectedMap[expected] = struct{}{}
					}
					count := 0
					for _, result := range resultsList {
						if _, exists := expectedMap[result]; exists {

							count++
						}
					}
					assert.Assert(t, count == 2 || count == 4, "Expected count to be 2 or 4, but got %v", count)
				} else if ok {
					count := 0
					for _, r := range expectedResults {
						if result == r {
							count++
						}
					}
					assert.Equal(t, count, 1)
				} else {
					assert.DeepEqual(t, tc.expected[i], result)
				}
			}
		})
	}
}

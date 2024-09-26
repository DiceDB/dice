package async

import (
	"github.com/google/go-cmp/cmp/cmpopts"
	"gotest.tools/v3/assert"
	"testing"
)

func TestHRANDFIELD(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()
	defer FireCommand(conn, "DEL key_hrandfield key_hrandfield02 key_hrandfield03")

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name:     "Basic HRANDFIELD operations",
			commands: []string{"HSET key_hrandfield field value", "HSET key_hrandfield field2 value2", "HRANDFIELD key_hrandfield"},
			expected: []interface{}{ONE, ONE, []string{"field", "field2"}},
		},
		{
			name:     "HRANDFIELD with count",
			commands: []string{"HSET key_hrandfield field3 value3", "HRANDFIELD key_hrandfield 2"},
			expected: []interface{}{ONE, []string{"field", "field2", "field3"}},
		},
		{
			name:     "HRANDFIELD with WITHVALUES",
			commands: []string{"HRANDFIELD key_hrandfield 2 WITHVALUES"},
			expected: []interface{}{[]string{"field", "value", "field2", "value2", "field3", "value3"}},
		},
		{
			name:     "HRANDFIELD on non-existent key",
			commands: []string{"HRANDFIELD key_hrandfield_nonexistent"},
			expected: []interface{}{"(nil)"},
		},
		{
			name:     "HRANDFIELD with wrong number of arguments",
			commands: []string{"HRANDFIELD"},
			expected: []interface{}{"ERR wrong number of arguments for 'hrandfield' command"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result := FireCommand(conn, cmd)
				expected := tc.expected[i]

				switch expected := expected.(type) {
				case []string:
					assertRandomFieldResult(t, result, expected)
				case int:
					assert.Equal(t, result, expected, "Unexpected result for command: %s", cmd)
				case string:
					assert.Equal(t, result, expected, "Unexpected result for command: %s", cmd)
				default:
					if str, ok := result.(string); ok {
						assert.Equal(t, str, expected, "Unexpected result for command: %s", cmd)
					} else {
						assert.DeepEqual(t, result, expected, cmpopts.EquateEmpty())
					}
				}
			}
		})
	}
}

// assertRandomFieldResult asserts that the result contains all expected values or a single valid result
func assertRandomFieldResult(t *testing.T, result interface{}, expected []string) {
	t.Helper()

	var resultsList []string
	switch r := result.(type) {
	case []interface{}:
		resultsList = make([]string, len(r))
		for i, v := range r {
			resultsList[i] = v.(string)
		}
	case string:
		resultsList = []string{r}
	default:
		t.Fatalf("Expected result to be []interface{} or string, got %T", result)
	}

	// generate a map of expected values for easy lookup
	expectedMap := make(map[string]struct{})
	for _, exp := range expected {
		expectedMap[exp] = struct{}{}
	}

	// count the number of results that are in the expected set
	count := 0
	for _, res := range resultsList {
		if _, exists := expectedMap[res]; exists {
			count++
		}
	}

	// assert that all results are in the expected set or that there is a single valid result
	assert.Assert(t, count == len(resultsList) || count == 1,
		"Expected all results to be in the expected set or a single valid result. Got %d out of %d",
		count, len(resultsList))
}

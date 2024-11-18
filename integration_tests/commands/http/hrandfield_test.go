package http

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
)

func TestHRANDFIELD(t *testing.T) {
	exec := NewHTTPCommandExecutor()
	exec.FireCommand(HTTPCommand{
		Command: "DEL",
		Body: map[string]interface{}{
			"keys": []interface{}{"key", "hrandfield"},
		},
	})
	testCases := []struct {
		name          string
		commands      []HTTPCommand
		expected      []interface{}
		delay         []time.Duration
		errorExpected bool
	}{
		{
			name: "Basic HRANDFIELD operations",
			commands: []HTTPCommand{
				{Command: "HSET", Body: map[string]interface{}{"key": "key_hrandfield", "field": "field1", "value": "value1"}},
				{Command: "HSET", Body: map[string]interface{}{"key": "key_hrandfield", "field": "field2", "value": "value2"}},
				{Command: "HRANDFIELD", Body: map[string]interface{}{"key": "key_hrandfield"}},
			},
			expected:      []interface{}{float64(1), float64(1), []string{"field1", "field2"}},
			delay:         []time.Duration{0, 0, 0},
			errorExpected: false,
		},
		{
			name: "HRANDFIELD with count",
			commands: []HTTPCommand{
				{Command: "HSET", Body: map[string]interface{}{"key": "key_hrandfield", "field": "field3", "value": "value3"}},
				{Command: "HRANDFIELD", Body: map[string]interface{}{"key": "key_hrandfield", "value": "2"}},
			},
			expected:      []interface{}{float64(1), []string{"field1", "field2", "field3"}},
			delay:         []time.Duration{0, 0},
			errorExpected: false,
		},
		{
			name: "HRANDFIELD with WITHVALUES",
			commands: []HTTPCommand{
				{Command: "HRANDFIELD", Body: map[string]interface{}{"values": []interface{}{"key_hrandfield", "2", "WITHVALUES"}}},
			},
			expected:      []interface{}{[]string{"field1", "value1", "field2", "value2", "field3", "value3"}},
			delay:         []time.Duration{0},
			errorExpected: false,
		},
		{
			name: "HRANDFIELD on non-existent key",
			commands: []HTTPCommand{
				{Command: "HRANDFIELD", Body: map[string]interface{}{"key": "key_hrandfield_nonexistent"}},
			},
			expected:      []interface{}{nil},
			delay:         []time.Duration{0},
			errorExpected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result, err := exec.FireCommand(cmd)
				if err != nil {
					// Check if the error message matches the expected result
					assert.Equal(t, tc.expected[i], err.Error(), "Error message mismatch for cmd %s", cmd)
				}

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
						assert.Equal(t, result, expected, cmpopts.EquateEmpty())
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
	assert.True(t, count == len(resultsList) || count == 1,
		"Expected all results to be in the expected set or a single valid result. Got %d out of %d",
		count, len(resultsList))
}

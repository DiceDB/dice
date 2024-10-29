package websocket

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
)

func TestHRANDFIELD(t *testing.T) {
	exec := NewWebsocketCommandExecutor()
	conn := exec.ConnectToServer()
	DeleteKey(t, conn, exec, "key_hrandfield")

	testCases := []struct {
		name   string
		cmds   []string
		expect []interface{}
		delays []time.Duration
	}{

		{
			name:   "Basic HRANDFIELD operations",
			cmds:   []string{"HSET key_hrandfield field value", "HSET key_hrandfield field2 value2", "HRANDFIELD key_hrandfield"},
			expect: []interface{}{float64(1), float64(1), []string{"field", "field2"}},
			delays: []time.Duration{0, 0, 0},
		},
		{
			name:   "HRANDFIELD with count",
			cmds:   []string{"HSET key_hrandfield field3 value3", "HRANDFIELD key_hrandfield 2"},
			expect: []interface{}{float64(1), []string{"field", "field2", "field3"}},
			delays: []time.Duration{0, 0},
		},
		{
			name:   "HRANDFIELD with WITHVALUES",
			cmds:   []string{"HRANDFIELD key_hrandfield 2 WITHVALUES"},
			expect: []interface{}{[]string{"field", "value", "field2", "value2", "field3", "value3"}},
			delays: []time.Duration{0},
		},
		{
			name:   "HRANDFIELD on non-existent key",
			cmds:   []string{"HRANDFIELD key_hrandfield_nonexistent"},
			expect: []interface{}{nil},
			delays: []time.Duration{0},
		},
		{
			name:   "HRANDFIELD with wrong number of arguments",
			cmds:   []string{"HRANDFIELD"},
			expect: []interface{}{"ERR wrong number of arguments for 'hrandfield' command"},
			delays: []time.Duration{0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.cmds {
				result, err := exec.FireCommandAndReadResponse(conn, cmd)
				assert.Nil(t, err)
				expected := tc.expect[i]

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

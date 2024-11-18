package async

import (
	"encoding/json"

	"testing"

	"github.com/dicedb/dice/testutils"
	"github.com/stretchr/testify/assert"
)

func compareJSON(t *testing.T, expected, actual string) {
	var expectedMap map[string]interface{}
	var actualMap map[string]interface{}

	err1 := json.Unmarshal([]byte(expected), &expectedMap)
	err2 := json.Unmarshal([]byte(actual), &actualMap)

	assert.Nil(t, err1)
	assert.Nil(t, err2)

	assert.Equal(t, expectedMap, actualMap)
}

func TestJSONToggle(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	simpleJSON := `{"name":true,"age":false}`
	complexJson := `{"field":true,"nested":{"field":false,"nested":{"field":true}}}`

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name:     "JSON.TOGGLE with existing key",
			commands: []string{`JSON.SET user $ ` + simpleJSON, "JSON.TOGGLE user $.name"},
			expected: []interface{}{"OK", []any{int64(0)}},
		},
		{
			name:     "JSON.TOGGLE with non-existing key",
			commands: []string{"JSON.TOGGLE user $.flag"},
			expected: []interface{}{"ERR could not perform this operation on a key that doesn't exist"},
		},
		{
			name:     "JSON.TOGGLE with invalid path",
			commands: []string{"JSON.TOGGLE user $.invalidPath"},
			expected: []interface{}{"ERR could not perform this operation on a key that doesn't exist"},
		},
		{
			name:     "JSON.TOGGLE with invalid command format",
			commands: []string{"JSON.TOGGLE testKey"},
			expected: []interface{}{"ERR wrong number of arguments for 'json.toggle' command"},
		},
		{
			name: "deeply nested JSON structure with multiple matching fields",
			commands: []string{
				`JSON.SET user $ ` + complexJson,
				"JSON.GET user",
				"JSON.TOGGLE user $..field",
				"JSON.GET user",
			},
			expected: []interface{}{
				"OK",
				`{"field":true,"nested":{"field":false,"nested":{"field":true}}}`,
				[]any{int64(0), int64(1), int64(0)}, // Toggle: true -> false, false -> true, true -> false
				`{"field":false,"nested":{"field":true,"nested":{"field":false}}}`,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			FireCommand(conn, "DEL user")
			for i, cmd := range tc.commands {
				result := FireCommand(conn, cmd)
				switch expected := tc.expected[i].(type) {
				case string:
					if isJSONString(expected) {
						compareJSON(t, expected, result.(string))
					} else {
						assert.Equal(t, expected, result)
					}
				case []interface{}:
					assert.True(t, testutils.UnorderedEqual(expected, result))
				default:
					assert.Equal(t, expected, result)
				}
			}
		})
	}
}

func isJSONString(s string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(s), &js) == nil
}

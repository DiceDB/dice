package http

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
func TestJSONTOGGLE(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	simpleJSON := `{"name":true,"age":false}`
	complexJson := `{"field":true,"nested":{"field":false,"nested":{"field":true}}}`

	testCases := []struct {
		name     string
		commands []HTTPCommand
		expected []interface{}
	}{
		{
			name: "JSON.TOGGLE with existing key",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "user", "path": "$", "value": simpleJSON}},
				{Command: "JSON.TOGGLE", Body: map[string]interface{}{"key": "user", "path": "$.name"}},
			},
			expected: []interface{}{"OK", []any{float64(0)}},
		},
		{
			name: "JSON.TOGGLE with non-existing key",
			commands: []HTTPCommand{
				{Command: "JSON.TOGGLE", Body: map[string]interface{}{"key": "user", "path": "$.flag"}},
			},
			expected: []interface{}{"ERR could not perform this operation on a key that doesn't exist"},
		},
		{
			name: "JSON.TOGGLE with invalid path",
			commands: []HTTPCommand{
				{Command: "JSON.TOGGLE", Body: map[string]interface{}{"key": "user", "path": "$.invalidPath"}},
			},
			expected: []interface{}{"ERR could not perform this operation on a key that doesn't exist"},
		},
		{
			name: "JSON.TOGGLE with invalid command format",
			commands: []HTTPCommand{
				{Command: "JSON.TOGGLE", Body: map[string]interface{}{"key": "testKey"}},
			},
			expected: []interface{}{"ERR wrong number of arguments for 'json.toggle' command"},
		},
		{
			name: "deeply nested JSON structure with multiple matching fields",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "user", "path": "$", "value": complexJson}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "user"}},
				{Command: "JSON.TOGGLE", Body: map[string]interface{}{"key": "user", "path": "$..field"}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "user"}},
			},
			expected: []interface{}{
				"OK",
				`{"field":true,"nested":{"field":false,"nested":{"field":true}}}`,
				[]any{float64(0), float64(1), float64(0)}, // Toggle: true -> false, false -> true, true -> false
				`{"field":false,"nested":{"field":true,"nested":{"field":false}}}`,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			exec.FireCommand(HTTPCommand{
				Command: "DEL",
				Body:    map[string]interface{}{"key": "user"},
			})
			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)
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

package http

import (
	"testing"

	"github.com/dicedb/dice/testutils"

	"github.com/stretchr/testify/assert"
)

func TestCopy(t *testing.T) {
	exec := NewHTTPCommandExecutor()
	simpleJSON := `{"name":"John","age":30}`

	testCases := []TestCase{
		{
			name: "COPY when source key doesn't exist",
			commands: []HTTPCommand{
				{Command: "COPY", Body: map[string]interface{}{"keys": []interface{}{"k1", "k2"}}},
			},
			expected: []interface{}{float64(0)},
		},
		{
			name: "COPY with no REPLACE",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "k1", "value": "v1"}},
				{Command: "COPY", Body: map[string]interface{}{"keys": []interface{}{"k1", "k2"}}},
				{Command: "GET", Body: map[string]interface{}{"key": "k1"}},
				{Command: "GET", Body: map[string]interface{}{"key": "k2"}},
			},
			expected: []interface{}{"OK", float64(1), "v1", "v1"},
		},
		{
			name: "COPY with REPLACE",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "k1", "value": "v1"}},
				{Command: "SET", Body: map[string]interface{}{"key": "k2", "value": "v2"}},
				{Command: "GET", Body: map[string]interface{}{"key": "k2"}},
				{Command: "COPY", Body: map[string]interface{}{"keys": []interface{}{"k1", "k2"}, "value": "REPLACE"}},
				{Command: "GET", Body: map[string]interface{}{"key": "k2"}},
			},
			expected: []interface{}{"OK", "OK", "v2", float64(1), "v1"},
		},
		{
			name: "COPY with JSON integer",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k1", "path": "$", "value": "2"}},
				{Command: "COPY", Body: map[string]interface{}{"keys": []interface{}{"k1", "k2"}}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k2"}},
			},
			expected: []interface{}{"OK", float64(1), "2"},
		},
		{
			name: "COPY with JSON boolean",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k1", "path": "$", "value": "true"}},
				{Command: "COPY", Body: map[string]interface{}{"keys": []interface{}{"k1", "k2"}}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k2"}},
			},
			expected: []interface{}{"OK", float64(1), "true"},
		},
		{
			name: "COPY with JSON array",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k1", "path": "$", "value": "[1,2,3]"}},
				{Command: "COPY", Body: map[string]interface{}{"keys": []interface{}{"k1", "k2"}}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k2"}},
			},
			expected: []interface{}{"OK", float64(1), "[1,2,3]"},
		},
		{
			name: "COPY with JSON simple JSON",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k1", "path": "$", "value": simpleJSON}},
				{Command: "COPY", Body: map[string]interface{}{"keys": []interface{}{"k1", "k2"}}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k2"}},
			},
			expected: []interface{}{"OK", float64(1), simpleJSON},
		},
		{
			name: "COPY with no expiry",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "k1", "value": "v1"}},
				{Command: "COPY", Body: map[string]interface{}{"keys": []interface{}{"k1", "k2"}}},
				{Command: "TTL", Body: map[string]interface{}{"key": "k1"}},
				{Command: "TTL", Body: map[string]interface{}{"key": "k2"}},
			},
			expected: []interface{}{"OK", float64(1), float64(-1), float64(-1)},
		},
		{
			name: "COPY with expiry making sure copy expires",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "k1", "value": "v1", "ex": 5}},
				{Command: "COPY", Body: map[string]interface{}{"keys": []interface{}{"k1", "k2"}}},
				{Command: "GET", Body: map[string]interface{}{"key": "k1"}},
				{Command: "GET", Body: map[string]interface{}{"key": "k2"}},
				{Command: "SLEEP", Body: map[string]interface{}{"key": 7}},
				{Command: "GET", Body: map[string]interface{}{"key": "k1"}},
				{Command: "GET", Body: map[string]interface{}{"key": "k2"}},
			},
			expected: []interface{}{"OK", float64(1), "v1", "v1", "OK", nil, nil},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"keys": []interface{}{"k1"}}})
			exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"keys": []interface{}{"k2"}}})
			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)

				if result == nil {
					assert.Equal(t, tc.expected[i], result, "Expected result to be nil for command %v", cmd)
					continue
				}

				if floatResult, ok := result.(float64); ok {
					assert.Equal(t, tc.expected[i], floatResult, "Mismatch for command %v", cmd)
					continue
				}

				if resultStr, ok := result.(string); ok {
					if testutils.IsJSONResponse(resultStr) {
						assert.JSONEq(t, tc.expected[i].(string), resultStr, "Mismatch in JSON response for command %v", cmd)
					} else {
						assert.Equal(t, tc.expected[i], resultStr, "Mismatch for command %v", cmd)
					}
				} else {
					t.Fatalf("command %v returned unexpected type: %T", cmd, result)
				}
			}

		})
	}

}

package http

import (
	"testing"

	"github.com/dicedb/dice/testutils"
	testifyAssert "github.com/stretchr/testify/assert"
	"gotest.tools/v3/assert"
)

func TestJSONARRPOP(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	arrayAtRoot := []interface{}{0, 1, 2, 3}
	nestedArray := map[string]interface{}{"a": 2, "b": []interface{}{0, 1, 2, 3}}
	arrayWithinArray := map[string]interface{}{"a": 2, "b": []interface{}{0, 1, 2, []interface{}{3, 4, 5}}}

	// Deleting the used keys
	exec.FireCommand(HTTPCommand{
		Command: "DEL",
		Body:    map[string]interface{}{"key": "k"},
	})

	testCases := []TestCase{
		{
			name: "update array at root path",
			commands: []HTTPCommand{
				{
					Command: "JSON.SET",
					Body:    map[string]interface{}{"key": "k", "path": "$", "json": arrayAtRoot},
				},
				{
					Command: "JSON.ARRPOP",
					Body:    map[string]interface{}{"key": "k", "path": "$", "index": "2"},
				},
				{
					Command: "JSON.GET",
					Body:    map[string]interface{}{"key": "k"},
				},
			},
			expected: []interface{}{"OK", float64(2), "[0,1,3]"},
		},
		{
			name: "update nested array",
			commands: []HTTPCommand{
				{
					Command: "JSON.SET",
					Body:    map[string]interface{}{"key": "k", "path": "$", "json": nestedArray},
				},
				{
					Command: "JSON.ARRPOP",
					Body:    map[string]interface{}{"key": "k", "path": "$.b", "index": "2"},
				},
				{
					Command: "JSON.GET",
					Body:    map[string]interface{}{"key": "k"},
				},
			},
			expected: []interface{}{"OK", []interface{}{float64(2)}, `{"a":2,"b":[0,1,3]}`},
		},
		{
			name: "update array with default index",
			commands: []HTTPCommand{
				{
					Command: "JSON.SET",
					Body:    map[string]interface{}{"key": "k", "path": "$", "json": arrayAtRoot},
				},
				{
					Command: "JSON.ARRPOP",
					Body:    map[string]interface{}{"key": "k", "path": "$"},
				},
				{
					Command: "JSON.GET",
					Body:    map[string]interface{}{"key": "k"},
				},
			},
			expected: []interface{}{"OK", float64(3), "[0,1,2]"},
		},
		{
			name: "update array within array",
			commands: []HTTPCommand{
				{
					Command: "JSON.SET",
					Body:    map[string]interface{}{"key": "k", "path": "$", "json": arrayWithinArray},
				},
				{
					Command: "JSON.ARRPOP",
					Body:    map[string]interface{}{"key": "k", "path": "$.b[3]", "index": "1"},
				},
				{
					Command: "JSON.GET",
					Body:    map[string]interface{}{"key": "k"},
				},
			},
			expected: []interface{}{"OK", []interface{}{float64(4)}, `{"a":2,"b":[0,1,2,[3,5]]}`},
		},
		{
			name: "non-array path",
			commands: []HTTPCommand{
				{
					Command: "JSON.SET",
					Body:    map[string]interface{}{"key": "k", "path": "$", "json": nestedArray},
				},
				{
					Command: "JSON.ARRPOP",
					Body:    map[string]interface{}{"key": "k", "path": "$.a", "index": "1"},
				},
			},
			expected: []interface{}{"OK", []interface{}{"(nil)"}},
		},
		{
			name: "invalid json path",
			commands: []HTTPCommand{
				{
					Command: "JSON.SET",
					Body:    map[string]interface{}{"key": "k", "path": "$", "json": arrayAtRoot},
				},
				{
					Command: "JSON.ARRPOP",
					Body:    map[string]interface{}{"key": "k", "path": "$..invalid*path", "index": "1"},
				},
			},
			expected: []interface{}{"OK", "ERR invalid JSONPath"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)
				jsonResult, isString := result.(string)

				if isString && testutils.IsJSONResponse(jsonResult) {
					testifyAssert.JSONEq(t, tc.expected[i].(string), jsonResult)
					continue
				}

				if slice, ok := tc.expected[i].([]interface{}); ok {
					assert.Assert(t, testutils.UnorderedEqual(slice, result))
				} else {
					assert.DeepEqual(t, tc.expected[i], result)
				}
			}
		})
	}

	// Deleting the used keys
	exec.FireCommand(HTTPCommand{
		Command: "DEL",
		Body:    map[string]interface{}{"key": "k"},
	})
}

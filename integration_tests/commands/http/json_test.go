package http

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/dicedb/dice/testutils"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
)

func TestJSONOperations(t *testing.T) {
	exec := NewHTTPCommandExecutor()
	simpleJsonString := `{"name":"John","age":30}`
	nestedJsonString := `{"name":"Alice","address":{"city":"New York","zip":"10001"},"array":[1,2,3,4,5]}`
	specialCharsJsonString := `{"key":"value with spaces","emoji":"😀"}`
	unicodeJsonString := `{"unicode":"こんにちは世界"}`
	escapedCharsJsonString := `{"escaped":"\"quoted\", \\backslash\\ and /forward/slash"}`
	complexJsonString := `{"inventory":{"mountain_bikes":[{"id":"bike:1","model":"Phoebe","price":1920,"specs":{"material":"carbon","weight":13.1},"colors":["black","silver"]},{"id":"bike:2","model":"Quaoar","price":2072,"specs":{"material":"aluminium","weight":7.9},"colors":["black","white"]},{"id":"bike:3","model":"Weywot","price":3264,"specs":{"material":"alloy","weight":13.8}}],"commuter_bikes":[{"id":"bike:4","model":"Salacia","price":1475,"specs":{"material":"aluminium","weight":16.6},"colors":["black","silver"]},{"id":"bike:5","model":"Mimas","price":3941,"specs":{"material":"alloy","weight":11.6}}]}}`

	var simpleJson map[string]interface{}
	var nestedJson map[string]interface{}
	var specialCharsJson map[string]interface{}
	var unicodeJson map[string]interface{}
	var escapedCharsJson map[string]interface{}
	var complexJson map[string]interface{}

	json.Unmarshal([]byte(simpleJsonString), &simpleJson)
	json.Unmarshal([]byte(nestedJsonString), &nestedJson)
	json.Unmarshal([]byte(specialCharsJsonString), &specialCharsJson)
	json.Unmarshal([]byte(unicodeJsonString), &unicodeJson)
	json.Unmarshal([]byte(escapedCharsJsonString), &escapedCharsJson)
	json.Unmarshal([]byte(complexJsonString), &complexJson)

	singleOrderedTestCases := []TestCase{
		{
			name: "Set and Get Integer",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "value": "2"}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", "2"},
		},
		{
			name: "Set and Get Boolean True",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "value": true}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", "true"},
		},
		{
			name: "Set and Get Boolean False",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "value": false}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", "false"},
		},
		{
			name: "Set and Get Simple JSON",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": simpleJson}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", simpleJsonString},
		},
		{
			name: "Set and Get Nested JSON",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": nestedJson}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", nestedJsonString},
		},
		{
			name: "Set and Get JSON Array",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": []interface{}{0, 1, 2, 3, 4}}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", `[0,1,2,3,4]`},
		},
		{
			name: "Set and Get JSON with Special Characters",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": specialCharsJson}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", specialCharsJsonString},
		},
		{
			name: "Set Non-JSON Value",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "k1", "value": "1"}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k1"}},
			},
			expected: []interface{}{"OK", "ERR Existing key has wrong Dice type"},
		},
		{
			name: "Set Empty JSON Object",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": map[string]interface{}{}}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", `{}`},
		},
		{
			name: "Set Empty JSON Array",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": []interface{}{}}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", `[]`},
		},
		{
			name: "Set JSON with Unicode",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": unicodeJson}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", unicodeJsonString},
		},
		{
			name: "Set JSON with Escaped Characters",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": escapedCharsJson}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", escapedCharsJsonString},
		},
		{
			name: "Set and Get Complex JSON",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": complexJson}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", complexJsonString},
		},
		{
			name: "Get Nested Array",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": complexJson}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k", "path": "$.inventory.mountain_bikes[*].model"}},
			},
			expected: []interface{}{"OK", `["Phoebe","Quaoar","Weywot"]`},
		},
		{
			name: "Get Nested Object",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": complexJson}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k", "path": "$.inventory.mountain_bikes[0].specs"}},
			},
			expected: []interface{}{"OK", `{"material":"carbon","weight":13.1}`},
		},
		{
			name: "Set Nested Value",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$.inventory.mountain_bikes[0].price", "value": 2000}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k", "path": "$.inventory.mountain_bikes[0].price"}},
			},
			expected: []interface{}{"OK", "2000"},
		},
	}

	multipleOrderedTestCases := []TestCase{
		{
			name: "Get All Prices",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": complexJson}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k", "path": "$..price"}},
			},
			expected: []interface{}{"OK", []interface{}{1475.0, 3941.0, 1920.0, 2072.0, 3264.0}},
		},
		{
			name: "Set Multiple Nested Values",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$.inventory.*[?(@.price<2000)].price", "value": 1500}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k", "path": "$..price"}},
			},
			expected: []interface{}{"OK", []interface{}{1500.0, 3941.0, 1500.0, 2072.0, 3264.0}},
		},
	}

	t.Run("Single Ordered Test Cases", func(t *testing.T) {
		for _, tc := range singleOrderedTestCases {
			t.Run(tc.name, func(t *testing.T) {
				for i, cmd := range tc.commands {
					result, _ := exec.FireCommand(cmd)

					if jsonResult, ok := result.(string); ok && testutils.IsJSONResponse(jsonResult) {
						assert.JSONEq(t, tc.expected[i].(string), jsonResult)
					} else {
						assert.Equal(t, tc.expected[i], result)
					}
				}
			})
		}
	})

	t.Run("Multiple Ordered Test Cases", func(t *testing.T) {
		for _, tc := range multipleOrderedTestCases {
			t.Run(tc.name, func(t *testing.T) {
				for i, cmd := range tc.commands {
					result, _ := exec.FireCommand(cmd)

					if jsonResult, ok := result.(string); ok && testutils.IsJSONResponse(jsonResult) {
						var jsonPayload []interface{}
						json.Unmarshal([]byte(jsonResult), &jsonPayload)
						assert.True(t, testutils.UnorderedEqual(tc.expected[i], jsonPayload))
					} else {
						assert.Equal(t, tc.expected[i], result)
					}
				}
			})
		}
	})

	// Deleting the used keys
	exec.FireCommand(HTTPCommand{
		Command: "DEL",
		Body: map[string]interface{}{
			"keys": []interface{}{"k", "k1"},
		},
	})
}

func TestJSONSetWithInvalidCases(t *testing.T) {
	exec := NewHTTPCommandExecutor()
	testCases := []TestCase{
		{
			name: "Set Invalid JSON",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "value": `{invalid:json}`}},
			},
			expected: []interface{}{"ERR invalid JSON"},
		},
		{
			name: "Set JSON with Wrong Number of Arguments",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"ERR wrong number of arguments for 'json.set' command"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)
				assert.True(t, strings.HasPrefix(result.(string), tc.expected[i].(string)), fmt.Sprintf("Expected: %s, Got: %s", tc.expected[i], result))
			}
		})
	}
}

func TestJSONSetWithNXAndXX(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	user1JsonString := `{"name": "John", "age": 30}`
	user2JsonString := `{"name": "Rahul", "age": 28}`

	var user1, user2 map[string]interface{}
	json.Unmarshal([]byte(user1JsonString), &user1)
	json.Unmarshal([]byte(user2JsonString), &user2)

	testCases := []TestCase{
		{
			name: "Set with XX on non-existent key",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": user1, "xx": "true"}},
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": user2, "nx": "true"}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{nil, "OK", user2JsonString},
		},
		{
			name: "Set with NX on existing key",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": user1}},
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": user2, "nx": "true"}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", nil, user1JsonString},
		},
		{
			name: "Set with XX on existing key",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": user1}},
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": user2, "xx": "true"}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", "OK", user2JsonString},
		},
		{
			name: "Set with NX on non-existent key",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": user1, "nx": "true"}},
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": user2, "nx": "true"}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", nil, user1JsonString},
		},
		{
			name: "Invalid combinations of NX and XX",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k1", "path": "$", "json": user1, "nx": "true", "xx": "true"}},
			},
			expected: []interface{}{"ERR syntax error"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)
				jsonResult, isString := result.(string)
				if isString && testutils.IsJSONResponse(jsonResult) {
					assert.JSONEq(t, tc.expected[i].(string), jsonResult)
				} else {
					assert.Equal(t, tc.expected[i], result)
				}
			}

			exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"keys": []interface{}{"k", "k1"}}})
		})

		// // Deleting the used keys
		// exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"keys": []interface{}{"k", "k1"}}})
	}
}

func TestJSONClearOperations(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []TestCase{
		{
			name: "jsonclear clear root path",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": map[string]interface{}{"a": 1}}},
				{Command: "JSON.CLEAR", Body: map[string]interface{}{"key": "k", "path": "$"}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", float64(1), "{}"},
		},
		{
			name: "jsonclear clear string type",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": map[string]interface{}{"name": "Tom"}}},
				{Command: "JSON.CLEAR", Body: map[string]interface{}{"key": "k", "path": "$.name"}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k", "path": "$.name"}},
			},
			expected: []interface{}{"OK", float64(0), `"Tom"`},
		},
		{
			name: "jsonclear clear array type",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": map[string]interface{}{"names": []interface{}{"Tom", "Jerry"}}}},
				{Command: "JSON.CLEAR", Body: map[string]interface{}{"key": "k", "path": "$.names"}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k", "path": "$.names"}},
			},
			expected: []interface{}{"OK", float64(1), "[]"},
		},
		{
			name: "jsonclear clear bool type",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": map[string]interface{}{"flag": true}}},
				{Command: "JSON.CLEAR", Body: map[string]interface{}{"key": "k", "path": "$.flag"}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k", "path": "$.flag"}},
			},
			expected: []interface{}{"OK", float64(0), "true"},
		},
		{
			name: "jsonclear clear null type",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": map[string]interface{}{"name": nil}}},
				{Command: "JSON.CLEAR", Body: map[string]interface{}{"key": "k", "path": "$.name"}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k", "path": "$.name"}},
			},
			expected: []interface{}{"OK", float64(0), "null"},
		},
		{
			name: "jsonclear clear integer type",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": map[string]interface{}{"age": 30}}},
				{Command: "JSON.CLEAR", Body: map[string]interface{}{"key": "k", "path": "$.age"}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k", "path": "$.age"}},
			},
			expected: []interface{}{"OK", float64(1), "0"},
		},
		{
			name: "jsonclear clear float type",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": map[string]interface{}{"price": 3.14}}},
				{Command: "JSON.CLEAR", Body: map[string]interface{}{"key": "k", "path": "$.price"}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k", "path": "$.price"}},
			},
			expected: []interface{}{"OK", float64(1), "0"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)
				assert.Equal(t, tc.expected[i], result)
			}
		})

	}

	// Deleting the used keys
	exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"keys": []interface{}{"k", "k1"}}})
}

func TestJSONDelOperations(t *testing.T) {
	exec := NewHTTPCommandExecutor()
	testCases := []TestCase{
		{
			name: "Delete root path",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": map[string]interface{}{"name": "Rahul"}}},
				{Command: "JSON.DEL", Body: map[string]interface{}{"key": "k", "path": "$"}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", float64(1), nil},
		},
		{
			name: "Delete nested field",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": map[string]interface{}{"name": "Tom", "address": map[string]interface{}{"city": "New York", "zip": "10001"}}}},
				{Command: "JSON.DEL", Body: map[string]interface{}{"key": "k", "path": "$.address.city"}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", float64(1), `{"name":"Tom","address":{"zip":"10001"}}`},
		},
		{
			name: "del string type",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": map[string]interface{}{"flag": true, "name": "Tom"}}},
				{Command: "JSON.DEL", Body: map[string]interface{}{"key": "k", "path": "$.name"}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", float64(1), `{"flag":true}`},
		},
		{
			name: "del bool type",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": map[string]interface{}{"flag": true, "name": "Tom"}}},
				{Command: "JSON.DEL", Body: map[string]interface{}{"key": "k", "path": "$.flag"}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", float64(1), `{"name":"Tom"}`},
		},
		{
			name: "del null type",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": map[string]interface{}{"name": nil, "age": 28}}},
				{Command: "JSON.DEL", Body: map[string]interface{}{"key": "k", "path": "$.name"}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", float64(1), `{"age":28}`},
		},
		{
			name: "del array type",
			commands: []HTTPCommand{
				{
					Command: "JSON.SET",
					Body: map[string]interface{}{
						"key":  "k",
						"path": "$",
						"json": map[string]interface{}{
							"names":  []interface{}{"Rahul", "Tom"},
							"bosses": map[string]interface{}{"names": []interface{}{"Jerry", "Rocky"}, "hobby": "swim"},
						},
					},
				},
				{Command: "JSON.DEL", Body: map[string]interface{}{"key": "k", "path": "$..names"}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", float64(2), `{"bosses":{"hobby":"swim"}}`},
		},
		{
			name: "del integer type",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": map[string]interface{}{"name": "Tom", "age": 28}}},
				{Command: "JSON.DEL", Body: map[string]interface{}{"key": "k", "path": "$.age"}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", float64(1), `{"name":"Tom"}`},
		},
		{
			name: "del float type",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": map[string]interface{}{"name": "sugar", "price": 3.14}}},
				{Command: "JSON.DEL", Body: map[string]interface{}{"key": "k", "path": "$.price"}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", float64(1), `{"name":"sugar"}`},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)
				jsonResult, isString := result.(string)
				if isString && testutils.IsJSONResponse(jsonResult) {
					assert.JSONEq(t, tc.expected[i].(string), jsonResult)
				} else {
					assert.Equal(t, tc.expected[i], result)
				}
			}
		})
	}

	// Deleting the used keys
	exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"key": "k"}})
}

func TestJSONForgetOperations(t *testing.T) {
	exec := NewHTTPCommandExecutor()
	testCases := []TestCase{
		{
			name: "forget root path",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": map[string]interface{}{"name": "Rahul"}}},
				{Command: "JSON.FORGET", Body: map[string]interface{}{"key": "k", "path": "$"}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", float64(1), nil},
		},
		{
			name: "forget nested field",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": map[string]interface{}{"name": "Tom", "address": map[string]interface{}{"city": "New York", "zip": "10001"}}}},
				{Command: "JSON.FORGET", Body: map[string]interface{}{"key": "k", "path": "$.address.city"}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", float64(1), `{"name":"Tom","address":{"zip":"10001"}}`},
		},
		{
			name: "forget string type",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": map[string]interface{}{"flag": true, "name": "Tom"}}},
				{Command: "JSON.FORGET", Body: map[string]interface{}{"key": "k", "path": "$.name"}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", float64(1), `{"flag":true}`},
		},
		{
			name: "forget bool type",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": map[string]interface{}{"flag": true, "name": "Tom"}}},
				{Command: "JSON.FORGET", Body: map[string]interface{}{"key": "k", "path": "$.flag"}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", float64(1), `{"name":"Tom"}`},
		},
		{
			name: "forget null type",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": map[string]interface{}{"name": nil, "age": 28}}},
				{Command: "JSON.FORGET", Body: map[string]interface{}{"key": "k", "path": "$.name"}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", float64(1), `{"age":28}`},
		},
		{
			name: "forget array type",
			commands: []HTTPCommand{
				{
					Command: "JSON.SET",
					Body: map[string]interface{}{
						"key":  "k",
						"path": "$",
						"json": map[string]interface{}{
							"names":  []interface{}{"Rahul", "Tom"},
							"bosses": map[string]interface{}{"names": []interface{}{"Jerry", "Rocky"}, "hobby": "swim"},
						},
					},
				},
				{Command: "JSON.FORGET", Body: map[string]interface{}{"key": "k", "path": "$..names"}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", float64(2), `{"bosses":{"hobby":"swim"}}`},
		},
		{
			name: "forget integer type",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": map[string]interface{}{"name": "Tom", "age": 28}}},
				{Command: "JSON.FORGET", Body: map[string]interface{}{"key": "k", "path": "$.age"}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", float64(1), `{"name":"Tom"}`},
		},
		{
			name: "forget float type",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": map[string]interface{}{"name": "sugar", "price": 3.14}}},
				{Command: "JSON.FORGET", Body: map[string]interface{}{"key": "k", "path": "$.price"}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", float64(1), `{"name":"sugar"}`},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)
				jsonResult, isString := result.(string)
				if isString && testutils.IsJSONResponse(jsonResult) {
					assert.JSONEq(t, tc.expected[i].(string), jsonResult)
				} else {
					assert.Equal(t, tc.expected[i], result)
				}
			}
		})
	}

	// Deleting the used keys
	exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"key": "k"}})
}

func TestJsonStrlen(t *testing.T) {
	exec := NewHTTPCommandExecutor()
	testCases := []TestCase{
		{
			name: "jsonstrlen with root path",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": []string{"hello", "world"}}},
				{Command: "JSON.STRLEN", Body: map[string]interface{}{"key": "k", "path": "$"}},
			},
			expected: []interface{}{"OK", []interface{}{nil}},
		},
		{
			name: "jsonstrlen nested",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": map[string]interface{}{"name": "jerry", "partner": map[string]interface{}{"name": "tom"}}}},
				{Command: "JSON.STRLEN", Body: map[string]interface{}{"key": "k", "path": "$..name"}},
			},
			expected: []interface{}{"OK", []interface{}{float64(5), float64(3)}},
		},
		{
			name: "jsonstrlen with no path and object at root",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": map[string]interface{}{"name": "bhima", "age": 10}}},
				{Command: "JSON.STRLEN", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", "ERR wrong type of path value - expected string but found object"},
		},
		{
			name: "jsonstrlen with no path and object at boolean",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": true}},
				{Command: "JSON.STRLEN", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", "ERR wrong type of path value - expected string but found boolean"},
		},
		{
			name: "jsonstrlen with no path and object at array",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": []int{1, 2, 3, 4}}},
				{Command: "JSON.STRLEN", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", "ERR wrong type of path value - expected string but found array"},
		},
		{
			name: "jsonstrlen with no path and object at integer",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": 1}},
				{Command: "JSON.STRLEN", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", "ERR wrong type of path value - expected string but found integer"},
		},
		{
			name: "jsonstrlen with no path and object at number",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": 1.9}},
				{Command: "JSON.STRLEN", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", "ERR wrong type of path value - expected string but found number"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)

				if stringResult, ok := result.(string); ok {
					assert.Equal(t, tc.expected[i], stringResult)
				} else {
					assert.True(t, testutils.UnorderedEqual(tc.expected[i], result.([]interface{})))
				}
			}
		})
	}

	// Deleting the used keys
	exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"key": "k"}})
}

func TestJSONMGET(t *testing.T) {
	exec := NewHTTPCommandExecutor()
	setupData := map[string]string{
		"xx":   `["hehhhe","hello"]`,
		"yy":   `{"name":"jerry","partner":{"name":"jerry","language":["rust"]},"partner2":{"language":["rust"]}}`,
		"zz":   `{"name":"tom","partner":{"name":"tom","language":["rust"]},"partner2":{"age":12,"language":["rust"]}}`,
		"doc1": `{"a":1,"b":2,"nested":{"a":3},"c":null}`,
		"doc2": `{"a":4,"b":5,"nested":{"a":6},"c":null}`,
	}

	for key, value := range setupData {
		var jsonPayload interface{}
		json.Unmarshal([]byte(value), &jsonPayload)
		resp, _ := exec.FireCommand(HTTPCommand{
			Command: "JSON.SET",
			Body: map[string]interface{}{
				"key":  key,
				"path": "$",
				"json": jsonPayload,
			},
		})

		fmt.Printf("expacting: %s with got: %s\n", "OK", resp)
		assert.Equal(t, "OK", resp)
	}

	testCases := []TestCase{
		{
			name: "MGET with root path",
			commands: []HTTPCommand{
				{Command: "JSON.MGET", Body: map[string]interface{}{"keys": []interface{}{"xx", "yy", "zz"}, "path": "$"}},
			},
			expected: []interface{}{[]interface{}{setupData["xx"], setupData["yy"], setupData["zz"]}},
		},
		{
			name: "MGET with specific path",
			commands: []HTTPCommand{
				{Command: "JSON.MGET", Body: map[string]interface{}{"keys": []interface{}{"xx", "yy", "zz"}, "path": "$.name"}},
			},
			expected: []interface{}{[]interface{}{nil, `"jerry"`, `"tom"`}},
		},
		{
			name: "MGET with nested path",
			commands: []HTTPCommand{
				{Command: "JSON.MGET", Body: map[string]interface{}{"keys": []interface{}{"xx", "yy", "zz"}, "path": "$.partner2.age"}},
			},
			expected: []interface{}{[]interface{}{nil, nil, "12"}},
		},
		{
			name: "MGET error",
			commands: []HTTPCommand{
				{Command: "JSON.MGET", Body: map[string]interface{}{"path": ""}},
			},
			expected: []interface{}{"ERR wrong number of arguments for 'json.mget' command"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)

				results, ok := result.([]interface{})
				if ok {
					expectedResults := tc.expected[i].([]interface{})
					assert.Equal(t, len(expectedResults), len(results))

					for j := range results {
						// Check if the expected result is a string and not nil
						expectedVal := expectedResults[j]
						resultVal := results[j]

						// Handle nil comparisons
						if expectedVal == nil || resultVal == nil {
							assert.Equal(t, expectedVal, resultVal)
							continue
						}

						expectedStr, isString := expectedVal.(string)
						resultStr, resultIsString := resultVal.(string)

						if isString && resultIsString && testutils.IsJSONResponse(expectedStr) {
							assert.JSONEq(t, expectedStr, resultStr)
						} else {
							assert.Equal(t, expectedVal, resultVal)
						}
					}
				} else {
					// Handle non-slice result
					assert.Equal(t, tc.expected[i], result)
				}
			}
		})
	}

	t.Run("MGET with recursive path", func(t *testing.T) {
		result, _ := exec.FireCommand(HTTPCommand{Command: "JSON.MGET", Body: map[string]interface{}{"keys": []interface{}{"doc1", "doc2"}, "path": "$..a"}})
		results, ok := result.([]interface{})
		assert.True(t, ok, "Expected result to be a slice of interface{}")
		expectedResults := [][]int{{1, 3}, {4, 6}}
		assert.Equal(t, len(expectedResults), len(results), "Expected 2 results")

		for i, result := range results {
			testutils.UnorderedEqual(expectedResults[i], result)
		}
	})

	// Deleting the used keys
	for key := range setupData {
		exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"key": key}})
	}
}

func TestJsonARRAPPEND(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	a := []int{1, 2}
	b := map[string]interface{}{
		"name":     "jerry",
		"partner":  map[string]interface{}{"name": "tom", "score": []int{10}},
		"partner2": map[string]interface{}{"score": []int{10, 20}},
	}
	c := map[string]interface{}{
		"name":     []string{"jerry"},
		"partner":  map[string]interface{}{"name": "tom", "score": []int{10}},
		"partner2": map[string]interface{}{"name": 12, "score": "rust"},
	}

	testCases := []TestCase{
		{
			name: "JSON.ARRAPPEND with root path",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": a}},
				{Command: "JSON.ARRAPPEND", Body: map[string]interface{}{"key": "k", "path": "$", "value": 3}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", []interface{}{3.0}, `[1,2,3]`},
		},
		{
			name: "JSON.ARRAPPEND nested",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": b}},
				{Command: "JSON.ARRAPPEND", Body: map[string]interface{}{"key": "k", "path": "$..score", "value": 10}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", []interface{}{2.0, 3.0}, `{"name":"jerry","partner":{"name":"tom","score":[10,10]},"partner2":{"score":[10,20,10]}}`},
		},
		{
			name: "JSON.ARRAPPEND nested with nil",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": c}},
				{Command: "JSON.ARRAPPEND", Body: map[string]interface{}{"key": "k", "path": "$..score", "value": 10}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", []interface{}{2.0, nil}, `{"name":["jerry"],"partner":{"name":"tom","score":[10,10]},"partner2":{"name":12,"score":"rust"}}`},
		},
		{
			name: "JSON.ARRAPPEND with different datatypes",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": c}},
				{Command: "JSON.ARRAPPEND", Body: map[string]interface{}{"key": "k", "path": "$.name", "value": 1}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", []interface{}{2.0}, `{"name":["jerry", 1],"partner":{"name":"tom","score":[10]},"partner2":{"name":12,"score":"rust"}}`},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)

				// because the order of keys is not guaranteed, we need to check if the result is an array
				if slice, ok := tc.expected[i].([]interface{}); ok {
					assert.True(t, testutils.UnorderedEqual(slice, result))
				} else if testutils.IsJSONResponse(tc.expected[i].(string)) {
					assert.JSONEq(t, tc.expected[i].(string), result.(string))
				} else {
					assert.Equal(t, tc.expected[i], result)
				}
			}
		})
	}

	// Deleting the used keys
	exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"key": "k"}})
}

func TestJsonNummultby(t *testing.T) {
	exec := NewHTTPCommandExecutor()
	a := map[string]interface{}{
		"a": "b",
		"b": []interface{}{
			map[string]interface{}{"a": 2},
			map[string]interface{}{"a": 5},
			map[string]interface{}{"a": "c"},
		},
	}
	invalidArgMessage := "ERR wrong number of arguments for 'json.nummultby' command"

	testCases := []TestCase{
		{
			name: "Invalid number of arguments",
			commands: []HTTPCommand{
				{Command: "JSON.NUMMULTBY", Body: map[string]interface{}{"key": "k"}},
				{Command: "JSON.NUMMULTBY", Body: map[string]interface{}{"path": "$"}},
				{Command: "JSON.NUMMULTBY", Body: map[string]interface{}{"value": "k"}},
			},
			expected: []interface{}{invalidArgMessage, invalidArgMessage, invalidArgMessage},
		},
		{
			name: "MultBy at non-existent key",
			commands: []HTTPCommand{
				{Command: "JSON.NUMMULTBY", Body: map[string]interface{}{"key": "k", "path": "$", "value": 1}},
			},
			expected: []interface{}{"ERR could not perform this operation on a key that doesn't exist"},
		},
		{
			name: "Invalid value of multiplier on non-existent key",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": a}},
				{Command: "JSON.NUMMULTBY", Body: map[string]interface{}{"key": "k", "path": "$.fe", "value": "a"}},
			},
			expected: []interface{}{"OK", "[]"},
		},
		{
			name: "Invalid value of multiplier on existent key",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": a}},
				{Command: "JSON.NUMMULTBY", Body: map[string]interface{}{"key": "k", "path": "$.a", "value": "a"}},
			},
			expected: []interface{}{"OK", "ERR expected value at line 1 column 1"},
		},
		{
			name: "MultBy at recursive path",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": a}},
				{Command: "JSON.NUMMULTBY", Body: map[string]interface{}{"key": "k", "path": "$..a", "value": 2}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k", "path": "$"}},
			},
			expected: []interface{}{"OK", []interface{}{4.0, 10.0, nil, nil}, `{"a":"b","b":[{"a":4},{"a":10},{"a":"c"}]}`},
		},
		{
			name: "MultBy at root path",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": a}},
				{Command: "JSON.NUMMULTBY", Body: map[string]interface{}{"key": "k", "path": "$.a", "value": 2}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k", "path": "$"}},
			},
			expected: []interface{}{"OK", []interface{}{nil}, `{"a":"b","b":[{"a":2},{"a":5},{"a":"c"}]}`},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)

				if slice, ok := tc.expected[i].([]interface{}); ok {
					var resultPayload []interface{}
					sonic.UnmarshalString(result.(string), &resultPayload)
					assert.True(t, testutils.UnorderedEqual(slice, resultPayload))
				} else if testutils.IsJSONResponse(tc.expected[i].(string)) {
					assert.JSONEq(t, tc.expected[i].(string), result.(string))
				} else {
					assert.Equal(t, tc.expected[i], result)
				}
			}
		})
	}

	// Deleting the used keys
	exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"key": "k"}})
}

func TestJsonObjLen(t *testing.T) {
	exec := NewHTTPCommandExecutor()
	setupData := map[string]string{
		"a": `{"name":"jerry","partner":{"name":"tom","language":["rust"]}}`,
		"b": `{"name":"jerry","partner":{"name":"tom","language":["rust"]},"partner2":{"name":"spike","language":["go","rust"]}}`,
		"c": `{"name":"jerry","partner":{"name":"tom","language":["rust"]},"partner2":{"name":12,"language":["rust"]}}`,
		"d": `["this","is","an","array"]`,
	}

	for key, value := range setupData {
		var jsonPayload interface{}
		json.Unmarshal([]byte(value), &jsonPayload)
		resp, _ := exec.FireCommand(HTTPCommand{
			Command: "JSON.SET",
			Body: map[string]interface{}{
				"key":  key,
				"path": "$",
				"json": jsonPayload,
			},
		})
		assert.Equal(t, resp, "OK")
	}

	testCases := []TestCase{
		{
			name: "JSON.OBJLEN with root path",
			commands: []HTTPCommand{
				{Command: "JSON.OBJLEN", Body: map[string]interface{}{"key": "a", "path": "$"}},
			},
			expected: []interface{}{[]interface{}{2.0}},
		},
		{
			name: "JSON.OBJLEN with nested path",
			commands: []HTTPCommand{
				{Command: "JSON.OBJLEN", Body: map[string]interface{}{"key": "b", "path": "$.partner"}},
			},
			expected: []interface{}{[]interface{}{2.0}},
		},
		{
			name: "JSON.OBJLEN with non-object path",
			commands: []HTTPCommand{
				{Command: "JSON.OBJLEN", Body: map[string]interface{}{"key": "d", "path": "$"}},
			},
			expected: []interface{}{[]interface{}{nil}},
		},
		{
			name: "JSON.OBJLEN with nested non-object path",
			commands: []HTTPCommand{
				{Command: "JSON.OBJLEN", Body: map[string]interface{}{"key": "c", "path": "$.partner2.name"}},
			},
			expected: []interface{}{[]interface{}{nil}},
		},
		{
			name: "JSON.OBJLEN nested objects",
			commands: []HTTPCommand{
				{Command: "JSON.OBJLEN", Body: map[string]interface{}{"key": "b", "path": "$..language"}},
			},
			expected: []interface{}{[]interface{}{nil, nil}},
		},
		{
			name: "JSON.OBJLEN invalid json path",
			commands: []HTTPCommand{
				{Command: "JSON.OBJLEN", Body: map[string]interface{}{"key": "b", "path": "$..language*something"}},
			},
			expected: []interface{}{"ERR Path '$..language*something' does not exist"},
		},
		{
			name: "JSON.OBJLEN with non-existent key",
			commands: []HTTPCommand{
				{Command: "JSON.OBJLEN", Body: map[string]interface{}{"key": "non_existing_key", "path": "$"}},
			},
			expected: []interface{}{nil},
		},
		{
			name: "JSON.OBJLEN with empty path",
			commands: []HTTPCommand{
				{Command: "JSON.OBJLEN", Body: map[string]interface{}{"key": "a"}},
			},
			expected: []interface{}{2.0},
		},
		{
			name: "JSON.OBJLEN invalid json path",
			commands: []HTTPCommand{
				{Command: "JSON.OBJLEN", Body: map[string]interface{}{"key": "c", "path": "$[1"}},
			},
			expected: []interface{}{"ERR Path '$[1' does not exist"},
		},
		{
			name: "JSON.OBJLEN with legacy path - root",
			commands: []HTTPCommand{
				{Command: "json.objlen", Body: map[string]interface{}{"key": "c", "path": "."}},
			},
			expected: []interface{}{3.0},
		},
		{
			name: "JSON.OBJLEN with legacy path - inner existing path",
			commands: []HTTPCommand{
				{Command: "json.objlen", Body: map[string]interface{}{"key": "c", "path": ".partner2"}},
			},
			expected: []interface{}{2.0},
		},
		{
			name: "JSON.OBJLEN with legacy path - inner existing path v2",
			commands: []HTTPCommand{
				{Command: "json.objlen", Body: map[string]interface{}{"key": "c", "path": "partner"}},
			},
			expected: []interface{}{2.0},
		},
		{
			name: "JSON.OBJLEN with legacy path - inner non-existent path",
			commands: []HTTPCommand{
				{Command: "json.objlen", Body: map[string]interface{}{"key": "c", "path": ".idonotexist"}},
			},
			expected: []interface{}{nil},
		},
		{
			name: "JSON.OBJLEN with legacy path - inner non-existent path v2",
			commands: []HTTPCommand{
				{Command: "json.objlen", Body: map[string]interface{}{"key": "c", "path": "idonotexist"}},
			},
			expected: []interface{}{nil},
		},
		{
			name: "JSON.OBJLEN with legacy path - inner existent path with nonJSON object",
			commands: []HTTPCommand{
				{Command: "json.objlen", Body: map[string]interface{}{"key": "c", "path": ".name"}},
			},
			expected: []interface{}{"WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
		{
			name: "JSON.OBJLEN with legacy path - inner existent path recursive object",
			commands: []HTTPCommand{
				{Command: "json.objlen", Body: map[string]interface{}{"key": "c", "path": "..partner"}},
			},
			expected: []interface{}{2.0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)

				if slice, ok := tc.expected[i].([]interface{}); ok {
					assert.True(t, testutils.UnorderedEqual(slice, result))
				} else {
					assert.Equal(t, tc.expected[i], result)
				}
			}
		})
	}

	// Deleting the used keys
	for key := range setupData {
		exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"key": key}})
	}
}

func TestJSONNumIncrBy(t *testing.T) {
	exec := NewHTTPCommandExecutor()
	invalidArgMessage := "ERR wrong number of arguments for 'json.numincrby' command"

	testCases := []TestCase{
		{
			name: "Invalid number of arguments",
			commands: []HTTPCommand{
				{Command: "JSON.NUMINCRBY", Body: map[string]interface{}{"key": "k"}},
				{Command: "JSON.NUMINCRBY", Body: map[string]interface{}{"path": "$"}},
				{Command: "JSON.NUMINCRBY", Body: map[string]interface{}{"value": "k"}},
			},
			expected: []interface{}{invalidArgMessage, invalidArgMessage, invalidArgMessage},
		},
		{
			name: "Non-existent key",
			commands: []HTTPCommand{
				{Command: "JSON.NUMINCRBY", Body: map[string]interface{}{"key": "non_existant_key", "path": "$", "value": 1}},
			},
			expected: []interface{}{"ERR could not perform this operation on a key that doesn't exist"},
		},
		{
			name: "Invalid value of increment",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "value": 1}},
				{Command: "JSON.NUMINCRBY", Body: map[string]interface{}{"key": "k", "path": "$", "value": "@"}},
			},
			expected: []interface{}{"OK", "ERR expected value at line 1 column 1"},
		},
		{
			name: "incrby at non root path",
			commands: []HTTPCommand{
				{
					Command: "JSON.SET",
					Body: map[string]interface{}{
						"key":  "k",
						"path": "$",
						"json": map[string]interface{}{
							"a": "b",
							"b": []interface{}{
								map[string]interface{}{"a": 2.2},
								map[string]interface{}{"a": 5},
								map[string]interface{}{"a": "c"},
							},
						},
					},
				},
				{Command: "JSON.NUMINCRBY", Body: map[string]interface{}{"key": "k", "path": "$..a", "value": 2}},
				{Command: "JSON.NUMINCRBY", Body: map[string]interface{}{"key": "k", "path": "$.a", "value": 2}},
				{Command: "JSON.NUMINCRBY", Body: map[string]interface{}{"key": "k", "path": "$..a", "value": -1}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k", "path": "$"}},
			},
			expected: []interface{}{"OK", []interface{}{nil, 4.2, 7.0, nil}, []interface{}{nil}, []interface{}{nil, 3.2, 6.0, nil}, `{"a":"b","b":[{"a":3.2},{"a":6},{"a":"c"}]}`},
		},
		{
			name: "incrby at root path",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "value": 1}},
				{Command: "JSON.NUMINCRBY", Body: map[string]interface{}{"key": "k", "path": "$", "value": 2}},
				{Command: "JSON.NUMINCRBY", Body: map[string]interface{}{"key": "k", "path": "$", "value": -1}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k", "path": "$"}},
			},
			expected: []interface{}{"OK", []interface{}{3.0}, []interface{}{2.0}, "2"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)

				if slice, ok := tc.expected[i].([]interface{}); ok {
					var resultPayload []interface{}
					sonic.UnmarshalString(result.(string), &resultPayload)
					assert.True(t, testutils.UnorderedEqual(slice, resultPayload))
				} else if testutils.IsJSONResponse(tc.expected[i].(string)) {
					assert.JSONEq(t, tc.expected[i].(string), result.(string))
				} else {
					assert.Equal(t, tc.expected[i], result)
				}
			}
		})
	}

	// Deleting the used keys
	exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"key": "k"}})
}

func TestJsonARRINSERT(t *testing.T) {
	exec := NewHTTPCommandExecutor()
	a := []interface{}{1, 2}
	b := map[string]interface{}{
		"name":     "tom",
		"score":    []interface{}{10, 20},
		"partner2": map[string]interface{}{"score": []interface{}{10, 20}},
	}

	testCases := []TestCase{
		{
			name: "JSON.ARRINSERT index out of bounds",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": a}},
				{Command: "JSON.ARRINSERT", Body: map[string]interface{}{"key": "k", "path": "$", "index": 4, "value": 3}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", "ERR index out of bounds", "[1,2]"},
		},
		{
			name: "JSON.ARRINSERT index is not integer",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": a}},
				{Command: "JSON.ARRINSERT", Body: map[string]interface{}{"key": "k", "path": "$", "index": "ss", "value": 3}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", "ERR value is not an integer or out of range", "[1,2]"},
		},
		{
			name: "JSON.ARRINSERT with positive index in root path",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": a}},
				{Command: "JSON.ARRINSERT", Body: map[string]interface{}{"key": "k", "path": "$", "index": 2, "values": []int{3, 4, 5}}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", []interface{}{5.0}, "[1,2,3,4,5]"},
		},
		{
			name: "JSON.ARRINSERT with negative index in root path",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": a}},
				{Command: "JSON.ARRINSERT", Body: map[string]interface{}{"key": "k", "path": "$", "index": -2, "values": []int{3, 4, 5}}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", []interface{}{5.0}, "[3,4,5,1,2]"},
		},
		{
			name: "JSON.ARRINSERT nested with positive index",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": b}},
				{Command: "JSON.ARRINSERT", Body: map[string]interface{}{"key": "k", "path": "$..score", "index": 1, "values": []interface{}{5, 6, true}}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", []interface{}{5.0, 5.0}, `{"name":"tom","score":[10,5,6,true,20],"partner2":{"score":[10,5,6,true,20]}}`},
		},
		{
			name: "JSON.ARRINSERT nested with negative index",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "k", "path": "$", "json": b}},
				{Command: "JSON.ARRINSERT", Body: map[string]interface{}{"key": "k", "path": "$..score", "index": -2, "values": []interface{}{5, 6, true}}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{"OK", []interface{}{5.0, 5.0}, `{"name":"tom","score":[5,6,true,10,20],"partner2":{"score":[5,6,true,10,20]}}`},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)

				// because the order of keys is not guaranteed, we need to check if the result is an array
				if slice, ok := tc.expected[i].([]interface{}); ok {
					assert.True(t, testutils.UnorderedEqual(slice, result))
				} else if testutils.IsJSONResponse(tc.expected[i].(string)) {
					assert.JSONEq(t, tc.expected[i].(string), result.(string))
				} else {
					assert.Equal(t, tc.expected[i], result)
				}
			}
		})
	}

	// Deleting the used keys
	exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"key": "k"}})
}

func TestJsonObjKeys(t *testing.T) {
	exec := NewHTTPCommandExecutor()
	setupData := map[string]string{
		"a": `{"name":"jerry","partner":{"name":"tom","language":["rust"]},"partner2":{"language":["rust"]}}`,
		"b": `{"name":"jerry","partner":{"name":"tom","language":["rust"]},"partner2":{"name":12,"language":["rust"]}}`,
		"c": `{"name":"jerry","partner":{"name":"tom","language":["rust"]},"partner2":{"name":12,"language":["rust"],"extra_key":"value"}}`,
		"d": `{"a":[3],"nested":{"a":{"b":2,"c":1}}}`,
	}

	for key, value := range setupData {
		var jsonPayload interface{}
		json.Unmarshal([]byte(value), &jsonPayload)
		resp, _ := exec.FireCommand(HTTPCommand{
			Command: "JSON.SET",
			Body: map[string]interface{}{
				"key":  key,
				"path": "$",
				"json": jsonPayload,
			},
		})
		assert.Equal(t, resp, "OK")
	}

	testCases := []TestCase{
		{
			name: "JSON.OBJKEYS root object",
			commands: []HTTPCommand{
				{Command: "JSON.OBJKEYS", Body: map[string]interface{}{"key": "a", "path": "$"}},
			},
			expected: []interface{}{
				[]interface{}{[]interface{}{"name", "partner", "partner2"}},
			},
		},
		{
			name: "JSON.OBJKEYS with nested path",
			commands: []HTTPCommand{
				{Command: "JSON.OBJKEYS", Body: map[string]interface{}{"key": "b", "path": "$.partner"}},
			},
			expected: []interface{}{
				[]interface{}{[]interface{}{"name", "language"}},
			},
		},
		{
			name: "JSON.OBJKEYS with non-object path",
			commands: []HTTPCommand{
				{Command: "JSON.OBJKEYS", Body: map[string]interface{}{"key": "c", "path": "$.name"}},
			},
			expected: []interface{}{[]interface{}{nil}},
		},
		{
			name: "JSON.OBJKEYS with nested non-object path",
			commands: []HTTPCommand{
				{Command: "JSON.OBJKEYS", Body: map[string]interface{}{"key": "b", "path": "$.partner.language"}},
			},
			expected: []interface{}{[]interface{}{nil}},
		},
		{
			name: "JSON.OBJKEYS with invalid json path - 1",
			commands: []HTTPCommand{
				{Command: "JSON.OBJKEYS", Body: map[string]interface{}{"key": "b", "path": "$..invalidpath*somethingrandomadded"}},
			},
			expected: []interface{}{"ERR parse error at 16 in $..invalidpath*somethingrandomadded"},
		},
		{
			name: "JSON.OBJKEYS with invalid json path - 2",
			commands: []HTTPCommand{
				{Command: "JSON.OBJKEYS", Body: map[string]interface{}{"key": "c", "path": "$[1"}},
			},
			expected: []interface{}{"ERR expected a number at 4 in $[1"},
		},
		{
			name: "JSON.OBJKEYS with invalid json path - 3",
			commands: []HTTPCommand{
				{Command: "JSON.OBJKEYS", Body: map[string]interface{}{"key": "c", "path": "$[random"}},
			},
			expected: []interface{}{"ERR parse error at 3 in $[random"},
		},
		{
			name: "JSON.OBJKEYS with only key",
			commands: []HTTPCommand{
				{Command: "JSON.OBJKEYS", Body: map[string]interface{}{"key": "a"}},
			},
			expected: []interface{}{
				[]interface{}{[]interface{}{"name", "partner", "partner2"}},
			},
		},
		{
			name: "JSON.OBJKEYS with non-existing key",
			commands: []HTTPCommand{
				{Command: "JSON.OBJKEYS", Body: map[string]interface{}{"key": "thisdoesnotexist"}},
			},
			expected: []interface{}{"ERR could not perform this operation on a key that doesn't exist"},
		},
		{
			name: "JSON.OBJKEYS with multiple json path",
			commands: []HTTPCommand{
				{Command: "JSON.OBJKEYS", Body: map[string]interface{}{"key": "d", "path": "$..a"}},
			},
			expected: []interface{}{
				[]interface{}{
					[]interface{}{"b", "c"},
					nil,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)

				if slice, ok := tc.expected[i].([]interface{}); ok {
					assert.Equal(t, slice, tc.expected[i], cmpopts.SortSlices(func(a, b interface{}) bool {
						return fmt.Sprintf("%v", a) < fmt.Sprintf("%v", b)
					}))
				} else {
					if _, ok := result.([]interface{}); ok {
						assert.ElementsMatch(t, tc.expected[i].([]interface{}), result.([]interface{}))
					} else {
							// handle the case where result is not a []interface{}
							assert.Equal(t, tc.expected[i], result)
					}

				}
			}
		})
	}

	// Deleting the used keys
	for key := range setupData {
		exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"key": key}})
	}
}

func TestJsonARRTRIM(t *testing.T) {
	exec := NewHTTPCommandExecutor()
	a := `[0,1,2]`
	b := `{"connection":{"wireless":true,"names":[0,1,2,3,4]},"names":[0,1,2,3,4]}`

	testCases := []TestCase{
		{
			name: "JSON.ARRTRIM not array",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "b", "path": "$", "json": json.RawMessage(b)}},
				{Command: "JSON.ARRTRIM", Body: map[string]interface{}{"key": "b", "path": "$", "index": 0, "value": 10}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "b"}},
			},
			expected: []interface{}{"OK", []interface{}{nil}, b},
		},
		{
			name: "JSON.ARRTRIM stop index out of bounds",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "a", "path": "$", "json": json.RawMessage(a)}},
				{Command: "JSON.ARRTRIM", Body: map[string]interface{}{"key": "a", "path": "$", "index": -10, "value": 10}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "a"}},
			},

			expected: []interface{}{"OK", []interface{}{float64(3)}, "[0,1,2]"},
		},
		{
			name: "JSON.ARRTRIM start & stop are positive",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "a", "path": "$", "json": json.RawMessage(a)}},
				{Command: "JSON.ARRTRIM", Body: map[string]interface{}{"key": "a", "path": "$", "index": 1, "value": 2}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "a"}},
			},
			expected: []interface{}{"OK", []interface{}{float64(2)}, "[1,2]"},
		},
		{
			name: "JSON.ARRTRIM start & stop are negative",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "a", "path": "$", "json": json.RawMessage(a)}},
				{Command: "JSON.ARRTRIM", Body: map[string]interface{}{"key": "a", "path": "$", "index": -2, "value": -1}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "a"}},
			},
			expected: []interface{}{"OK", []interface{}{float64(2)}, "[1,2]"},
		},
		{
			name: "JSON.ARRTRIM subpath trim",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "b", "path": "$", "json": json.RawMessage(b)}},
				{Command: "JSON.ARRTRIM", Body: map[string]interface{}{"key": "b", "path": "$..names", "index": 1, "value": 4}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "b"}},
			},
			expected: []interface{}{"OK", []interface{}{float64(4), float64(4)}, `{"connection":{"wireless":true,"names":[1,2,3,4]},"names":[1,2,3,4]}`},
		},
		{
			name: "JSON.ARRTRIM subpath not array",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "b", "path": "$", "json": json.RawMessage(b)}},
				{Command: "JSON.ARRTRIM", Body: map[string]interface{}{"key": "b", "path": "$.connection", "index": 0, "value": 1}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "b"}},
			},
			expected: []interface{}{"OK", []interface{}{nil}, b},
		},
		{
			name: "JSON.ARRTRIM positive start larger than stop",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "b", "path": "$", "json": json.RawMessage(b)}},
				{Command: "JSON.ARRTRIM", Body: map[string]interface{}{"key": "b", "path": "$.names", "index": 3, "value": 1}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "b"}},
			},
			expected: []interface{}{"OK", []interface{}{float64(0)}, `{"names":[],"connection":{"wireless":true,"names":[0,1,2,3,4]}}`},
		},
		{
			name: "JSON.ARRTRIM negative start larger than stop",
			commands: []HTTPCommand{
				{Command: "JSON.SET", Body: map[string]interface{}{"key": "b", "path": "$", "json": json.RawMessage(b)}},
				{Command: "JSON.ARRTRIM", Body: map[string]interface{}{"key": "b", "path": "$.names", "index": -1, "value": -3}},
				{Command: "JSON.GET", Body: map[string]interface{}{"key": "b"}},
			},
			expected: []interface{}{"OK", []interface{}{float64(0)}, `{"names":[],"connection":{"wireless":true,"names":[0,1,2,3,4]}}`},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)

				if slice, ok := tc.expected[i].([]interface{}); ok {
					assert.True(t, testutils.UnorderedEqual(slice, result))
				} else if testutils.IsJSONResponse(tc.expected[i].(string)) {
					assert.JSONEq(t, tc.expected[i].(string), result.(string))
				} else {
					assert.Equal(t, tc.expected[i], result)
				}
			}
		})
	}

	// Clean up the keys
	exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"key": "a"}})
	exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"key": "b"}})
}

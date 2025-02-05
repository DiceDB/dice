// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/dicedb/dice/testutils"
	"github.com/dicedb/dicedb-go"

	"github.com/stretchr/testify/assert"
)

type IntegrationTestCase struct {
	name       string
	setupData  string
	commands   []string
	expected   []interface{}
	assertType []string
	cleanUp    []string
}

func runIntegrationTests(t *testing.T, client *dicedb.Client, testCases []IntegrationTestCase, preTestChecksCommand string, postTestChecksCommand string) {
	for _, tc := range testCases {
		if preTestChecksCommand != "" {
			resp := client.FireString(preTestChecksCommand)
			assert.Equal(t, int64(0), resp)
		}

		t.Run(tc.name, func(t *testing.T) {
			if tc.setupData != "" {
				result := client.FireString(tc.setupData)
				assert.Equal(t, "OK", result)
			}

			cleanupAndPostTestChecks := func() {
				for _, cmd := range tc.cleanUp {
					client.FireString(cmd)
				}

				if postTestChecksCommand != "" {
					resp := client.FireString(postTestChecksCommand)
					assert.Equal(t, int64(0), resp)
				}
			}
			defer cleanupAndPostTestChecks()

			for i := 0; i < len(tc.commands); i++ {
				cmd := tc.commands[i]
				out := tc.expected[i]
				result := client.FireString(cmd)

				switch tc.assertType[i] {
				case "equal":
					assert.Equal(t, out, result)
				case "perm_equal":
					assert.True(t, testutils.ArraysArePermutations(testutils.ConvertToArray(out.(string)), testutils.ConvertToArray(result.GetVStr())))
				case "range":
					assert.True(t, result.GetVInt() <= out.(int64) && result.GetVInt() > 0, "Expected %v to be within 0 to %v", result, out)
				case "json_equal":
					assert.JSONEq(t, out.(string), result.GetVStr())
				}
			}
		})
	}
}

func TestJSONOperations(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	simpleJSON := `{"name":"John","age":30}`
	nestedJSON := `{"name":"Alice","address":{"city":"New York","zip":"10001"},"array":[1,2,3,4,5]}`
	specialCharsJSON := `{"key":"value with spaces","emoji":"😀"}`
	unicodeJSON := `{"unicode":"こんにちは世界"}`
	escapedCharsJSON := `{"escaped":"\"quoted\", \\backslash\\ and /forward/slash"}`
	complexJSON := `{"inventory":{"mountain_bikes":[{"id":"bike:1","model":"Phoebe","price":1920,"specs":{"material":"carbon","weight":13.1},"colors":["black","silver"]},{"id":"bike:2","model":"Quaoar","price":2072,"specs":{"material":"aluminium","weight":7.9},"colors":["black","white"]},{"id":"bike:3","model":"Weywot","price":3264,"specs":{"material":"alloy","weight":13.8}}],"commuter_bikes":[{"id":"bike:4","model":"Salacia","price":1475,"specs":{"material":"aluminium","weight":16.6},"colors":["black","silver"]},{"id":"bike:5","model":"Mimas","price":3941,"specs":{"material":"alloy","weight":11.6}}]}}`

	// Background:
	// Ordering in JSON objects is not guaranteed
	// Ordering in JSON arrays is guaranteed
	// Ordering of arrays which are constructed using a key of an object is not maintained across objects

	// Single ordered test cases will cover all JSON operations which have a single possible order of expected elements
	// different JSON Key orderings are not considered as different JSONs, hence will be considered in this category
	// What goes here:
	// - Cases where the possible order of the result is a single permutation
	// 		- JSON Arrays fetched without any JSONPath
	// 		- JSON Objects (key ordering is taken care of)
	// ref: https://github.com/DiceDB/dice/pull/365
	singleOrderedTestCases := []struct {
		name     string
		setCmd   string
		getCmd   string
		expected string
	}{
		{
			name:     "Set and Get Integer",
			setCmd:   `JSON.SET tools $ 2`,
			getCmd:   `JSON.GET tools`,
			expected: "2",
		},
		{
			name:     "Set and Get Boolean True",
			setCmd:   `JSON.SET booleanTrue $ true`,
			getCmd:   `JSON.GET booleanTrue`,
			expected: "true",
		},
		{
			name:     "Set and Get Boolean False",
			setCmd:   `JSON.SET booleanFalse $ false`,
			getCmd:   `JSON.GET booleanFalse`,
			expected: "false",
		},
		{
			name:     "Set and Get Simple JSON",
			setCmd:   `JSON.SET user $ ` + simpleJSON,
			getCmd:   `JSON.GET user`,
			expected: simpleJSON,
		},
		{
			name:     "Set and Get Nested JSON",
			setCmd:   `JSON.SET user:2 $ ` + nestedJSON,
			getCmd:   `JSON.GET user:2`,
			expected: nestedJSON,
		},
		{
			name:     "Set and Get JSON Array",
			setCmd:   `JSON.SET numbers $ [1,2,3,4,5]`,
			getCmd:   `JSON.GET numbers`,
			expected: `[1,2,3,4,5]`,
		},
		{
			name:     "Set and Get JSON with Special Characters",
			setCmd:   `JSON.SET special $ ` + specialCharsJSON,
			getCmd:   `JSON.GET special`,
			expected: specialCharsJSON,
		},
		{
			name:     "Get JSON with Wrong Number of Arguments",
			setCmd:   ``,
			getCmd:   `JSON.GET`,
			expected: "ERR wrong number of arguments for 'json.get' command",
		},
		{
			name:     "Set Non-JSON Value",
			setCmd:   `SET nonJson "not a json"`,
			getCmd:   `JSON.GET nonJson`,
			expected: "WRONGTYPE Operation against a key holding the wrong kind of value",
		},
		{
			name:     "Set Empty JSON Object",
			setCmd:   `JSON.SET empty $ {}`,
			getCmd:   `JSON.GET empty`,
			expected: `{}`,
		},
		{
			name:     "Set Empty JSON Array",
			setCmd:   `JSON.SET emptyArray $ []`,
			getCmd:   `JSON.GET emptyArray`,
			expected: `[]`,
		},
		{
			name:     "Set JSON with Unicode",
			setCmd:   `JSON.SET unicode $ ` + unicodeJSON,
			getCmd:   `JSON.GET unicode`,
			expected: unicodeJSON,
		},
		{
			name:     "Set JSON with Escaped Characters",
			setCmd:   `JSON.SET escaped $ ` + escapedCharsJSON,
			getCmd:   `JSON.GET escaped`,
			expected: escapedCharsJSON,
		},
		{
			name:     "Set and Get Complex JSON",
			setCmd:   `JSON.SET inventory $ ` + complexJSON,
			getCmd:   `JSON.GET inventory`,
			expected: complexJSON,
		},
		{
			name:     "Get Nested Array",
			setCmd:   `JSON.SET inventory $ ` + complexJSON,
			getCmd:   `JSON.GET inventory $.inventory.mountain_bikes[*].model`,
			expected: `["Phoebe","Quaoar","Weywot"]`,
		},
		{
			name:     "Get Nested Object",
			setCmd:   `JSON.SET inventory $ ` + complexJSON,
			getCmd:   `JSON.GET inventory $.inventory.mountain_bikes[0].specs`,
			expected: `{"material":"carbon","weight":13.1}`,
		},
		{
			name:     "Set Nested Value",
			setCmd:   `JSON.SET inventory $.inventory.mountain_bikes[0].price 2000`,
			getCmd:   `JSON.GET inventory $.inventory.mountain_bikes[0].price`,
			expected: `2000`,
		},
		{
			name:     "Get JSON with non-existent path",
			setCmd:   `JSON.SET user $ ` + simpleJSON,
			getCmd:   `JSON.GET user $.nonExistent`,
			expected: `(nil)`,
		},
	}

	// Multiple test cases will address JSON operations where the order of elements can vary, but all orders are "valid" and to be accepted
	// The variation in order is due to the inherent nature of JSON objects.
	// When dealing with a resultant array that contains elements from multiple objects, the order of these elements can have several valid permutations.
	// Which then means, the overall order of elements in the resultant array is not fixed, although each sub-array within it is guaranteed to be ordered.
	// What goes here:
	// - Cases where the possible order of the resultant array is multiple permutations
	// ref: https://github.com/DiceDB/dice/pull/365
	multipleOrderedTestCases := []struct {
		name     string
		setCmd   string
		getCmd   string
		expected []string
	}{
		{
			name:     "Get All Prices",
			setCmd:   `JSON.SET inventory $ ` + complexJSON,
			getCmd:   `JSON.GET inventory $..price`,
			expected: []string{`[1475,3941,1920,2072,3264]`, `[1920,2072,3264,1475,3941]`}, // Ordering agnostic
		},
		{
			name:     "Set Multiple Nested Values",
			setCmd:   `JSON.SET inventory $.inventory.*[?(@.price<2000)].price 1500`,
			getCmd:   `JSON.GET inventory $..price`,
			expected: []string{`[1500,3941,1500,2072,3264]`, `[1500,2072,3264,1500,3941]`}, // Ordering agnostic
		},
	}

	t.Run("Single Ordered Test Cases", func(t *testing.T) {
		for _, tc := range singleOrderedTestCases {
			t.Run(tc.name, func(t *testing.T) {
				if tc.setCmd != "" {
					result := client.FireString(tc.setCmd)
					assert.Equal(t, "OK", result)
				}

				if tc.getCmd != "" {
					result := client.FireString(tc.getCmd)
					if testutils.IsJSONResponse(result.GetVStr()) {
						assert.JSONEq(t, tc.expected, result.GetVStr())
					} else {
						assert.Equal(t, tc.expected, result)
					}
				}
			})
		}
	})

	t.Run("Multiple Ordered Test Cases", func(t *testing.T) {
		for _, tc := range multipleOrderedTestCases {
			t.Run(tc.name, func(t *testing.T) {
				if tc.setCmd != "" {
					result := client.FireString(tc.setCmd)
					assert.Equal(t, "OK", result)
				}

				if tc.getCmd != "" {
					result := client.FireString(tc.getCmd)
					testutils.AssertJSONEqualList(t, tc.expected, result.GetVStr())
				}
			})
		}
	})
}

func TestJSONSetWithInvalidJSON(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []struct {
		name     string
		command  string
		expected string
	}{
		{
			name:     "Set Invalid JSON",
			command:  `JSON.SET invalid $ {invalid:json}`,
			expected: "ERR invalid JSON",
		},
		{
			name:     "Set JSON with Wrong Number of Arguments",
			command:  `JSON.SET`,
			expected: "ERR wrong number of arguments for 'json.set' command",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := client.FireString(tc.command)
			assert.True(t, strings.HasPrefix(result.GetVStr(), tc.expected), fmt.Sprintf("Expected: %s, Got: %s", tc.expected, result))
		})
	}
}

func TestUnsupportedJSONPathPatterns(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()
	complexJSON := `{"inventory":{"mountain_bikes":[{"id":"bike:1","model":"Phoebe","price":1920,"specs":{"material":"carbon","weight":13.1},"colors":["black","silver"]},{"id":"bike:2","model":"Quaoar","price":2072,"specs":{"material":"aluminium","weight":7.9},"colors":["black","white"]},{"id":"bike:3","model":"Weywot","price":3264,"specs":{"material":"alloy","weight":13.8}}],"commuter_bikes":[{"id":"bike:4","model":"Salacia","price":1475,"specs":{"material":"aluminium","weight":16.6},"colors":["black","silver"]},{"id":"bike:5","model":"Mimas","price":3941,"specs":{"material":"alloy","weight":11.6}}]}}`

	setupCmd := `JSON.SET bikes:inventory $ ` + complexJSON
	result := client.FireString(setupCmd)
	assert.Equal(t, "OK", result)

	testCases := []struct {
		name     string
		command  string
		expected string
	}{
		{
			name:     "Regex in JSONPath",
			command:  `JSON.GET bikes:inventory '$..[?(@.specs.material =~ "(?i)al")].model'`,
			expected: "ERR invalid JSONPath",
		},
		{
			name:     "Using @ for referencing other fields",
			command:  `JSON.GET bikes:inventory '$.inventory.mountain_bikes[?(@.specs.material =~ @.regex_pat)].model'`,
			expected: "ERR invalid JSONPath",
		},
		{
			name:     "Complex condition with multiple comparisons",
			command:  `JSON.GET bikes:inventory '$..mountain_bikes[?(@.price < 3000 && @.specs.weight < 10)]'`,
			expected: "ERR invalid JSONPath",
		},
		{
			name:     "Get all colors",
			command:  `JSON.GET bikes:inventory '$..[*].colors'`,
			expected: "ERR invalid JSONPath",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := client.FireString(tc.command)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestJSONSetWithNXAndXX(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	client.FireString("DEL user")

	user1 := `{"name":"John","age":30}`
	user2 := `{"name":"Rahul","age":28}`

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name:     "Set with XX on non-existent key",
			commands: []string{"JSON.SET user $ " + user1 + " XX", "JSON.SET user $ " + user1 + " NX", "JSON.GET user"},
			expected: []interface{}{"(nil)", "OK", user1},
		},
		{
			name:     "Set with NX on existing key",
			commands: []string{"DEL user", "JSON.SET user $ " + user1, "JSON.SET user $ " + user1 + " NX"},
			expected: []interface{}{int64(1), "OK", "(nil)"},
		},
		{
			name:     "Set with XX on existing key",
			commands: []string{"JSON.SET user $ " + user2 + " XX", "JSON.GET user", "DEL user"},
			expected: []interface{}{"OK", user2, int64(1)},
		},
		{
			name:     "Set with NX on non-existent key",
			commands: []string{"JSON.SET user $ " + user2 + " NX", "JSON.SET user $ " + user1 + " NX", "JSON.GET user"},
			expected: []interface{}{"OK", "(nil)", user2},
		},
		{
			name:     "Invalid combinations of NX and XX",
			commands: []string{"JSON.SET user $ " + user2 + " NX NX", "JSON.SET user $ " + user2 + " NX XX", "JSON.SET user $ " + user2 + " NX Abcd", "JSON.SET user $ " + user2 + " NX "},
			expected: []interface{}{"ERR syntax error", "ERR syntax error", "ERR syntax error", "(nil)"},
		},
		{
			name:     "Invalid combinations of XX",
			commands: []string{"JSON.SET user $ " + user2 + " XX XX", "JSON.SET user $ " + user2 + " XX NX", "JSON.SET user $ " + user2 + " XX Abcd", "JSON.SET user $ " + user2 + " XX "},
			expected: []interface{}{"ERR syntax error", "ERR syntax error", "ERR syntax error", "OK"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result := client.FireString(cmd)
				jsonResult := result.GetVStr()
				if testutils.IsJSONResponse(jsonResult) {
					assert.JSONEq(t, tc.expected[i].(string), jsonResult)
				} else {
					assert.Equal(t, tc.expected[i], result)
				}
			}
		})
	}
}

func TestJSONDel(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	client.FireString("DEL user")

	// preTestChecksCommand := "DEL user"
	// postTestChecksCommand := "DEL user"

	testCases := []IntegrationTestCase{
		{
			name:      "Delete root path",
			setupData: `JSON.SET user $ {"age":13,"high":1.60,"flag":true,"name":"jerry","pet":null,"language":["python","golang"],"partner":{"name":"tom","language":["rust"]}}`,
			commands: []string{
				"JSON.DEL user $",
				"JSON.GET user $",
			},
			expected:   []interface{}{int64(1), "(nil)"},
			assertType: []string{"equal", "equal"},
			cleanUp:    []string{"DEL user"},
		},
		{
			name:      "Delete nested field",
			setupData: `JSON.SET user $ {"name":"Tom","address":{"city":"New York","zip":"10001"}}`,
			commands: []string{
				"JSON.DEL user $.address.city",
				"JSON.GET user $",
			},
			expected:   []interface{}{int64(1), `{"name":"Tom","address":{"zip":"10001"}}`},
			assertType: []string{"equal", "jsoneq"},
			cleanUp:    []string{"DEL user"},
		},
		{
			name:      "del string type",
			setupData: `JSON.SET user $ {"flag":true,"name":"Tom"}`,
			commands: []string{
				"JSON.DEL user $.name",
				"JSON.GET user $"},
			expected:   []interface{}{int64(1), `{"flag":true}`},
			assertType: []string{"equal", "jsoneq"},
			cleanUp:    []string{"DEL user"},
		},
		{
			name:      "del bool type",
			setupData: `JSON.SET user $ {"flag":true,"name":"Tom"}`,
			commands: []string{
				"JSON.DEL user $.flag",
				"JSON.GET user $"},
			expected:   []interface{}{int64(1), `{"name":"Tom"}`},
			assertType: []string{"equal", "jsoneq"},
			cleanUp:    []string{"DEL user"},
		},
		{
			name:      "del null type",
			setupData: `JSON.SET user $ {"name":null,"age":28}`,
			commands: []string{
				"JSON.DEL user $.name",
				"JSON.GET user $"},
			expected:   []interface{}{int64(1), `{"age":28}`},
			assertType: []string{"equal", "jsoneq"},
			cleanUp:    []string{"DEL user"},
		},
		{
			name:      "del array type",
			setupData: `JSON.SET user $ {"names":["Rahul","Tom"],"bosses":{"names":["Jerry","Rocky"],"hobby":"swim"}}`,
			commands: []string{
				"JSON.DEL user $..names",
				"JSON.GET user $"},
			expected:   []interface{}{int64(2), `{"bosses":{"hobby":"swim"}}`},
			assertType: []string{"equal", "jsoneq"},
			cleanUp:    []string{"DEL user"},
		},
		{
			name:      "del integer type",
			setupData: `JSON.SET user $ {"age":28,"name":"Tom"}`,
			commands: []string{
				"JSON.DEL user $.age",
				"JSON.GET user $"},
			expected:   []interface{}{int64(1), `{"name":"Tom"}`},
			assertType: []string{"equal", "jsoneq"},
			cleanUp:    []string{"DEL user"},
		},
		{
			name:      "del float type",
			setupData: `JSON.SET user $ {"price":3.14,"name":"sugar"}`,
			commands: []string{
				"JSON.DEL user $.price",
				"JSON.GET user $"},
			expected:   []interface{}{int64(1), `{"name":"sugar"}`},
			assertType: []string{"equal", "jsoneq"},
			cleanUp:    []string{"DEL user"},
		},
		{
			name:      "delete key with []",
			setupData: `JSON.SET user $ {"key[0]":"value","array":["a","b"]}`,
			commands: []string{
				`JSON.DEL user ["key[0]"]`,
				"JSON.GET user $"},
			expected:   []interface{}{int64(1), `{"array": ["a","b"]}`},
			assertType: []string{"equal", "jsoneq"},
			cleanUp:    []string{"DEL user"},
		},
	}

	runIntegrationTests(t, client, testCases, "", "")
}

func TestJSONMGET(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()
	defer client.FireString("DEL xx yy zz p q r t u v doc1 doc2")

	setupData := map[string]string{
		"xx":   `["hehhhe","hello"]`,
		"yy":   `{"name":"jerry","partner":{"name":"jerry","language":["rust"]},"partner2":{"language":["rust"]}}`,
		"zz":   `{"name":"tom","partner":{"name":"tom","language":["rust"]},"partner2":{"age":12,"language":["rust"]}}`,
		"doc1": `{"a":1,"b":2,"nested":{"a":3},"c":null}`,
		"doc2": `{"a":4,"b":5,"nested":{"a":6},"c":null}`,
	}

	for key, value := range setupData {
		resp := client.FireString(fmt.Sprintf("JSON.SET %s $ %s", key, value))
		assert.Equal(t, "OK", resp)
	}

	testCases := []struct {
		name     string
		command  string
		expected []interface{}
	}{
		{
			name:     "MGET with root path",
			command:  "JSON.MGET xx yy zz $",
			expected: []interface{}{setupData["xx"], setupData["yy"], setupData["zz"]},
		},
		{
			name:     "MGET with specific path",
			command:  "JSON.MGET xx yy zz $.name",
			expected: []interface{}{"(nil)", `"jerry"`, `"tom"`},
		},
		{
			name:     "MGET with nested path",
			command:  "JSON.MGET xx yy zz $.partner2.age",
			expected: []interface{}{"(nil)", "(nil)", "12"},
		},
		{
			name:     "MGET error",
			command:  "JSON.MGET t",
			expected: []interface{}{"ERR wrong number of arguments for 'json.mget' command"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := client.FireString(tc.command)
			results := result.GetVList()
			assert.Equal(t, len(tc.expected), len(results))
			for i := range results {
				if testutils.IsJSONResponse(tc.expected[i].(string)) {
					assert.JSONEq(t, tc.expected[i].(string), results[i].GetStringValue())
				} else {
					assert.Equal(t, tc.expected[i], results[i])
				}
			}
		})
	}

	t.Run("MGET with recursive path", testJSONMGETRecursive(client))
}

func sliceContainsItem(slice []int, item int) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func testJSONMGETRecursive(client *dicedb.Client) func(*testing.T) {
	return func(t *testing.T) {
		result := client.FireString("JSON.MGET doc1 doc2 $..a")
		results := result.GetVList()
		assert.Equal(t, 2, len(results), "Expected 2 results")

		expectedSets := [][]int{
			{1, 3},
			{4, 6},
		}

		for i, res := range results {
			var actualSet []int
			err := sonic.UnmarshalString(res.GetStringValue(), &actualSet)
			assert.Nil(t, err, "Failed to unmarshal JSON")

			assert.True(t, len(actualSet) == len(expectedSets[i]),
				"Mismatch in number of elements for set %d", i)

			for _, expected := range expectedSets[i] {
				assert.True(t, sliceContainsItem(actualSet, expected),
					"Set %d does not contain expected value %d", i, expected)
			}
		}
	}
}

func TestJSONForget(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	preTestChecksCommand := `DEL user`
	postTestChecksCommand := `DEL user`

	testCases := []IntegrationTestCase{
		{
			name:      "Forget root path",
			setupData: `JSON.SET user $ {"age":13,"high":1.60,"flag":true,"name":"jerry","pet":null,"language":["python","golang"],"partner":{"name":"tom","language":["rust"]}}`,
			commands: []string{
				"JSON.FORGET user $",
				"JSON.GET user $",
			},
			expected:   []interface{}{int64(1), "(nil)"},
			assertType: []string{"equal", "equal"},
			cleanUp:    []string{"DEL user"},
		},
		{
			name:      "Forget nested field",
			setupData: `JSON.SET user $ {"name":"Tom","address":{"city":"New York","zip":"10001"}}`,
			commands: []string{
				"JSON.FORGET user $.address.city",
				"JSON.GET user $",
			},
			expected:   []interface{}{int64(1), `{"name":"Tom","address":{"zip":"10001"}}`},
			assertType: []string{"equal", "jsoneq"},
			cleanUp:    []string{"DEL user"},
		},
		{
			name:      "forget string type",
			setupData: `JSON.SET user $ {"flag":true,"name":"Tom"}`,
			commands: []string{
				"JSON.FORGET user $.name",
				"JSON.GET user $"},
			expected:   []interface{}{int64(1), `{"flag":true}`},
			assertType: []string{"equal", "jsoneq"},
			cleanUp:    []string{"DEL user"},
		},
		{
			name:      "forget bool type",
			setupData: `JSON.SET user $ {"flag":true,"name":"Tom"}`,
			commands: []string{
				"JSON.FORGET user $.flag",
				"JSON.GET user $"},
			expected:   []interface{}{int64(1), `{"name":"Tom"}`},
			assertType: []string{"equal", "jsoneq"},
			cleanUp:    []string{"DEL user"},
		},
		{
			name:      "forget null type",
			setupData: `JSON.SET user $ {"name":null,"age":28}`,
			commands: []string{
				"JSON.FORGET user $.name",
				"JSON.GET user $"},
			expected:   []interface{}{int64(1), `{"age":28}`},
			assertType: []string{"equal", "jsoneq"},
			cleanUp:    []string{"DEL user"},
		},
		{
			name:      "forget array type",
			setupData: `JSON.SET user $ {"names":["Rahul","Tom"],"bosses":{"names":["Jerry","Rocky"],"hobby":"swim"}}`,
			commands: []string{
				"JSON.FORGET user $..names",
				"JSON.GET user $"},
			expected:   []interface{}{int64(2), `{"bosses":{"hobby":"swim"}}`},
			assertType: []string{"equal", "jsoneq"},
			cleanUp:    []string{"DEL user"},
		},
		{
			name:      "forget integer type",
			setupData: `JSON.SET user $ {"age":28,"name":"Tom"}`,
			commands: []string{
				"JSON.FORGET user $.age",
				"JSON.GET user $"},
			expected:   []interface{}{int64(1), `{"name":"Tom"}`},
			assertType: []string{"equal", "jsoneq"},
			cleanUp:    []string{"DEL user"},
		},
		{
			name:      "forget float type",
			setupData: `JSON.SET user $ {"price":3.14,"name":"sugar"}`,
			commands: []string{
				"JSON.FORGET user $.price",
				"JSON.GET user $"},
			expected:   []interface{}{int64(1), `{"name":"sugar"}`},
			assertType: []string{"equal", "jsoneq"},
			cleanUp:    []string{"DEL user"},
		},
		{
			name:      "forget array element",
			setupData: `JSON.SET user $ {"names":["Rahul","Tom"],"bosses":{"names":["Jerry","Rocky"],"hobby":"swim"}}`,
			commands: []string{
				"JSON.FORGET user $.names[0]",
				"JSON.GET user $"},
			expected:   []interface{}{int64(1), `{"names":["Tom"],"bosses":{"names":["Jerry","Rocky"],"hobby":"swim"}}`},
			assertType: []string{"equal", "jsoneq"},
			cleanUp:    []string{"DEL user"},
		},
	}

	runIntegrationTests(t, client, testCases, preTestChecksCommand, postTestChecksCommand)
}

func TestJSONToggle(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	preTestChecksCommand := "DEL user"
	postTestChecksCommand := "DEL user"

	simpleJSON := `{"name":"DiceDB","hasAccess":false}`
	complexJson := `{"field":true,"nested":{"field":false,"nested":{"field":true}}}`

	testCases := []IntegrationTestCase{
		{
			name:       "JSON.TOGGLE with existing key",
			setupData:  `JSON.SET user $ ` + simpleJSON,
			commands:   []string{"JSON.TOGGLE user $.hasAccess"},
			expected:   []interface{}{[]any{int64(1)}},
			assertType: []string{"jsoneq"},
			cleanUp:    []string{"DEL user"},
		},
		{
			name:       "JSON.TOGGLE with non-existing key",
			setupData:  "",
			commands:   []string{"JSON.TOGGLE user $.flag"},
			expected:   []interface{}{"ERR could not perform this operation on a key that doesn't exist"},
			assertType: []string{"equal"},
			cleanUp:    []string{},
		},
		{
			name:       "JSON.TOGGLE with invalid path",
			setupData:  "",
			commands:   []string{"JSON.TOGGLE user $.invalidPath"},
			expected:   []interface{}{"ERR could not perform this operation on a key that doesn't exist"},
			assertType: []string{"equal"},
			cleanUp:    []string{},
		},
		{
			name:       "JSON.TOGGLE with invalid command format",
			setupData:  "",
			commands:   []string{"JSON.TOGGLE testKey"},
			expected:   []interface{}{"ERR wrong number of arguments for 'json.toggle' command"},
			assertType: []string{"equal"},
			cleanUp:    []string{},
		},
		{
			name:      "deeply nested JSON structure with multiple matching fields",
			setupData: `JSON.SET user $ ` + complexJson,
			commands: []string{
				"JSON.GET user",
				"JSON.TOGGLE user $..field",
				"JSON.GET user",
			},
			expected: []interface{}{
				`{"field":true,"nested":{"field":false,"nested":{"field":true}}}`,
				[]any{int64(0), int64(1), int64(0)}, // Toggle: true -> false, false -> true, true -> false
				`{"field":false,"nested":{"field":true,"nested":{"field":false}}}`,
			},
			assertType: []string{"jsoneq", "jsoneq", "jsoneq"},
			cleanUp:    []string{"DEL user"},
		},
	}

	runIntegrationTests(t, client, testCases, preTestChecksCommand, postTestChecksCommand)
}

func TestJSONNumIncrBy(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	preTestChecksCommand := "DEL foo"
	postTestChecksCommand := "DEL foo"

	invalidArgMessage := "ERR wrong number of arguments for 'json.numincrby' command"

	testCases := []IntegrationTestCase{
		{
			name:       "Invalid number of arguments",
			setupData:  "",
			commands:   []string{"JSON.NUMINCRBY ", "JSON.NUMINCRBY foo", "JSON.NUMINCRBY foo $"},
			expected:   []interface{}{invalidArgMessage, invalidArgMessage, invalidArgMessage},
			assertType: []string{"equal", "equal", "equal"},
			cleanUp:    []string{},
		},
		{
			name:       "Non-existent key",
			setupData:  "",
			commands:   []string{"JSON.NUMINCRBY foo $ 1"},
			expected:   []interface{}{"ERR could not perform this operation on a key that doesn't exist"},
			assertType: []string{"equal"},
			cleanUp:    []string{},
		},
		{
			name:       "Invalid value of increment",
			setupData:  "JSON.SET foo $ 1",
			commands:   []string{"JSON.GET foo $", "JSON.NUMINCRBY foo $ @", "JSON.NUMINCRBY foo $ 122@"},
			expected:   []interface{}{"1", "ERR expected value at line 1 column 1", "ERR trailing characters at line 1 column 4"},
			assertType: []string{"equal", "equal", "equal"},
			cleanUp:    []string{"DEL foo"},
		},
		{
			name:       "incrby at non root path",
			setupData:  fmt.Sprintf("JSON.SET %s $ %s", "foo", `{"a":"b","b":[{"a":2.2},{"a":5},{"a":"c"}]}`),
			commands:   []string{"JSON.NUMINCRBY foo $..a 2", "JSON.NUMINCRBY foo $.a 2", "JSON.GET foo", "JSON.NUMINCRBY foo $..a -2", "JSON.GET foo"},
			expected:   []interface{}{"[null,4.2,7,null]", "[null]", "{\"a\":\"b\",\"b\":[{\"a\":4.2},{\"a\":7},{\"a\":\"c\"}]}", "[null,2.2,5,null]", "{\"a\":\"b\",\"b\":[{\"a\":2.2},{\"a\":5},{\"a\":\"c\"}]}"},
			assertType: []string{"perm_equal", "perm_equal", "json_equal", "perm_equal", "json_equal"},
			cleanUp:    []string{"DEL foo"},
		},
		{
			name:       "incrby at root path",
			setupData:  "JSON.SET foo $ 1",
			commands:   []string{"JSON.NUMINCRBY foo $ 1", "JSON.GET foo $", "JSON.NUMINCRBY foo $ -1", "JSON.GET foo $"},
			expected:   []interface{}{"[2]", "2", "[1]", "1"},
			assertType: []string{"equal", "equal", "equal", "equal"},
			cleanUp:    []string{"DEL foo"},
		},
		{
			name:       "incrby at root path",
			setupData:  "JSON.SET foo $ 1",
			commands:   []string{"expire foo 10", "JSON.NUMINCRBY foo $ 1", "ttl foo", "JSON.GET foo $", "JSON.NUMINCRBY foo $ -1", "JSON.GET foo $"},
			expected:   []interface{}{int64(1), "[2]", int64(10), "2", "[1]", "1"},
			assertType: []string{"equal", "equal", "range", "equal", "equal", "equal"},
			cleanUp:    []string{"DEL foo"},
		},
	}

	runIntegrationTests(t, client, testCases, preTestChecksCommand, postTestChecksCommand)
}

func TestJsonNumMultBy(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	preTestChecksCommand := "DEL docu"
	postTestChecksCommand := "DEL docu"

	a := `{"a":"b","b":[{"a":2},{"a":5},{"a":"c"}]}`
	invalidArgMessage := "ERR wrong number of arguments for 'json.nummultby' command"

	testCases := []IntegrationTestCase{
		{
			name:       "Invalid number of arguments",
			commands:   []string{"JSON.NUMMULTBY ", "JSON.NUMMULTBY docu", "JSON.NUMMULTBY docu $"},
			expected:   []interface{}{invalidArgMessage, invalidArgMessage, invalidArgMessage},
			assertType: []string{"equal", "equal", "equal"},
		},
		{
			name:       "MultBy at non-existent key",
			commands:   []string{"JSON.NUMMULTBY docu $ 1"},
			expected:   []interface{}{"ERR could not perform this operation on a key that doesn't exist"},
			assertType: []string{"equal"},
		},
		{
			name:       "Invalid value of multiplier on non-existent key",
			setupData:  "JSON.SET docu $ " + a,
			commands:   []string{"JSON.NUMMULTBY docu $.fe x"},
			expected:   []interface{}{"[]"},
			assertType: []string{"equal"},
			cleanUp:    []string{"DEL docu"},
		},
		{
			name:       "Invalid value of multiplier on existent key",
			setupData:  "JSON.SET docu $ " + a,
			commands:   []string{"JSON.NUMMULTBY docu $.a x"},
			expected:   []interface{}{"ERR expected value at line 1 column 1"},
			assertType: []string{"equal"},
			cleanUp:    []string{"DEL docu"},
		},
		{
			name:       "MultBy at recursive path",
			setupData:  "JSON.SET docu $ " + a,
			commands:   []string{"JSON.NUMMULTBY docu $..a 2"},
			expected:   []interface{}{"[4,null,10,null]"},
			assertType: []string{"perm_equal"},
			cleanUp:    []string{"DEL docu"},
		},
		{
			name:       "MultBy at root path",
			setupData:  "JSON.SET docu $ " + a,
			commands:   []string{"JSON.NUMMULTBY docu $.a 2"},
			expected:   []interface{}{"[null]"},
			assertType: []string{"perm_equal"},
			cleanUp:    []string{"DEL docu"},
		},
	}

	runIntegrationTests(t, client, testCases, preTestChecksCommand, postTestChecksCommand)
}

func TestJsonStrlen(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()
	client.FireString("DEL doc")

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name: "jsonstrlen with root path",
			commands: []string{
				`JSON.SET doc $ ["hello","world"]`,
				"JSON.STRLEN doc $",
			},
			expected: []interface{}{"OK", []interface{}{"(nil)"}},
		},
		{
			name: "jsonstrlen nested",
			commands: []string{
				`JSON.SET doc $ {"name":"jerry","partner":{"name":"tom"}}`,
				"JSON.STRLEN doc $..name",
			},
			expected: []interface{}{"OK", []interface{}{int64(5), int64(3)}},
		},
		{
			name: "jsonstrlen with no path and object at root",
			commands: []string{
				`JSON.SET doc $ {"name":"bhima","age":10}`,
				"JSON.STRLEN doc",
			},
			expected: []interface{}{"OK", "ERR wrong type of path value - expected string but found object"},
		},
		{
			name: "jsonstrlen with no path and object at boolean",
			commands: []string{
				`JSON.SET doc $ true`,
				"JSON.STRLEN doc",
			},
			expected: []interface{}{"OK", "ERR wrong type of path value - expected string but found boolean"},
		},
		{
			name: "jsonstrlen with no path and object at array",
			commands: []string{
				`JSON.SET doc $ [1,2,3,4]`,
				"JSON.STRLEN doc",
			},
			expected: []interface{}{"OK", "ERR wrong type of path value - expected string but found array"},
		},
		{
			name: "jsonstrlen with no path and object at integer",
			commands: []string{
				`JSON.SET doc $ 1`,
				"JSON.STRLEN doc",
			},
			expected: []interface{}{"OK", "ERR wrong type of path value - expected string but found integer"},
		},
		{
			name: "jsonstrlen with no path and object at number",
			commands: []string{
				`JSON.SET doc $ 1.9`,
				"JSON.STRLEN doc",
			},
			expected: []interface{}{"OK", "ERR wrong type of path value - expected string but found number"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result := client.FireString(cmd)
				stringResult := result.GetVStr()
				assert.Equal(t, tc.expected[i], stringResult)
			}
		})
	}
}

func TestJsonSTRAPPEND(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	simpleJSON := `{"name":"John","age":30}`

	testCases := []struct {
		name     string
		setCmd   string
		getCmd   string
		expected interface{}
	}{
		{
			name:     "STRAPPEND to nested string",
			setCmd:   `JSON.SET doc $ ` + simpleJSON,
			getCmd:   `JSON.STRAPPEND doc $.name " Doe"`,
			expected: int64(8),
		},
		{
			name:     "STRAPPEND to multiple paths",
			setCmd:   `JSON.SET doc $ ` + simpleJSON,
			getCmd:   `JSON.STRAPPEND doc $..name "baz"`,
			expected: int64(7),
		},
		{
			name:     "STRAPPEND to non-string",
			setCmd:   `JSON.SET doc $ ` + simpleJSON,
			getCmd:   `JSON.STRAPPEND doc $.age " years"`,
			expected: "(nil)",
		},
		{
			name:     "STRAPPEND with empty string",
			setCmd:   `JSON.SET doc $ ` + simpleJSON,
			getCmd:   `JSON.STRAPPEND doc $.name ""`,
			expected: int64(4),
		},
		{
			name:     "STRAPPEND to non-existent path",
			setCmd:   `JSON.SET doc $ ` + simpleJSON,
			getCmd:   `JSON.STRAPPEND doc $.nonexistent " test"`,
			expected: []interface{}{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Use FLUSHDB to clear all keys before each test
			result := client.FireString("FLUSHDB")
			assert.Equal(t, "OK", result)

			result = client.FireString(tc.setCmd)
			assert.Equal(t, "OK", result)

			result = client.FireString(tc.getCmd)
			assert.Equal(t, tc.expected, result)

		})
	}
}

func TestJSONClearOperations(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	// deleteTestKeys([]string{"user"}, store)
	client.FireString("DEL user")

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name: "jsonclear root path",
			commands: []string{
				`JSON.SET user $ {"age":13,"high":1.60,"flag":true,"name":"jerry","pet":null,"language":["python","golang"],"partner":{"name":"tom","language":["rust"]}}`,
				"JSON.CLEAR user $",
				"JSON.GET user $",
			},
			expected: []interface{}{"OK", int64(1), "{}"},
		},
		{
			name: "jsonclear string type",
			commands: []string{
				`JSON.SET user $ {"name":"Tom","age":30}`,
				"JSON.CLEAR user $.name",
				"JSON.GET user $.name",
			},
			expected: []interface{}{"OK", int64(0), `"Tom"`},
		},
		{
			name: "jsonclear array type",
			commands: []string{
				`JSON.SET user $ {"names":["Rahul","Tom"],"ages":[25,30]}`,
				"JSON.CLEAR user $.names",
				"JSON.GET user $.names",
			},
			expected: []interface{}{"OK", int64(1), "[]"},
		},
		{
			name: "jsonclear bool type",
			commands: []string{
				`JSON.SET user $  {"flag":true,"name":"Tom"}`,
				"JSON.CLEAR user $.flag",
				"JSON.GET user $.flag"},
			expected: []interface{}{"OK", int64(0), "true"},
		},
		{
			name: "jsonclear null type",
			commands: []string{
				`JSON.SET user $ {"name":null,"age":28}`,
				"JSON.CLEAR user $.pet",
				"JSON.GET user $.name"},
			expected: []interface{}{"OK", int64(0), "null"},
		},
		{
			name: "jsonclear integer type",
			commands: []string{
				`JSON.SET user $ {"age":28,"name":"Tom"}`,
				"JSON.CLEAR user $.age",
				"JSON.GET user $.age"},
			expected: []interface{}{"OK", int64(1), "0"},
		},
		{
			name: "jsonclear float type",
			commands: []string{
				`JSON.SET user $ {"price":3.14,"name":"sugar"}`,
				"JSON.CLEAR user $.price",
				"JSON.GET user $.price"},
			expected: []interface{}{"OK", int64(1), "0"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result := client.FireString(cmd)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}
}

func TestJsonObjLen(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	a := `{"name":"jerry","partner":{"name":"tom","language":["rust"]}}`
	b := `{"name":"jerry","partner":{"name":"tom","language":["rust"]},"partner2":{"name":"spike","language":["go","rust"]}}`
	c := `{"name":"jerry","partner":{"name":"tom","language":["rust"]},"partner2":{"name":12,"language":["rust"]}}`
	d := `["this","is","an","array"]`

	defer func() {
		resp := client.FireString("DEL obj")
		assert.Equal(t, int64(1), resp)
	}()

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name:     "JSON.OBJLEN with root path",
			commands: []string{"json.set obj $ " + a, "json.objlen obj $"},
			expected: []interface{}{"OK", []interface{}{int64(2)}},
		},

		{
			name:     "JSON.OBJLEN with nested path",
			commands: []string{"json.set obj $ " + b, "json.objlen obj $.partner"},
			expected: []interface{}{"OK", []interface{}{int64(2)}},
		},
		{
			name:     "JSON.OBJLEN with non-object path",
			commands: []string{"json.set obj $ " + d, "json.objlen obj $"},
			expected: []interface{}{"OK", []interface{}{"(nil)"}},
		},
		{
			name:     "JSON.OBJLEN with nested non-object path",
			commands: []string{"json.set obj $ " + c, "json.objlen obj $.partner2.name"},
			expected: []interface{}{"OK", []interface{}{"(nil)"}},
		},
		{
			name:     "JSON.OBJLEN nested objects",
			commands: []string{"json.set obj $ " + b, "json.objlen obj $..language"},
			expected: []interface{}{"OK", []interface{}{"(nil)", "(nil)"}},
		},
		{
			name:     "JSON.OBJLEN invalid json path",
			commands: []string{"json.set obj $ " + b, "json.objlen obj $..language*something"},
			expected: []interface{}{"OK", "ERR Path '$..language*something' does not exist"},
		},
		{
			name:     "JSON.OBJLEN with non-existent key",
			commands: []string{"json.set obj $ " + b, "json.objlen non_existing_key $"},
			expected: []interface{}{"OK", "ERR could not perform this operation on a key that doesn't exist"},
		},
		{
			name:     "JSON.OBJLEN with empty path",
			commands: []string{"json.set obj $ " + a, "json.objlen obj"},
			expected: []interface{}{"OK", int64(2)},
		},
		{
			name:     "JSON.OBJLEN invalid json path2",
			commands: []string{"json.set obj $ " + c, "json.objlen obj $[1"},
			expected: []interface{}{"OK", "ERR Path '$[1' does not exist"},
		},
		{
			name:     "JSON.OBJLEN invalid json path",
			commands: []string{"json.set obj $ " + c, "json.objlen"},
			expected: []interface{}{"OK", "ERR wrong number of arguments for 'json.objlen' command"},
		},
		{
			name:     "JSON.OBJLEN with legacy path - root",
			commands: []string{"json.set obj $ " + c, "json.objlen obj ."},
			expected: []interface{}{"OK", int64(3)},
		},
		{
			name:     "JSON.OBJLEN with legacy path - inner existing path",
			commands: []string{"json.set obj $ " + c, "json.objlen obj .partner", "json.objlen obj .partner2"},
			expected: []interface{}{"OK", int64(2), int64(2)},
		},
		{
			name:     "JSON.OBJLEN with legacy path - inner existing path v2",
			commands: []string{"json.set obj $ " + c, "json.objlen obj partner", "json.objlen obj partner2"},
			expected: []interface{}{"OK", int64(2), int64(2)},
		},
		{
			name:     "JSON.OBJLEN with legacy path - inner non-existent path",
			commands: []string{"json.set obj $ " + c, "json.objlen obj .idonotexist"},
			expected: []interface{}{"OK", "(nil)"},
		},
		{
			name:     "JSON.OBJLEN with legacy path - inner non-existent path v2",
			commands: []string{"json.set obj $ " + c, "json.objlen obj idonotexist"},
			expected: []interface{}{"OK", "(nil)"},
		},
		{
			name:     "JSON.OBJLEN with legacy path - inner existent path with nonJSON object",
			commands: []string{"json.set obj $ " + c, "json.objlen obj .name"},
			expected: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
		{
			name:     "JSON.OBJLEN with legacy path - inner existent path recursive object",
			commands: []string{"json.set obj $ " + c, "json.objlen obj ..partner"},
			expected: []interface{}{"OK", int64(2)},
		},
	}

	for _, tcase := range testCases {
		client.FireString("DEL obj")
		t.Run(tcase.name, func(t *testing.T) {
			for i := 0; i < len(tcase.commands); i++ {
				cmd := tcase.commands[i]
				out := tcase.expected[i]
				result := client.FireString(cmd)
				assert.Equal(t, out, result, "Expected out and result to be deeply equal")
			}
		})
	}
}

func TestJSONARRPOP(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()
	client.FireString("DEL key")

	arrayAtRoot := `[0,1,2,3]`
	nestedArray := `{"a":2,"b":[0,1,2,3]}`

	testCases := []struct {
		name        string
		commands    []string
		expected    []interface{}
		assertType  []string
		jsonResp    []bool
		nestedArray bool
		path        string
	}{
		{
			name:       "update array at root path",
			commands:   []string{"json.set key $ " + arrayAtRoot, "json.arrpop key $ 2", "json.get key"},
			expected:   []interface{}{"OK", int64(2), "[0,1,3]"},
			assertType: []string{"equal", "equal", "deep_equal"},
		},
		{
			name:       "update nested array",
			commands:   []string{"json.set key $ " + nestedArray, "json.arrpop key $.b 2", "json.get key"},
			expected:   []interface{}{"OK", []interface{}{int64(2)}, `{"a":2,"b":[0,1,3]}`},
			assertType: []string{"equal", "deep_equal", "na"},
		},
	}

	for _, tcase := range testCases {
		t.Run(tcase.name, func(t *testing.T) {
			for i := 0; i < len(tcase.commands); i++ {
				cmd := tcase.commands[i]
				out := tcase.expected[i]
				result := client.FireString(cmd)

				jsonResult := result.GetVStr()

				if testutils.IsJSONResponse(jsonResult) {
					assert.JSONEq(t, out.(string), jsonResult)
					continue
				}

				if tcase.assertType[i] == "equal" {
					assert.Equal(t, out, result)
				} else if tcase.assertType[i] == "deep_equal" {
					// TODO: fix this
					// assert.True(t, testutils.ArraysArePermutations(out.([]interface{}), result.GetVList()))
				}
			}
		})
	}
}

func TestJsonARRAPPEND(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()
	a := `[1,2]`
	b := `{"name":"jerry","partner":{"name":"tom","score":[10]},"partner2":{"score":[10,20]}}`
	c := `{"name":["jerry"],"partner":{"name":"tom","score":[10]},"partner2":{"name":12,"score":"rust"}}`

	testCases := []struct {
		name       string
		commands   []string
		expected   []interface{}
		assertType []string
	}{

		{
			name:       "JSON.ARRAPPEND with root path",
			commands:   []string{"json.set a $ " + a, `json.arrappend a $ 3`},
			expected:   []interface{}{"OK", []interface{}{int64(3)}},
			assertType: []string{"equal", "deep_equal"},
		},
		{
			name:       "JSON.ARRAPPEND nested",
			commands:   []string{"JSON.SET doc $ " + b, `JSON.ARRAPPEND doc $..score 10`},
			expected:   []interface{}{"OK", []interface{}{int64(2), int64(3)}},
			assertType: []string{"equal", "deep_equal"},
		},
		{
			name:       "JSON.ARRAPPEND nested with nil",
			commands:   []string{"JSON.SET doc $ " + c, `JSON.ARRAPPEND doc $..score 10`},
			expected:   []interface{}{"OK", []interface{}{int64(2), "(nil)"}},
			assertType: []string{"equal", "deep_equal"},
		},
		{
			name:       "JSON.ARRAPPEND with different datatypes",
			commands:   []string{"JSON.SET doc $ " + c, "JSON.ARRAPPEND doc $.name 1"},
			expected:   []interface{}{"OK", []interface{}{int64(2)}},
			assertType: []string{"equal", "deep_equal"},
		},
	}
	for _, tcase := range testCases {
		client.FireString("DEL a")
		client.FireString("DEL doc")
		t.Run(tcase.name, func(t *testing.T) {
			for i := 0; i < len(tcase.commands); i++ {
				cmd := tcase.commands[i]
				out := tcase.expected[i]
				result := client.FireString(cmd)
				if tcase.assertType[i] == "equal" {
					assert.Equal(t, out, result)
				} else if tcase.assertType[i] == "deep_equal" {
					// TODO: fix this
					// assert.True(t, testutils.ArraysArePermutations(out.([]interface{}), result.([]interface{})))
				}
			}
		})
	}
}

func TestJsonARRINSERT(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()
	a := `[1,2]`
	b := `{"name":"tom","score":[10,20],"partner2":{"score":[10,20]}}`

	testCases := []struct {
		name       string
		commands   []string
		expected   []interface{}
		assertType []string
	}{
		{
			name:       "JSON.ARRINSERT index out of bounds",
			commands:   []string{"json.set a $ " + a, `JSON.ARRINSERT a $ 4 3`, "JSON.GET a"},
			expected:   []interface{}{"OK", "ERR index out of bounds", "[1,2]"},
			assertType: []string{"equal", "equal", "equal"},
		},
		{
			name:       "JSON.ARRINSERT index is not integer",
			commands:   []string{"json.set a $ " + a, `JSON.ARRINSERT a $ ss 3`, "JSON.GET a"},
			expected:   []interface{}{"OK", "ERR value is not an integer or out of range", "[1,2]"},
			assertType: []string{"equal", "equal", "equal"},
		},
		{
			name:       "JSON.ARRINSERT with positive index in root path",
			commands:   []string{"json.set a $ " + a, `JSON.ARRINSERT a $ 2 3 4 5`, "JSON.GET a"},
			expected:   []interface{}{"OK", []interface{}{int64(5)}, "[1,2,3,4,5]"},
			assertType: []string{"equal", "deep_equal", "equal"},
		},

		{
			name:       "JSON.ARRINSERT with positive index in root path",
			commands:   []string{"json.set a $ " + a, `JSON.ARRINSERT a $ 2 3 4 5`, "JSON.GET a"},
			expected:   []interface{}{"OK", []interface{}{int64(5)}, "[1,2,3,4,5]"},
			assertType: []string{"equal", "deep_equal", "equal"},
		},
		{
			name:       "JSON.ARRINSERT with negative index in root path",
			commands:   []string{"json.set a $ " + a, `JSON.ARRINSERT a $ -2 3 4 5`, "JSON.GET a"},
			expected:   []interface{}{"OK", []interface{}{int64(5)}, "[3,4,5,1,2]"},
			assertType: []string{"equal", "deep_equal", "equal"},
		},
		{
			name:       "JSON.ARRINSERT nested with positive index",
			commands:   []string{"JSON.SET b $ " + b, `JSON.ARRINSERT b $..score 1 5 6 true`, "JSON.GET b"},
			expected:   []interface{}{"OK", []interface{}{int64(5), int64(5)}, `{"name":"tom","score":[10,5,6,true,20],"partner2":{"score":[10,5,6,true,20]}}`},
			assertType: []string{"equal", "deep_equal", "jsoneq"},
		},
		{
			name:       "JSON.ARRINSERT nested with negative index",
			commands:   []string{"JSON.SET b $ " + b, `JSON.ARRINSERT b $..score -2 5 6 true`, "JSON.GET b"},
			expected:   []interface{}{"OK", []interface{}{int64(5), int64(5)}, `{"name":"tom","score":[5,6,true,10,20],"partner2":{"score":[5,6,true,10,20]}}`},
			assertType: []string{"equal", "deep_equal", "jsoneq"},
		},
	}
	for _, tcase := range testCases {
		t.Run(tcase.name, func(t *testing.T) {
			for i := 0; i < len(tcase.commands); i++ {
				cmd := tcase.commands[i]
				out := tcase.expected[i]
				result := client.FireString(cmd)
				if tcase.assertType[i] == "equal" {
					assert.Equal(t, out, result)
				} else if tcase.assertType[i] == "deep_equal" {
					// TODO: fix this
					// assert.True(t, testutils.ArraysArePermutations(out.([]interface{}), result.([]interface{})))
				} else if tcase.assertType[i] == "jsoneq" {
					assert.JSONEq(t, out.(string), result.GetVStr())
				}
			}
		})
	}
}

func TestJsonObjKeys(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()
	a := `{"name":"jerry","partner":{"name":"tom","language":["rust"]},"partner2":{"language":["rust"]}}`
	b := `{"name":"jerry","partner":{"name":"tom","language":["rust"]},"partner2":{"name":12,"language":["rust"]}}`
	c := `{"name":"jerry","partner":{"name":"tom","language":["rust"]},"partner2":{"name":12,"language":["rust"],"extra_key":"value"}}`
	d := `{"a":[3],"nested":{"a":{"b":2,"c":1}}}`

	testCases := []struct {
		name       string
		commands   []string
		expected   []interface{}
		assertType []string
	}{
		{
			name:     "JSON.OBJKEYS root object",
			commands: []string{"json.set doc $ " + a, "json.objkeys doc $"},
			expected: []interface{}{"OK", []interface{}{
				[]interface{}{"name", "partner", "partner2"},
			}},
			assertType: []string{"equal", "nested_perm_equal"},
		},
		{
			name:     "JSON.OBJKEYS with nested path",
			commands: []string{"json.set doc $ " + b, "json.objkeys doc $.partner"},
			expected: []interface{}{"OK", []interface{}{
				[]interface{}{"name", "language"},
			}},
			assertType: []string{"equal", "nested_perm_equal"},
		},
		{
			name:     "JSON.OBJKEYS with non-object path",
			commands: []string{"json.set doc $ " + c, "json.objkeys doc $.name"},
			expected: []interface{}{"OK", []interface{}{
				"(nil)",
			}},
			assertType: []string{"equal", "nested_perm_equal"},
		},
		{
			name:     "JSON.OBJKEYS with nested non-object path",
			commands: []string{"json.set doc $ " + b, "json.objkeys doc $.partner.language"},
			expected: []interface{}{"OK", []interface{}{
				"(nil)",
			}},
			assertType: []string{"equal", "nested_perm_equal"},
		},
		{
			name:     "JSON.OBJKEYS with invalid json path - 1",
			commands: []string{"json.set doc $ " + b, "json.objkeys doc $..invalidpath*somethingrandomadded"},
			expected: []interface{}{
				"OK",
				"ERR parse error at 16 in $..invalidpath*somethingrandomadded",
			},
			assertType: []string{"equal", "equal"},
		},
		{
			name:       "JSON.OBJKEYS with invalid json path - 2",
			commands:   []string{"json.set doc $ " + c, "json.objkeys doc $[1"},
			expected:   []interface{}{"OK", "ERR expected a number at 4 in $[1"},
			assertType: []string{"equal", "equal"},
		},
		{
			name:       "JSON.OBJKEYS with invalid json path - 3",
			commands:   []string{"json.set doc $ " + c, "json.objkeys doc $[random"},
			expected:   []interface{}{"OK", "ERR parse error at 3 in $[random"},
			assertType: []string{"equal", "equal"},
		},
		{
			name:       "JSON.OBJKEYS with only command",
			commands:   []string{"json.set doc $ " + c, "json.objkeys"},
			expected:   []interface{}{"OK", "ERR wrong number of arguments for 'json.objkeys' command"},
			assertType: []string{"equal", "equal"},
		},
		{
			name:       "JSON.OBJKEYS with non-existing key",
			commands:   []string{"json.set doc $ " + c, "json.objkeys thisdoesnotexist $"},
			expected:   []interface{}{"OK", "ERR could not perform this operation on a key that doesn't exist"},
			assertType: []string{"equal", "equal"},
		},
		{
			name:     "JSON.OBJKEYS with empty path",
			commands: []string{"json.set doc $ " + c, "json.objkeys doc"},
			expected: []interface{}{"OK", []interface{}{
				"name", "partner", "partner2",
			}},
			assertType: []string{"equal", "nested_perm_equal"},
		},
		{
			name:     "JSON.OBJKEYS with multiple json path",
			commands: []string{"json.set doc $ " + d, "json.objkeys doc $..a"},
			expected: []interface{}{"OK", []interface{}{
				[]interface{}{"b", "c"},
				"(nil)",
			}},
			assertType: []string{"equal", "nested_perm_equal"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			for i := 0; i < len(tc.commands); i++ {
				cmd := tc.commands[i]
				out := tc.expected[i]
				result := client.FireString(cmd)

				if tc.assertType[i] == "equal" {
					assert.Equal(t, out, result)
				} else if tc.assertType[i] == "perm_equal" {
					// TODO: fix this
					// assert.True(t, testutils.ArraysArePermutations(out.([]interface{}), result.([]interface{})))
				} else if tc.assertType[i] == "json_equal" {
					assert.JSONEq(t, out.(string), result.GetVStr())
				} else if tc.assertType[i] == "nested_perm_equal" {
					assert.ElementsMatch(t,
						sortNestedSlices(out.([]interface{})),
						// TODO: fix this
						// sortNestedSlices(result.([]interface{})),
						"Mismatch in JSON object keys",
					)
				}
			}
		})
	}

}

func sortNestedSlices(data []interface{}) []interface{} {
	result := make([]interface{}, len(data))
	for i, item := range data {
		if slice, ok := item.([]interface{}); ok {
			sorted := make([]interface{}, len(slice))
			copy(sorted, slice)
			sort.Slice(sorted, func(i, j int) bool {
				return fmt.Sprintf("%v", sorted[i]) < fmt.Sprintf("%v", sorted[j])
			})
			result[i] = sorted
		} else {
			result[i] = item
		}
	}
	return result
}

func TestJsonARRTRIM(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()
	a := `[0,1,2]`
	b := `{"connection":{"wireless":true,"names":[0,1,2,3,4]},"names":[0,1,2,3,4]}`

	client.FireString("DEL a b")
	defer client.FireString("DEL a b")

	testCases := []struct {
		name       string
		commands   []string
		expected   []interface{}
		assertType []string
	}{
		{
			name:       "JSON.ARRTRIM not array",
			commands:   []string{"JSON.SET b $ " + b, `JSON.ARRTRIM b $ 0 10`, "JSON.GET b"},
			expected:   []interface{}{"OK", []interface{}{"(nil)"}, b},
			assertType: []string{"equal", "deep_equal", "jsoneq"},
		},
		{
			name:       "JSON.ARRTRIM stop index out of bounds",
			commands:   []string{"JSON.SET a $ " + a, `JSON.ARRTRIM a $ -10 10`, "JSON.GET a"},
			expected:   []interface{}{"OK", []interface{}{int64(3)}, "[0,1,2]"},
			assertType: []string{"equal", "deep_equal", "equal"},
		},
		{
			name:       "JSON.ARRTRIM start&stop are positive",
			commands:   []string{"JSON.SET a $ " + a, `JSON.ARRTRIM a $ 1 2`, "JSON.GET a"},
			expected:   []interface{}{"OK", []interface{}{int64(2)}, "[1,2]"},
			assertType: []string{"equal", "deep_equal", "equal"},
		},
		{
			name:       "JSON.ARRTRIM start&stop are negative",
			commands:   []string{"JSON.SET a $ " + a, `JSON.ARRTRIM a $ -2 -1 `, "JSON.GET a"},
			expected:   []interface{}{"OK", []interface{}{int64(2)}, "[1,2]"},
			assertType: []string{"equal", "deep_equal", "equal"},
		},

		{
			name:       "JSON.ARRTRIM subpath trim",
			commands:   []string{"JSON.SET b $ " + b, `JSON.ARRTRIM b $..names 1 4`, "JSON.GET b"},
			expected:   []interface{}{"OK", []interface{}{int64(4), int64(4)}, `{"connection":{"wireless":true,"names":[1,2,3,4]},"names":[1,2,3,4]}`},
			assertType: []string{"equal", "deep_equal", "jsoneq"},
		},
		{
			name:       "JSON.ARRTRIM subpath not array",
			commands:   []string{"JSON.SET b $ " + b, `JSON.ARRTRIM b $.connection 0 1`, "JSON.GET b"},
			expected:   []interface{}{"OK", []interface{}{"(nil)"}, b},
			assertType: []string{"equal", "deep_equal", "jsoneq"},
		},
		{
			name:       "JSON.ARRTRIM positive start larger than stop",
			commands:   []string{"JSON.SET b $ " + b, `JSON.ARRTRIM b $.names 3 1`, "JSON.GET b"},
			expected:   []interface{}{"OK", []interface{}{int64(0)}, `{"names":[],"connection":{"wireless":true,"names":[0,1,2,3,4]}}`},
			assertType: []string{"equal", "deep_equal", "jsoneq"},
		},
		{
			name:       "JSON.ARRTRIM negative start larger than stop",
			commands:   []string{"JSON.SET b $ " + b, `JSON.ARRTRIM b $.names -1 -3`, "JSON.GET b"},
			expected:   []interface{}{"OK", []interface{}{int64(0)}, `{"names":[],"connection":{"wireless":true,"names":[0,1,2,3,4]}}`},
			assertType: []string{"equal", "deep_equal", "jsoneq"},
		},
	}
	for _, tcase := range testCases {
		t.Run(tcase.name, func(t *testing.T) {
			for i := 0; i < len(tcase.commands); i++ {
				cmd := tcase.commands[i]
				out := tcase.expected[i]
				result := client.FireString(cmd)
				if tcase.assertType[i] == "equal" {
					assert.Equal(t, out, result)
				} else if tcase.assertType[i] == "deep_equal" {
					// TODO: fix this
					// assert.True(t, testutils.ArraysArePermutations(out.([]interface{}), result.([]interface{})))
				} else if tcase.assertType[i] == "jsoneq" {
					assert.JSONEq(t, out.(string), result.GetVStr())
				}
			}
		})
	}
}

func TestJSONARRINDEX(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	client.FireString("DEL key")
	defer client.FireString("DEL key")

	normalArray := `[0,1,2,3,4,3]`
	nestedArray := `{"arrays":[{"arr":[1,2,3]},{"arr":[2,3,4]},{"arr":[1]}]}`
	nestedArray2 := `{"a":[3],"nested":{"a":{"b":2,"c":1}}}`

	tests := []struct {
		name       string
		commands   []string
		expected   []interface{}
		assertType []string
	}{
		{
			name:       "should return error if key is not present",
			commands:   []string{"json.set key $ " + normalArray, "json.arrindex nonExistentKey $ 3"},
			expected:   []interface{}{"OK", "ERR could not perform this operation on a key that doesn't exist"},
			assertType: []string{"equal", "equal"},
		},
		{
			name:       "should return error if json path is invalid",
			commands:   []string{"json.set key $ " + normalArray, "json.arrindex key $invalid_path 3"},
			expected:   []interface{}{"OK", "ERR Path '$invalid_path' does not exist"},
			assertType: []string{"equal", "equal"},
		},
		{
			name:       "should return error if provided path does not have any data",
			commands:   []string{"json.set key $ " + normalArray, "json.arrindex key $.some_path 3"},
			expected:   []interface{}{"OK", []interface{}{}},
			assertType: []string{"equal", "equal"},
		},
		{
			name:       "should return error if invalid start index provided",
			commands:   []string{"json.set key $ " + normalArray, "json.arrindex key $ 3 abc"},
			expected:   []interface{}{"OK", "ERR Couldn't parse as integer"},
			assertType: []string{"equal", "equal"},
		},
		{
			name:       "should return error if invalid stop index provided",
			commands:   []string{"json.set key $ " + normalArray, "json.arrindex key $ 3 4 abc"},
			expected:   []interface{}{"OK", "ERR Couldn't parse as integer"},
			assertType: []string{"equal", "equal"},
		},
		{
			name:       "should return array index when given element is present",
			commands:   []string{"json.set key $ " + normalArray, "json.arrindex key $ 3"},
			expected:   []interface{}{"OK", []interface{}{int64(3)}},
			assertType: []string{"equal", "equal"},
		},
		{
			name:       "should return -1 when given element is not present",
			commands:   []string{"json.set key $ " + normalArray, "json.arrindex key $ 10"},
			expected:   []interface{}{"OK", []interface{}{int64(-1)}},
			assertType: []string{"equal", "equal"},
		},
		{
			name:       "should return array index with start optional param provided",
			commands:   []string{"json.set key $ " + normalArray, "json.arrindex key $ 3 4"},
			expected:   []interface{}{"OK", []interface{}{int64(5)}},
			assertType: []string{"equal", "equal"},
		},
		{
			name:       "should return array index with start and stop optional param provided",
			commands:   []string{"json.set key $ " + normalArray, "json.arrindex key $ 4 4 5"},
			expected:   []interface{}{"OK", []interface{}{int64(4)}},
			assertType: []string{"equal", "equal"},
		},
		{
			name:       "should return -1 with start and stop optional param provided where start > stop",
			commands:   []string{"json.set key $ " + normalArray, "json.arrindex key $ 3 2 1"},
			expected:   []interface{}{"OK", []interface{}{int64(-1)}},
			assertType: []string{"equal", "equal"},
		},
		{
			name:       "should return -1 with start (out of boud) and stop (out of bound) optional param provided",
			commands:   []string{"json.set key $ " + normalArray, "json.arrindex key $ 3 6 10"},
			expected:   []interface{}{"OK", []interface{}{int64(-1)}},
			assertType: []string{"equal", "equal"},
		},
		{
			name:       "should return list of array indexes for nested json",
			commands:   []string{"json.set key $ " + nestedArray, "json.arrindex key $.arrays.*.arr 3"},
			expected:   []interface{}{"OK", []interface{}{int64(2), int64(1), int64(-1)}},
			assertType: []string{"equal", "deep_equal"},
		},
		{
			name:       "should return list of array indexes for multiple json path",
			commands:   []string{"json.set key $ " + nestedArray, "json.arrindex key $..arr 3"},
			expected:   []interface{}{"OK", []interface{}{int64(2), int64(1), int64(-1)}},
			assertType: []string{"equal", "deep_equal"},
		},
		{
			name:       "should return array of length 1 for nested json path, with index",
			commands:   []string{"json.set key $ " + nestedArray, "json.arrindex key $.arrays[1].arr 3"},
			expected:   []interface{}{"OK", []interface{}{int64(1)}},
			assertType: []string{"equal", "deep_equal"},
		},
		{
			name:       "should return empty array for nonexistent path in nested json",
			commands:   []string{"json.set key $ " + nestedArray, "json.arrindex key $..arr1 3"},
			expected:   []interface{}{"OK", []interface{}{}},
			assertType: []string{"equal", "deep_equal"},
		},
		{
			name:       "should return -1 for each nonexisting value in nested json",
			commands:   []string{"json.set key $ " + nestedArray, "json.arrindex key $..arr 5"},
			expected:   []interface{}{"OK", []interface{}{int64(-1), int64(-1), int64(-1)}},
			assertType: []string{"equal", "deep_equal"},
		},
		{
			name:       "should return nil for non-array path and -1 for array path if value DNE",
			commands:   []string{"json.set key $ " + nestedArray2, "json.arrindex key $..a 2"},
			expected:   []interface{}{"OK", []interface{}{int64(-1), "(nil)"}},
			assertType: []string{"equal", "deep_equal"},
		},
		{
			name:       "should return nil for non-array path if value DNE and valid index for array path if value exists",
			commands:   []string{"json.set key $ " + nestedArray2, "json.arrindex key $..a 3"},
			expected:   []interface{}{"OK", []interface{}{int64(0), "(nil)"}},
			assertType: []string{"equal", "deep_equal"},
		},
		{
			name:       "should handle stop index - 0 which should be last index inclusive",
			commands:   []string{"json.set key $ " + nestedArray, "json.arrindex key $..arr 3 1 0", "json.arrindex key $..arr 3 2 0"},
			expected:   []interface{}{"OK", []interface{}{int64(2), int64(1), int64(-1)}, []interface{}{int64(2), int64(-1), int64(-1)}},
			assertType: []string{"equal", "deep_equal", "deep_equal"},
		},
		{
			name:       "should handle stop index - -1 which should be last index exclusive",
			commands:   []string{"json.set key $ " + nestedArray, "json.arrindex key $..arr 3 1 -1", "json.arrindex key $..arr 3 2 -1"},
			expected:   []interface{}{"OK", []interface{}{int64(-1), int64(1), int64(-1)}, []interface{}{int64(-1), int64(-1), int64(-1)}},
			assertType: []string{"equal", "deep_equal", "deep_equal"},
		},
		{
			name:       "should handle negative start index",
			commands:   []string{"json.set key $ " + nestedArray, "json.arrindex key $..arr 3 -1"},
			expected:   []interface{}{"OK", []interface{}{int64(2), int64(-1), int64(-1)}},
			assertType: []string{"equal", "deep_equal"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client.FireString("DEL key")

			for i, cmd := range tc.commands {
				result := client.FireString(cmd)
				expected := tc.expected[i]
				if tc.assertType[i] == "equal" {
					assert.Equal(t, result, expected)
				} else if tc.assertType[i] == "deep_equal" {
					// TODO: fix this
					// assert.ElementsMatch(t, result.([]interface{}), expected.([]interface{}))
				}
			}
		})
	}
}

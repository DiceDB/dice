package async

import (
	"fmt"
	"github.com/google/go-cmp/cmp/cmpopts"
	"net"
	"strings"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/dicedb/dice/testutils"
	testifyAssert "github.com/stretchr/testify/assert"
	"gotest.tools/v3/assert"
)

func TestJSONOperations(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

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
			expected: "ERR Existing key has wrong Dice type",
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
					result := FireCommand(conn, tc.setCmd)
					assert.Equal(t, "OK", result)
				}

				if tc.getCmd != "" {
					result := FireCommand(conn, tc.getCmd)
					if testutils.IsJSONResponse(result.(string)) {
						testifyAssert.JSONEq(t, tc.expected, result.(string))
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
					result := FireCommand(conn, tc.setCmd)
					assert.Equal(t, "OK", result)
				}

				if tc.getCmd != "" {
					result := FireCommand(conn, tc.getCmd)
					testutils.AssertJSONEqualList(t, tc.expected, result.(string))
				}
			})
		}
	})
}

func TestJSONSetWithInvalidJSON(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

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
			result := FireCommand(conn, tc.command)
			assert.Check(t, strings.HasPrefix(result.(string), tc.expected), fmt.Sprintf("Expected: %s, Got: %s", tc.expected, result))
		})
	}
}

func TestUnsupportedJSONPathPatterns(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()
	complexJSON := `{"inventory":{"mountain_bikes":[{"id":"bike:1","model":"Phoebe","price":1920,"specs":{"material":"carbon","weight":13.1},"colors":["black","silver"]},{"id":"bike:2","model":"Quaoar","price":2072,"specs":{"material":"aluminium","weight":7.9},"colors":["black","white"]},{"id":"bike:3","model":"Weywot","price":3264,"specs":{"material":"alloy","weight":13.8}}],"commuter_bikes":[{"id":"bike:4","model":"Salacia","price":1475,"specs":{"material":"aluminium","weight":16.6},"colors":["black","silver"]},{"id":"bike:5","model":"Mimas","price":3941,"specs":{"material":"alloy","weight":11.6}}]}}`

	setupCmd := `JSON.SET bikes:inventory $ ` + complexJSON
	result := FireCommand(conn, setupCmd)
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
			result := FireCommand(conn, tc.command)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestJSONSetWithNXAndXX(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	FireCommand(conn, "DEL user")

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
				result := FireCommand(conn, cmd)
				jsonResult, isString := result.(string)
				if isString && testutils.IsJSONResponse(jsonResult) {
					testifyAssert.JSONEq(t, tc.expected[i].(string), jsonResult)
				} else {
					assert.Equal(t, tc.expected[i], result)
				}
			}
		})
	}
}

func TestJSONClearOperations(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	// deleteTestKeys([]string{"user"}, store)
	FireCommand(conn, "DEL user")

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name: "Clear root path",
			commands: []string{
				`JSON.SET user $ {"age":13,"high":1.60,"flag":true,"name":"jerry","pet":null,"language":["python","golang"],"partner":{"name":"tom","language":["rust"]}}`,
				"JSON.CLEAR user $",
				"JSON.GET user $",
			},
			expected: []interface{}{"OK", int64(1), "{}"},
		},
		{
			name: "Clear string type",
			commands: []string{
				`JSON.SET user $ {"name":"Tom","age":30}`,
				"JSON.CLEAR user $.name",
				"JSON.GET user $.name",
			},
			expected: []interface{}{"OK", int64(0), `"Tom"`},
		},
		{
			name: "Clear array type",
			commands: []string{
				`JSON.SET user $ {"names":["Rahul","Tom"],"ages":[25,30]}`,
				"JSON.CLEAR user $.names",
				"JSON.GET user $.names",
			},
			expected: []interface{}{"OK", int64(1), "[]"},
		},
		{
			name: "clear bool type",
			commands: []string{
				`JSON.SET user $  {"flag":true,"name":"Tom"}`,
				"JSON.CLEAR user $.flag",
				"JSON.GET user $.flag"},
			expected: []interface{}{"OK", int64(0), "true"},
		},
		{
			name: "clear null type",
			commands: []string{
				`JSON.SET user $ {"name":null,"age":28}`,
				"JSON.CLEAR user $.pet",
				"JSON.GET user $.name"},
			expected: []interface{}{"OK", int64(0), "null"},
		},
		{
			name: "clear integer type",
			commands: []string{
				`JSON.SET user $ {"age":28,"name":"Tom"}`,
				"JSON.CLEAR user $.age",
				"JSON.GET user $.age"},
			expected: []interface{}{"OK", int64(1), "0"},
		},
		{
			name: "clear float type",
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
				result := FireCommand(conn, cmd)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}
}

func TestJSONDelOperations(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	FireCommand(conn, "DEL user")

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name: "Delete root path",
			commands: []string{
				`JSON.SET user $ {"age":13,"high":1.60,"flag":true,"name":"jerry","pet":null,"language":["python","golang"],"partner":{"name":"tom","language":["rust"]}}`,
				"JSON.DEL user $",
				"JSON.GET user $",
			},
			expected: []interface{}{"OK", int64(1), "(nil)"},
		},
		{
			name: "Delete nested field",
			commands: []string{
				`JSON.SET user $ {"name":"Tom","address":{"city":"New York","zip":"10001"}}`,
				"JSON.DEL user $.address.city",
				"JSON.GET user $",
			},
			expected: []interface{}{"OK", int64(1), `{"name":"Tom","address":{"zip":"10001"}}`},
		},
		{
			name: "del string type",
			commands: []string{
				`JSON.SET user $ {"flag":true,"name":"Tom"}`,
				"JSON.DEL user $.name",
				"JSON.GET user $"},
			expected: []interface{}{"OK", int64(1), `{"flag":true}`},
		},
		{
			name: "del bool type",
			commands: []string{
				`JSON.SET user $ {"flag":true,"name":"Tom"}`,
				"JSON.DEL user $.flag",
				"JSON.GET user $"},
			expected: []interface{}{"OK", int64(1), `{"name":"Tom"}`},
		},
		{
			name: "del null type",
			commands: []string{
				`JSON.SET user $ {"name":null,"age":28}`,
				"JSON.DEL user $.name",
				"JSON.GET user $"},
			expected: []interface{}{"OK", int64(1), `{"age":28}`},
		},
		{
			name: "del array type",
			commands: []string{
				`JSON.SET user $ {"names":["Rahul","Tom"],"bosses":{"names":["Jerry","Rocky"],"hobby":"swim"}}`,
				"JSON.DEL user $..names",
				"JSON.GET user $"},
			expected: []interface{}{"OK", int64(2), `{"bosses":{"hobby":"swim"}}`},
		},
		{
			name: "del integer type",
			commands: []string{
				`JSON.SET user $ {"age":28,"name":"Tom"}`,
				"JSON.DEL user $.age",
				"JSON.GET user $"},
			expected: []interface{}{"OK", int64(1), `{"name":"Tom"}`},
		},
		{
			name: "del float type",
			commands: []string{
				`JSON.SET user $ {"price":3.14,"name":"sugar"}`,
				"JSON.DEL user $.price",
				"JSON.GET user $"},
			expected: []interface{}{"OK", int64(1), `{"name":"sugar"}`},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result := FireCommand(conn, cmd)
				stringResult, ok := result.(string)
				if ok && testutils.IsJSONResponse(stringResult) {
					testifyAssert.JSONEq(t, tc.expected[i].(string), stringResult)
				} else {
					assert.Equal(t, tc.expected[i], result)
				}
			}
		})
	}
}

func TestJSONForgetOperations(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	FireCommand(conn, "FORGET user")

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name: "Forget root path",
			commands: []string{
				`JSON.SET user $ {"age":13,"high":1.60,"flag":true,"name":"jerry","pet":null,"language":["python","golang"],"partner":{"name":"tom","language":["rust"]}}`,
				"JSON.FORGET user $",
				"JSON.GET user $",
			},
			expected: []interface{}{"OK", int64(1), "(nil)"},
		},
		{
			name: "Forget nested field",
			commands: []string{
				`JSON.SET user $ {"name":"Tom","address":{"city":"New York","zip":"10001"}}`,
				"JSON.FORGET user $.address.city",
				"JSON.GET user $",
			},
			expected: []interface{}{"OK", int64(1), `{"name":"Tom","address":{"zip":"10001"}}`},
		},
		{
			name: "forget string type",
			commands: []string{`JSON.SET user $ {"flag":true,"name":"Tom"}`,
				"JSON.FORGET user $.name",
				"JSON.GET user $"},
			expected: []interface{}{"OK", int64(1), `{"flag":true}`},
		},
		{
			name: "forget bool type",
			commands: []string{`JSON.SET user $ {"flag":true,"name":"Tom"}`,
				"JSON.FORGET user $.flag",
				"JSON.GET user $"},
			expected: []interface{}{"OK", int64(1), `{"name":"Tom"}`},
		},
		{
			name: "forget null type",
			commands: []string{`JSON.SET user $ {"name":null,"age":28}`,
				"JSON.FORGET user $.name",
				"JSON.GET user $"},
			expected: []interface{}{"OK", int64(1), `{"age":28}`},
		},
		{
			name: "forget array type",
			commands: []string{`JSON.SET user $ {"names":["Rahul","Tom"],"bosses":{"names":["Jerry","Rocky"],"hobby":"swim"}}`,
				"JSON.FORGET user $..names",
				"JSON.GET user $"},
			expected: []interface{}{"OK", int64(2), `{"bosses":{"hobby":"swim"}}`},
		},
		{
			name: "forget integer type",
			commands: []string{`JSON.SET user $ {"age":28,"name":"Tom"}`,
				"JSON.FORGET user $.age",
				"JSON.GET user $"},
			expected: []interface{}{"OK", int64(1), `{"name":"Tom"}`},
		},
		{
			name: "forget float type",
			commands: []string{`JSON.SET user $ {"price":3.14,"name":"sugar"}`,
				"JSON.FORGET user $.price",
				"JSON.GET user $"},
			expected: []interface{}{"OK", int64(1), `{"name":"sugar"}`},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result := FireCommand(conn, cmd)
				stringResult, ok := result.(string)
				if ok && testutils.IsJSONResponse(stringResult) {
					testifyAssert.JSONEq(t, tc.expected[i].(string), stringResult)
				} else {
					assert.Equal(t, tc.expected[i], result)
				}
			}
		})
	}
}

func arraysArePermutations[T comparable](a, b []T) bool {
	// If lengths are different, they cannot be permutations
	if len(a) != len(b) {
		return false
	}

	// Count occurrences of each element in array 'a'
	countA := make(map[T]int)
	for _, elem := range a {
		countA[elem]++
	}

	// Subtract occurrences based on array 'b'
	for _, elem := range b {
		countA[elem]--
		if countA[elem] < 0 {
			return false
		}
	}

	// Check if all counts are zero
	for _, count := range countA {
		if count != 0 {
			return false
		}
	}

	return true
}

func TestJsonStrlen(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()
	FireCommand(conn, "DEL doc")

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name: "STRLEN with root path",
			commands: []string{
				`JSON.SET doc $ ["hello","world"]`,
				"JSON.STRLEN doc $",
			},
			expected: []interface{}{"OK", []interface{}{"(nil)"}},
		},
		{
			name: "STRLEN nested",
			commands: []string{
				`JSON.SET doc $ {"name":"jerry","partner":{"name":"tom"}}`,
				"JSON.STRLEN doc $..name",
			},
			expected: []interface{}{"OK", []interface{}{int64(5), int64(3)}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result := FireCommand(conn, cmd)
				stringResult, ok := result.(string)
				if ok {
					assert.Equal(t, tc.expected[i], stringResult)
				} else {
					assert.Assert(t, arraysArePermutations(tc.expected[i].([]interface{}), result.([]interface{})))
				}
			}
		})
	}
}

func TestJSONMGET(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()
	defer FireCommand(conn, "DEL xx yy zz p q r t u v doc1 doc2")

	setupData := map[string]string{
		"xx":   `["hehhhe","hello"]`,
		"yy":   `{"name":"jerry","partner":{"name":"jerry","language":["rust"]},"partner2":{"language":["rust"]}}`,
		"zz":   `{"name":"tom","partner":{"name":"tom","language":["rust"]},"partner2":{"age":12,"language":["rust"]}}`,
		"doc1": `{"a":1,"b":2,"nested":{"a":3},"c":null}`,
		"doc2": `{"a":4,"b":5,"nested":{"a":6},"c":null}`,
	}

	for key, value := range setupData {
		resp := FireCommand(conn, fmt.Sprintf("JSON.SET %s $ %s", key, value))
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
			result := FireCommand(conn, tc.command)
			results, ok := result.([]interface{})
			if ok {
				assert.Equal(t, len(tc.expected), len(results))
				for i := range results {
					if testutils.IsJSONResponse(tc.expected[i].(string)) {
						testifyAssert.JSONEq(t, tc.expected[i].(string), results[i].(string))
					} else {
						assert.Equal(t, tc.expected[i], results[i])
					}
				}
			} else {
				assert.Equal(t, tc.expected[0], result)
			}
		})
	}

	t.Run("MGET with recursive path", testJSONMGETRecursive(conn))
}

func testJSONMGETRecursive(conn net.Conn) func(*testing.T) {
	return func(t *testing.T) {
		result := FireCommand(conn, "JSON.MGET doc1 doc2 $..a")
		results, ok := result.([]interface{})
		assert.Assert(t, ok, "Expected result to be a slice of interface{}")
		assert.Equal(t, 2, len(results), "Expected 2 results")

		expectedSets := [][]int{
			{1, 3},
			{4, 6},
		}

		for i, res := range results {
			var actualSet []int
			err := sonic.UnmarshalString(res.(string), &actualSet)
			assert.NilError(t, err, "Failed to unmarshal JSON")

			assert.Assert(t, len(actualSet) == len(expectedSets[i]),
				"Mismatch in number of elements for set %d", i)

			for _, expected := range expectedSets[i] {
				assert.Assert(t, sliceContainsItem(actualSet, expected),
					"Set %d does not contain expected value %d", i, expected)
			}
		}
	}
}

// sliceContainsItem checks if a slice sliceContainsItem a given item
func sliceContainsItem(slice []int, item int) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func TestJsonARRAPPEND(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()
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
		FireCommand(conn, "DEL a")
		FireCommand(conn, "DEL doc")
		t.Run(tcase.name, func(t *testing.T) {
			for i := 0; i < len(tcase.commands); i++ {
				cmd := tcase.commands[i]
				out := tcase.expected[i]
				result := FireCommand(conn, cmd)
				if tcase.assertType[i] == "equal" {
					assert.Equal(t, out, result)
				} else if tcase.assertType[i] == "deep_equal" {
					assert.Assert(t, arraysArePermutations(out.([]interface{}), result.([]interface{})))
				}
			}
		})
	}
}

func deStringify(input string) []string {
	input = strings.Replace(input, "[", "", -1)
	input = strings.Replace(input, "]", "", -1)
	arrayString := strings.Split(input, ",")
	for i, v := range arrayString {
		arrayString[i] = strings.TrimSpace(v)
	}
	return arrayString
}

func TestJsonNummultby(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	a := `{"a":"b","b":[{"a":2},{"a":5},{"a":"c"}]}`
	invalidArgMessage := "ERR wrong number of arguments for 'json.nummultby' command"

	testCases := []struct {
		name        string
		commands    []string
		expected   []interface{}
		assertType []string
	}{
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
			commands:   []string{"JSON.SET docu $ " + a, "JSON.NUMMULTBY docu $.fe x"},
			expected:   []interface{}{"OK", "[]"},
			assertType: []string{"equal", "equal"},
		},
		{
			name:       "Invalid value of multiplier on existent key",
			commands:   []string{"JSON.SET docu $ " + a, "JSON.NUMMULTBY docu $.a x"},
			expected:   []interface{}{"OK", "ERR expected value at line 1 column 1"},
			assertType: []string{"equal", "equal"},
		},
		{
			name:       "MultBy at recursive path",
			commands:   []string{"JSON.SET docu $ " + a, "JSON.NUMMULTBY docu $..a 2"},
			expected:   []interface{}{"OK", "[4,10,null,null]"},
			assertType: []string{"equal", "deep_equal"},
		},
		{
			name:       "MultBy at root path",
			commands:   []string{"JSON.SET docu $ " + a, "JSON.NUMMULTBY docu $.a 2"},
			expected:   []interface{}{"OK", "[null]"},
			assertType: []string{"equal", "deep_equal"},
		},
	}

	for _, tcase := range testCases {
		FireCommand(conn, "DEL docu")
		t.Run(tcase.name, func(t *testing.T) {
			for i := 0; i < len(tcase.commands); i++ {
				cmd := tcase.commands[i]
				out := tcase.expected[i]
				result := FireCommand(conn, cmd)
				if tcase.assertType[i] == "equal" {
					assert.Equal(t, out, result)
				} else if tcase.assertType[i] == "deep_equal" {
					assert.Assert(t, arraysArePermutations(deStringify(out.(string)), deStringify(result.(string))))
				}
			}
		})
	}
}

func TestJsonObjLen(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	a := `{"name":"jerry","partner":{"name":"tom","language":["rust"]}}`
	b := `{"name":"jerry","partner":{"name":"tom","language":["rust"]},"partner2":{"name":"spike","language":["go","rust"]}}`
	c := `{"name":"jerry","partner":{"name":"tom","language":["rust"]},"partner2":{"name":12,"language":["rust"]}}`
	d := `["this","is","an","array"]`

	defer func() {
		resp := FireCommand(conn, "DEL obj")
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
			expected: []interface{}{"OK", "ERR parse error at 13 in $..language*something"},
		},
		{
			name:     "JSON.OBJLEN with non-existant key",
			commands: []string{"json.set obj $ " + b, "json.objlen non_existing_key $"},
			expected: []interface{}{"OK", "(nil)"},
		},
		{
			name:     "JSON.OBJLEN with empty path",
			commands: []string{"json.set obj $ " + a, "json.objlen obj"},
			expected: []interface{}{"OK", int64(2)},
		},
		{
			name:     "JSON.OBJLEN invalid json path",
			commands: []string{"json.set obj $ " + c, "json.objlen obj $[1"},
			expected: []interface{}{"OK", "ERR expected a number at 4 in $[1"},
		},
		{
			name:     "JSON.OBJLEN invalid json path",
			commands: []string{"json.set obj $ " + c, "json.objlen"},
			expected: []interface{}{"OK", "ERR wrong number of arguments for 'json.objlen' command"},
		},
	}

	for _, tcase := range testCases {
		FireCommand(conn, "DEL obj")
		t.Run(tcase.name, func(t *testing.T) {
			for i := 0; i < len(tcase.commands); i++ {
				cmd := tcase.commands[i]
				out := tcase.expected[i]
				result := FireCommand(conn, cmd)
				assert.DeepEqual(t, out, result)
			}
		})
	}
}

func convertToArray(input string) []string {
	input = strings.Trim(input, `"[`)
	input = strings.Trim(input, `]"`)
	elements := strings.Split(input, ",")
	for i, element := range elements {
		elements[i] = strings.TrimSpace(element)
	}
	return elements
}

func TestJSONNumIncrBy(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()
	invalidArgMessage := "ERR wrong number of arguments for 'json.numincrby' command"
	testCases := []struct {
		name        string
		setupData   string
		commands    []string
		expected   []interface{}
		assertType []string
		cleanUp    []string
	}{
		{
			name:       "Invalid number of arguments",
			setupData:  "",
			commands:   []string{"JSON.NUMINCRBY ", "JSON.NUMINCRBY foo", "JSON.NUMINCRBY foo $"},
			expected:   []interface{}{invalidArgMessage, invalidArgMessage, invalidArgMessage},
			assertType: []string{"equal", "equal", "equal"},
			cleanUp:    []string{},
		},
		{
			name:       "Non-existant key",
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

	for _, tc := range testCases {
		FireCommand(conn, "DEL foo")
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupData != "" {
				assert.Equal(t, FireCommand(conn, tc.setupData), "OK")
			}
			for i := 0; i < len(tc.commands); i++ {
				cmd := tc.commands[i]
				out := tc.expected[i]
				result := FireCommand(conn, cmd)
				switch tc.assertType[i] {
				case "equal":
					assert.Equal(t, out, result)
				case "perm_equal":
					assert.Assert(t, arraysArePermutations(convertToArray(out.(string)), convertToArray(result.(string))))
				case "range":
					assert.Assert(t, result.(int64) <= tc.expected[i].(int64) && result.(int64) > 0, "Expected %v to be within 0 to %v", result, tc.expected[i])
				case "json_equal":
					testifyAssert.JSONEq(t, out.(string), result.(string))
				}
			}
			for i := 0; i < len(tc.cleanUp); i++ {
				FireCommand(conn, tc.cleanUp[i])
			}
		})
	}
}

func TestJsonARRINSERT(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()
	a := `[1,2]`
	b := `{"name":"tom","score":[10,20],"partner2":{"score":[10,20]}}`

	testCases := []struct {
		name       string
		commands   []string
		expected   []interface{}
		assertType []string
	}{
		{
			name:       "JSON.ARRINSERT index out if bounds",
			commands:   []string{"json.set a $ " + a, `JSON.ARRINSERT a $ 4 3`, "JSON.GET a"},
			expected:   []interface{}{"OK", "ERR index out of bounds", "[1,2]"},
			assertType: []string{"equal", "equal", "equal"},
		},
		{
			name:       "JSON.ARRINSERT index is not integer",
			commands:   []string{"json.set a $ " + a, `JSON.ARRINSERT a $ ss 3`, "JSON.GET a"},
			expected:   []interface{}{"OK", "ERR Couldn't parse as integer", "[1,2]"},
			assertType: []string{"equal", "equal", "equal"},
		},
		{
			name:       "JSON.ARRINSERT with postive index in root path",
			commands:   []string{"json.set a $ " + a, `JSON.ARRINSERT a $ 2 3 4 5`, "JSON.GET a"},
			expected:   []interface{}{"OK", []interface{}{int64(5)}, "[1,2,3,4,5]"},
			assertType: []string{"equal", "deep_equal", "equal"},
		},

		{
			name:       "JSON.ARRINSERT with postive index in root path",
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
			name:       "JSON.ARRINSERT nested with postive index",
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
				result := FireCommand(conn, cmd)
				if tcase.assertType[i] == "equal" {
					assert.Equal(t, out, result)
				} else if tcase.assertType[i] == "deep_equal" {
					assert.Assert(t, arraysArePermutations(out.([]interface{}), result.([]interface{})))
				} else if tcase.assertType[i] == "jsoneq" {
					testifyAssert.JSONEq(t, out.(string), result.(string))
				}
			}
		})
	}
}

func TestJsonObjKeys(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()
	a := `{"name":"jerry","partner":{"name":"tom","language":["rust"]},"partner2":{"language":["rust"]}}`
	b := `{"name":"jerry","partner":{"name":"tom","language":["rust"]},"partner2":{"name":12,"language":["rust"]}}`
	c := `{"name":"jerry","partner":{"name":"tom","language":["rust"]},"partner2":{"name":12,"language":["rust"],"extra_key":"value"}}`
	d := `{"a":[3],"nested":{"a":{"b":2,"c":1}}}`

	testCases := []struct {
		name        string
		setCommand  string
		testCommand string
		expected    []interface{}
	}{
		{
			name:        "JSON.OBJKEYS root object",
			setCommand:  "json.set doc $ " + a,
			testCommand: "json.objkeys doc $",
			expected: []interface{}{
				[]interface{}{"name", "partner", "partner2"},
			},
		},
		{
			name:        "JSON.OBJKEYS with nested path",
			setCommand:  "json.set doc $ " + b,
			testCommand: "json.objkeys doc $.partner",
			expected: []interface{}{
				[]interface{}{"name", "language"},
			},
		},
		{
			name:        "JSON.OBJKEYS with non-object path",
			setCommand:  "json.set doc $ " + c,
			testCommand: "json.objkeys doc $.name",
			expected: []interface{}{
				"(nil)",
			},
		},
		{
			name:        "JSON.OBJKEYS with nested non-object path",
			setCommand:  "json.set doc $ " + b,
			testCommand: "json.objkeys doc $.partner.language",
			expected: []interface{}{
				"(nil)",
			},
		},
		{
			name:        "JSON.OBJKEYS with invalid json path - 1",
			setCommand:  "json.set doc $ " + b,
			testCommand: "json.objkeys doc $..invalidpath*somethingrandomadded",
			expected:    []interface{}{"ERR parse error at 16 in $..invalidpath*somethingrandomadded"},
		},
		{
			name:        "JSON.OBJKEYS with invalid json path - 2",
			setCommand:  "json.set doc $ " + c,
			testCommand: "json.objkeys doc $[1",
			expected:    []interface{}{"ERR expected a number at 4 in $[1"},
		},
		{
			name:        "JSON.OBJKEYS with invalid json path - 3",
			setCommand:  "json.set doc $ " + c,
			testCommand: "json.objkeys doc $[random",
			expected:    []interface{}{"ERR parse error at 3 in $[random"},
		},
		{
			name:        "JSON.OBJKEYS with only command",
			setCommand:  "json.set doc $ " + c,
			testCommand: "json.objkeys",
			expected:    []interface{}{"ERR wrong number of arguments for 'json.objkeys' command"},
		},
		{
			name:        "JSON.OBJKEYS with non-existing key",
			setCommand:  "json.set doc $ " + c,
			testCommand: "json.objkeys thisdoesnotexist $",
			expected:    []interface{}{"ERR could not perform this operation on a key that doesn't exist"},
		},
		{
			name:        "JSON.OBJKEYS with empty path",
			setCommand:  "json.set doc $ " + c,
			testCommand: "json.objkeys doc",
			expected: []interface{}{
				"name", "partner", "partner2",
			},
		},
		{
			name:        "JSON.OBJKEYS with multiple json path",
			setCommand:  "json.set doc $ " + d,
			testCommand: "json.objkeys doc $..a",
			expected: []interface{}{
				[]interface{}{"b", "c"},
				"(nil)",
			},
		},
	}

	for _, tc := range testCases {
		FireCommand(conn, "DEL doc")
		t.Run(tc.name, func(t *testing.T) {
			FireCommand(conn, tc.setCommand)
			expected := tc.expected
			out := FireCommand(conn, tc.testCommand)

			_, isString := out.(string)
			if isString {
				outInterface := []interface{}{out}
				assert.DeepEqual(t, outInterface, expected)
			} else {
				assert.DeepEqual(t, out.([]interface{}), expected,
					cmpopts.SortSlices(func(a, b interface{}) bool {
						return fmt.Sprintf("%v", a) < fmt.Sprintf("%v", b)
					}))
			}
		})
	}

}

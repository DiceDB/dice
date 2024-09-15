package commands

import (
	"fmt"
	"strings"
	"testing"

	"github.com/dicedb/dice/internal/server/utils"
	"github.com/dicedb/dice/testutils"
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

	for _, tc := range singleOrderedTestCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setCmd != utils.EmptyStr {
				result := FireCommand(conn, tc.setCmd)
				assert.Equal(t, "OK", result)
			}

			if tc.getCmd != utils.EmptyStr {
				result := FireCommand(conn, tc.getCmd)
				if testutils.IsJSONResponse(result.(string)) {
					testutils.AssertJSONEqual(t, tc.expected, result.(string))
				} else {
					assert.Equal(t, tc.expected, result)
				}
			}
		})
	}

	for _, tc := range multipleOrderedTestCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setCmd != utils.EmptyStr {
				result := FireCommand(conn, tc.setCmd)
				assert.Equal(t, "OK", result)
			}

			if tc.getCmd != utils.EmptyStr {
				result := FireCommand(conn, tc.getCmd)
				testutils.AssertJSONEqualList(t, tc.expected, result.(string))
			}
		})
	}
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

	// deleteTestKeys([]string{"user"}, store)
	FireCommand(conn, "DEL user")

	user1 := `{"name":"John","age":30}`
	user2 := `{"name":"Rahul","age":28}`

	testCases := []struct {
		commands []string
		expected []interface{}
	}{
		{
			commands: []string{"JSON.SET user $ " + user1 + " XX", "JSON.SET user $ " + user1 + " NX", "JSON.GET user"},
			expected: []interface{}{"(nil)", "OK", user1},
		},
		{
			commands: []string{"DEL user", "JSON.SET user $ " + user1, "JSON.SET user $ " + user1 + " NX"},
			expected: []interface{}{int64(1), "OK", "(nil)"},
		},
		{
			commands: []string{"JSON.SET user $ " + user2 + " XX", "JSON.GET user", "DEL user"},
			expected: []interface{}{"OK", user2, int64(1)},
		},
		{
			commands: []string{"JSON.SET user $ " + user2 + " NX", "JSON.SET user $ " + user1 + " NX", "JSON.GET user"},
			expected: []interface{}{"OK", "(nil)", user2},
		},
		{
			commands: []string{"JSON.SET user $ " + user2 + " NX NX", "JSON.SET user $ " + user2 + " NX XX", "JSON.SET user $ " + user2 + " NX Abcd", "JSON.SET user $ " + user2 + " NX "},
			expected: []interface{}{"ERR syntax error", "ERR syntax error", "ERR syntax error", "(nil)"},
		},
		{
			commands: []string{"JSON.SET user $ " + user2 + " XX XX", "JSON.SET user $ " + user2 + " XX NX", "JSON.SET user $ " + user2 + " XX Abcd", "JSON.SET user $ " + user2 + " XX "},
			expected: []interface{}{"ERR syntax error", "ERR syntax error", "ERR syntax error", "OK"},
		},
	}

	for _, tcase := range testCases {
		for i := 0; i < len(tcase.commands); i++ {
			cmd := tcase.commands[i]
			out := tcase.expected[i]
			result := FireCommand(conn, cmd)
			jsonResult, isString := result.(string)
			if isString && testutils.IsJSONResponse(jsonResult) {
				testutils.AssertJSONEqual(t, out.(string), jsonResult)
			} else {
				assert.Equal(t, out, result)
			}
		}
	}
}

func TestJSONClearOperations(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	// deleteTestKeys([]string{"user"}, store)
	FireCommand(conn, "DEL user")

	stringClearTestJson := `{"flag":true,"name":"Tom"}`
	booleanClearTestJson := `{"flag":true,"name":"Tom"}`
	arrayClearTestJson := `{"names":["Rahul","Tom"],"bosses":{"names":["Jerry","Rocky"]}}`
	integerClearTestJson := `{"age":28,"name":"Tom"}`
	floatClearTestJson := `{"price":3.14,"name":"sugar"}`
	nullClearTestJson := `{"name":null,"age":28}`
	multiClearTestJson := `{"age":13,"high":1.60,"flag":true,"name":"jerry","pet":null,"language":["python","golang"],"partner":{"name":"tom","language":["rust"]}}`

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name:     "clear root path",
			commands: []string{"JSON.SET user $ " + multiClearTestJson, "JSON.CLEAR user $", "JSON.GET user $"},
			expected: []interface{}{"OK", int64(1), "{}"},
		},
		{
			name:     "clear string type",
			commands: []string{"JSON.SET user $ " + stringClearTestJson, "JSON.CLEAR user $.name", "JSON.GET user $.name"},
			expected: []interface{}{"OK", int64(0), `"Tom"`},
		},
		{
			name:     "clear bool type",
			commands: []string{"JSON.SET user $ " + booleanClearTestJson, "JSON.CLEAR user $.flag", "JSON.GET user $.flag"},
			expected: []interface{}{"OK", int64(0), "true"},
		},
		{
			name:     "clear null type",
			commands: []string{"JSON.SET user $ " + nullClearTestJson, "JSON.CLEAR user $.pet", "JSON.GET user $.name"},
			expected: []interface{}{"OK", int64(0), "null"},
		},
		{
			name:     "clear array type",
			commands: []string{"JSON.SET user $ " + arrayClearTestJson, "JSON.CLEAR user $..names", "JSON.GET user $..names"},
			expected: []interface{}{"OK", int64(2), "[[],[]]"},
		},
		{
			name:     "clear integer type",
			commands: []string{"JSON.SET user $ " + integerClearTestJson, "JSON.CLEAR user $.age", "JSON.GET user $.age"},
			expected: []interface{}{"OK", int64(1), "0"},
		},
		{
			name:     "clear float type",
			commands: []string{"JSON.SET user $ " + floatClearTestJson, "JSON.CLEAR user $.price", "JSON.GET user $.price"},
			expected: []interface{}{"OK", int64(1), "0"},
		},
	}

	for _, tcase := range testCases {
		t.Run(tcase.name, func(t *testing.T) {
			for i := 0; i < len(tcase.commands); i++ {
				cmd := tcase.commands[i]
				out := tcase.expected[i]
				result := FireCommand(conn, cmd)
				assert.Equal(t, out, result)

			}
		})
	}
}

func TestJSONDelOperations(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	FireCommand(conn, "DEL user")

	stringDelTestJson := `{"flag":true,"name":"Tom"}`
	booleanDelTestJson := `{"flag":true,"name":"Tom"}`
	arrayDelTestJson := `{"names":["Rahul","Tom"],"bosses":{"names":["Jerry","Rocky"],"hobby":"swim"}}`
	integerDelTestJson := `{"age":28,"name":"Tom"}`
	floatDelTestJson := `{"price":3.14,"name":"sugar"}`
	nullDelTestJson := `{"name":null,"age":28}`
	multiDelTestJson := `{"age":13,"high":1.60,"flag":true,"name":"jerry","pet":null,"language":["python","golang"],"partner":{"name":"tom","language":["rust"]}}`

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name:     "del root path",
			commands: []string{"JSON.SET user $ " + multiDelTestJson, "JSON.DEL user $", "JSON.GET user $"},
			expected: []interface{}{"OK", int64(1), "(nil)"},
		},
		{
			name:     "del string type",
			commands: []string{"JSON.SET user $ " + stringDelTestJson, "JSON.DEL user $.name", "JSON.GET user $"},
			expected: []interface{}{"OK", int64(1), `{"flag":true}`},
		},
		{
			name:     "del bool type",
			commands: []string{"JSON.SET user $ " + booleanDelTestJson, "JSON.DEL user $.flag", "JSON.GET user $"},
			expected: []interface{}{"OK", int64(1), `{"name":"Tom"}`},
		},
		{
			name:     "del null type",
			commands: []string{"JSON.SET user $ " + nullDelTestJson, "JSON.DEL user $.name", "JSON.GET user $"},
			expected: []interface{}{"OK", int64(1), `{"age":28}`},
		},
		{
			name:     "del array type",
			commands: []string{"JSON.SET user $ " + arrayDelTestJson, "JSON.DEL user $..names", "JSON.GET user $"},
			expected: []interface{}{"OK", int64(2), `{"bosses":{"hobby":"swim"}}`},
		},
		{
			name:     "del integer type",
			commands: []string{"JSON.SET user $ " + integerDelTestJson, "JSON.DEL user $.age", "JSON.GET user $"},
			expected: []interface{}{"OK", int64(1), `{"name":"Tom"}`},
		},
		{
			name:     "del float type",
			commands: []string{"JSON.SET user $ " + floatDelTestJson, "JSON.DEL user $.price", "JSON.GET user $"},
			expected: []interface{}{"OK", int64(1), `{"name":"sugar"}`},
		},
	}

	for _, tcase := range testCases {
		t.Run(tcase.name, func(t *testing.T) {
			for i := 0; i < len(tcase.commands); i++ {
				cmd := tcase.commands[i]
				out := tcase.expected[i]
				result := FireCommand(conn, cmd)
				jsonResult, isString := result.(string)
				if isString && testutils.IsJSONResponse(jsonResult) {
					testutils.AssertJSONEqual(t, out.(string), jsonResult)
				} else {
					assert.Equal(t, out, result)
				}

			}
		})
	}
}

func TestJSONForgetOperations(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	FireCommand(conn, "FORGET user")

	stringForgetTestJson := `{"flag":true,"name":"Tom"}`
	booleanForgetTestJson := `{"flag":true,"name":"Tom"}`
	arrayForgetTestJson := `{"names":["Rahul","Tom"],"bosses":{"names":["Jerry","Rocky"],"hobby":"swim"}}`
	integerForgetTestJson := `{"age":28,"name":"Tom"}`
	floatForgetTestJson := `{"price":3.14,"name":"sugar"}`
	nullForgetTestJson := `{"name":null,"age":28}`
	multiForgetTestJson := `{"age":13,"high":1.60,"flag":true,"name":"jerry","pet":null,"language":["python","golang"],"partner":{"name":"tom","language":["rust"]}}`

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name:     "forget root path",
			commands: []string{"JSON.SET user $ " + multiForgetTestJson, "JSON.FORGET user $", "JSON.GET user $"},
			expected: []interface{}{"OK", int64(1), "(nil)"},
		},
		{
			name:     "forget string type",
			commands: []string{"JSON.SET user $ " + stringForgetTestJson, "JSON.FORGET user $.name", "JSON.GET user $"},
			expected: []interface{}{"OK", int64(1), `{"flag":true}`},
		},
		{
			name:     "forget bool type",
			commands: []string{"JSON.SET user $ " + booleanForgetTestJson, "JSON.FORGET user $.flag", "JSON.GET user $"},
			expected: []interface{}{"OK", int64(1), `{"name":"Tom"}`},
		},
		{
			name:     "forget null type",
			commands: []string{"JSON.SET user $ " + nullForgetTestJson, "JSON.FORGET user $.name", "JSON.GET user $"},
			expected: []interface{}{"OK", int64(1), `{"age":28}`},
		},
		{
			name:     "forget array type",
			commands: []string{"JSON.SET user $ " + arrayForgetTestJson, "JSON.FORGET user $..names", "JSON.GET user $"},
			expected: []interface{}{"OK", int64(2), `{"bosses":{"hobby":"swim"}}`},
		},
		{
			name:     "forget integer type",
			commands: []string{"JSON.SET user $ " + integerForgetTestJson, "JSON.FORGET user $.age", "JSON.GET user $"},
			expected: []interface{}{"OK", int64(1), `{"name":"Tom"}`},
		},
		{
			name:     "forget float type",
			commands: []string{"JSON.SET user $ " + floatForgetTestJson, "JSON.FORGET user $.price", "JSON.GET user $"},
			expected: []interface{}{"OK", int64(1), `{"name":"sugar"}`},
		},
	}

	for _, tcase := range testCases {
		t.Run(tcase.name, func(t *testing.T) {
			for i := 0; i < len(tcase.commands); i++ {
				cmd := tcase.commands[i]
				out := tcase.expected[i]
				result := FireCommand(conn, cmd)
				jsonResult, isString := result.(string)
				if isString && testutils.IsJSONResponse(jsonResult) {
					testutils.AssertJSONEqual(t, out.(string), jsonResult)
				} else {
					assert.Equal(t, out, result)
				}

			}
		})
	}
}

func arraysArePermutations(a, b []interface{}) bool {
	// If lengths are different, they cannot be permutations
	if len(a) != len(b) {
		return false
	}

	// Count occurrences of each element in array 'a'
	countA := make(map[interface{}]int)
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
	a := `["hehhhe","hello"]`
	b := `{"name":"jerry","partner":{"name":"tom","language":["rust"]},"partner2":{"language":["rust"]}}`
	c := `{"name":"jerry","partner":{"name":"tom","language":["rust"]},"partner2":{"name":12,"language":["rust"]}}`

	testCases := []struct {
		name        string
		commands    []string
		expected    []interface{}
		assert_type []string
	}{

		{
			name:        "JSON.STRLEN with root path",
			commands:    []string{"json.set a $ " + a, "json.strlen a $"},
			expected:    []interface{}{"OK", []interface{}{"(nil)"}},
			assert_type: []string{"equal", "deep_equal"},
		},
		{
			name:        "JSON.STRLEN nested",
			commands:    []string{"JSON.SET doc $ " + b, "JSON.STRLEN doc $..name"},
			expected:    []interface{}{"OK", []interface{}{int64(3), int64(5)}},
			assert_type: []string{"equal", "deep_equal"},
		},
		{
			name:        "JSON.STRLEN nested with nil",
			commands:    []string{"JSON.SET doc $ " + c, "JSON.STRLEN doc $..name"},
			expected:    []interface{}{"OK", []interface{}{int64(3), int64(5), "(nil)"}},
			assert_type: []string{"equal", "deep_equal"},
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
				if tcase.assert_type[i] == "equal" {
					assert.Equal(t, out, result)
				} else if tcase.assert_type[i] == "deep_equal" {
					assert.Assert(t, arraysArePermutations(out.([]interface{}), result.([]interface{})))
				}
			}
		})
	}
}

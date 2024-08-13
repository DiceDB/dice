package tests

import (
	"fmt"
	"strings"
	"testing"

	"github.com/dicedb/dice/internal/constants"
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

	testCases := []struct {
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
			expected: "ERR wrong number of arguments for 'JSON.GET' command",
		},
		{
			name:     "Set Non-JSON Value",
			setCmd:   `SET nonJson "not a json"`,
			getCmd:   `JSON.GET nonJson`,
			expected: "the operation is not permitted on this type",
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
			name:     "Get All Prices",
			setCmd:   `JSON.SET inventory $ ` + complexJSON,
			getCmd:   `JSON.GET inventory $..price`,
			expected: `[1475,3941,1920,2072,3264]`,
		},
		{
			name:     "Set Nested Value",
			setCmd:   `JSON.SET inventory $.inventory.mountain_bikes[0].price 2000`,
			getCmd:   `JSON.GET inventory $.inventory.mountain_bikes[0].price`,
			expected: `2000`,
		},
		{
			name:     "Set Multiple Nested Values",
			setCmd:   `JSON.SET inventory $.inventory.*[?(@.price<2000)].price 1500`,
			getCmd:   `JSON.GET inventory $..price`,
			expected: `[1500,3941,2000,2072,3264]`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setCmd != constants.EmptyStr {
				result := fireCommand(conn, tc.setCmd)
				assert.Equal(t, "OK", result)
			}

			if tc.getCmd != constants.EmptyStr {
				result := fireCommand(conn, tc.getCmd)
				if testutils.IsJSONResponse(result.(string)) {
					testutils.AssertJSONEqual(t, tc.expected, result.(string))
				} else {
					assert.Equal(t, tc.expected, result)
				}
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
			expected: "ERR wrong number of arguments for 'JSON.SET' command",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := fireCommand(conn, tc.command)
			assert.Check(t, strings.HasPrefix(result.(string), tc.expected), fmt.Sprintf("Expected: %s, Got: %s", tc.expected, result))
		})
	}
}

func TestUnsupportedJSONPathPatterns(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()
	complexJSON := `{"inventory":{"mountain_bikes":[{"id":"bike:1","model":"Phoebe","price":1920,"specs":{"material":"carbon","weight":13.1},"colors":["black","silver"]},{"id":"bike:2","model":"Quaoar","price":2072,"specs":{"material":"aluminium","weight":7.9},"colors":["black","white"]},{"id":"bike:3","model":"Weywot","price":3264,"specs":{"material":"alloy","weight":13.8}}],"commuter_bikes":[{"id":"bike:4","model":"Salacia","price":1475,"specs":{"material":"aluminium","weight":16.6},"colors":["black","silver"]},{"id":"bike:5","model":"Mimas","price":3941,"specs":{"material":"alloy","weight":11.6}}]}}`

	setupCmd := `JSON.SET bikes:inventory $ ` + complexJSON
	result := fireCommand(conn, setupCmd)
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
			result := fireCommand(conn, tc.command)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestJSONSetWithNXAndXX(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	deleteTestKeys([]string{"user"})

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
			result := fireCommand(conn, cmd)
			jsonResult, isString := result.(string)
			if isString && testutils.IsJSONResponse(jsonResult) {
				testutils.AssertJSONEqual(t, out.(string), jsonResult)
			} else {
				assert.Equal(t, out, result)
			}
		}
	}
}

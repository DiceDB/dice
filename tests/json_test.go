package tests

import (
	"testing"

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
			name:     "Set Invalid JSON",
			setCmd:   `JSON.SET invalid $ {invalid:json}`,
			getCmd:   ``,
			expected: "ERR invalid JSON: invalid character 'i' looking for beginning of object key string",
		},
		{
			name:     "Set JSON with Wrong Number of Arguments",
			setCmd:   `JSON.SET`,
			getCmd:   ``,
			expected: "ERR wrong number of arguments for 'JSON.SET' command",
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
			if tc.setCmd != "" {
				result := fireCommand(conn, tc.setCmd)
				if tc.name != "Set Invalid JSON" && tc.name != "Set JSON with Wrong Number of Arguments" {
					assert.Equal(t, "OK", result)
				} else {
					assert.Equal(t, tc.expected, result)
				}
			}

			if tc.getCmd != "" {
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

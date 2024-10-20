package resp

import (
	"gotest.tools/v3/assert"
	"testing"
)

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
				result := FireCommand(conn, cmd)
				assert.Equal(t, tc.expected[i], result)
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
			expected: []interface{}{"OK", "ERR Path '$..language*something' does not exist"},
		},
		{
			name:     "JSON.OBJLEN with non-existent key",
			commands: []string{"json.set obj $ " + b, "json.objlen non_existing_key $"},
			expected: []interface{}{"OK", "(nil)"},
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
			commands: []string{"json.set obj $ " + c, "json.objlen obj .partner", "json.objlen obj .partner2",},
			expected: []interface{}{"OK", int64(2), int64(2)},
		},
		{
			name:     "JSON.OBJLEN with legacy path - inner existing path v2",
			commands: []string{"json.set obj $ " + c, "json.objlen obj partner", "json.objlen obj partner2",},
			expected: []interface{}{"OK", int64(2), int64(2)},
		},
		{
			name:     "JSON.OBJLEN with legacy path - inner non-existent path",
			commands: []string{"json.set obj $ " + c, "json.objlen obj .idonotexist",},
			expected: []interface{}{"OK", "(nil)"},
		},
		{
			name:     "JSON.OBJLEN with legacy path - inner non-existent path v2",
			commands: []string{"json.set obj $ " + c, "json.objlen obj idonotexist",},
			expected: []interface{}{"OK", "(nil)"},
		},
		{
			name:     "JSON.OBJLEN with legacy path - inner existent path with nonJSON object",
			commands: []string{"json.set obj $ " + c, "json.objlen obj .name",},
			expected: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
		{
			name:     "JSON.OBJLEN with legacy path - inner existent path recursive object",
			commands: []string{"json.set obj $ " + c, "json.objlen obj ..partner",},
			expected: []interface{}{"OK", int64(2)},
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

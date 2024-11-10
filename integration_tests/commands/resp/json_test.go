package resp

import (
	"fmt"
	"sort"
	"testing"

	"github.com/dicedb/dice/testutils"
	"github.com/stretchr/testify/assert"
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
					assert.True(t, arraysArePermutations(tc.expected[i].([]interface{}), result.([]interface{})))
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
		FireCommand(conn, "DEL obj")
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

func TestJSONARRPOP(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()
	FireCommand(conn, "DEL key")

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
				result := FireCommand(conn, cmd)

				jsonResult, isString := result.(string)

				if isString && testutils.IsJSONResponse(jsonResult) {
					assert.JSONEq(t, out.(string), jsonResult)
					continue
				}

				if tcase.assertType[i] == "equal" {
					assert.Equal(t, out, result)
				} else if tcase.assertType[i] == "deep_equal" {
					assert.True(t, arraysArePermutations(out.([]interface{}), result.([]interface{})))
				}
			}
		})
	}
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
					assert.True(t, arraysArePermutations(out.([]interface{}), result.([]interface{})))
				}
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
				result := FireCommand(conn, cmd)
				if tcase.assertType[i] == "equal" {
					assert.Equal(t, out, result)
				} else if tcase.assertType[i] == "deep_equal" {
					assert.True(t, arraysArePermutations(out.([]interface{}), result.([]interface{})))
				} else if tcase.assertType[i] == "jsoneq" {
					assert.JSONEq(t, out.(string), result.(string))
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
				assert.Equal(t, outInterface, expected)
			} else {
						assert.ElementsMatch(t, 
                sortNestedSlices(expected), 
                sortNestedSlices(out.([]interface{})),
                "Mismatch in JSON object keys")
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
	conn := getLocalConnection()
	defer conn.Close()
	a := `[0,1,2]`
	b := `{"connection":{"wireless":true,"names":[0,1,2,3,4]},"names":[0,1,2,3,4]}`

	FireCommand(conn, "DEL a b")
	defer FireCommand(conn, "DEL a b")

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
				result := FireCommand(conn, cmd)
				if tcase.assertType[i] == "equal" {
					assert.Equal(t, out, result)
				} else if tcase.assertType[i] == "deep_equal" {
					assert.True(t, arraysArePermutations(out.([]interface{}), result.([]interface{})))
				} else if tcase.assertType[i] == "jsoneq" {
					assert.JSONEq(t, out.(string), result.(string))
				}
			}
		})
	}
}

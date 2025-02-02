// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package websocket

import (
	"fmt"
	"sort"
	"testing"

	"github.com/dicedb/dice/testutils"
	"github.com/gorilla/websocket"
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

func runIntegrationTests(t *testing.T, exec *WebsocketCommandExecutor, conn *websocket.Conn, testCases []IntegrationTestCase, preTestChecksCommand string, postTestChecksCommand string) {
	for _, tc := range testCases {
		if preTestChecksCommand != "" {
			resp, _ := exec.FireCommandAndReadResponse(conn, preTestChecksCommand)
			assert.Equal(t, float64(0), resp)
		}

		t.Run(tc.name, func(t *testing.T) {
			if tc.setupData != "" {
				result, _ := exec.FireCommandAndReadResponse(conn, tc.setupData)
				assert.Equal(t, "OK", result)
			}

			cleanupAndPostTestChecks := func() {
				for _, cmd := range tc.cleanUp {
					exec.FireCommandAndReadResponse(conn, cmd)
				}

				if postTestChecksCommand != "" {
					resp, _ := exec.FireCommandAndReadResponse(conn, postTestChecksCommand)
					assert.Equal(t, float64(0), resp)
				}
			}
			defer cleanupAndPostTestChecks()

			for i := 0; i < len(tc.commands); i++ {
				cmd := tc.commands[i]
				out := tc.expected[i]
				result, _ := exec.FireCommandAndReadResponse(conn, cmd)

				switch tc.assertType[i] {
				case "equal":
					assert.Equal(t, out, result)
				case "perm_equal":
					assert.True(t, testutils.ArraysArePermutations(testutils.ConvertToArray(out.(string)), testutils.ConvertToArray(result.(string))))
				case "range":
					assert.True(t, result.(float64) <= out.(float64) && result.(float64) > 0, "Expected %v to be within 0 to %v", result, out)
				case "json_equal":
					assert.JSONEq(t, out.(string), result.(string))
				case "deep_equal":
					assert.ElementsMatch(t, result.([]interface{}), out.([]interface{}))
				}
			}
		})
	}
}

func TestJSONClearOperations(t *testing.T) {
	exec := NewWebsocketCommandExecutor()
	conn := exec.ConnectToServer()
	defer conn.Close()

	DeleteKey(t, conn, exec, "user")

	defer func() {
		resp, err := exec.FireCommandAndReadResponse(conn, "DEL user")
		assert.Nil(t, err)
		assert.Equal(t, float64(1), resp)
	}()

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
			expected: []interface{}{"OK", float64(1), "{}"},
		},
		{
			name: "jsonclear string type",
			commands: []string{
				`JSON.SET user $ {"name":"Tom","age":30}`,
				"JSON.CLEAR user $.name",
				"JSON.GET user $.name",
			},
			expected: []interface{}{"OK", float64(0), `"Tom"`},
		},
		{
			name: "jsonclear array type",
			commands: []string{
				`JSON.SET user $ {"names":["Rahul","Tom"],"ages":[25,30]}`,
				"JSON.CLEAR user $.names",
				"JSON.GET user $.names",
			},
			expected: []interface{}{"OK", float64(1), "[]"},
		},
		{
			name: "jsonclear bool type",
			commands: []string{
				`JSON.SET user $ {"flag":true,"name":"Tom"}`,
				"JSON.CLEAR user $.flag",
				"JSON.GET user $.flag"},
			expected: []interface{}{"OK", float64(0), "true"},
		},
		{
			name: "jsonclear null type",
			commands: []string{
				`JSON.SET user $ {"name":null,"age":28}`,
				"JSON.CLEAR user $.pet",
				"JSON.GET user $.name"},
			expected: []interface{}{"OK", float64(0), "null"},
		},
		{
			name: "jsonclear integer type",
			commands: []string{
				`JSON.SET user $ {"age":28,"name":"Tom"}`,
				"JSON.CLEAR user $.age",
				"JSON.GET user $.age"},
			expected: []interface{}{"OK", float64(1), "0"},
		},
		{
			name: "jsonclear float64 type",
			commands: []string{
				`JSON.SET user $ {"price":3.14,"name":"sugar"}`,
				"JSON.CLEAR user $.price",
				"JSON.GET user $.price"},
			expected: []interface{}{"OK", float64(1), "0"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result, err := exec.FireCommandAndReadResponse(conn, cmd)
				assert.Nil(t, err)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}
}

func TestJsonStrlen(t *testing.T) {
	exec := NewWebsocketCommandExecutor()
	conn := exec.ConnectToServer()

	defer conn.Close()

	DeleteKey(t, conn, exec, "doc")

	defer func() {
		resp, err := exec.FireCommandAndReadResponse(conn, "DEL doc")
		assert.Nil(t, err)
		assert.Equal(t, float64(1), resp)
	}()

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
			expected: []interface{}{"OK", []interface{}{nil}},
		},
		{
			name: "jsonstrlen nested",
			commands: []string{
				`JSON.SET doc $ {"name":"jerry","partner":{"name":"tom"}}`,
				"JSON.STRLEN doc $..name",
			},
			expected: []interface{}{"OK", []interface{}{float64(5), float64(3)}},
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
				result, err := exec.FireCommandAndReadResponse(conn, cmd)
				assert.Nil(t, err, "error: %v", err)
				stringResult, ok := result.(string)
				if ok {
					assert.Equal(t, tc.expected[i], stringResult)
				} else {
					assert.True(t, testutils.ArraysArePermutations(tc.expected[i].([]interface{}), result.([]interface{})))
				}
			}
		})
	}
}

func TestJsonObjLen(t *testing.T) {
	exec := NewWebsocketCommandExecutor()
	conn := exec.ConnectToServer()
	defer conn.Close()

	DeleteKey(t, conn, exec, "obj")

	a := `{"name":"jerry","partner":{"name":"tom","language":["rust"]}}`
	b := `{"name":"jerry","partner":{"name":"tom","language":["rust"]},"partner2":{"name":"spike","language":["go","rust"]}}`
	c := `{"name":"jerry","partner":{"name":"tom","language":["rust"]},"partner2":{"name":12,"language":["rust"]}}`
	d := `["this","is","an","array"]`

	defer func() {
		resp, err := exec.FireCommandAndReadResponse(conn, "DEL obj")
		assert.Nil(t, err)
		assert.Equal(t, float64(1), resp)
	}()

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name:     "JSON.OBJLEN with root path",
			commands: []string{"json.set obj $ " + a, "json.objlen obj $"},
			expected: []interface{}{"OK", []interface{}{float64(2)}},
		},
		{
			name:     "JSON.OBJLEN with nested path",
			commands: []string{"json.set obj $ " + b, "json.objlen obj $.partner"},
			expected: []interface{}{"OK", []interface{}{float64(2)}},
		},
		{
			name:     "JSON.OBJLEN with non-object path",
			commands: []string{"json.set obj $ " + d, "json.objlen obj $"},
			expected: []interface{}{"OK", []interface{}{nil}},
		},
		{
			name:     "JSON.OBJLEN with nested non-object path",
			commands: []string{"json.set obj $ " + c, "json.objlen obj $.partner2.name"},
			expected: []interface{}{"OK", []interface{}{nil}},
		},
		{
			name:     "JSON.OBJLEN nested objects",
			commands: []string{"json.set obj $ " + b, "json.objlen obj $..language"},
			expected: []interface{}{"OK", []interface{}{nil, nil}},
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
			expected: []interface{}{"OK", float64(2)},
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
			expected: []interface{}{"OK", float64(3)},
		},
		{
			name:     "JSON.OBJLEN with legacy path - inner existing path",
			commands: []string{"json.set obj $ " + c, "json.objlen obj .partner", "json.objlen obj .partner2"},
			expected: []interface{}{"OK", float64(2), float64(2)},
		},
		{
			name:     "JSON.OBJLEN with legacy path - inner existing path v2",
			commands: []string{"json.set obj $ " + c, "json.objlen obj partner", "json.objlen obj partner2"},
			expected: []interface{}{"OK", float64(2), float64(2)},
		},
		{
			name:     "JSON.OBJLEN with legacy path - inner non-existent path",
			commands: []string{"json.set obj $ " + c, "json.objlen obj .idonotexist"},
			expected: []interface{}{"OK", nil},
		},
		{
			name:     "JSON.OBJLEN with legacy path - inner non-existent path v2",
			commands: []string{"json.set obj $ " + c, "json.objlen obj idonotexist"},
			expected: []interface{}{"OK", nil},
		},
		{
			name:     "JSON.OBJLEN with legacy path - inner existent path with nonJSON object",
			commands: []string{"json.set obj $ " + c, "json.objlen obj .name"},
			expected: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
		{
			name:     "JSON.OBJLEN with legacy path - inner existent path recursive object",
			commands: []string{"json.set obj $ " + c, "json.objlen obj ..partner"},
			expected: []interface{}{"OK", float64(2)},
		},
	}

	for _, tcase := range testCases {
		DeleteKey(t, conn, exec, "obj")
		t.Run(tcase.name, func(t *testing.T) {
			for i := 0; i < len(tcase.commands); i++ {
				cmd := tcase.commands[i]
				out := tcase.expected[i]
				result, err := exec.FireCommandAndReadResponse(conn, cmd)
				assert.Nil(t, err)
				assert.Equal(t, out, result)
			}
		})
	}
}

func TestJsonARRTRIM(t *testing.T) {
	exec := NewWebsocketCommandExecutor()
	conn := exec.ConnectToServer()
	defer conn.Close()

	a := `[0,1,2]`
	b := `{"connection":{"wireless":true,"names":[0,1,2,3,4]},"names":[0,1,2,3,4]}`

	defer func() {
		resp1, err := exec.FireCommandAndReadResponse(conn, "DEL a")
		assert.Nil(t, err)
		resp2, err := exec.FireCommandAndReadResponse(conn, "DEL b")
		assert.Nil(t, err)
		assert.Equal(t, float64(1), resp1)
		assert.Equal(t, float64(1), resp2)
	}()

	testCases := []struct {
		name       string
		commands   []string
		expected   []interface{}
		assertType []string
	}{
		{
			name:       "JSON.ARRTRIM not array",
			commands:   []string{"JSON.SET b $ " + b, `JSON.ARRTRIM b $ 0 10`, "JSON.GET b"},
			expected:   []interface{}{"OK", []interface{}{nil}, b},
			assertType: []string{"equal", "deep_equal", "jsoneq"},
		},
		{
			name:       "JSON.ARRTRIM stop index out of bounds",
			commands:   []string{"JSON.SET a $ " + a, `JSON.ARRTRIM a $ -10 10`, "JSON.GET a"},
			expected:   []interface{}{"OK", []interface{}{float64(3)}, "[0,1,2]"},
			assertType: []string{"equal", "deep_equal", "equal"},
		},
		{
			name:       "JSON.ARRTRIM start&stop are positive",
			commands:   []string{"JSON.SET a $ " + a, `JSON.ARRTRIM a $ 1 2`, "JSON.GET a"},
			expected:   []interface{}{"OK", []interface{}{float64(2)}, "[1,2]"},
			assertType: []string{"equal", "deep_equal", "equal"},
		},
		{
			name:       "JSON.ARRTRIM start&stop are negative",
			commands:   []string{"JSON.SET a $ " + a, `JSON.ARRTRIM a $ -2 -1 `, "JSON.GET a"},
			expected:   []interface{}{"OK", []interface{}{float64(2)}, "[1,2]"},
			assertType: []string{"equal", "deep_equal", "equal"},
		},
		{
			name:       "JSON.ARRTRIM subpath trim",
			commands:   []string{"JSON.SET b $ " + b, `JSON.ARRTRIM b $..names 1 4`, "JSON.GET b"},
			expected:   []interface{}{"OK", []interface{}{float64(4), float64(4)}, `{"connection":{"wireless":true,"names":[1,2,3,4]},"names":[1,2,3,4]}`},
			assertType: []string{"equal", "deep_equal", "jsoneq"},
		},
		{
			name:       "JSON.ARRTRIM subpath not array",
			commands:   []string{"JSON.SET b $ " + b, `JSON.ARRTRIM b $.connection 0 1`, "JSON.GET b"},
			expected:   []interface{}{"OK", []interface{}{nil}, b},
			assertType: []string{"equal", "deep_equal", "jsoneq"},
		},
		{
			name:       "JSON.ARRTRIM positive start larger than stop",
			commands:   []string{"JSON.SET b $ " + b, `JSON.ARRTRIM b $.names 3 1`, "JSON.GET b"},
			expected:   []interface{}{"OK", []interface{}{float64(0)}, `{"names":[],"connection":{"wireless":true,"names":[0,1,2,3,4]}}`},
			assertType: []string{"equal", "deep_equal", "jsoneq"},
		},
		{
			name:       "JSON.ARRTRIM negative start larger than stop",
			commands:   []string{"JSON.SET b $ " + b, `JSON.ARRTRIM b $.names -1 -3`, "JSON.GET b"},
			expected:   []interface{}{"OK", []interface{}{float64(0)}, `{"names":[],"connection":{"wireless":true,"names":[0,1,2,3,4]}}`},
			assertType: []string{"equal", "deep_equal", "jsoneq"},
		},
	}

	for _, tcase := range testCases {
		t.Run(tcase.name, func(t *testing.T) {
			for i := 0; i < len(tcase.commands); i++ {
				cmd := tcase.commands[i]
				out := tcase.expected[i]
				result, err := exec.FireCommandAndReadResponse(conn, cmd)
				assert.Nil(t, err)
				if tcase.assertType[i] == "equal" {
					assert.Equal(t, out, result)
				} else if tcase.assertType[i] == "deep_equal" {
					assert.True(t, testutils.ArraysArePermutations(out.([]interface{}), result.([]interface{})))
				} else if tcase.assertType[i] == "jsoneq" {
					assert.JSONEq(t, out.(string), result.(string))
				}
			}
		})
	}
}

func TestJsonARRINSERT(t *testing.T) {
	exec := NewWebsocketCommandExecutor()
	conn := exec.ConnectToServer()
	defer conn.Close()

	a := `[1,2]`
	b := `{"name":"tom","score":[10,20],"partner2":{"score":[10,20]}}`

	defer func() {
		resp1, err := exec.FireCommandAndReadResponse(conn, "DEL a")
		assert.Nil(t, err)
		resp2, err := exec.FireCommandAndReadResponse(conn, "DEL b")
		assert.Nil(t, err)
		assert.Equal(t, float64(1), resp1)
		assert.Equal(t, float64(1), resp2)
	}()

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
			expected:   []interface{}{"OK", []interface{}{float64(5)}, "[1,2,3,4,5]"},
			assertType: []string{"equal", "deep_equal", "equal"},
		},
		{
			name:       "JSON.ARRINSERT with negative index in root path",
			commands:   []string{"json.set a $ " + a, `JSON.ARRINSERT a $ -2 3 4 5`, "JSON.GET a"},
			expected:   []interface{}{"OK", []interface{}{float64(5)}, "[3,4,5,1,2]"},
			assertType: []string{"equal", "deep_equal", "equal"},
		},
		{
			name:       "JSON.ARRINSERT nested with positive index",
			commands:   []string{"JSON.SET b $ " + b, `JSON.ARRINSERT b $..score 1 5 6 true`, "JSON.GET b"},
			expected:   []interface{}{"OK", []interface{}{float64(5), float64(5)}, `{"name":"tom","score":[10,5,6,true,20],"partner2":{"score":[10,5,6,true,20]}}`},
			assertType: []string{"equal", "deep_equal", "jsoneq"},
		},
		{
			name:       "JSON.ARRINSERT nested with negative index",
			commands:   []string{"JSON.SET b $ " + b, `JSON.ARRINSERT b $..score -2 5 6 true`, "JSON.GET b"},
			expected:   []interface{}{"OK", []interface{}{float64(5), float64(5)}, `{"name":"tom","score":[5,6,true,10,20],"partner2":{"score":[5,6,true,10,20]}}`},
			assertType: []string{"equal", "deep_equal", "jsoneq"},
		},
	}

	for _, tcase := range testCases {
		t.Run(tcase.name, func(t *testing.T) {
			for i := 0; i < len(tcase.commands); i++ {
				cmd := tcase.commands[i]
				out := tcase.expected[i]
				result, err := exec.FireCommandAndReadResponse(conn, cmd)
				assert.Nil(t, err)
				if tcase.assertType[i] == "equal" {
					assert.Equal(t, out, result)
				} else if tcase.assertType[i] == "deep_equal" {
					assert.True(t, testutils.ArraysArePermutations(out.([]interface{}), result.([]interface{})))
				} else if tcase.assertType[i] == "jsoneq" {
					assert.JSONEq(t, out.(string), result.(string))
				}
			}
		})
	}
}

func TestJsonObjKeyslmao(t *testing.T) {
	exec := NewWebsocketCommandExecutor()
	conn := exec.ConnectToServer()
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
			expected:    []interface{}{nil},
		},
		{
			name:        "JSON.OBJKEYS with nested non-object path",
			setCommand:  "json.set doc $ " + b,
			testCommand: "json.objkeys doc $.partner.language",
			expected:    []interface{}{nil},
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
				nil,
			},
		},
	}

	for _, tcase := range testCases {
		t.Run(tcase.name, func(t *testing.T) {
			_, err := exec.FireCommandAndReadResponse(conn, tcase.setCommand)
			assert.Nil(t, err)
			expected := tcase.expected
			out, _ := exec.FireCommandAndReadResponse(conn, tcase.testCommand)

			sortNested := func(data []interface{}) {
				for _, elem := range data {
					if innerSlice, ok := elem.([]interface{}); ok {
						sort.Slice(innerSlice, func(i, j int) bool {
							return innerSlice[i].(string) < innerSlice[j].(string)
						})
					}
				}
			}

			if expected != nil {
				sortNested(expected)
			}
			if outSlice, ok := out.([]interface{}); ok {
				sortNested(outSlice)
				assert.ElementsMatch(t, expected, outSlice)
			} else {
				outInterface := []interface{}{out}
				assert.ElementsMatch(t, expected, outInterface)
			}
		})
	}
}

func TestJSONARRPOP(t *testing.T) {
	exec := NewWebsocketCommandExecutor()
	conn := exec.ConnectToServer()
	defer conn.Close()

	DeleteKey(t, conn, exec, "key")

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
			expected:   []interface{}{"OK", float64(2), "[0,1,3]"},
			assertType: []string{"equal", "equal", "deep_equal"},
		},
		{
			name:       "update nested array",
			commands:   []string{"json.set key $ " + nestedArray, "json.arrpop key $.b 2", "json.get key"},
			expected:   []interface{}{"OK", []interface{}{float64(2)}, `{"a":2,"b":[0,1,3]}`},
			assertType: []string{"equal", "deep_equal", "na"},
		},
	}

	for _, tcase := range testCases {
		t.Run(tcase.name, func(t *testing.T) {
			for i := 0; i < len(tcase.commands); i++ {
				cmd := tcase.commands[i]
				out := tcase.expected[i]
				result, err := exec.FireCommandAndReadResponse(conn, cmd)
				assert.Nil(t, err)
				jsonResult, isString := result.(string)

				if isString && testutils.IsJSONResponse(jsonResult) {
					assert.JSONEq(t, out.(string), jsonResult)
					continue
				}

				if tcase.assertType[i] == "equal" {
					assert.Equal(t, out, result)
				} else if tcase.assertType[i] == "deep_equal" {
					assert.True(t, testutils.ArraysArePermutations(tcase.expected[i].([]interface{}), result.([]interface{})))
				}
			}
		})
	}
}

func TestJsonARRAPPEND(t *testing.T) {
	exec := NewWebsocketCommandExecutor()
	conn := exec.ConnectToServer()
	defer conn.Close()

	DeleteKey(t, conn, exec, "key")
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
			expected:   []interface{}{"OK", []interface{}{float64(3)}},
			assertType: []string{"equal", "deep_equal"},
		},
		{
			name:       "JSON.ARRAPPEND nested",
			commands:   []string{"JSON.SET doc $ " + b, `JSON.ARRAPPEND doc $..score 10`},
			expected:   []interface{}{"OK", []interface{}{float64(2), float64(3)}},
			assertType: []string{"equal", "deep_equal"},
		},
		{
			name:       "JSON.ARRAPPEND nested with nil",
			commands:   []string{"JSON.SET doc $ " + c, `JSON.ARRAPPEND doc $..score 10`},
			expected:   []interface{}{"OK", []interface{}{float64(2), nil}},
			assertType: []string{"equal", "deep_equal"},
		},
		{
			name:       "JSON.ARRAPPEND with different datatypes",
			commands:   []string{"JSON.SET doc $ " + c, "JSON.ARRAPPEND doc $.name 1"},
			expected:   []interface{}{"OK", []interface{}{float64(2)}},
			assertType: []string{"equal", "deep_equal"},
		},
	}
	for _, tcase := range testCases {
		t.Run(tcase.name, func(t *testing.T) {
			DeleteKey(t, conn, exec, "a")
			DeleteKey(t, conn, exec, "doc")
			for i := 0; i < len(tcase.commands); i++ {
				cmd := tcase.commands[i]
				out := tcase.expected[i]
				result, err := exec.FireCommandAndReadResponse(conn, cmd)
				assert.Nil(t, err)
				jsonResult, isString := result.(string)

				if isString && testutils.IsJSONResponse(jsonResult) {
					assert.JSONEq(t, out.(string), jsonResult)
					continue
				}

				if tcase.assertType[i] == "equal" {
					assert.Equal(t, out, result)
				} else if tcase.assertType[i] == "deep_equal" {
					assert.True(t, testutils.ArraysArePermutations(tcase.expected[i].([]interface{}), result.([]interface{})))
				}
			}
		})
	}
}

func TestJSONDel(t *testing.T) {
	exec := NewWebsocketCommandExecutor()
	conn := exec.ConnectToServer()
	defer conn.Close()

	defer func() {
		resp, _ := exec.FireCommandAndReadResponse(conn, "DEL user")
		assert.Equal(t, float64(0), resp)
	}()

	preTestChecksCommand := "DEL user"
	postTestChecksCommand := "DEL user"

	testCases := []IntegrationTestCase{
		{
			name:      "Delete root path",
			setupData: `JSON.SET user $ {"age":13,"high":1.60,"flag":true,"name":"jerry","pet":null,"language":["python","golang"],"partner":{"name":"tom","language":["rust"]}}`,
			commands: []string{
				"JSON.DEL user $",
				"JSON.GET user $",
			},
			expected:   []interface{}{float64(1), nil},
			assertType: []string{"equal", "equal"},
			cleanUp:    []string{"DEL user"},
		},
		{
			name:      "Delete nested field",
			setupData: `JSON.SET user $ {"age":13,"high":1.60,"flag":true,"name":"jerry","pet":null,"language":["python","golang"],"partner":{"name":"tom","language":["rust"]}}`,
			commands: []string{
				"JSON.DEL user $.partner.name",
				"JSON.GET user $",
			},
			expected:   []interface{}{float64(1), `{"age":13,"high":1.60,"flag":true,"name":"jerry","pet":null,"language":["python","golang"],"partner":{"language":["rust"]}}`},
			assertType: []string{"equal", "jsoneq"},
			cleanUp:    []string{"DEL user"},
		},
		{
			name:      "del string type",
			setupData: `JSON.SET user $ {"flag":true,"name":"Tom"}`,
			commands: []string{
				"JSON.DEL user $.name",
				"JSON.GET user $"},
			expected:   []interface{}{float64(1), `{"flag":true}`},
			assertType: []string{"equal", "jsoneq"},
			cleanUp:    []string{"DEL user"},
		},
		{
			name:      "del bool type",
			setupData: `JSON.SET user $ {"flag":true,"name":"Tom"}`,
			commands: []string{
				"JSON.DEL user $.flag",
				"JSON.GET user $"},
			expected:   []interface{}{float64(1), `{"name":"Tom"}`},
			assertType: []string{"equal", "jsoneq"},
			cleanUp:    []string{"DEL user"},
		},
		{
			name:      "del null type",
			setupData: `JSON.SET user $ {"name":null,"age":28}`,
			commands: []string{
				"JSON.DEL user $.name",
				"JSON.GET user $"},
			expected:   []interface{}{float64(1), `{"age":28}`},
			assertType: []string{"equal", "jsoneq"},
			cleanUp:    []string{"DEL user"},
		},
		{
			name:      "del array type",
			setupData: `JSON.SET user $ {"names":["Rahul","Tom"],"bosses":{"names":["Jerry","Rocky"],"hobby":"swim"}}`,
			commands: []string{
				"JSON.DEL user $..names",
				"JSON.GET user $"},
			expected:   []interface{}{float64(2), `{"bosses":{"hobby":"swim"}}`},
			assertType: []string{"equal", "jsoneq"},
			cleanUp:    []string{"DEL user"},
		},
		{
			name:      "del integer type",
			setupData: `JSON.SET user $ {"age":28,"name":"Tom"}`,
			commands: []string{
				"JSON.DEL user $.age",
				"JSON.GET user $"},
			expected:   []interface{}{float64(1), `{"name":"Tom"}`},
			assertType: []string{"equal", "jsoneq"},
			cleanUp:    []string{"DEL user"},
		},
		{
			name:      "del float type",
			setupData: `JSON.SET user $ {"price":3.14,"name":"sugar"}`,
			commands: []string{
				"JSON.DEL user $.price",
				"JSON.GET user $"},
			expected:   []interface{}{float64(1), `{"name":"sugar"}`},
			assertType: []string{"equal", "jsoneq"},
			cleanUp:    []string{"DEL user"},
		},
		{
			name:      "delete key with []",
			setupData: `JSON.SET user $ {"key[0]":"value","array":["a","b"]}`,
			commands: []string{
				`JSON.DEL user ["key[0]"]`,
				"JSON.GET user $"},
			expected:   []interface{}{float64(1), `{"array": ["a","b"]}`},
			assertType: []string{"equal", "jsoneq"},
			cleanUp:    []string{"DEL user"},
		},
	}

	runIntegrationTests(t, exec, conn, testCases, preTestChecksCommand, postTestChecksCommand)
}

func TestJSONForget(t *testing.T) {
	exec := NewWebsocketCommandExecutor()
	conn := exec.ConnectToServer()
	defer conn.Close()

	defer func() {
		resp, _ := exec.FireCommandAndReadResponse(conn, "DEL user")
		assert.Equal(t, float64(0), resp)
	}()

	preTestChecksCommand := "DEL user"
	postTestChecksCommand := "DEL user"

	testCases := []IntegrationTestCase{
		{
			name:      "Forget root path",
			setupData: `JSON.SET user $ {"age":13,"high":1.60,"flag":true,"name":"jerry","pet":null,"language":["python","golang"],"partner":{"name":"tom","language":["rust"]}}`,
			commands: []string{
				"JSON.FORGET user $",
				"JSON.GET user $",
			},
			expected:   []interface{}{float64(1), nil},
			assertType: []string{"equal", "equal"},
			cleanUp:    []string{"DEL user"},
		},
		{
			name:      "Forget nested field",
			setupData: `JSON.SET user $ {"age":13,"high":1.60,"flag":true,"name":"jerry","pet":null,"language":["python","golang"],"partner":{"name":"tom","language":["rust"]}}`,
			commands: []string{
				"JSON.FORGET user $.partner.name",
				"JSON.GET user $",
			},
			expected:   []interface{}{float64(1), `{"age":13,"high":1.60,"flag":true,"name":"jerry","pet":null,"language":["python","golang"],"partner":{"language":["rust"]}}`},
			assertType: []string{"equal", "jsoneq"},
			cleanUp:    []string{"DEL user"},
		},

		{
			name:      "forget string type",
			setupData: `JSON.SET user $ {"flag":true,"name":"Tom"}`,
			commands: []string{
				"JSON.FORGET user $.name",
				"JSON.GET user $"},
			expected:   []interface{}{float64(1), `{"flag":true}`},
			assertType: []string{"equal", "jsoneq"},
			cleanUp:    []string{"DEL user"},
		},
		{
			name:      "forget bool type",
			setupData: `JSON.SET user $ {"flag":true,"name":"Tom"}`,
			commands: []string{
				"JSON.FORGET user $.flag",
				"JSON.GET user $"},
			expected:   []interface{}{float64(1), `{"name":"Tom"}`},
			assertType: []string{"equal", "jsoneq"},
			cleanUp:    []string{"DEL user"},
		},
		{
			name:      "forget null type",
			setupData: `JSON.SET user $ {"name":null,"age":28}`,
			commands: []string{
				"JSON.FORGET user $.name",
				"JSON.GET user $"},
			expected:   []interface{}{float64(1), `{"age":28}`},
			assertType: []string{"equal", "jsoneq"},
			cleanUp:    []string{"DEL user"},
		},
		{
			name:      "forget array type",
			setupData: `JSON.SET user $ {"names":["Rahul","Tom"],"bosses":{"names":["Jerry","Rocky"],"hobby":"swim"}}`,
			commands: []string{
				"JSON.FORGET user $..names",
				"JSON.GET user $"},
			expected:   []interface{}{float64(2), `{"bosses":{"hobby":"swim"}}`},
			assertType: []string{"equal", "jsoneq"},
			cleanUp:    []string{"DEL user"},
		},
		{
			name:      "forget integer type",
			setupData: `JSON.SET user $ {"age":28,"name":"Tom"}`,
			commands: []string{
				"JSON.FORGET user $.age",
				"JSON.GET user $"},
			expected:   []interface{}{float64(1), `{"name":"Tom"}`},
			assertType: []string{"equal", "jsoneq"},
			cleanUp:    []string{"DEL user"},
		},
		{
			name:      "forget float type",
			setupData: `JSON.SET user $ {"price":3.14,"name":"sugar"}`,
			commands: []string{
				"JSON.FORGET user $.price",
				"JSON.GET user $"},
			expected:   []interface{}{float64(1), `{"name":"sugar"}`},
			assertType: []string{"equal", "jsoneq"},
			cleanUp:    []string{"DEL user"},
		},
		{
			name:      "forget array element",
			setupData: `JSON.SET user $ {"names":["Rahul","Tom"],"bosses":{"names":["Jerry","Rocky"],"hobby":"swim"}}`,
			commands: []string{
				"JSON.FORGET user $.names[0]",
				"JSON.GET user $"},
			expected:   []interface{}{float64(1), `{"names":["Tom"],"bosses":{"names":["Jerry","Rocky"],"hobby":"swim"}}`},
			assertType: []string{"equal", "jsoneq"},
			cleanUp:    []string{"DEL user"},
		},
	}

	runIntegrationTests(t, exec, conn, testCases, preTestChecksCommand, postTestChecksCommand)
}

func TestJSONToggle(t *testing.T) {
	exec := NewWebsocketCommandExecutor()
	conn := exec.ConnectToServer()
	defer conn.Close()

	defer func() {
		resp, _ := exec.FireCommandAndReadResponse(conn, "DEL user")
		assert.Equal(t, float64(0), resp)
	}()

	preTestChecksCommand := "DEL user"
	postTestChecksCommand := "DEL user"

	simpleJSON := `{"name":"DiceDB","hasAccess":false}`
	complexJson := `{"field":true,"nested":{"field":false,"nested":{"field":true}}}`

	testCases := []IntegrationTestCase{
		{
			name:       "JSON.TOGGLE with existing key",
			setupData:  `JSON.SET user $ ` + simpleJSON,
			commands:   []string{"JSON.TOGGLE user $.hasAccess"},
			expected:   []interface{}{[]interface{}{float64(1)}},
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
			cleanUp:    []string{"DEL user"},
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
				complexJson,
				[]any{float64(0), float64(1), float64(0)}, // Toggle: true -> false, false -> true, true -> false
				`{"field":false,"nested":{"field":true,"nested":{"field":false}}}`,
			},
			assertType: []string{"jsoneq", "jsoneq", "jsoneq"},
			cleanUp:    []string{"DEL user"},
		},
	}

	runIntegrationTests(t, exec, conn, testCases, preTestChecksCommand, postTestChecksCommand)
}

func TestJSONNumIncrBy(t *testing.T) {
	exec := NewWebsocketCommandExecutor()
	conn := exec.ConnectToServer()
	defer conn.Close()

	defer func() {
		resp, _ := exec.FireCommandAndReadResponse(conn, "DEL foo")
		assert.Equal(t, float64(0), resp)
	}()

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
			assertType: []string{"perm_equal", "json_equal", "perm_equal", "json_equal"},
			cleanUp:    []string{"DEL foo"},
		},
		{
			name:       "incrby at root path",
			setupData:  "JSON.SET foo $ 1",
			commands:   []string{"expire foo 10", "JSON.NUMINCRBY foo $ 1", "ttl foo", "JSON.GET foo $", "JSON.NUMINCRBY foo $ -1", "JSON.GET foo $"},
			expected:   []interface{}{float64(1), "[2]", float64(10), "2", "[1]", "1"},
			assertType: []string{"equal", "equal", "range", "equal", "equal", "equal"},
			cleanUp:    []string{"DEL foo"},
		},
	}
	runIntegrationTests(t, exec, conn, testCases, preTestChecksCommand, postTestChecksCommand)
}

func TestJsonNumMultBy(t *testing.T) {
	exec := NewWebsocketCommandExecutor()
	conn := exec.ConnectToServer()
	defer conn.Close()

	defer func() {
		resp, _ := exec.FireCommandAndReadResponse(conn, "DEL docu")
		assert.Equal(t, float64(0), resp)
	}()

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

	runIntegrationTests(t, exec, conn, testCases, preTestChecksCommand, postTestChecksCommand)
}

func TestJSONARRINDEX(t *testing.T) {
	exec := NewWebsocketCommandExecutor()
	conn := exec.ConnectToServer()
	defer conn.Close()

	preTestChecksCommand := "DEL key"
	postTestChecksCommand := "DEL key"

	defer exec.FireCommand(conn, postTestChecksCommand)

	normalArray := `[0,1,2,3,4,3]`
	nestedArray := `{"arrays":[{"arr":[1,2,3]},{"arr":[2,3,4]},{"arr":[1]}]}`
	nestedArray2 := `{"a":[3],"nested":{"a":{"b":2,"c":1}}}`

	tests := []IntegrationTestCase{
		{
			name:       "should return error if key is not present",
			commands:   []string{"json.set key $ " + normalArray, "json.arrindex nonExistentKey $ 3"},
			expected:   []interface{}{"OK", "ERR could not perform this operation on a key that doesn't exist"},
			assertType: []string{"equal", "equal"},
			cleanUp:    []string{"DEL key"},
		},
		{
			name:       "should return error if json path is invalid",
			commands:   []string{"json.set key $ " + normalArray, "json.arrindex key $invalid_path 3"},
			expected:   []interface{}{"OK", "ERR Path '$invalid_path' does not exist"},
			assertType: []string{"equal", "equal"},
			cleanUp:    []string{"DEL key"},
		},
		{
			name:       "should return error if provided path does not have any data",
			commands:   []string{"json.set key $ " + normalArray, "json.arrindex key $.some_path 3"},
			expected:   []interface{}{"OK", []interface{}{}},
			assertType: []string{"equal", "equal"},
			cleanUp:    []string{"DEL key"},
		},
		{
			name:       "should return error if invalid start index provided",
			commands:   []string{"json.set key $ " + normalArray, "json.arrindex key $ 3 abc"},
			expected:   []interface{}{"OK", "ERR Couldn't parse as integer"},
			assertType: []string{"equal", "equal"},
			cleanUp:    []string{"DEL key"},
		},
		{
			name:       "should return error if invalid stop index provided",
			commands:   []string{"json.set key $ " + normalArray, "json.arrindex key $ 3 4 abc"},
			expected:   []interface{}{"OK", "ERR Couldn't parse as integer"},
			assertType: []string{"equal", "equal"},
			cleanUp:    []string{"DEL key"},
		},
		{
			name:       "should return array index when given element is present",
			commands:   []string{"json.set key $ " + normalArray, "json.arrindex key $ 3"},
			expected:   []interface{}{"OK", []interface{}{float64(3)}},
			assertType: []string{"equal", "equal"},
			cleanUp:    []string{"DEL key"},
		},
		{
			name:       "should return -1 when given element is not present",
			commands:   []string{"json.set key $ " + normalArray, "json.arrindex key $ 10"},
			expected:   []interface{}{"OK", []interface{}{float64(-1)}},
			assertType: []string{"equal", "equal"},
			cleanUp:    []string{"DEL key"},
		},
		{
			name:       "should return array index with start optional param provided",
			commands:   []string{"json.set key $ " + normalArray, "json.arrindex key $ 3 4"},
			expected:   []interface{}{"OK", []interface{}{float64(5)}},
			assertType: []string{"equal", "equal"},
			cleanUp:    []string{"DEL key"},
		},
		{
			name:       "should return array index with start and stop optional param provided",
			commands:   []string{"json.set key $ " + normalArray, "json.arrindex key $ 4 4 5"},
			expected:   []interface{}{"OK", []interface{}{float64(4)}},
			assertType: []string{"equal", "equal"},
			cleanUp:    []string{"DEL key"},
		},
		{
			name:       "should return -1 with start and stop optional param provided where start > stop",
			commands:   []string{"json.set key $ " + normalArray, "json.arrindex key $ 3 2 1"},
			expected:   []interface{}{"OK", []interface{}{float64(-1)}},
			assertType: []string{"equal", "equal"},
			cleanUp:    []string{"DEL key"},
		},
		{
			name:       "should return -1 with start (out of boud) and stop (out of bound) optional param provided",
			commands:   []string{"json.set key $ " + normalArray, "json.arrindex key $ 3 6 10"},
			expected:   []interface{}{"OK", []interface{}{float64(-1)}},
			assertType: []string{"equal", "equal"},
			cleanUp:    []string{"DEL key"},
		},
		{
			name:       "should return list of array indexes for nested json",
			commands:   []string{"json.set key $ " + nestedArray, "json.arrindex key $.arrays.*.arr 3"},
			expected:   []interface{}{"OK", []interface{}{float64(2), float64(1), float64(-1)}},
			assertType: []string{"equal", "deep_equal"},
			cleanUp:    []string{"DEL key"},
		},
		{
			name:       "should return list of array indexes for multiple json path",
			commands:   []string{"json.set key $ " + nestedArray, "json.arrindex key $..arr 3"},
			expected:   []interface{}{"OK", []interface{}{float64(2), float64(1), float64(-1)}},
			assertType: []string{"equal", "deep_equal"},
			cleanUp:    []string{"DEL key"},
		},
		{
			name:       "should return array of length 1 for nested json path, with index",
			commands:   []string{"json.set key $ " + nestedArray, "json.arrindex key $.arrays[1].arr 3"},
			expected:   []interface{}{"OK", []interface{}{float64(1)}},
			assertType: []string{"equal", "deep_equal"},
			cleanUp:    []string{"DEL key"},
		},
		{
			name:       "should return empty array for nonexistent path in nested json",
			commands:   []string{"json.set key $ " + nestedArray, "json.arrindex key $..arr1 3"},
			expected:   []interface{}{"OK", []interface{}{}},
			assertType: []string{"equal", "deep_equal"},
			cleanUp:    []string{"DEL key"},
		},
		{
			name:       "should return -1 for each nonexisting value in nested json",
			commands:   []string{"json.set key $ " + nestedArray, "json.arrindex key $..arr 5"},
			expected:   []interface{}{"OK", []interface{}{float64(-1), float64(-1), float64(-1)}},
			assertType: []string{"equal", "deep_equal"},
			cleanUp:    []string{"DEL key"},
		},
		{
			name:       "should return nil for non-array path and -1 for array path if value DNE",
			commands:   []string{"json.set key $ " + nestedArray2, "json.arrindex key $..a 2"},
			expected:   []interface{}{"OK", []interface{}{float64(-1), nil}},
			assertType: []string{"equal", "deep_equal"},
			cleanUp:    []string{"DEL key"},
		},
		{
			name:       "should return nil for non-array path if value DNE and valid index for array path if value exists",
			commands:   []string{"json.set key $ " + nestedArray2, "json.arrindex key $..a 3"},
			expected:   []interface{}{"OK", []interface{}{float64(0), nil}},
			assertType: []string{"equal", "deep_equal"},
			cleanUp:    []string{"DEL key"},
		},
		{
			name:       "should handle stop index - 0 which should be last index inclusive",
			commands:   []string{"json.set key $ " + nestedArray, "json.arrindex key $..arr 3 1 0", "json.arrindex key $..arr 3 2 0"},
			expected:   []interface{}{"OK", []interface{}{float64(2), float64(1), float64(-1)}, []interface{}{float64(2), float64(-1), float64(-1)}},
			assertType: []string{"equal", "deep_equal", "deep_equal"},
			cleanUp:    []string{"DEL key"},
		},
		{
			name:       "should handle stop index - -1 which should be last index exclusive",
			commands:   []string{"json.set key $ " + nestedArray, "json.arrindex key $..arr 3 1 -1", "json.arrindex key $..arr 3 2 -1"},
			expected:   []interface{}{"OK", []interface{}{float64(-1), float64(1), float64(-1)}, []interface{}{float64(-1), float64(-1), float64(-1)}},
			assertType: []string{"equal", "deep_equal", "deep_equal"},
			cleanUp:    []string{"DEL key"},
		},
		{
			name:       "should handle negative start index",
			commands:   []string{"json.set key $ " + nestedArray, "json.arrindex key $..arr 3 -1"},
			expected:   []interface{}{"OK", []interface{}{float64(2), float64(-1), float64(-1)}},
			assertType: []string{"equal", "deep_equal"},
			cleanUp:    []string{"DEL key"},
		},
	}

	runIntegrationTests(t, exec, conn, tests, preTestChecksCommand, postTestChecksCommand)
}

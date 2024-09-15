package commands

import (
	"testing"

	"github.com/dicedb/dice/testutils"
	"gotest.tools/v3/assert"
)

func TestEvalJSONTOGGLE(t *testing.T) {
    conn := getLocalConnection()
    defer conn.Close()

	simpleJSON :=`{"name":true,"age":false}`
	complexJson :=`{"field":true,"nested":{"field":false,"nested":{"field":true}}}`

    testCases := []struct {
        name     string
        commands []string
        expected []interface{}
    }{
        {
			name:     "JSON.TOGGLE with existing key",
			commands: []string{`JSON.SET user $ `+ simpleJSON, "JSON.TOGGLE user $.name"},
			expected: []interface{}{"OK", []any{int64(0)}},
		},
        {
            name:     "JSON.TOGGLE with non-existing key",
            commands: []string{"JSON.TOGGLE user $.flag"},
            expected: []interface{}{"ERR could not perform this operation on a key that doesn't exist"},
        },
        {
			name:     "JSON.TOGGLE with invalid path",
			commands: []string{`JSON.SET testkey $ ` + simpleJSON, "JSON.TOGGLE user $.invalidPath"},
			expected: []interface{}{"WRONGTYPE Operation against a key holding the wrong kind of value", "ERR could not perform this operation on a key that doesn't exist"},
		},
        {
            name:     "JSON.TOGGLE with invalid command format",
            commands: []string{"JSON.TOGGLE testKey"},
            expected: []interface{}{"ERR wrong number of arguments for 'json.toggle' command"},
        },
		{
            name:     "deeply nested JSON structure with multiple matching fields",
            commands: []string{
                `JSON.SET user $ `+ complexJson,
				"JSON.GET user",
				"JSON.TOGGLE user $..field",
				"JSON.GET user",
            },
            expected: []interface{}{"OK",
			"{\"field\":true,\"nested\":{\"field\":false,\"nested\":{\"field\":true}}}",
			[]any{int64(0), int64(1), int64(0)},
			"{\"field\":false,\"nested\":{\"field\":true,\"nested\":{\"field\":false}}}"},
        },
    }

    for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			FireCommand(conn, "DEL user") 
	
			for i, cmd := range tc.commands {
				result := FireCommand(conn, cmd)
				if expectedSlice, ok := tc.expected[i].([]interface{}); ok {
					assert.Assert(t, testutils.UnorderedEqual(expectedSlice, result))
				} else {
					assert.DeepEqual(t, tc.expected[i], result)
				}
			}
		})
	}
}
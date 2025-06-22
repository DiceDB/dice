package ironhawk

import (
	"testing"

	"gotest.tools/v3/assert"
)
	 
var getKeysAndFlagsTestCases = []struct {
	name     string
	inCmd    string
	expected interface{}
}{	 
	{
	 	name:  "Set command",
		inCmd: "set 1 2",
		expected: []interface{}{
			[]interface{}{
				1,
				[]interface{}{
					"OW",
					"update",
				},
			},
		},
	},
	{
		name:  "Get Command in Set",
		inCmd: "set 1 2 get",
		expected: []interface{}{
			[]interface{}{
				1,
				[]interface{}{
					"RW",
					"access",
					"update",
				},
			},
		},
	},
	{
		name:  "DEL Command",
		inCmd: "DEL 1 2",
		expected: []interface{}{
			[]interface{}{
				1,
				[]interface{}{
					"RM",
					"delete",
				},
			},
			[]interface{}{
				2,
				[]interface{}{
					"RM",
					"delete",
				},
			},
		},
	},
	{
		name:  "DEL Command",
		inCmd: "DEL 1 2",
		expected: []interface{}{
			[]interface{}{
				1,
				[]interface{}{
					"RM",
					"delete",
				},
			},
			[]interface{}{
				2,
				[]interface{}{
					"RM",
					"delete",
				},
			},
		},
	},
	{
		name:     "PING Command",
		inCmd:    "PING",
		expected: "ERR the command has no key arguments",
	},
}

func TestCommandGetKeysAndFlags(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()
	for _, tc := range getKeysAndFlagsTestCases {
		t.Run(tc.name, func(t *testing.T) {
			result := conn.FireString("COMMAND GETKEYSANDFLAGS " + tc.inCmd)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func BenchmarkGetKeysAndFlagsMatch(b *testing.B) {
	conn := getLocalConnection()
	defer conn.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tc := range getKeysAndFlagsTestCases {
			conn.FireString("COMMAND GETKEYS " + tc.inCmd)
		}
	}
}

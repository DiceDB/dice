package async

import (
	"testing"

	"gotest.tools/v3/assert"
)

var THREE int64 = 3
var FOUR int64 = 4

func TestHSCAN(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	testCases := []TestCase{
		{
			commands: []string{"HSCAN empty_hash 0"},
			expected: []interface{}{[]interface{}{"0", []interface{}{}}},
		},
		{
			commands: []string{
				"HSET test_hash field1 value1 field2 value2 field3 value3",
				"HSCAN test_hash 0",
			},
			expected: []interface{}{
				THREE,
				[]interface{}{
					"0",
					[]interface{}{"field1", "value1", "field2", "value2", "field3", "value3"},
				},
			},
		},
		{
			commands: []string{
				"HSET pattern_hash foo1 bar1 foo2 bar2 baz1 qux1 baz2 qux2",
				"HSCAN pattern_hash 0 MATCH foo*",
			},
			expected: []interface{}{
				FOUR,
				[]interface{}{
					"0",
					[]interface{}{"foo1", "bar1", "foo2", "bar2"},
				},
			},
		},
		{
			commands: []string{
				"HSET large_hash field1 value1 field2 value2",
				"HSCAN large_hash 0 COUNT 2",
			},
			expected: []interface{}{
				TWO,
				[]interface{}{"0", []interface{}{"field1", "value1", "field2", "value2"}},
			},
		},
		{
			commands: []string{
				"SET wrong_type_key string_value",
				"HSCAN wrong_type_key 0",
			},
			expected: []interface{}{
				"OK",
				"WRONGTYPE Operation against a key holding the wrong kind of value",
			},
		},
		{
			commands: []string{"HSCAN"},
			expected: []interface{}{"ERR wrong number of arguments for 'hscan' command"},
		},
		{
			commands: []string{
				"HSET test_hash1 field1 value1 field2 value2 field3 value3 field4 value4",
				"HSCAN test_hash1 0 COUNT 2",
				"HSCAN test_hash1 2 COUNT 2",
			},
			expected: []interface{}{
				FOUR,
				[]interface{}{"2", []interface{}{"field1", "value1", "field2", "value2"}},
				[]interface{}{"0", []interface{}{"field3", "value3", "field4", "value4"}},
			},
		},
	}

	for _, tc := range testCases {
		for i, cmd := range tc.commands {
			result := FireCommand(conn, cmd)
			assert.DeepEqual(t, tc.expected[i], result)
		}
	}
}

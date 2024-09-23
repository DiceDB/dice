package commands

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestHINCRBY(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	testcases := []TestCase{
		{
			commands: []string{"HINCRBY key field"},
			expected: []interface{}{"ERR wrong number of arguments for 'hincrby' command"},
		},
		{
			commands: []string{"HINCRBY key field-1 10", "HINCRBY key field-2 20"},
			expected: []interface{}{int64(10), int64(20)},
		},
		{
			commands: []string{"HINCRBY key field-1 10", "HINCRBY key field-1 20"},
			expected: []interface{}{int64(20), int64(40)},
		},
		{
			commands: []string{"HSET key field 10", "HINCRBY key field 5"},
			expected: []interface{}{int64(1), int64(15)},
		},
		{
			commands: []string{"HSET key field value", "HINCRBY key field 20"},
			expected: []interface{}{int64(0), "ERR hash value is not an integer"},
		},
		{
			commands: []string{"HINCRBY key field value", "HINCRBY key field 9999999999999999999999999999999999999"},
			expected: []interface{}{"ERR value is not an integer or out of range", "ERR value is not an integer or out of range"},
		},
	}

	for _, tc := range testcases {
		for i, cmd := range tc.commands {
			result := FireCommand(conn, cmd)
			assert.DeepEqual(t, tc.expected[i], result)
		}
	}
}

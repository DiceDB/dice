package resp

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestVals(t *testing.T){
	conn := getLocalConnection()
	defer conn.Close()

	testCases := []TestCase{
		{
			name: "One or more vals exist",
			commands: []string{"HSET key field value", "HSET key field1 value1","HVALS key"},
			expected: []interface{}{int64(1), int64(1), []interface{}{"value", "value1"}},
		},
		{
			name: "No values exist",
			commands: []string{"HVALS key"},
			expected: []interface{}{[]interface{}{}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			FireCommand(conn, "HDEL key field")
			FireCommand(conn, "HDEL key field1")

			for i, cmd := range tc.commands {
				result := FireCommand(conn, cmd)
				assert.DeepEqual(t, tc.expected[i], result)
			}
		})
	}
}
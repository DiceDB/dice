package async

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestAPPEND(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name: "Append After Set and Delete",
			commands: []string{
				"SET key value",
				"APPEND key value",
				"GET key",
				"APPEND key 100",
				"GET key",
				"DEL key",
				"APPEND key value",
				"GET key",
			},
			expected: []interface{}{"OK", int64(10), "valuevalue", int64(13), "valuevalue100", int64(1), int64(5), "value"},
		},
		{
			name: "Append to Integer Values",
			commands: []string{
				"DEL key",
				"APPEND key 1",
				"APPEND key 2",
				"GET key",
				"SET key 1",
				"APPEND key 2",
				"GET key",
			},
			expected: []interface{}{int64(0), int64(1), int64(2), "12", "OK", int64(2), "12"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			FireCommand(conn, "DEL key")

			for i, cmd := range tc.commands {
				result := FireCommand(conn, cmd)
				assert.DeepEqual(t, tc.expected[i], result)
			}
		})
	}
}

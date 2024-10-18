package resp

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestGETRANGE(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	testCases := []TestCase{
		{
			name:     "Get range on a string",
			commands: []string{"SET test1 shankar", "GETRANGE test1 0 7"},
			expected: []interface{}{"OK", "shankar"},
		},
		{
			name:     "Get range on a non existent key",
			commands: []string{"GETRANGE test2 0 7"},
			expected: []interface{}{""},
		},
		{
			name:     "Get range on wrong key type",
			commands: []string{"LPUSH test3 shankar", "GETRANGE test3 0 7"},
			expected: []interface{}{int64(1), "WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
		{
			name:     "GETRANGE against string value: 0, -1",
			commands: []string{"SET test4 apple", "GETRANGE test4 0 -1"},
			expected: []interface{}{"OK", "apple"},
		},
		{
			name:     "GETRANGE against string value: 5, 3",
			commands: []string{"SET test5 apple", "GETRANGE test5 5 3"},
			expected: []interface{}{"OK", ""},
		},
		{
			name:     "GETRANGE against integer value: -1, -100",
			commands: []string{"SET test6 apple", "GETRANGE test6 -1 -100"},
			expected: []interface{}{"OK", ""},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// deleteTestKeys([]string{"k"}, store)
			FireCommand(conn, "DEL k")

			for i, cmd := range tc.commands {
				result := FireCommand(conn, cmd)
				assert.DeepEqual(t, tc.expected[i], result)
			}
		})
	}
}

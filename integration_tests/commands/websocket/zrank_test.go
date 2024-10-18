package websocket

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestZRANK(t *testing.T) {
	exec := NewWebsocketCommandExecutor()
	conn := exec.ConnectToServer()

	// Clean up before and after tests
	exec.FireCommand(conn, "DEL myset")
	defer exec.FireCommand(conn, "DEL myset")

	// Initialize the sorted set with members and their scores
	exec.FireCommand(conn, "ZADD myset 1 member1 2 member2 3 member3 4 member4 5 member5")

	testCases := []TestCase{
		{
			name:     "ZRANK of existing member",
			commands: []string{"ZRANK myset member1"},
			expected: []interface{}{float64(0)},
		},
		{
			name:     "ZRANK of non-existing member",
			commands: []string{"ZRANK myset member6"},
			expected: []interface{}{"(nil)"},
		},
		{
			name:     "ZRANK with WITHSCORE option for existing member",
			commands: []string{"ZRANK myset member3 WITHSCORE"},
			expected: []interface{}{[]interface{}{float64(2), float64(3)}},
		},
		{
			name:     "ZRANK with WITHSCORE option for non-existing member",
			commands: []string{"ZRANK myset member6 WITHSCORE"},
			expected: []interface{}{"(nil)"},
		},
		{
			name:     "ZRANK on non-existing myset",
			commands: []string{"ZRANK nonexisting member1"},
			expected: []interface{}{"(nil)"},
		},
		{
			name:     "ZRANK with wrong number of arguments",
			commands: []string{"ZRANK myset"},
			expected: []interface{}{"ERR wrong number of arguments for 'zrank' command"},
		},
		{
			name:     "ZRANK with invalid option",
			commands: []string{"ZRANK myset member1 INVALID_OPTION"},
			expected: []interface{}{"ERR syntax error"},
		},
	}

	for _, tc := range testCases {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result := exec.FireCommand(conn, cmd)
				assert.DeepEqual(t, tc.expected[i], result)
			}
		})
	}
}

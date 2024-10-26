package websocket

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestZRANK(t *testing.T) {
	exec := NewWebsocketCommandExecutor()
	conn := exec.ConnectToServer()

	// Clean up before and after tests
	DeleteKey(t, conn, exec, "myset")
	defer func() {
		resp, err := exec.FireCommandAndReadResponse(conn, "DEL myset")
		assert.Nil(t, err)
		assert.Equal(t, float64(1), resp, "Cleanup failed")
	}()

	// Initialize the sorted set with members and their scores
	_, err := exec.FireCommandAndReadResponse(conn, "ZADD myset 1 member1 2 member2 3 member3 4 member4 5 member5")
	assert.Nil(t, err)

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
			expected: []interface{}{[]interface{}{float64(2), "3"}},
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
				result, err := exec.FireCommandAndReadResponse(conn, cmd)
				assert.Nil(t, err)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}
}

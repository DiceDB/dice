package async

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestZADD(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	FireCommand(conn, "DEL key")
	defer FireCommand(conn, "DEL key")

	testCases := []TestCase{
		{
			name:     "ZADD with two new members",
			commands: []string{"ZADD key 1 member1 2 member2"},
			expected: []interface{}{int64(2)},
		},
		{
			name:     "ZADD with three new members",
			commands: []string{"ZADD key 3 member3 4 member4 5 member5"},
			expected: []interface{}{int64(3)},
		},
		{
			name:     "ZADD with existing members",
			commands: []string{"ZADD key 1 member1 2 member2 3 member3 4 member4 5 member5"},
			expected: []interface{}{int64(0)},
		},
		{
			name:     "ZADD with mixed new and existing members",
			commands: []string{"ZADD key 1 member1 2 member2 3 member3 4 member4 5 member5 6 member6"},
			expected: []interface{}{int64(1)},
		},
		{
			name:     "ZADD without any members",
			commands: []string{"ZADD key 1"},
			expected: []interface{}{"ERR wrong number of arguments for 'zadd' command"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result := FireCommand(conn, cmd)
				assert.DeepEqual(t, tc.expected[i], result)
			}
		})
	}
}

func TestZRANGE(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	FireCommand(conn, "DEL key")
	defer FireCommand(conn, "DEL key")

	FireCommand(conn, "ZADD key 1 member1 2 member2 3 member3 4 member4 5 member5 6 member6")
	defer FireCommand(conn, "DEL key")

	testCases := []TestCase{
		{
			name:     "ZRANGE with mixed indices",
			commands: []string{"ZRANGE key 0 -1"},
			expected: []interface{}{[]interface{}{"member1", "member2", "member3", "member4", "member5", "member6"}},
		},
		{
			name:     "ZRANGE with positive indices #1",
			commands: []string{"ZRANGE key 0 2"},
			expected: []interface{}{[]interface{}{"member1", "member2", "member3"}},
		},
		{
			name:     "ZRANGE with positive indices #2",
			commands: []string{"ZRANGE key 2 4"},
			expected: []interface{}{[]interface{}{"member3", "member4", "member5"}},
		},
		{
			name:     "ZRANGE with all positive indices",
			commands: []string{"ZRANGE key 0 10"},
			expected: []interface{}{[]interface{}{"member1", "member2", "member3", "member4", "member5", "member6"}},
		},
		{
			name:     "ZRANGE with out of bound indices",
			commands: []string{"ZRANGE key 10 20"},
			expected: []interface{}{[]interface{}{}},
		},
		{
			name:     "ZRANGE with positive indices and scores",
			commands: []string{"ZRANGE key 0 10 WITHSCORES"},
			expected: []interface{}{[]interface{}{"member1", "1", "member2", "2", "member3", "3", "member4", "4", "member5", "5", "member6", "6"}},
		},
		{
			name:     "ZRANGE with positive indices and scores in reverse order",
			commands: []string{"ZRANGE key 0 10 REV WITHSCORES"},
			expected: []interface{}{[]interface{}{"member6", "6", "member5", "5", "member4", "4", "member3", "3", "member2", "2", "member1", "1"}},
		},
		{
			name:     "ZRANGE with negative indices",
			commands: []string{"ZRANGE key -1 -1"},
			expected: []interface{}{[]interface{}{"member6"}},
		},
		{
			name:     "ZRANGE with negative indices and scores",
			commands: []string{"ZRANGE key -8 -5 WITHSCORES"},
			expected: []interface{}{[]interface{}{"member1", "1", "member2", "2"}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result := FireCommand(conn, cmd)
				assert.DeepEqual(t, tc.expected[i], result)
			}
		})
	}
}

func TestZRANK(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	FireCommand(conn, "DEL key")
	defer FireCommand(conn, "DEL key")

	// Initialize the sorted set with members and their scores
	FireCommand(conn, "ZADD key 1 member1 2 member2 3 member3 4 member4 5 member5")

	testCases := []TestCase{
		{
			name:     "ZRANK of existing member",
			commands: []string{"ZRANK key member1"},
			expected: []interface{}{int64(0)},
		},
		{
			name:     "ZRANK of non-existing member",
			commands: []string{"ZRANK key member6"},
			expected: []interface{}{"(nil)"},
		},
		{
			name:     "ZRANK with WITHSCORE option for existing member",
			commands: []string{"ZRANK key member3 WITHSCORE"},
			expected: []interface{}{[]interface{}{int64(2), int64(3)}},
		},
		{
			name:     "ZRANK with WITHSCORE option for non-existing member",
			commands: []string{"ZRANK key member6 WITHSCORE"},
			expected: []interface{}{"(nil)"},
		},
		{
			name:     "ZRANK on non-existing key",
			commands: []string{"ZRANK nonexisting member1"},
			expected: []interface{}{"(nil)"},
		},
		{
			name:     "ZRANK with wrong number of arguments",
			commands: []string{"ZRANK key"},
			expected: []interface{}{"ERR wrong number of arguments for 'zrank' command"},
		},
		{
			name:     "ZRANK with invalid option",
			commands: []string{"ZRANK key member1 INVALID_OPTION"},
			expected: []interface{}{"ERR syntax error"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result := FireCommand(conn, cmd)
				assert.DeepEqual(t, tc.expected[i], result)
			}
		})
	}
}

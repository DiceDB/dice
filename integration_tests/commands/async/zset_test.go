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

func TestZCOUNT(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	FireCommand(conn, "DEL key")
	defer FireCommand(conn, "DEL key")

	FireCommand(conn, "ZADD key 10 member1 20 member2 30 member3 40 member4")

	testCases := []TestCase{
		{
			name:     "ZCOUNT with matching members in range 15-35",
			commands: []string{"ZCOUNT key 15 35"},
			expected: []interface{}{int64(2)}, // member2 and member3
		},
		{
			name:     "ZCOUNT with full range -inf to +inf",
			commands: []string{"ZCOUNT key -inf +inf"},
			expected: []interface{}{int64(4)}, // All members
		},
		{
			name:     "ZCOUNT with no matching members",
			commands: []string{"ZCOUNT key 50 100"},
			expected: []interface{}{int64(0)},
		},
		{
			name:     "ZCOUNT with single member at boundary",
			commands: []string{"ZCOUNT key 20 20"},
			expected: []interface{}{int64(1)}, // Only member2
		},
		{
			name:     "ZCOUNT with reversed range",
			commands: []string{"ZCOUNT key 40 10"},
			expected: []interface{}{int64(0)}, // Invalid range
		},
		{
			name:     "ZCOUNT on non-existent key",
			commands: []string{"ZCOUNT nonExistingKey 0 100"},
			expected: []interface{}{int64(0)},
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

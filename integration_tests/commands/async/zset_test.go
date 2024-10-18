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

func TestZPOPMIN(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	testCases := []TestCase{
		{
			name:     "ZPOPMIN on non-existing key with/without count argument",
			commands: []string{"ZPOPMIN NON_EXISTENT_KEY"},
			expected: []interface{}{[]interface{}{}},
		},
		{
			name:     "ZPOPMIN with wrong type of key with/without count argument",
			commands: []string{"SET stringkey string_value", "ZPOPMIN stringkey", "DEL stringkey"},
			expected: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value", int64(1)},
		},
		{
			name:     "ZPOPMIN on existing key (without count argument)",
			commands: []string{"ZADD myzset 1 member1 2 member2 3 member3", "ZPOPMIN myzset", "DEL myzset"},
			expected: []interface{}{int64(3), []interface{}{"member1", "1"}, int64(1)},
		},
		{
			name:     "ZPOPMIN with normal count argument",
			commands: []string{"ZADD myzset 1 member1 2 member2 3 member3", "ZPOPMIN myzset 2", "DEL myzset"},
			expected: []interface{}{int64(3), []interface{}{"member1", "1", "member2", "2"}, int64(1)},
		},
		{
			name:     "ZPOPMIN with count argument but multiple members have the same score",
			commands: []string{"ZADD myzset 1 member1 1 member2 1 member3", "ZPOPMIN myzset 2", "DEL myzset"},
			expected: []interface{}{int64(3), []interface{}{"member1", "1", "member2", "1"}, int64(1)},
		},
		{
			name:     "ZPOPMIN with negative count argument",
			commands: []string{"ZADD myzset 1 member1 2 member2 3 member3", "ZPOPMIN myzset -1", "DEL myzset"},
			expected: []interface{}{int64(3), []interface{}{}, int64(1)},
		},
		{
			name:     "ZPOPMIN with invalid count argument",
			commands: []string{"ZADD myzset 1 member1", "ZPOPMIN myzset INCORRECT_COUNT_ARGUMENT", "DEL myzset"},
			expected: []interface{}{int64(1), "ERR value is not an integer or out of range", int64(1)},
		},
		{
			name:     "ZPOPMIN with count argument greater than length of sorted set",
			commands: []string{"ZADD myzset 1 member1 2 member2 3 member3", "ZPOPMIN myzset 10", "DEL myzset"},
			expected: []interface{}{int64(3), []interface{}{"member1", "1", "member2", "2", "member3", "3"}, int64(1)},
		},
		{
			name:     "ZPOPMIN on empty sorted set",
			commands: []string{"ZADD myzset 1 member1", "ZPOPMIN myzset 1", "ZPOPMIN myzset", "DEL myzset"},
			expected: []interface{}{int64(1), []interface{}{"member1", "1"}, []interface{}{}, int64(1)},
		},
		{
			name:     "ZPOPMIN with floating-point scores",
			commands: []string{"ZADD myzset 1.5 member1 2.7 member2 3.8 member3", "ZPOPMIN myzset", "DEL myzset"},
			expected: []interface{}{int64(3), []interface{}{"member1", "1.5"}, int64(1)},
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

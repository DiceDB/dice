package resp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestZPOPMax(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	testCases := []TestCase{
		{
			name:     "ZPOPMAX on non-existing key with/without count argument",
			commands: []string{"ZPOPMAX NON_EXISTENT_KEY"},
			expected: []interface{}{[]interface{}{}},
		},
		{
			name:     "ZPOPMAX with wrong type of key with/without count argument",
			commands: []string{"SET stringkey string_value", "ZPOPMAX stringkey"},
			expected: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
		{
			name:     "ZPOPMAX on existing key (without count argument)",
			commands: []string{"ZADD sortedSet 1 member1 2 member2 3 member3", "ZPOPMAX sortedSet", "ZCOUNT sortedSet  1 10"},
			expected: []interface{}{int64(3), []interface{}{"member3", "3"}, int64(2)},
		},
		{
			name:     "ZPOPMAX with normal count argument",
			commands: []string{"ZADD sortedSet 1 member1 2 member2 3 member3", "ZPOPMAX sortedSet 2", "ZCOUNT sortedSet  1 2"},
			expected: []interface{}{int64(3), []interface{}{"member3", "3", "member2", "2"}, int64(1)},
		},
		{
			name:     "ZPOPMAX with count argument but multiple members have the same score",
			commands: []string{"ZADD sortedSet 1 member1 1 member2 1 member3", "ZPOPMAX sortedSet 2", "ZCOUNT sortedSet 1 1"},
			expected: []interface{}{int64(3), []interface{}{"member3", "1", "member2", "1"}, int64(1)},
		},
		{
			name:     "ZPOPMAX with negative count argument",
			commands: []string{"ZADD sortedSet 1 member1 2 member2 3 member3", "ZPOPMAX sortedSet -1", "ZCOUNT sortedSet  0.6 3.231"},
			expected: []interface{}{int64(3), []interface{}{}, int64(3)},
		},
		{
			name:     "ZPOPMAX with invalid count argument",
			commands: []string{"ZADD sortedSet 1 member1", "ZPOPMAX sortedSet INCORRECT_COUNT_ARGUMENT", "ZCOUNT sortedSet  1 10"},
			expected: []interface{}{int64(1), "ERR value is out of range, must be positive", int64(1)},
		},
		{
			name:     "ZPOPMAX with count argument greater than length of sorted set",
			commands: []string{"ZADD sortedSet 1 member1 2 member2 3 member3", "ZPOPMAX sortedSet 10", "ZCOUNT sortedSet  1 10"},
			expected: []interface{}{int64(3), []interface{}{"member3", "3", "member2", "2", "member1", "1"}, int64(0)},
		},
		{
			name:     "ZPOPMAX on empty sorted set",
			commands: []string{"ZADD sortedSet 1 member1", "ZPOPMAX sortedSet 1", "ZPOPMAX sortedSet", "ZCOUNT sortedSet  0 10000"},
			expected: []interface{}{int64(1), []interface{}{"member1", "1"}, []interface{}{}, int64(0)},
		},
		{
			name:     "ZPOPMAX with floating-point scores",
			commands: []string{"ZADD sortedSet 1.5 member1 2.7 member2 3.8 member3", "ZPOPMAX sortedSet", "ZCOUNT sortedSet 1.499 2.711"},
			expected: []interface{}{int64(3), []interface{}{"member3", "3.8"}, int64(2)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result := FireCommand(conn, cmd)
				assert.Equal(t, tc.expected[i], result)
			}
			FireCommand(conn, "DEL sortedSet")
		})
	}
}

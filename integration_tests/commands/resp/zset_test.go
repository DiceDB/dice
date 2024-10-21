package resp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
			commands: []string{"ZADD myzset 1 member1 2 member2 3 member3", "ZPOPMIN myzset"},
			expected: []interface{}{int64(3), []interface{}{"member1", "1"}},
		},
		{
			name:     "ZPOPMIN with normal count argument",
			commands: []string{"ZADD myzset 1 member1 2 member2 3 member3", "ZPOPMIN myzset 2"},
			expected: []interface{}{int64(3), []interface{}{"member1", "1", "member2", "2"}},
		},
		{
			name:     "ZPOPMIN with count argument but multiple members have the same score",
			commands: []string{"ZADD myzset 1 member1 1 member2 1 member3", "ZPOPMIN myzset 2"},
			expected: []interface{}{int64(3), []interface{}{"member1", "1", "member2", "1"}},
		},
		{
			name:     "ZPOPMIN with negative count argument",
			commands: []string{"ZADD myzset 1 member1 2 member2 3 member3", "ZPOPMIN myzset -1"},
			expected: []interface{}{int64(3), []interface{}{}},
		},
		{
			name:     "ZPOPMIN with invalid count argument",
			commands: []string{"ZADD myzset 1 member1", "ZPOPMIN myzset INCORRECT_COUNT_ARGUMENT"},
			expected: []interface{}{int64(1), "ERR value is not an integer or out of range"},
		},
		{
			name:     "ZPOPMIN with count argument greater than length of sorted set",
			commands: []string{"ZADD myzset 1 member1 2 member2 3 member3", "ZPOPMIN myzset 10"},
			expected: []interface{}{int64(3), []interface{}{"member1", "1", "member2", "2", "member3", "3"}},
		},
		{
			name:     "ZPOPMIN on empty sorted set",
			commands: []string{"ZADD myzset 1 member1", "ZPOPMIN myzset 1", "ZPOPMIN myzset"},
			expected: []interface{}{int64(1), []interface{}{"member1", "1"}, []interface{}{}},
		},
		{
			name:     "ZPOPMIN with floating-point scores",
			commands: []string{"ZADD myzset 1.5 member1 2.7 member2 3.8 member3", "ZPOPMIN myzset"},
			expected: []interface{}{int64(3), []interface{}{"member1", "1.5"}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			FireCommand(conn, "DEL myzset")
			for i, cmd := range tc.commands {
				result := FireCommand(conn, cmd)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}
}

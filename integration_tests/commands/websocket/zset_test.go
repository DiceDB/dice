package websocket

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestZPOPMIN(t *testing.T) {
	exec := NewWebsocketCommandExecutor()

	testCases := []TestCase{
		{
			name:     "ZPOPMIN on non-existing key with/without count argument",
			commands: []string{"ZPOPMIN NON_EXISTENT_KEY"},
			expected: []interface{}{[]interface{}{}},
		},
		{
			name:     "ZPOPMIN with wrong type of key with/without count argument",
			commands: []string{"SET stringkey string_value", "ZPOPMIN stringkey", "DEL stringkey"},
			expected: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value", float64(1)},
		},
		{
			name:     "ZPOPMIN on existing key (without count argument)",
			commands: []string{"ZADD myzset 1 member1 2 member2 3 member3", "ZPOPMIN myzset"},
			expected: []interface{}{float64(3), []interface{}{"member1", "1"}},
		},
		{
			name:     "ZPOPMIN with normal count argument",
			commands: []string{"ZADD myzset 1 member1 2 member2 3 member3", "ZPOPMIN myzset 2"},
			expected: []interface{}{float64(3), []interface{}{"member1", "1", "member2", "2"}},
		},
		{
			name:     "ZPOPMIN with count argument but multiple members have the same score",
			commands: []string{"ZADD myzset 1 member1 1 member2 1 member3", "ZPOPMIN myzset 2"},
			expected: []interface{}{float64(3), []interface{}{"member1", "1", "member2", "1"}},
		},
		{
			name:     "ZPOPMIN with negative count argument",
			commands: []string{"ZADD myzset 1 member1 2 member2 3 member3", "ZPOPMIN myzset -1"},
			expected: []interface{}{float64(3), []interface{}{}},
		},
		{
			name:     "ZPOPMIN with invalid count argument",
			commands: []string{"ZADD myzset 1 member1", "ZPOPMIN myzset INCORRECT_COUNT_ARGUMENT"},
			expected: []interface{}{float64(1), "ERR value is not an integer or out of range"},
		},
		{
			name:     "ZPOPMIN with count argument greater than length of sorted set",
			commands: []string{"ZADD myzset 1 member1 2 member2 3 member3", "ZPOPMIN myzset 10"},
			expected: []interface{}{float64(3), []interface{}{"member1", "1", "member2", "2", "member3", "3"}},
		},
		{
			name:     "ZPOPMIN on empty sorted set",
			commands: []string{"ZADD myzset 1 member1", "ZPOPMIN myzset 1", "ZPOPMIN myzset"},
			expected: []interface{}{float64(1), []interface{}{"member1", "1"}, []interface{}{}},
		},
		{
			name:     "ZPOPMIN with floating-point scores",
			commands: []string{"ZADD myzset 1.5 member1 2.7 member2 3.8 member3", "ZPOPMIN myzset"},
			expected: []interface{}{float64(3), []interface{}{"member1", "1.5"}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			conn := exec.ConnectToServer()

			DeleteKey(t, conn, exec, "myzset")

			for i, cmd := range tc.commands {
				result, err := exec.FireCommandAndReadResponse(conn, cmd)
				assert.Nil(t, err)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}
}

func TestZCOUNT(t *testing.T) {
	exec := NewWebsocketCommandExecutor()

	testCases := []TestCase{
		{
			name:     "ZCOUNT on non-existing key",
			commands: []string{"ZCOUNT NON_EXISTENT_KEY 0 10"},
			expected: []interface{}{float64(0)},
		},
		{
			name:     "ZCOUNT on key with wrong type",
			commands: []string{"SET stringkey string_value", "ZCOUNT stringkey 0 10", "DEL stringkey"},
			expected: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value", float64(1)},
		},
		{
			name:     "ZCOUNT on existing key with valid range",
			commands: []string{"ZADD myzset 1 member1 5 member2 10 member3", "ZCOUNT myzset 1 5"},
			expected: []interface{}{float64(3), float64(2)}, // ZADD returns 3, ZCOUNT should return 2
		},
		{
			name:     "ZCOUNT with min and max outside the range of elements",
			commands: []string{"ZADD myzset 1 member1 5 member2 10 member3", "ZCOUNT myzset -10 -1"},
			expected: []interface{}{float64(3), float64(0)}, // ZADD returns 3, ZCOUNT should return 0
		},
		{
			name:     "ZCOUNT with min greater than max",
			commands: []string{"ZADD myzset 1 member1 5 member2 10 member3", "ZCOUNT myzset 10 5"},
			expected: []interface{}{float64(3), float64(0)}, // ZCOUNT with invalid range should return 0
		},
		{
			name:     "ZCOUNT with negative scores and valid range",
			commands: []string{"ZADD myzset -5 member1 0 member2 5 member3", "ZCOUNT myzset -10 0"},
			expected: []interface{}{float64(3), float64(2)}, // ZADD returns 3, ZCOUNT should return 2
		},
		{
			name:     "ZCOUNT with floating-point scores",
			commands: []string{"ZADD myzset 1.5 member1 2.7 member2 3.8 member3", "ZCOUNT myzset 1 3"},
			expected: []interface{}{float64(3), float64(2)}, // ZCOUNT should count 2 elements within the range
		},
		{
			name:     "ZCOUNT with exact matching min and max",
			commands: []string{"ZADD myzset 1 member1 2 member2 3 member3", "ZCOUNT myzset 2 2"},
			expected: []interface{}{float64(3), float64(1)}, // ZCOUNT should return 1 (member2)
		},
		{
			name:     "ZCOUNT on an empty sorted set",
			commands: []string{"ZADD myzset 1 member1", "DEL myzset", "ZCOUNT myzset 0 10"},
			expected: []interface{}{float64(1), float64(1), float64(0)}, // DEL returns 1, ZCOUNT returns 0
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			conn := exec.ConnectToServer()

			DeleteKey(t, conn, exec, "myzset")

			//posrcleanup

			t.Cleanup(func() {
				DeleteKey(t, conn, exec, "myzset")
			})

			for i, cmd := range tc.commands {
				result, err := exec.FireCommandAndReadResponse(conn, cmd)
				assert.Nil(t, err)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}
}

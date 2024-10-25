package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestZPOPMIN(t *testing.T) {
	exec := NewHTTPCommandExecutor()
	testCases := []TestCase{
		{
			name: "ZPOPMIN on non-existing key with/without count argument",
			commands: []HTTPCommand{
				{Command: "ZPOPMIN", Body: map[string]interface{}{"key": "NON_EXISTENT_KEY"}},
			},
			expected: []interface{}{[]interface{}{}},
		},
		{
			name: "ZPOPMIN with wrong type of key with/without count argument",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "stringkey", "value": "string_value"}},
				{Command: "ZPOPMIN", Body: map[string]interface{}{"key": "stringkey"}},
			},
			expected: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value", float64(1)},
		},
		{
			name: "ZPOPMIN on existing key (without count argument)",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset", "values": [...]string{"1", "member1", "2", "member2", "3", "member3"}}},
				{Command: "ZPOPMIN", Body: map[string]interface{}{"key": "myzset"}},
			},
			expected: []interface{}{float64(3), []interface{}{"member1", "1"}, float64(1)},
		},
		{
			name: "ZPOPMIN with normal count argument",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset", "values": [...]string{"1", "member1", "2", "member2", "3", "member3"}}},
				{Command: "ZPOPMIN", Body: map[string]interface{}{"key": "myzset", "value": int64(2)}},
			},
			expected: []interface{}{float64(3), []interface{}{"member1", "1", "member2", "2"}, float64(1)},
		},
		{
			name: "ZPOPMIN with count argument but multiple members have the same score",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset", "values": [...]string{"1", "member1", "1", "member2", "1", "member3"}}},
				{Command: "ZPOPMIN", Body: map[string]interface{}{"key": "myzset", "value": int64(2)}},
			},
			expected: []interface{}{float64(3), []interface{}{"member1", "1", "member2", "1"}, float64(1)},
		},
		{
			name: "ZPOPMIN with negative count argument",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset", "values": [...]string{"1", "member1", "2", "member2", "3", "member3"}}},
				{Command: "ZPOPMIN", Body: map[string]interface{}{"key": "myzset", "value": int64(-1)}},
			},
			expected: []interface{}{float64(3), []interface{}{}, float64(1)},
		},
		{
			name: "ZPOPMIN with invalid count argument",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset", "values": [...]string{"1", "member1"}}},
				{Command: "ZPOPMIN", Body: map[string]interface{}{"key": "myzset", "value": "INCORRECT_COUNT_ARGUMENT"}},
			},
			expected: []interface{}{float64(1), "ERR value is not an integer or out of range", float64(1)},
		},
		{
			name: "ZPOPMIN with count argument greater than length of sorted set",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset", "values": [...]string{"1", "member1", "2", "member2", "3", "member3"}}},
				{Command: "ZPOPMIN", Body: map[string]interface{}{"key": "myzset", "value": int64(10)}},
			},
			expected: []interface{}{float64(3), []interface{}{"member1", "1", "member2", "2", "member3", "3"}, float64(1)},
		},
		{
			name: "ZPOPMIN on empty sorted set",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset", "values": [...]string{"1", "member1"}}},
				{Command: "ZPOPMIN", Body: map[string]interface{}{"key": "myzset", "value": int64(1)}},
				{Command: "ZPOPMIN", Body: map[string]interface{}{"key": "myzset"}},
			},
			expected: []interface{}{float64(1), []interface{}{"member1", "1"}, []interface{}{}, float64(1)},
		},
		{
			name: "ZPOPMIN with floating-point scores",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset", "values": [...]string{"1.5", "member1", "2.7", "member2", "3.8", "member3"}}},
				{Command: "ZPOPMIN", Body: map[string]interface{}{"key": "myzset"}},
			},
			expected: []interface{}{float64(3), []interface{}{"member1", "1.5"}, float64(1)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			exec.FireCommand(HTTPCommand{
				Command: "DEL",
				Body:    map[string]interface{}{"key": "myzset"},
			})
			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)

				assert.Equal(t, tc.expected[i], result)
			}
		})
	}
}

func TestZCOUNT(t *testing.T) {
	exec := NewHTTPCommandExecutor()
	testCases := []struct {
		name     string
		commands []HTTPCommand
		expected []interface{}
	}{
		{
			name: "ZCOUNT on non-existent key",
			commands: []HTTPCommand{
				{Command: "ZCOUNT", Body: map[string]interface{}{"key": "NON_EXISTENT_KEY", "values": [...]string{"0", "100"}}},
			},
			expected: []interface{}{float64(0)}, // Expecting count of 0
		},
		{
			name: "ZCOUNT on empty sorted set",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset", "values": [...]string{"10", "member1"}}},
				{Command: "ZCOUNT", Body: map[string]interface{}{"key": "myzset", "values": [...]string{"0", "5"}}},
			},
			expected: []interface{}{float64(1), float64(0)}, // Expecting ZADD to return 1, ZCOUNT to return 0
		},
		{
			name: "ZCOUNT with invalid range",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset", "values": [...]string{"10", "member1"}}},
				{Command: "ZCOUNT", Body: map[string]interface{}{"key": "myzset", "values": [...]string{"20", "5"}}},
			},
			expected: []interface{}{float64(1), float64(0)}, // Expecting ZADD to return 1, ZCOUNT to return 0
		},
		{
			name: "ZCOUNT with valid key and range",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset", "values": [...]string{"10", "member1", "20", "member2", "30", "member3"}}},
				{Command: "ZCOUNT", Body: map[string]interface{}{"key": "myzset", "values": [...]string{"15", "25"}}},
			},
			expected: []interface{}{float64(3), float64(1)}, // Expecting ZADD to return 3, ZCOUNT to return 1
		},
		{
			name: "ZCOUNT with min and max values outside existing members",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset", "values": [...]string{"5", "member1", "15", "member2", "25", "member3"}}},
				{Command: "ZCOUNT", Body: map[string]interface{}{"key": "myzset", "values": [...]string{"30", "50"}}},
			},
			expected: []interface{}{float64(3), float64(0)}, // Expecting ZADD to return 3, ZCOUNT to return 0
		},
		{
			name: "ZCOUNT with multiple members having the same score",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset", "values": [...]string{"10", "member1", "10", "member2", "10", "member3"}}},
				{Command: "ZCOUNT", Body: map[string]interface{}{"key": "myzset", "values": [...]string{"0", "20"}}},
			},
			expected: []interface{}{float64(3), float64(3)}, // Expecting ZADD to return 3, ZCOUNT to return 3
		},
		{
			name: "ZCOUNT with count argument exceeding the number of members",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset", "values": [...]string{"1", "member1", "2", "member2"}}},
				{Command: "ZCOUNT", Body: map[string]interface{}{"key": "myzset", "values": [...]string{"3", "10"}}},
			},
			expected: []interface{}{float64(2), float64(0)}, // Expecting ZADD to return 2, ZCOUNT to return 0
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			exec.FireCommand(HTTPCommand{
				Command: "DEL",
				Body:    map[string]interface{}{"key": "myzset"},
			})

			// Ensure post-test cleanup
			t.Cleanup(func() {
				exec.FireCommand(HTTPCommand{
					Command: "DEL",
					Body:    map[string]interface{}{"key": "myzset"},
				})
				t.Log("Pre-test cleanup executed: Deleted key 'myzset'")
			})

			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)

				assert.Equal(t, tc.expected[i], result)
			}
		})
	}
}

package http

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestZPOPMAX(t *testing.T) {
	exec := NewHTTPCommandExecutor()
	testCases := []TestCase{
		{
			name: "ZPOPMAX on non-existing key with/without count argument",
			commands: []HTTPCommand{
				{Command: "ZPOPMAX", Body: map[string]interface{}{"key": "invalidTest"}},
			},
			expected: []interface{}{[]interface{}{}},
		},
		{
			name: "ZPOPMAX with wrong type of key with/without count argument",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "sortedSet", "value": "testString"}},
				{Command: "ZPOPMAX", Body: map[string]interface{}{"key": "sortedSet"}},
			},
			expected: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
		{
			name: "ZPOPMAX on existing key (without count argument)",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "sortedSet", "values": [...]string{"1", "member1", "2", "member2", "3", "member3"}}},
				{Command: "ZPOPMAX", Body: map[string]interface{}{"key": "sortedSet"}},
			},
			expected: []interface{}{float64(3), []interface{}{"member3", "3"}},
		},
		{
			name: "ZPOPMAX with normal count argument",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "sortedSet", "values": [...]string{"1", "member1", "2", "member2", "3", "member3"}}},
				{Command: "ZPOPMAX", Body: map[string]interface{}{"key": "sortedSet", "value": int64(2)}},
			},
			expected: []interface{}{float64(3), []interface{}{"member3", "3", "member2", "2"}},
		},
		{
			name: "ZPOPMAX with count argument but multiple members have the same score",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "sortedSet", "values": [...]string{"1", "member1", "1", "member2", "1", "member3"}}},
				{Command: "ZPOPMAX", Body: map[string]interface{}{"key": "sortedSet", "value": int64(2)}},
			},
			expected: []interface{}{float64(3), []interface{}{"member3", "1", "member2", "1"}},
		},
		{
			name: "ZPOPMAX with negative count argument",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "sortedSet", "values": [...]string{"1", "member1", "2", "member2", "3", "member3"}}},
				{Command: "ZPOPMAX", Body: map[string]interface{}{"key": "sortedSet", "value": int64(-1)}},
			},
			expected: []interface{}{float64(3), []interface{}{}},
		},
		{
			name: "ZPOPMAX with invalid count argument",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "sortedSet", "values": [...]string{"1", "member1"}}},
				{Command: "ZPOPMAX", Body: map[string]interface{}{"key": "sortedSet", "value": "INCORRECT_COUNT_ARGUMENT"}},
			},
			expected: []interface{}{float64(1), "ERR value is not an integer or out of range"},
		},
		{
			name: "ZPOPMAX with count argument greater than length of sorted set",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "sortedSet", "values": [...]string{"1", "member1", "2", "member2", "3", "member3"}}},
				{Command: "ZPOPMAX", Body: map[string]interface{}{"key": "sortedSet", "value": int64(10)}},
			},
			expected: []interface{}{float64(3), []interface{}{"member3", "3", "member2", "2", "member1", "1"}},
		},
		{
			name: "ZPOPMAX with floating-point scores",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "sortedSet", "values": [...]string{"1.5", "member1", "2.7", "member2", "3.8", "member3"}}},
				{Command: "ZPOPMAX", Body: map[string]interface{}{"key": "sortedSet"}},
			},
			expected: []interface{}{float64(3), []interface{}{"member3", "3.8"}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			exec.FireCommand(HTTPCommand{
				Command: "DEL",
				Body:    map[string]interface{}{"key": "sortedSet"},
			})
			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)
				assert.DeepEqual(t, tc.expected[i], result)
			}
		})
	}
}

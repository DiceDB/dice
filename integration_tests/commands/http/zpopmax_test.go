package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
				{Command: "ZCOUNT", Body: map[string]interface{}{"key": "sortedSet", "values": [...]string{"1", "2.98"}}},
			},
			expected: []interface{}{float64(3), []interface{}{"member3", "3"}, float64(2)},
		},
		{
			name: "ZPOPMAX with normal count argument",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "sortedSet", "values": [...]string{"1", "member1", "2", "member2", "3", "member3"}}},
				{Command: "ZPOPMAX", Body: map[string]interface{}{"key": "sortedSet", "value": int64(2)}},
				{Command: "ZCOUNT", Body: map[string]interface{}{"key": "sortedSet", "values": [...]string{"0.44", "2"}}},
			},
			expected: []interface{}{float64(3), []interface{}{"member3", "3", "member2", "2"}, float64(1)},
		},
		{
			name: "ZPOPMAX with count argument but multiple members have the same score",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "sortedSet", "values": [...]string{"1", "member1", "1", "member2", "1", "member3"}}},
				{Command: "ZPOPMAX", Body: map[string]interface{}{"key": "sortedSet", "value": int64(2)}},
				{Command: "ZCOUNT", Body: map[string]interface{}{"key": "sortedSet", "values": [...]string{"1", "2"}}},
			},
			expected: []interface{}{float64(3), []interface{}{"member3", "1", "member2", "1"}, float64(1)},
		},
		{
			name: "ZPOPMAX with negative count argument",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "sortedSet", "values": [...]string{"1", "member1", "2", "member2", "3", "member3"}}},
				{Command: "ZPOPMAX", Body: map[string]interface{}{"key": "sortedSet", "value": int64(-1)}},
				{Command: "ZCOUNT", Body: map[string]interface{}{"key": "sortedSet", "values": [...]string{"1", "1000"}}},
			},
			expected: []interface{}{float64(3), []interface{}{}, float64(3)},
		},
		{
			name: "ZPOPMAX with invalid count argument",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "sortedSet", "values": [...]string{"1", "member1"}}},
				{Command: "ZPOPMAX", Body: map[string]interface{}{"key": "sortedSet", "value": "INCORRECT_COUNT_ARGUMENT"}},
				{Command: "ZCOUNT", Body: map[string]interface{}{"key": "sortedSet", "values": [...]string{"1", "2"}}},
			},
			expected: []interface{}{float64(1), "ERR value is out of range, must be positive", float64(1)},
		},
		{
			name: "ZPOPMAX with count argument greater than length of sorted set",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "sortedSet", "values": [...]string{"1", "member1", "2", "member2", "3", "member3"}}},
				{Command: "ZPOPMAX", Body: map[string]interface{}{"key": "sortedSet", "value": int64(10)}},
				{Command: "ZCOUNT", Body: map[string]interface{}{"key": "sortedSet", "values": [...]string{"1", "2"}}},
			},
			expected: []interface{}{float64(3), []interface{}{"member3", "3", "member2", "2", "member1", "1"}, float64(0)},
		},
		{
			name: "ZPOPMAX with floating-point scores",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "sortedSet", "values": [...]string{"1.5", "member1", "2.7", "member2", "3.8", "member3"}}},
				{Command: "ZPOPMAX", Body: map[string]interface{}{"key": "sortedSet"}},
				{Command: "ZCOUNT", Body: map[string]interface{}{"key": "sortedSet", "values": [...]string{"1.3", "3.6"}}},
			},
			expected: []interface{}{float64(3), []interface{}{"member3", "3.8"}, float64(2)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)
				assert.Equal(t, tc.expected[i], result)
			}
			exec.FireCommand(HTTPCommand{
				Command: "DEL",
				Body:    map[string]interface{}{"key": "sortedSet"},
			})
		})
	}
}

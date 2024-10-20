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

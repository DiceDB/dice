// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

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
				{Command: "ZCOUNT", Body: map[string]interface{}{"key": "myzset", "values": [...]string{"1", "2.98"}}},
			},
			expected: []interface{}{float64(3), []interface{}{"member1", "1"}, float64(1)},
		},
		{
			name: "ZPOPMIN with normal count argument",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset", "values": [...]string{"1", "member1", "2", "member2", "3", "member3"}}},
				{Command: "ZPOPMIN", Body: map[string]interface{}{"key": "myzset", "value": int64(2)}},
				{Command: "ZCOUNT", Body: map[string]interface{}{"key": "myzset", "values": [...]string{"0.44", "2"}}},
			},
			expected: []interface{}{float64(3), []interface{}{"member1", "1", "member2", "2"}, float64(0)},
		},
		{
			name: "ZPOPMIN with count argument but multiple members have the same score",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset", "values": [...]string{"1", "member1", "1", "member2", "1", "member3"}}},
				{Command: "ZPOPMIN", Body: map[string]interface{}{"key": "myzset", "value": int64(2)}},
				{Command: "ZCOUNT", Body: map[string]interface{}{"key": "myzset", "values": [...]string{"1", "2"}}},
			},
			expected: []interface{}{float64(3), []interface{}{"member1", "1", "member2", "1"}, float64(1)},
		},
		{
			name: "ZPOPMIN with negative count argument",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset", "values": [...]string{"1", "member1", "2", "member2", "3", "member3"}}},
				{Command: "ZPOPMIN", Body: map[string]interface{}{"key": "myzset", "value": int64(-1)}},
				{Command: "ZCOUNT", Body: map[string]interface{}{"key": "myzset", "values": [...]string{"1", "1000"}}},
			},
			expected: []interface{}{float64(3), []interface{}{}, float64(3)},
		},
		{
			name: "ZPOPMIN with invalid count argument",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset", "values": [...]string{"1", "member1"}}},
				{Command: "ZPOPMIN", Body: map[string]interface{}{"key": "myzset", "value": "INCORRECT_COUNT_ARGUMENT"}},
				{Command: "ZCOUNT", Body: map[string]interface{}{"key": "myzset", "values": [...]string{"1", "2"}}},
			},
			expected: []interface{}{float64(1), "ERR value is not an integer or out of range", float64(1)},
		},
		{
			name: "ZPOPMIN with count argument greater than length of sorted set",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset", "values": [...]string{"1", "member1", "2", "member2", "3", "member3"}}},
				{Command: "ZPOPMIN", Body: map[string]interface{}{"key": "myzset", "value": int64(10)}},
				{Command: "ZCOUNT", Body: map[string]interface{}{"key": "myzset", "values": [...]string{"1", "2"}}},
			},
			expected: []interface{}{float64(3), []interface{}{"member1", "1", "member2", "2", "member3", "3"}, float64(0)},
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
				{Command: "ZCOUNT", Body: map[string]interface{}{"key": "myzset", "values": [...]string{"1.3", "3.6"}}},
			},
			expected: []interface{}{float64(3), []interface{}{"member1", "1.5"}, float64(1)},
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
				Body:    map[string]interface{}{"key": "myzset"},
			})
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

func TestZADD(t *testing.T) {
	exec := NewHTTPCommandExecutor()
	testCases := []struct {
		name     string
		commands []HTTPCommand
		expected []interface{}
	}{
		{
			name: "ZADD with two new members",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"1", "member1", "2", "member2"}}},
			},
			expected: []interface{}{float64(2)},
		},
		{
			name: "ZADD with three new members",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"3", "member3", "4", "member4", "5", "member5"}}},
			},
			expected: []interface{}{float64(3)},
		},
		{
			name: "ZADD with existing members",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"1", "member1", "2", "member2", "3", "member3", "4", "member4", "5", "member5"}}},
			},
			expected: []interface{}{float64(5)},
		},
		{
			name: "ZADD with mixed new and existing members",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"1", "member1", "2", "member2", "3", "member3", "4", "member4", "5", "member5", "6", "member6"}}},
			},
			expected: []interface{}{float64(6)},
		},
		{
			name: "ZADD without any members",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"1"}}},
			},
			expected: []interface{}{"ERR wrong number of arguments for 'zadd' command"},
		},

		{
			name: "ZADD XX option without existing key",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"XX", "10", "member9"}}},
			},
			expected: []interface{}{float64(0)},
		},
		{
			name: "ZADD XX with existing key and member2",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"XX", "20", "member2"}}},
			},
			expected: []interface{}{float64(0)},
		},
		{
			name: "ZADD XX updates existing elements scores",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"XX", "15", "member1", "25", "member2"}}},
			},
			expected: []interface{}{float64(0)},
		},
		{
			name: "ZADD GT and XX only updates existing elements when new scores are greater",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"GT", "XX", "20", "member1"}}},
			},
			expected: []interface{}{float64(0)},
		},
		{
			name: "ZADD LT and XX only updates existing elements when new scores are lower",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"LT", "XX", "20", "member1"}}},
			},
			expected: []interface{}{float64(0)},
		},
		{
			name: "ZADD NX and XX not compatible",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"NX", "XX", "20", "member1"}}},
			},
			expected: []interface{}{"ERR XX and NX options at the same time are not compatible"},
		},
		{
			name: "ZADD XX and CH compatible",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"XX", "CH", "20", "member1"}}},
			},
			expected: []interface{}{float64(0)},
		},
		{
			name: "ZADD INCR and XX compatible",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"XX", "INCR", "20", "member1"}}},
			},
			expected: []interface{}{float64(0)},
		},
		{
			name: "ZADD INCR and XX not compatible because of more than one member",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"XX", "INCR", "20", "member1", "25", "member2"}}},
			},
			expected: []interface{}{"ERR INCR option supports a single increment-element pair"},
		},
		{
			name: "ZADD XX, LT and GT are not compatible",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"XX", "LT", "GT", "20", "member1"}}},
			},
			expected: []interface{}{"ERR GT, LT, and/or NX options at the same time are not compatible"},
		},
		{
			name: "ZADD XX, LT, GT, CH, INCR are not compatible",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"XX", "LT", "GT", "INCR", "CH", "20", "member1"}}},
			},
			expected: []interface{}{"ERR GT, LT, and/or NX options at the same time are not compatible"},
		},
		{
			name: "ZADD XX, GT and CH compatible",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"XX", "GT", "CH", "60", "member1", "30", "member2"}}},
			},
			expected: []interface{}{float64(0)},
		},
		{
			name: "ZADD XX, LT and CH compatible",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"XX", "LT", "CH", "4", "member1", "1", "member2"}}},
			},
			expected: []interface{}{float64(0)},
		},
		{
			name: "ZADD XX with existing key and new member",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"XX", "20", "member20"}}},
			},
			expected: []interface{}{float64(0)},
		},
		{
			name: "ZADD XX wont update as new members",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"XX", "15", "member18", "25", "member20"}}},
			},
			expected: []interface{}{float64(0)},
		},
		{
			name: "ZADD XX and GT wont add new member",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"GT", "XX", "20", "member18"}}},
			},
			expected: []interface{}{float64(0)},
		},
		{
			name: "ZADD XX and LT and new member wont update",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"LT", "XX", "20", "member18"}}},
			},
			expected: []interface{}{float64(0)},
		},
		{
			name: "ZADD XX and CH and new member wont work",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"XX", "CH", "20", "member18"}}},
			},
			expected: []interface{}{float64(0)},
		},
		{
			name: "ZADD XX, LT, CH, new member wont update",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"XX", "LT", "CH", "50", "member18", "40", "member20"}}},
			},
			expected: []interface{}{float64(0)},
		},
		{
			name: "ZADD XX, GT and CH, new member wont update",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"XX", "GT", "CH", "60", "member18", "30", "member20"}}},
			},
			expected: []interface{}{float64(0)},
		},

		{
			name: "ZADD NX existing key new member",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"NX", "10", "member9"}}},
			},
			expected: []interface{}{float64(1)},
		},
		{
			name: "ZADD NX existing key old member",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"NX", "20", "member2"}}},
			},
			expected: []interface{}{float64(1)},
		},
		{
			name: "ZADD NX existing key one new member and one old member",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"NX", "15", "member1", "25", "member11"}}},
			},
			expected: []interface{}{float64(2)},
		},
		{
			name: "ZADD NX and XX not compatible",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"NX", "XX", "20", "member1"}}},
			},
			expected: []interface{}{"ERR XX and NX options at the same time are not compatible"},
		},
		{
			name: "ZADD NX XX CH not compatible",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"NX", "XX", "CH", "20", "member1"}}},
			},
			expected: []interface{}{"ERR XX and NX options at the same time are not compatible"},
		},
		{
			name: "ZADD NX XX CH INCR not compatible",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"NX", "XX", "CH", "INCR", "20", "member1"}}},
			},
			expected: []interface{}{"ERR XX and NX options at the same time are not compatible"},
		},
		{
			name: "ZADD NX XX LT not compatible",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"NX", "XX", "LT", "20", "member1"}}},
			},
			expected: []interface{}{"ERR XX and NX options at the same time are not compatible"},
		},
		{
			name: "ZADD NX XX GT not compatible",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"NX", "XX", "GT", "20", "member1"}}},
			},
			expected: []interface{}{"ERR XX and NX options at the same time are not compatible"},
		},
		{
			name: "ZADD NX XX LT CH not compatible",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"NX", "XX", "LT", "CH", "20", "member1"}}},
			},
			expected: []interface{}{"ERR XX and NX options at the same time are not compatible"},
		},
		{
			name: "ZADD NX XX LT CH INCR not compatible",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"NX", "XX", "LT", "CH", "INCR", "20", "member1"}}},
			},
			expected: []interface{}{"ERR XX and NX options at the same time are not compatible"},
		},
		{
			name: "ZADD NX XX GT CH not compatible",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"NX", "XX", "GT", "CH", "20", "member1"}}},
			},
			expected: []interface{}{"ERR XX and NX options at the same time are not compatible"},
		},
		{
			name: "ZADD NX XX GT CH INCR not compatible",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"NX", "XX", "GT", "CH", "INCR", "20", "member1"}}},
			},
			expected: []interface{}{"ERR XX and NX options at the same time are not compatible"},
		},
		{
			name: "ZADD NX XX INCR not compatible",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"NX", "XX", "INCR", "20", "member1"}}},
			},
			expected: []interface{}{"ERR XX and NX options at the same time are not compatible"},
		},
		{
			name: "ZADD NX XX INCR LT not compatible",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"NX", "XX", "INCR", "LT", "20", "member1"}}},
			},
			expected: []interface{}{"ERR XX and NX options at the same time are not compatible"},
		},
		{
			name: "ZADD NX XX INCR GT not compatible",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"NX", "XX", "INCR", "GT", "20", "member1"}}},
			},
			expected: []interface{}{"ERR XX and NX options at the same time are not compatible"},
		},
		{
			name: "ZADD NX XX LT GT not compatible",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"NX", "XX", "LT", "GT", "20", "member1"}}},
			},
			expected: []interface{}{"ERR XX and NX options at the same time are not compatible"},
		},
		{
			name: "ZADD NX XX LT GT CH not compatible",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"NX", "XX", "LT", "GT", "CH", "20", "member1"}}},
			},
			expected: []interface{}{"ERR XX and NX options at the same time are not compatible"},
		},
		{
			name: "ZADD NX XX LT GT CH INCR not compatible",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"NX", "XX", "LT", "GT", "CH", "INCR", "20", "member1"}}},
			},
			expected: []interface{}{"ERR XX and NX options at the same time are not compatible"},
		},

		// NX without XX and all LT GT CH and INCR - all errors
		{
			name: "ZADD NX and GT incompatible",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"NX", "GT", "20", "member1"}}},
			},
			expected: []interface{}{"ERR GT, LT, and/or NX options at the same time are not compatible"},
		},
		{
			name: "ZADD NX and LT incompatible",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"NX", "LT", "20", "member1"}}},
			},
			expected: []interface{}{"ERR GT, LT, and/or NX options at the same time are not compatible"},
		},
		{
			name: "ZADD NX, LT and GT incompatible",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"NX", "LT", "GT", "20", "member1"}}},
			},
			expected: []interface{}{"ERR GT, LT, and/or NX options at the same time are not compatible"},
		},
		{
			name: "ZADD NX, LT, GT and INCR incompatible",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"NX", "LT", "GT", "INCR", "20", "member1"}}},
			},
			expected: []interface{}{"ERR GT, LT, and/or NX options at the same time are not compatible"},
		},
		{
			name: "ZADD NX, LT, GT and CH incompatible",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"NX", "LT", "GT", "CH", "20", "member1"}}},
			},
			expected: []interface{}{"ERR GT, LT, and/or NX options at the same time are not compatible"},
		},
		{
			name: "ZADD NX, LT, GT, CH and INCR incompatible",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"NX", "LT", "GT", "CH", "INCR", "20", "member1"}}},
			},
			expected: []interface{}{"ERR GT, LT, and/or NX options at the same time are not compatible"},
		},
		{
			name: "ZADD NX, LT, CH not compatible",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"NX", "LT", "CH", "20", "member1"}}},
			},
			expected: []interface{}{"ERR GT, LT, and/or NX options at the same time are not compatible"},
		},
		{
			name: "ZADD NX, LT, INCR not compatible",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"NX", "LT", "INCR", "20", "member1"}}},
			},
			expected: []interface{}{"ERR GT, LT, and/or NX options at the same time are not compatible"},
		},
		{
			name: "ZADD NX, LT, CH, INCR not compatible",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"NX", "LT", "CH", "INCR", "20", "member1"}}},
			},
			expected: []interface{}{"ERR GT, LT, and/or NX options at the same time are not compatible"},
		},
		{
			name: "ZADD NX, GT, CH not compatible",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"NX", "GT", "CH", "20", "member1"}}},
			},
			expected: []interface{}{"ERR GT, LT, and/or NX options at the same time are not compatible"},
		},
		{
			name: "ZADD NX, GT, INCR not compatible",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"NX", "GT", "INCR", "20", "member1"}}},
			},
			expected: []interface{}{"ERR GT, LT, and/or NX options at the same time are not compatible"},
		},
		{
			name: "ZADD NX, GT, CH, INCR not compatible",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"NX", "GT", "CH", "INCR", "20", "member1"}}},
			},
			expected: []interface{}{"ERR GT, LT, and/or NX options at the same time are not compatible"},
		},
		{
			name: "ZADD NX, CH with new member returns CH based - if added or not",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"NX", "CH", "20", "member13"}}},
			},
			expected: []interface{}{float64(1)},
		},
		{
			name: "ZADD NX, CH with existing member returns CH based - if added or not",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"NX", "CH", "10", "member13"}}},
			},
			expected: []interface{}{float64(1)},
		},

		{
			name: "ZADD with GT with existing member",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"GT", "15", "member14"}}},
			},
			expected: []interface{}{float64(1)},
		},
		{
			name: "ZADD with GT with new member",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"GT", "15", "member15"}}},
			},
			expected: []interface{}{float64(1)},
		},
		{
			name: "ZADD GT and LT",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"GT", "LT", "15", "member15"}}},
			},
			expected: []interface{}{"ERR GT, LT, and/or NX options at the same time are not compatible"},
		},
		{
			name: "ZADD GT LT CH",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"GT", "LT", "CH", "15", "member15"}}},
			},
			expected: []interface{}{"ERR GT, LT, and/or NX options at the same time are not compatible"},
		},
		{
			name: "ZADD GT LT CH INCR",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"GT", "LT", "CH", "INCR", "15", "member15"}}},
			},
			expected: []interface{}{"ERR GT, LT, and/or NX options at the same time are not compatible"},
		},
		{
			name: "ZADD GT LT INCR",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"GT", "LT", "INCR", "15", "member15"}}},
			},
			expected: []interface{}{"ERR GT, LT, and/or NX options at the same time are not compatible"},
		},
		{
			name: "ZADD GT CH with existing member score less no change hence 0",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"GT", "CH", "10", "member15"}}},
			},
			expected: []interface{}{float64(1)},
		},
		{
			name: "ZADD GT CH with existing member score more, changed score hence 1",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"GT", "CH", "25", "member15"}}},
			},
			expected: []interface{}{float64(1)},
		},
		{
			name: "ZADD GT CH with existing member score equal, nothing returned",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"GT", "CH", "25", "member15"}}},
			},
			expected: []interface{}{float64(1)},
		},
		{
			name: "ZADD GT CH with new member score",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"GT", "CH", "5", "member19"}}},
			},
			expected: []interface{}{float64(1)},
		},
		{
			name: "ZADD GT with INCR if score less than current score after INCR returns nil",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"GT", "INCR", "-5", "member15"}}},
			},
			expected: []interface{}{float64(-5)},
		},
		{
			name: "ZADD GT with INCR updates existing member score if greater after INCR",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"GT", "INCR", "5", "member15"}}},
			},
			expected: []interface{}{float64(5)},
		},

		// ZADD with LT options
		{
			name: "ZADD with LT with existing member score greater",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"LT", "15", "member14"}}},
			},
			expected: []interface{}{float64(1)},
		},
		{
			name: "ZADD with LT with new member",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"LT", "15", "member23"}}},
			},
			expected: []interface{}{float64(1)},
		},
		{
			name: "ZADD LT with existing member score equal",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"LT", "15", "member14"}}},
			},
			expected: []interface{}{float64(1)},
		},
		{
			name: "ZADD LT with existing member score less",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"LT", "10", "member14"}}},
			},
			expected: []interface{}{float64(1)},
		},
		{
			name: "ZADD LT with INCR does not update if score is greater after INCR",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"LT", "INCR", "5", "member14"}}},
			},
			expected: []interface{}{float64(5)},
		},
		{
			name: "ZADD LT with INCR updates if updated score is less than current",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"LT", "INCR", "-1", "member14"}}},
			},
			expected: []interface{}{float64(-1)},
		},
		{
			name: "ZADD LT with CH updates existing member score if less, CH returns changed elements",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"LT", "CH", "5", "member1", "2", "member2"}}},
			},
			expected: []interface{}{float64(2)},
		},
		// ZADD with INCR options
		{
			name: "ZADD INCR with new members, inserts as it is",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"INCR", "15", "member24"}}},
			},
			expected: []interface{}{float64(15)},
		},
		{
			name: "ZADD INCR with existing members, increases the score",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"INCR", "5", "member24"}}},
			},
			expected: []interface{}{float64(5)},
		},
		// ZADD with CH options
		{
			name: "ZADD CH with one existing member update, returns count of updates",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"CH", "45", "member2"}}},
			},
			expected: []interface{}{float64(1)},
		},
		{
			name: "ZADD CH with multiple existing member updates, returns count of updates",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"CH", "50", "member2", "63", "member3"}}},
			},
			expected: []interface{}{float64(2)},
		},
		{
			name: "ZADD CH with 1 new and 1 existing member update, returns count of updates",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "myzset2", "values": [...]string{"CH", "50", "member2", "64", "member32"}}},
			},
			expected: []interface{}{float64(2)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			exec.FireCommand(HTTPCommand{
				Command: "DEL",
				Body:    map[string]interface{}{"key": "myzset2"},
			})

			t.Cleanup(func() {
				exec.FireCommand(HTTPCommand{
					Command: "DEL",
					Body:    map[string]interface{}{"key": "myzset2"},
				})
				t.Log("Pre-test cleanup executed: Deleted key 'myzset2'")
			})

			// Execute test commands and validate results
			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}
}

func TestZRANGE_HTTP(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []struct {
		name     string
		commands []HTTPCommand
		expected []interface{}
	}{
		{
			name: "ZRANGE with mixed indices",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "key", "values": [...]string{"1", "member1", "2", "member2", "3", "member3", "4", "member4", "5", "member5", "6", "member6"}}},
				{Command: "ZRANGE", Body: map[string]interface{}{"key": "key", "values": [...]string{"0", "-1"}}}, // Use start and stop instead of values
			},
			expected: []interface{}{float64(6), []interface{}{"member1", "member2", "member3", "member4", "member5", "member6"}},
		},
		{
			name: "ZRANGE with positive indices #1",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "key", "values": []string{"1", "member1", "2", "member2", "3", "member3", "4", "member4", "5", "member5", "6", "member6"}}},
				{Command: "ZRANGE", Body: map[string]interface{}{"key": "key", "values": [...]string{"0", "2"}}},
			},
			expected: []interface{}{float64(6), []interface{}{"member1", "member2", "member3"}},
		},
		{
			name: "ZRANGE with positive indices #2",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "key", "values": []string{"1", "member1", "2", "member2", "3", "member3", "4", "member4", "5", "member5", "6", "member6"}}},
				{Command: "ZRANGE", Body: map[string]interface{}{"key": "key", "values": [...]string{"2", "4"}}},
			},
			expected: []interface{}{float64(6), []interface{}{"member3", "member4", "member5"}},
		},
		{
			name: "ZRANGE with all positive indices",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "key", "values": []string{"1", "member1", "2", "member2", "3", "member3", "4", "member4", "5", "member5", "6", "member6"}}},
				{Command: "ZRANGE", Body: map[string]interface{}{"key": "key", "values": [...]string{"0", "10"}}},
			},
			expected: []interface{}{float64(6), []interface{}{"member1", "member2", "member3", "member4", "member5", "member6"}},
		},
		{
			name: "ZRANGE with out of bound indices",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "key", "values": []string{"1", "member1", "2", "member2", "3", "member3", "4", "member4", "5", "member5", "6", "member6"}}},
				{Command: "ZRANGE", Body: map[string]interface{}{"key": "key", "values": [...]string{"10", "20"}}},
			},
			expected: []interface{}{float64(6), []interface{}{}},
		},
		{
			name: "ZRANGE with positive indices and scores",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "key", "values": []string{"1", "member1", "2", "member2", "3", "member3", "4", "member4", "5", "member5", "6", "member6"}}},
				{Command: "ZRANGE", Body: map[string]interface{}{"key": "key", "values": [...]string{"0", "10", "WITHSCORES"}}},
			},
			expected: []interface{}{float64(6), []interface{}{"member1", "1", "member2", "2", "member3", "3", "member4", "4", "member5", "5", "member6", "6"}},
		},
		{
			name: "ZRANGE with positive indices and scores in reverse order",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "key", "values": []string{"1", "member1", "2", "member2", "3", "member3", "4", "member4", "5", "member5", "6", "member6"}}},
				{Command: "ZRANGE", Body: map[string]interface{}{"key": "key", "values": [...]string{"0", "10", "REV", "WITHSCORES"}}},
			},
			expected: []interface{}{float64(6), []interface{}{"member6", "6", "member5", "5", "member4", "4", "member3", "3", "member2", "2", "member1", "1"}},
		},
		{
			name: "ZRANGE with negative indices",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "key", "values": []string{"1", "member1", "2", "member2", "3", "member3", "4", "member4", "5", "member5", "6", "member6"}}},
				{Command: "ZRANGE", Body: map[string]interface{}{"key": "key", "values": [...]string{"-1", "-1"}}},
			},
			expected: []interface{}{float64(6), []interface{}{"member6"}},
		},
		{
			name: "ZRANGE with negative indices and scores",
			commands: []HTTPCommand{
				{Command: "ZADD", Body: map[string]interface{}{"key": "key", "values": []string{"1", "member1", "2", "member2", "3", "member3", "4", "member4", "5", "member5", "6", "member6"}}},
				{Command: "ZRANGE", Body: map[string]interface{}{"key": "key", "values": [...]string{"-8", "-5", "WITHSCORES"}}},
			},
			expected: []interface{}{float64(6), []interface{}{"member1", "1", "member2", "2"}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clear previous data
			exec.FireCommand(HTTPCommand{
				Command: "DEL",
				Body:    map[string]interface{}{"key": "key"},
			})

			// Execute commands and validate results
			for i, cmd := range tc.commands {
				result, err := exec.FireCommand(cmd)
				if err != nil {
					t.Fatalf("Error executing command %v: %v", cmd, err)
				}

				// Check if the result matches the expected value
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}
}

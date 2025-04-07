// // Copyright (c) 2022-present, DiceDB contributors
// // All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"errors"
	"testing"
	// "time"
)

func TestZCOUNT(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name:     "ZCOUNT with valid key and range",
			commands: []string{"ZADD myzset 5 member1 7 member2 9 member3", "ZCOUNT myzset 5 10"},
			expected: []interface{}{int64(3), int64(3)}, // All 3 members are in range
		},
		{
			name:     "ZCOUNT with no members in range",
			commands: []string{"ZADD myzset 5 member1 7 member2 9 member3", "ZCOUNT myzset 11 15"},
			expected: []interface{}{int64(3), int64(0)}, // No members in range 11-15
		},
		{
			name:     "ZCOUNT with non-existent key",
			commands: []string{"ZCOUNT myzset 1 5"},
			expected: []interface{}{int64(0)}, // Return 0 for non-existent key
		},
		{
			name:     "ZCOUNT with exact range boundaries",
			commands: []string{"ZADD myzset 5 member1 7 member2 9 member3", "ZCOUNT myzset 5 9"},
			expected: []interface{}{int64(3), int64(3)}, // All members are in range
		},
		{
			name:     "ZCOUNT with partial range match",
			commands: []string{"ZADD myzset 5 member1 7 member2 9 member3", "ZCOUNT myzset 6 8"},
			expected: []interface{}{int64(3), int64(1)}, // Only member2 with score 7 is in range
		},
		{
			name:     "ZCOUNT with wrong number of arguments (too few)",
			commands: []string{"ZADD myzset 5 member1", "ZCOUNT myzset 5"},
			expected: []interface{}{int64(1), errors.New("wrong number of arguments for 'ZCOUNT' command")},
		},
		{
			name:     "ZCOUNT with wrong number of arguments (too many)",
			commands: []string{"ZADD myzset 5 member1", "ZCOUNT myzset 5 10 extra"},
			expected: []interface{}{int64(1), errors.New("wrong number of arguments for 'ZCOUNT' command")},
		},
		{
			name:     "ZCOUNT with invalid number format for min",
			commands: []string{"ZADD myzset 5 member1", "ZCOUNT myzset invalid 10"},
			expected: []interface{}{int64(1), errors.New("value is not an integer or a float")},
		},
		{
			name:     "ZCOUNT with invalid number format for max",
			commands: []string{"ZADD myzset 5 member1", "ZCOUNT myzset 5 invalid"},
			expected: []interface{}{int64(1), errors.New("value is not an integer or a float")},
		},
		{
			name:     "ZCOUNT with wrong type operation",
			commands: []string{"SET wrongtype value", "ZCOUNT wrongtype 1 10"},
			expected: []interface{}{string("OK"), errors.New("wrongtype operation against a key holding the wrong kind of value")},
		},
		{
			name:     "ZCOUNT with negative scores",
			commands: []string{"ZADD myzset -10 neg_ten -5 neg_five 0 zero 5 five", "ZCOUNT myzset -8 2"},
			expected: []interface{}{int64(4), int64(2)}, // Should count -5 and 0
		},
		{
			name:     "ZCOUNT with decimal scores",
			commands: []string{"ZADD myzset 1.5 one_point_five 2.5 two_point_five 3.5 three_point_five", "ZCOUNT myzset 2 3"},
			expected: []interface{}{int64(3), int64(1)}, // Only 2.5 is in range
		},
		{
			name:     "ZCOUNT with min greater than max",
			commands: []string{"ZADD myzset 5 member1 7 member2 9 member3", "ZCOUNT myzset 10 5"},
			expected: []interface{}{int64(3), int64(0)}, // Invalid range should return 0
		},
		{
			name:     "ZCOUNT with single value range",
			commands: []string{"ZADD myzset 5 member1 7 member2 9 member3", "ZCOUNT myzset 7 7"},
			expected: []interface{}{int64(3), int64(1)}, // Only member2 with score 7
		},
		{
			name:     "ZCOUNT with duplicate scores",
			commands: []string{"ZADD myzset 5 member1 5 member2 5 member3 7 member4", "ZCOUNT myzset 5 5"},
			expected: []interface{}{int64(4), int64(3)}, // 3 members with score 5
		},
		{
			name:     "ZCOUNT with min and max on exact member scores",
			commands: []string{"ZADD exactzset 10 member1 20 member2 30 member3 40 member4", "ZCOUNT exactzset 10 30"},
			expected: []interface{}{int64(4), int64(3)}, // Scores 10, 20, 30
		},
		{
			name:     "ZCOUNT with score 0",
			commands: []string{"ZADD zerozset -5 neg_five 0 zero 5 five", "ZCOUNT zerozset -2 2"},
			expected: []interface{}{int64(3), int64(1)}, // Only zero is in range
		},
		{
			name:     "ZCOUNT with range outside all members",
			commands: []string{"ZADD myzset 5 member1 10 member2 15 member3", "ZCOUNT myzset 20 30"},
			expected: []interface{}{int64(3), int64(0)}, // No members in range
		},
	}

	runTestcases(t, client, testCases)
}

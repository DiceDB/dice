// // Copyright (c) 2022-present, DiceDB contributors
// // All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"errors"
	"testing"

	"github.com/dicedb/dicedb-go/wire"
	// "time"
)

func extractValueZCOUNT(res *wire.Result) interface{} {
	return res.GetZCOUNTRes().Count
}

func TestZCOUNT(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name:           "ZCOUNT with valid key and range",
			commands:       []string{"ZADD myzset 5 member1 7 member2 9 member3", "ZCOUNT myzset 5 10"},
			expected:       []interface{}{3, 3}, // All 3 members are in range
			valueExtractor: []ValueExtractorFn{extractValueZADD, extractValueZCOUNT},
		},
		{
			name:           "ZCOUNT with no members in range",
			commands:       []string{"ZADD s100 5 member1 7 member2 9 member3", "ZCOUNT s100 11 15"},
			expected:       []interface{}{3, 0}, // No members in range 11-15
			valueExtractor: []ValueExtractorFn{extractValueZADD, extractValueZCOUNT},
		},
		{
			name:           "ZCOUNT with non-existent key",
			commands:       []string{"ZCOUNT s101 1 5"},
			expected:       []interface{}{0}, // Return 0 for non-existent key
			valueExtractor: []ValueExtractorFn{extractValueZCOUNT},
		},
		{
			name:           "ZCOUNT with exact range boundaries",
			commands:       []string{"ZADD s105 5 member1 7 member2 9 member3", "ZCOUNT s105 5 9"},
			expected:       []interface{}{3, 3}, // All members are in range
			valueExtractor: []ValueExtractorFn{extractValueZADD, extractValueZCOUNT},
		},
		{
			name:           "ZCOUNT with partial range match",
			commands:       []string{"ZADD s102 5 member1 7 member2 9 member3", "ZCOUNT s102 6 8"},
			expected:       []interface{}{3, 1}, // Only member2 with score 7 is in range
			valueExtractor: []ValueExtractorFn{extractValueZADD, extractValueZCOUNT},
		},
		{
			name:           "ZCOUNT with wrong number of arguments (too few)",
			commands:       []string{"ZADD s103 5 member1", "ZCOUNT s103 5"},
			expected:       []interface{}{1, errors.New("wrong number of arguments for 'ZCOUNT' command")},
			valueExtractor: []ValueExtractorFn{extractValueZADD, nil},
		},
		{
			name:           "ZCOUNT with wrong number of arguments (too many)",
			commands:       []string{"ZADD s104 5 member1", "ZCOUNT s104 5 10 extra"},
			expected:       []interface{}{1, errors.New("wrong number of arguments for 'ZCOUNT' command")},
			valueExtractor: []ValueExtractorFn{extractValueZADD, nil},
		},
		{
			name:           "ZCOUNT with invalid number format for min",
			commands:       []string{"ZADD s109 5 member1", "ZCOUNT s109 invalid 10"},
			expected:       []interface{}{1, errors.New("value is not an integer or a float")},
			valueExtractor: []ValueExtractorFn{extractValueZADD, nil},
		},
		{
			name:           "ZCOUNT with invalid number format for max",
			commands:       []string{"ZADD s106 5 member1", "ZCOUNT s106 5 invalid"},
			expected:       []interface{}{1, errors.New("value is not an integer or a float")},
			valueExtractor: []ValueExtractorFn{extractValueZADD, nil},
		},
		{
			name:           "ZCOUNT with wrong type operation",
			commands:       []string{"SET wrongtype value", "ZCOUNT wrongtype 1 10"},
			expected:       []interface{}{string("OK"), errors.New("wrongtype operation against a key holding the wrong kind of value")},
			valueExtractor: []ValueExtractorFn{extractValueSET, nil},
		},
		{
			name:           "ZCOUNT with negative scores",
			commands:       []string{"DEL myzset", "ZADD myzset -10 neg_ten -5 neg_five 0 zero 5 five", "ZCOUNT myzset -8 2"},
			expected:       []interface{}{1, 4, 2}, // Should count -5 and 0
			valueExtractor: []ValueExtractorFn{extractValueDEL, extractValueZADD, extractValueZCOUNT},
		},
		{
			name:           "ZCOUNT with decimal scores",
			commands:       []string{"ZADD myzset1 1.5 one_point_five 2.5 two_point_five 3.5 three_point_five"},
			expected:       []interface{}{errors.New("value is not an integer or a float")},
			valueExtractor: []ValueExtractorFn{nil},
		},
		{
			name:           "ZCOUNT with min greater than max",
			commands:       []string{"ZADD s1 5 member1 7 member2 9 member3", "ZCOUNT s1 10 5"},
			expected:       []interface{}{3, 0},
			valueExtractor: []ValueExtractorFn{extractValueZADD, extractValueZCOUNT},
		},
		{
			name:           "ZCOUNT with single value range",
			commands:       []string{"ZADD s2 5 member1 7 member2 9 member3", "ZCOUNT s2 7 7"},
			expected:       []interface{}{3, 1}, // Only member2 with score 7
			valueExtractor: []ValueExtractorFn{extractValueZADD, extractValueZCOUNT},
		},
		{
			name:           "ZCOUNT with duplicate scores",
			commands:       []string{"ZADD s3 5 member1 5 member2 5 member3 7 member4", "ZCOUNT s3 5 5"},
			expected:       []interface{}{4, 3}, // 3 members with score 5
			valueExtractor: []ValueExtractorFn{extractValueZADD, extractValueZCOUNT},
		},
		{
			name:           "ZCOUNT with min and max on exact member scores",
			commands:       []string{"ZADD s4 10 member1 20 member2 30 member3 40 member4", "ZCOUNT s4 10 30"},
			expected:       []interface{}{4, 3}, // Scores 10, 20, 30
			valueExtractor: []ValueExtractorFn{extractValueZADD, extractValueZCOUNT},
		},
		{
			name:           "ZCOUNT with score 0",
			commands:       []string{"ZADD zerozset -5 neg_five 0 zero 5 five", "ZCOUNT zerozset -2 2"},
			expected:       []interface{}{3, 1}, // Only zero is in range
			valueExtractor: []ValueExtractorFn{extractValueZADD, extractValueZCOUNT},
		},
		{
			name:           "ZCOUNT with range outside all members",
			commands:       []string{"ZADD s9 5 member1 10 member2 15 member3", "ZCOUNT s9 20 30"},
			expected:       []interface{}{3, 0}, // No members in range
			valueExtractor: []ValueExtractorFn{extractValueZADD, extractValueZCOUNT},
		},
	}

	runTestcases(t, client, testCases)
}

// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"errors"
	"testing"

	"github.com/dicedb/dicedb-go/wire"
)

func extractValueZADD(res *wire.Result) interface{} {
	return res.GetZADDRes().Count
}

func TestZADD(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name:     "Call ZADD with no arguments",
			commands: []string{"ZADD"},
			expected: []interface{}{
				errors.New("wrong number of arguments for 'ZADD' command"),
			},
			valueExtractor: []ValueExtractorFn{nil},
		},
		{
			name:     "Call ZADD with just the key",
			commands: []string{"ZADD key"},
			expected: []interface{}{
				errors.New("wrong number of arguments for 'ZADD' command"),
			},
			valueExtractor: []ValueExtractorFn{nil},
		},
		{
			name:     "Call ZADD with just key and score",
			commands: []string{"ZADD key 1"},
			expected: []interface{}{
				errors.New("wrong number of arguments for 'ZADD' command"),
			},
			valueExtractor: []ValueExtractorFn{nil},
		},
		{
			name: "Call ZADD with NX, XX, and CH flags",
			commands: []string{
				"ZADD key NX 1 memberNX", // Add new member with NX
				"ZADD key NX 2 memberNX", // Try to update existing member with NX
				"ZADD key XX 3 memberNX", // Update existing member with XX
				"ZADD key XX 1 memberXX", // Try to add non-existing member with XX
				"ZADD key CH 4 memberNX", // Update score with CH flag
			},
			expected: []interface{}{
				1, // Member added with NX
				0, // Member not updated with NX
				0, // Member updated with XX
				0, // Member not added with XX
				1, // 1 change (score updated with CH)
			},
			valueExtractor: []ValueExtractorFn{extractValueZADD, extractValueZADD, extractValueZADD, extractValueZADD, extractValueZADD},
		},
		{
			name: "Call ZADD with GT and LT flags",
			commands: []string{
				"ZADD key 2 memberGT",    // Add new member
				"ZADD key GT 3 memberGT", // Update with higher score using GT
				"ZADD key GT 1 memberGT", // Try to update with lower score using GT
				"ZADD key LT 1 memberLT", // Add new member
				"ZADD key LT 0 memberLT", // Update with lower score using LT
				"ZADD key LT 2 memberLT", // Try to update with higher score using LT
			},
			expected: []interface{}{
				1, // Member added
				0, // Member updated with GT
				0, // Member not updated with GT
				1, // Member added
				0, // Member updated with LT
				0, // Member not updated with LT
			},
			valueExtractor: []ValueExtractorFn{extractValueZADD, extractValueZADD, extractValueZADD, extractValueZADD, extractValueZADD, extractValueZADD},
		},
		{
			name: "Call ZADD with INCR flag",
			commands: []string{
				"ZADD key1 INCR 1 memberINCR", // Add new member with INCR
				"ZADD key1 INCR 2 memberINCR", // Increment score of existing member
			},
			expected: []interface{}{
				1, // Incremented score returned
				3, // Incremented score returned
			},
			valueExtractor: []ValueExtractorFn{extractValueZADD, extractValueZADD},
		},
		{
			name: "Call ZADD with invalid flag combinations",
			commands: []string{
				"ZADD key NX XX 1 memberInvalid",    // Invalid combination of NX and XX
				"ZADD key GT LT 1 memberInvalid",    // Invalid combination of GT and LT
				"ZADD key INCR 1 member1 2 member2", // INCR with multiple score-member pairs
			},
			expected: []interface{}{
				errors.New("XX and NX options at the same time are not compatible"),
				errors.New("GT, LT, and/or NX options at the same time are not compatible"),
				errors.New("INCR option supports a single increment-element pair"),
			},
			valueExtractor: []ValueExtractorFn{nil, nil, nil},
		},
		{
			name: "Call ZADD with CH flag and multiple changes",
			commands: []string{
				"ZADD key CH 1 member1", // Add new member
				"ZADD key CH 2 member1", // Update score of existing member
				"ZADD key CH 3 member2", // Add another new member
			},
			expected: []interface{}{
				1, // 1 change (new member added)
				1, // 1 change (score updated)
				1, // 1 change (new member added)
			},
			valueExtractor: []ValueExtractorFn{extractValueZADD, extractValueZADD, extractValueZADD},
		},
	}
	runTestcases(t, client, testCases)
}

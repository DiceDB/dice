// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"errors"
	"testing"

	"github.com/dicedb/dicedb-go/wire"
)

func extractValueZREM(res *wire.Result) interface{} {
	return res.GetZREMRes().Count
}

func TestZREM(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name:     "Call ZREM with bad arguments",
			commands: []string{"ZREM", "ZREM key"},
			expected: []interface{}{
				errors.New("wrong number of arguments for 'ZREM' command"),
				errors.New("wrong number of arguments for 'ZREM' command"),
			},
			valueExtractor: []ValueExtractorFn{nil, nil},
		},
		{
			name:           "Call ZREM with non-existing key",
			commands:       []string{"ZREM nonExistingKey member1"},
			expected:       []interface{}{0},
			valueExtractor: []ValueExtractorFn{extractValueZREM},
		},
		{
			name: "Call ZREM on a key which is not a sorted set",
			commands: []string{
				"SET key value",
				"ZREM key member1",
			},
			expected: []interface{}{
				"OK",
				errors.New("wrongtype operation against a key holding the wrong kind of value"),
			},
			valueExtractor: []ValueExtractorFn{extractValueSET, nil},
		},
		{
			name: "Call ZREM with existing key and members",
			commands: []string{
				"ZADD key1 1 member1 2 member2 3 member3",
				"ZREM key1 member1 member2",
			},
			expected: []interface{}{
				3, // member1 added
				2, // member1,member2 removed
			},
			valueExtractor: []ValueExtractorFn{extractValueZADD, extractValueZREM},
		},
	}
	runTestcases(t, client, testCases)
}

// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"errors"
	"testing"

	"github.com/dicedb/dicedb-go/wire"
)

func extractValueZRANGE(res *wire.Result) interface{} {
	return res.GetZRANGERes().Elements
}

func TestZRANGE(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name:           "Test ZRANGE command with bad params",
			commands:       []string{"ZRANGE", "ZRANGE key", "ZRANGE key 1"},
			expected:       []interface{}{errors.New("wrong number of arguments for 'ZRANGE' command"), errors.New("wrong number of arguments for 'ZRANGE' command"), errors.New("wrong number of arguments for 'ZRANGE' command")},
			valueExtractor: []ValueExtractorFn{nil, nil, nil},
		},
		{
			name:           "Test ZRANGE command non numeric start and stop",
			commands:       []string{"ZRANGE key a b", "ZRANGE key 1 b"},
			expected:       []interface{}{errors.New("value is not an integer or a float"), errors.New("value is not an integer or a float")},
			valueExtractor: []ValueExtractorFn{nil, nil},
		},
		{
			name:           "Test ZRANGE command non existent key",
			commands:       []string{"ZRANGE key 1 2"},
			expected:       []interface{}{nil},
			valueExtractor: []ValueExtractorFn{nil},
		},
		{
			name:           "Test ZRANGE on non sorted set key",
			commands:       []string{"SET key value", "ZRANGE key 1 2"},
			expected:       []interface{}{"OK", errors.New("wrongtype operation against a key holding the wrong kind of value")},
			valueExtractor: []ValueExtractorFn{extractValueSET, nil},
		},
		{
			name: "Test ZRANGE on set",
			commands: []string{
				"ZADD z1 1 mem1 2 mem2",
				"ZRANGE z1 0 1",
			},
			expected: []interface{}{
				2,
				[]*wire.ZElement{{Member: "mem1", Score: 1, Rank: 1}},
			},
			valueExtractor: []ValueExtractorFn{extractValueZADD, extractValueZRANGE},
		},
	}

	runTestcases(t, client, testCases)
}

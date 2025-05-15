// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"errors"
	"testing"

	"github.com/dicedb/dicedb-go/wire"
)

func extractValueZCARD(res *wire.Result) interface{} {
	return res.GetZCARDRes().Count
}

func TestZCARD(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name:     "ZCARD with wrong number of arguments",
			commands: []string{"ZCARD", "ZCARD myzset more args"},
			expected: []interface{}{
				errors.New("wrong number of arguments for 'ZCARD' command"),
				errors.New("wrong number of arguments for 'ZCARD' command"),
			},
			valueExtractor: []ValueExtractorFn{nil, nil},
		},
		{
			name:     "ZCARD with wrong type of key",
			commands: []string{"SET string_key string_value", "ZCARD string_key"},
			expected: []interface{}{
				"OK",
				errors.New("wrongtype operation against a key holding the wrong kind of value"),
			},
			valueExtractor: []ValueExtractorFn{extractValueSET, nil},
		},
		{
			name:           "ZCARD with non-existent key",
			commands:       []string{"ZADD myzset 1 one", "ZCARD wrong_myzset"},
			expected:       []interface{}{1, 0},
			valueExtractor: []ValueExtractorFn{extractValueZADD, extractValueZCARD},
		},
		{
			name:           "ZCARD with sorted set holding single element",
			commands:       []string{"ZADD u2 1 one", "ZCARD u2"},
			expected:       []interface{}{1, 1},
			valueExtractor: []ValueExtractorFn{extractValueZADD, extractValueZCARD},
		},
		{
			name:           "ZCARD with sorted set holding multiple elements",
			commands:       []string{"ZADD u3 1 one 2 two", "ZCARD u3", "ZADD u3 3 three", "ZCARD u3", "ZREM u3 two", "ZCARD u3"},
			expected:       []interface{}{2, 2, 1, 3, 1, 2},
			valueExtractor: []ValueExtractorFn{extractValueZADD, extractValueZCARD, extractValueZADD, extractValueZCARD, extractValueZREM, extractValueZCARD},
		},
	}

	runTestcases(t, client, testCases)
}

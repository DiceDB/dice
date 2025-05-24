// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"errors"
	"testing"

	"github.com/dicedb/dicedb-go/wire"
)

func extractValueZRANK(res *wire.Result) interface{} {
	resp := res.GetZRANKRes()
	if resp.Element == nil {
		return int64(0)
	} else {
		return resp.Element.Rank
	}
}

func TestZRANK(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name:           "ZRANK of existing member",
			commands:       []string{"ZADD users1 20 bob 10 alice 30 charlie", "ZRANK users1 bob"},
			expected:       []interface{}{3, 2},
			valueExtractor: []ValueExtractorFn{extractValueZADD, extractValueZRANK},
		},
		{
			name:           "ZRANK of non-existing member",
			commands:       []string{"ZADD users2 20 bob 10 alice 30 charlie", "ZRANK users2 daniel"},
			expected:       []interface{}{3, 0},
			valueExtractor: []ValueExtractorFn{extractValueZADD, extractValueZRANK},
		},
		{
			name:           "ZRANK on non-existing key",
			commands:       []string{"ZRANK nonexisting member1"},
			expected:       []interface{}{0},
			valueExtractor: []ValueExtractorFn{extractValueZRANK},
		},
		{
			name:           "ZRANK with wrong number of arguments",
			commands:       []string{"ZRANK key"},
			expected:       []interface{}{errors.New("wrong number of arguments for 'ZRANK' command")},
			valueExtractor: []ValueExtractorFn{nil},
		},
		{
			name:     "ZRANK with invalid option or more than two args",
			commands: []string{"ZRANK key member1 INVALID_OPTION", "ZRANK key member1 member2"},
			expected: []interface{}{errors.New("wrong number of arguments for 'ZRANK' command"),
				errors.New("wrong number of arguments for 'ZRANK' command")},
			valueExtractor: []ValueExtractorFn{nil, nil},
		},
	}

	runTestcases(t, client, testCases)
}

// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"errors"
	"testing"

	"github.com/dicedb/dicedb-go/wire"
)

func extractValueHSET(res *wire.Result) interface{} {
	return res.GetHSETRes().Count
}

func TestHSET(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name:           "Set Field Value at Key stored in Hash",
			commands:       []string{"HSET k f v"},
			expected:       []interface{}{1},
			valueExtractor: []ValueExtractorFn{extractValueHSET},
		},
		{
			name:     "Set Hash on non-hash Key",
			commands: []string{"SET key f", "HSET key f v"},
			expected: []interface{}{"OK",
				errors.New("wrongtype operation against a key holding the wrong kind of value"),
			},
			valueExtractor: []ValueExtractorFn{extractValueSET, nil},
		},
		{
			name:     "Set Hash with no Field and Value",
			commands: []string{"HSET k"},
			expected: []interface{}{
				errors.New("wrong number of arguments for 'HSET' command"),
			},
			valueExtractor: []ValueExtractorFn{nil, nil},
		},
	}
	runTestcases(t, client, testCases)
}

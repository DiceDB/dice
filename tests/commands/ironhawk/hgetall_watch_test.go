// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"errors"
	"testing"

	"github.com/dicedb/dicedb-go/wire"
)

func extractValueHGETALLWATCH(res *wire.Result) interface{} {
	return res.Message
}

func TestHGETALLWATCH(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name:     "HGetAll watch subscription without key arg",
			commands: []string{"HGETALL.WATCH"},
			expected: []interface{}{
				errors.New("wrong number of arguments for 'HGETALL.WATCH' command"),
			},
			valueExtractor: []ValueExtractorFn{nil},
		},
		{
			name:     "HGetAll watch subscription without field arg",
			commands: []string{"HGETALL.WATCH k1"},
			expected: []interface{}{
				"OK",
			},
			valueExtractor: []ValueExtractorFn{extractValueHGETALLWATCH},
		},
	}

	runTestcases(t, client, testCases)
}

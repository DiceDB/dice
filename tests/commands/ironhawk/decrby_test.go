// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"testing"

	"github.com/dicedb/dicedb-go/wire"
)

func extractValueDECRBY(result *wire.Result) interface{} {
	return result.GetDECRBYRes().GetValue()
}

func TestDECRBY(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	// TODO: Add test cases for DECRBY with non-integer values
	// TODO: Add test cases for non existent key
	// TODO: Add test cases for DECR with negative values
	// TODO: Add test cases for DECR with min and max int64 values
	testCases := []TestCase{
		{
			name: "DECRBY",
			commands: []string{
				"SET key1 5",
				"DECRBY key1 2",
				"DECRBY key1 2",
				"DECRBY key1 1",
				"DECRBY key1 1",
			},
			expected:       []interface{}{"OK", 3, 1, 0, -1},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueDECRBY, extractValueDECRBY, extractValueDECRBY, extractValueDECRBY},
		},
	}
	runTestcases(t, client, testCases)
}

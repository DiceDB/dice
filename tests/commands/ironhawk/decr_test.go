// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"testing"

	"github.com/dicedb/dicedb-go/wire"
)

func extractValueDECR(result *wire.Result) interface{} {
	return result.GetDECRRes().GetValue()
}

func TestDECR(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	// TODO: Add test cases for DECR with non-integer values
	// TODO: Add test cases for non existent key
	// TODO: Add test cases for DECR with negative values
	// TODO: Add test cases for DECR with min and max int64 values
	testCases := []TestCase{
		{
			name:           "DECR",
			commands:       []string{"SET key1 2", "DECR key1", "DECR key1", "DECR key1"},
			expected:       []interface{}{"OK", 1, 0, -1},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueDECR, extractValueDECR, extractValueDECR},
		},
	}
	runTestcases(t, client, testCases)
}

// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"errors"
	"testing"

	"github.com/dicedb/dicedb-go/wire"
)

func extractValueTYPE(res *wire.Result) interface{} {
	return res.GetTYPERes().Type
}

func TestTYPE(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name:     "TYPE with invalid number of arguments",
			commands: []string{"TYPE"},
			expected: []interface{}{
				errors.New("wrong number of arguments for 'TYPE' command"),
			},
			valueExtractor: []ValueExtractorFn{nil},
		},
		{
			name:           "TYPE for non-existent key",
			commands:       []string{"TYPE k1"},
			expected:       []interface{}{"none"},
			valueExtractor: []ValueExtractorFn{extractValueTYPE},
		},
		{
			name:           "TYPE for key with String value",
			commands:       []string{"SET k1 v1", "TYPE k1"},
			expected:       []interface{}{"OK", "string"},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueTYPE},
		},
	}

	runTestcases(t, client, testCases)
}

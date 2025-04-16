// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"errors"
	"testing"

	"github.com/dicedb/dicedb-go/wire"
)

func extractValueECHO(result *wire.Result) interface{} {
	return result.GetECHORes().GetMessage()
}

func TestEcho(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	// TODO: Add tests where argument is a string with special characters and spaces etc
	testCases := []TestCase{
		{
			name:     "ECHO with invalid number of arguments",
			commands: []string{"ECHO"},
			expected: []interface{}{
				errors.New("wrong number of arguments for 'ECHO' command"),
			},
			valueExtractor: []ValueExtractorFn{extractValueECHO},
		},
		{
			name:           "ECHO with one argument",
			commands:       []string{"ECHO hello"},
			expected:       []interface{}{"hello"},
			valueExtractor: []ValueExtractorFn{extractValueECHO},
		},
	}
	runTestcases(t, client, testCases)
}

// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"errors"
	"testing"
	"time"

	"github.com/dicedb/dicedb-go/wire"
)

func extractValueGET(result *wire.Result) interface{} {
	return result.GetGETRes().Value
}

func TestGET(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name:           "Get with expiration",
			commands:       []string{"SET k v EX 2", "GET k", "GET k"},
			expected:       []interface{}{"OK", "v", ""},
			delay:          []time.Duration{0, 0, 4 * time.Second},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueGET, extractValueGET},
		},
		{
			name:           "Get without expiration",
			commands:       []string{"SET k v", "GET k"},
			expected:       []interface{}{"OK", "v"},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueGET},
		},
		{
			name:           "Get with non existent key",
			commands:       []string{"GET nek"},
			expected:       []interface{}{""},
			valueExtractor: []ValueExtractorFn{extractValueGET},
		},
		{
			name:     "GET with no keys or arguments",
			commands: []string{"GET"},
			expected: []interface{}{
				errors.New("wrong number of arguments for 'GET' command"),
			},
			valueExtractor: []ValueExtractorFn{nil},
		},
	}

	runTestcases(t, client, testCases)
}

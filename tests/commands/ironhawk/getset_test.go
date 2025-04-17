// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"errors"
	"testing"
	"time"

	"github.com/dicedb/dicedb-go/wire"
)

func extractValueGETSET(res *wire.Result) interface{} {
	return res.GetGETSETRes().Value
}

func TestGETSET(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name:           "SET with expiration and GETSET",
			commands:       []string{"SET k v EX 2", "GETSET k v2", "TTL k"},
			expected:       []interface{}{"OK", "v", -1},
			delay:          []time.Duration{0, 0, 3 * time.Second},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueGETSET, extractValueTTL},
		},
		{
			name:           "GETSET without expiration",
			commands:       []string{"SET k1 v", "GET k1", "GETSET k1 v2", "GET k1"},
			expected:       []interface{}{"OK", "v", "v", "v2"},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueGET, extractValueGETSET, extractValueGET},
		},
		{
			name:           "GETSET with non existent key",
			commands:       []string{"GETSET nek v", "GET nek"},
			expected:       []interface{}{"", "v"},
			valueExtractor: []ValueExtractorFn{extractValueGETSET, extractValueGET},
		},
		{
			name:     "GETSET with no keys or arguments",
			commands: []string{"GETSET"},
			expected: []interface{}{
				errors.New("wrong number of arguments for 'GETSET' command"),
			},
			valueExtractor: []ValueExtractorFn{nil},
		},
	}

	runTestcases(t, client, testCases)
}

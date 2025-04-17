// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"errors"
	"testing"
	"time"

	"github.com/dicedb/dicedb-go/wire"
)

func extractValueTTL(res *wire.Result) interface{} {
	return res.GetTTLRes().Seconds
}

func TestTTL(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name:           "TTL Simple Value",
			commands:       []string{"SET k101 v1", "GETEX k101 EX 5", "TTL k101"},
			expected:       []interface{}{"OK", "v1", 4},
			delay:          []time.Duration{0, 0, 500 * time.Millisecond},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueGETEX, extractValueTTL},
		},
		{
			name:           "TTL on Non-Existent Key",
			commands:       []string{"TTL foo1"},
			expected:       []interface{}{-2},
			valueExtractor: []ValueExtractorFn{extractValueTTL},
		},
		{
			name:     "TTL with negative expiry",
			commands: []string{"SET foo bar", "GETEX foo EX -5"},
			expected: []interface{}{"OK",
				errors.New("invalid value for a parameter in 'GETEX' command for EX parameter"),
			},
			valueExtractor: []ValueExtractorFn{extractValueSET, nil},
		},
		{
			name:           "TTL without Expiry",
			commands:       []string{"SET foo2 bar", "GET foo2", "TTL foo2"},
			expected:       []interface{}{"OK", "bar", -1},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueGET, extractValueTTL},
		},
		{
			name:           "TTL after DEL",
			commands:       []string{"SET foo bar", "GETEX foo EX 5", "DEL foo", "TTL foo"},
			expected:       []interface{}{"OK", "bar", 1, -2},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueGETEX, extractValueDEL, extractValueTTL},
		},
		{
			name:           "Multiple TTL updates",
			commands:       []string{"SET foo bar", "GETEX foo EX 10", "GETEX foo EX 5", "TTL foo"},
			expected:       []interface{}{"OK", "bar", "bar", 4},
			delay:          []time.Duration{0, 0, 0, 500 * time.Millisecond},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueGETEX, extractValueGETEX, extractValueTTL},
		},
		{
			name:           "TTL with Persist",
			commands:       []string{"SET foo3 bar", "GETEX foo3 persist", "TTL foo3"},
			expected:       []interface{}{"OK", "bar", -1},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueGETEX, extractValueTTL},
		},
		{
			name:           "TTL with Expire and Expired Key",
			commands:       []string{"SET foo bar", "GETEX foo ex 2", "TTL foo", "GET foo"},
			expected:       []interface{}{"OK", "bar", 1, ""},
			delay:          []time.Duration{0, 0, 500 * time.Millisecond, 5 * time.Second},
			valueExtractor: []ValueExtractorFn{extractValueSET, extractValueGETEX, extractValueTTL, extractValueGET},
		},
	}

	runTestcases(t, client, testCases)
}

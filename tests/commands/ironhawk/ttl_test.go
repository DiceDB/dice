// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"errors"
	"testing"
	"time"
)

func TestTTL(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name:     "TTL Simple Value",
			commands: []string{"SET foo bar", "GETEX foo EX 5", "TTL foo"},
			expected: []interface{}{"OK", "bar", 5},
			delay:    []time.Duration{0, 0, 0},
		},
		{
			name:     "TTL on Non-Existent Key",
			commands: []string{"TTL foo1"},
			expected: []interface{}{-2},
		},
		{
			name:     "TTL with negative expiry",
			commands: []string{"SET foo bar", "GETEX foo EX -5"},
			expected: []interface{}{"OK",
				errors.New("invalid value for a parameter in 'GETEX' command for EX parameter"),
			},
		},
		{
			name:     "TTL without Expiry",
			commands: []string{"SET foo2 bar", "GET foo2", "TTL foo2"},
			expected: []interface{}{"OK", "bar", -1},
		},
		{
			name:     "TTL after DEL",
			commands: []string{"SET foo bar", "GETEX foo EX 5", "DEL foo", "TTL foo"},
			expected: []interface{}{"OK", "bar", 1, -2},
		},
		{
			name:     "Multiple TTL updates",
			commands: []string{"SET foo bar", "GETEX foo EX 10", "GETEX foo EX 5", "TTL foo"},
			expected: []interface{}{"OK", "bar", "bar", 5},
			delay:    []time.Duration{0, 0, 0, 0},
		},
		{
			name:     "TTL with Persist",
			commands: []string{"SET foo3 bar", "GETEX foo3 persist", "TTL foo3"},
			expected: []interface{}{"OK", "bar", -1},
		},
		{
			name:     "TTL with Expire and Expired Key",
			commands: []string{"SET foo bar", "GETEX foo ex 3", "TTL foo", "GET foo"},
			expected: []interface{}{"OK", "bar", 3, nil},
			delay:    []time.Duration{0, 0, 0, 5 * time.Second},
		},
	}

	runTestcases(t, client, testCases)
}

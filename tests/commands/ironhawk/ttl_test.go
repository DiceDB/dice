// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"testing"
	"time"
)

func TestTTL(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name:     "TTL Simple Value",
			commands: []string{"SET foo bar", "GETEX foo EX 5", "GETEX foo", "TTL foo"},
			expected: []interface{}{"OK", "bar", "bar", 5},
		},
		{
			name:     "TTL on Non-Existent Key",
			commands: []string{"TTL foo1"},
			expected: []interface{}{-2},
		},
		{
			name:     "TTL without Expiry",
			commands: []string{"SET foo2 bar", "GET foo2", "TTL foo2"},
			expected: []interface{}{"OK", "bar", -1},
		},
		{
			name:     "TTL with Persist",
			commands: []string{"SET foo3 bar", "GETEX foo3 persist", "TTL foo3"},
			expected: []interface{}{"OK", "bar", -1},
		},
		{
			name:     "TTL with Expire and Expired Key",
			commands: []string{"SET foo bar", "GETEX foo ex 5", "GET foo", "TTL foo", "TTL foo"},
			expected: []interface{}{"OK", "bar", "bar", 5, -2},
			delay:    []time.Duration{0, 0, 0, 0, 5 * time.Second},
		},
	}

	runTestcases(t, client, testCases)
}

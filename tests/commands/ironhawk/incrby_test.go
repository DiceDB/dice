// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"fmt"
	"math"
	"testing"
)

func TestINCRBY(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name: "INCRBY multiple integer keys",
			commands: []string{
				"SET key 3", "GET key", "INCRBY key 2", "INCRBY key 1", "GET key",
			},
			expected: []interface{}{
				"OK", 3, 5, 6, 6,
			},
		},
		{
			name: "INCRBY negetive values",
			commands: []string{
				"SET key 100",
				"INCRBY key -2",
				"INCRBY key -10",
				"INCRBY key -88",
				"INCRBY key -100",
				"GET key",
			},
			expected: []interface{}{
				"OK", 98, 88, 0, -100, -100,
			},
		},
		{
			name: "INCRBY non-existent key and expect keys to be created",
			commands: []string{
				"SET key 3",
				"INCRBY unsetKey 2",
				"GET key",
				"GET unsetKey",
			},
			expected: []interface{}{
				"OK", 2, 3, 2, -100, -100,
			},
		},
		{
			name: "INCRBY max int64 and expect min int64 (rollover)",
			commands: []string{
				"SET key " + fmt.Sprintf("%d", math.MaxInt64-1),
				"INCRBY key 1",
				"INCRBY key 1",
				"GET key",
			},
			expected: []interface{}{
				"OK", math.MaxInt64, math.MinInt64, math.MinInt64,
			},
		},
		{
			name: "INCRBY min int64 with -1 and expect max int64 (rollover)",
			commands: []string{
				"SET key " + fmt.Sprintf("%d", math.MinInt64+1),
				"INCRBY key -1",
				"INCRBY key -1",
				"GET key",
			},
			expected: []interface{}{
				"OK", math.MinInt64, math.MaxInt64, math.MaxInt64,
			},
		},
		{
			name: "INCRBY with string value and expect type error",
			commands: []string{
				"SET key 1",
				"INCRBY key abc",
			},
			expected: []interface{}{
				"OK", fmt.Errorf("ERR value is not an integer or out of range"),
			},
		},
	}
	runTestcases(t, client, testCases)
}

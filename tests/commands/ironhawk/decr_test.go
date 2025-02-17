// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"fmt"
	"math"
	"testing"
)

func TestDECR(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name: "Decrement multiple keys",
			commands: []string{"SET k1 3", "DECR k1", "DECR k1", "DECR k1", "DECR k2", "GET k1", "GET k2", "SET k3 " + fmt.Sprint(math.MinInt64+1), "DECR k3", "DECR k3"},
			expected: []interface{}{"OK", 2, 1, 0, -1, 0, -1, "OK", math.MinInt64, math.MaxInt64},
			keysUsed: []string{"k1", "k2", "k3"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, key := range tc.keysUsed {
				client.FireString("DEL " + key)
			}
			
			for i, cmd := range tc.commands {
				result := client.FireString(cmd)
				assertEqual(t, tc.expected[i], result)
			}
		})
	}
}

func TestDECRBY(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name: "Decrement multiple keys",
			commands: []string{"SET k1 3", "SET k3 " + fmt.Sprint(math.MinInt64+1), "DECRBY k1 2", "DECRBY k1 1", "DECRBY k4 1", "DECRBY k3 1", "DECRBY k3 " + fmt.Sprint(math.MinInt64), "DECRBY k5 abc", "GET k1", "GET k4"},
			expected: []interface{}{"OK", "OK", 1, 0, -1, math.MinInt64, 0, "ERR value is not an integer or out of range", 0, -1},
			keysUsed: []string{"k1", "k3", "k4", "k5"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, key := range tc.keysUsed {
				client.FireString("DEL " + key)
			}
			
			for i, cmd := range tc.commands {
				result := client.FireString(cmd)
				assertEqual(t, tc.expected[i], result)
			}
		})
	}
}

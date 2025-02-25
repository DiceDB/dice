// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"testing"
)

func TestDECRBY(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name: "DECRBY",
			commands: []string{
				"SET key1 5",
				"DECRBY key1 2",
				"DECRBY key1 2",
				"DECRBY key1 1",
				"DECRBY key1 1",
			},
			expected: []interface{}{
				"OK",
				3,
				1,
				0,
				-1,
			},
		},
	}
	runTestcases(t, client, testCases)
}

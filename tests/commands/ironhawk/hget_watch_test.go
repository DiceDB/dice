// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"errors"
	"testing"
)

func TestHGETWATCH(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name:     "HGet watch subscription without key arg",
			commands: []string{"HGET.WATCH"},
			expected: []interface{}{
				errors.New("wrong number of arguments for 'HGET.WATCH' command"),
			},
		},
		{
			name:     "HGet watch subscription without field arg",
			commands: []string{"HGET.WATCH k1"},
			expected: []interface{}{
				errors.New("wrong number of arguments for 'HGET.WATCH' command"),
			},
		},
	}

	runTestcases(t, client, testCases)
}

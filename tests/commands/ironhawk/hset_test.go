// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"errors"
	"testing"
)

func TestHSET(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name:     "Set Field Value at Key stored in Hash",
			commands: []string{"HSET k f v"},
			expected: []interface{}{1},
		},
		{
			name:     "Set Hash on non-hash Key",
			commands: []string{"SET key f", "HSET key f v"},
			expected: []interface{}{"OK",
				errors.New("wrongtype operation against a key holding the wrong kind of value"),
			},
		},
		{
			name:     "Set Hash with no Field and Value",
			commands: []string{"HSET k"},
			expected: []interface{}{
				errors.New("wrong number of arguments for 'HSET' command"),
			},
		},
	}
	runTestcases(t, client, testCases)
}

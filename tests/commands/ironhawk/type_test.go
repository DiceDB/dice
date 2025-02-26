// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"errors"
	"testing"
)

func TestType(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name:     "TYPE with invalid number of arguments",
			commands: []string{"TYPE"},
			expected: []interface{}{
				errors.New("wrong number of arguments for 'TYPE' command"),
			},
		},
		{
			name:     "TYPE for non-existent key",
			commands: []string{"TYPE k1"},
			expected: []interface{}{"none"},
		},
		{
			name:     "TYPE for key with String value",
			commands: []string{"SET k1 v1", "TYPE k1"},
			expected: []interface{}{"OK", "string"},
		},
	}

	runTestcases(t, client, testCases)
}

// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"errors"
	"testing"
	"time"
)

func TestGETDEL(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name:     "GETDEL",
			commands: []string{"SET k v", "GETDEL k", "GETDEL k", "GET k"},
			expected: []interface{}{"OK", "v", nil, nil},
		},
		{
			name:     "GETDEL with expiration, checking if key exist and is already expired, then it should return null",
			commands: []string{"GETDEL k", "SET k v EX 2", "GETDEL k"},
			expected: []interface{}{nil, "OK", nil},
			delay:    []time.Duration{0, 0, 3 * time.Second},
		},
		{
			name: "GETDEL with expiration, checking if key exist and is not yet expired, then it should return its " +
				"value",
			commands: []string{"SET k v EX 40", "GETDEL k"},
			expected: []interface{}{"OK", "v"},
			delay:    []time.Duration{0, 2 * time.Second},
		},
		{
			name:     "GETDEL with invalid command",
			commands: []string{"GETDEL", "GETDEL k v"},
			expected: []interface{}{
				errors.New("wrong number of arguments for 'GETDEL' command"),
				errors.New("wrong number of arguments for 'GETDEL' command"),
			},
		},
	}

	runTestcases(t, client, testCases)
}

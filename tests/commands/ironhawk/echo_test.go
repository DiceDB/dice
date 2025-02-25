// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"errors"
	"testing"
)

func TestEcho(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name:     "ECHO with invalid number of arguments",
			commands: []string{"ECHO"},
			expected: []interface{}{
				errors.New("wrong number of arguments for 'ECHO' command"),
			},
		},
		{
			name:     "ECHO with one argument",
			commands: []string{"ECHO hello"},
			expected: []interface{}{"hello"},
		},
		{
			name: "ECHO with multiple arguments",
			commands: []string{"ECHO hello world"},
			expected: []interface{}{
				errors.New("wrong number of arguments for 'ECHO' command"),
			},
		},
	}
	runTestcases(t, client, testCases)
}

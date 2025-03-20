// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"errors"
	"testing"
)

func TestPing(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name:     "PING no arguments",
			commands: []string{"PING"},
			expected: []interface{}{"PONG"},
		},
		{
			name:     "PING with one argument",
			commands: []string{"PING hello"},
			expected: []interface{}{"PONG hello"},
		},
		{
			name:     "PING with two arguments",
			commands: []string{"PING hello world"},
			expected: []interface{}{
				errors.New("wrong number of arguments for 'PING' command"),
			},
		},
	}
	runTestcases(t, client, testCases)
}

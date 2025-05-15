// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"errors"
	"testing"
)

func TestZCOUNTWATCH(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name:     "Get watch subscription without key arg",
			commands: []string{"ZCOUNT.WATCH", "ZCOUNT.WATCH users", "ZCOUNT.WATCH users 1"},
			expected: []interface{}{
				errors.New("wrong number of arguments for 'ZCOUNT.WATCH' command"),
				errors.New("wrong number of arguments for 'ZCOUNT.WATCH' command"),
				errors.New("wrong number of arguments for 'ZCOUNT.WATCH' command"),
			},
			valueExtractor: []ValueExtractorFn{nil, nil, nil},
		},
	}

	runTestcases(t, client, testCases)
}

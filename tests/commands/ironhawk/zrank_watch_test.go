// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"errors"
	"testing"
)

func TestZRANKWATCH(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name:     "Get watch subscription without key arg",
			commands: []string{"ZRANK.WATCH", "ZRANK.WATCH users"},
			expected: []interface{}{
				errors.New("wrong number of arguments for 'ZRANK.WATCH' command"),
				errors.New("wrong number of arguments for 'ZRANK.WATCH' command"),
			},
			valueExtractor: []ValueExtractorFn{nil, nil},
		},
	}

	runTestcases(t, client, testCases)
}

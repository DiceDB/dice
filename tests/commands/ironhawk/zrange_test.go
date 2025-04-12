// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"errors"
	"testing"
)

func TestZrange(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name:     "Test ZRANGE command with bad params",
			commands: []string{"ZRANGE", "ZRANGE key", "ZRANGE key 1"},
			expected: []interface{}{errors.New("wrong number of arguments for 'ZRANGE' command"), errors.New("wrong number of arguments for 'ZRANGE' command"), errors.New("wrong number of arguments for 'ZRANGE' command")},
		},
		{
			name:     "Test ZRANGE command non numeric start and stop",
			commands: []string{"ZRANGE key a b", "ZRANGE key 1 b"},
			expected: []interface{}{errors.New("value is not an integer or a float"), errors.New("value is not an integer or a float")},
		},
		{
			name:     "Test ZRANGE command non existent key",
			commands: []string{"ZRANGE key 1 2"},
			expected: []interface{}{nil},
		},
		{
			name:     "Test ZRANGE on non sorted set key",
			commands: []string{"SET key value", "ZRANGE key 1 2"},
			expected: []interface{}{"OK", errors.New("wrongtype operation against a key holding the wrong kind of value")},
		},
		{
			name: "Test ZRANGE on set",
			commands: []string{
				"ZADD sorted_set 1 mem1 2 mem2",
				"ZRANGE sorted_set 0 1",
				"ZRANGE sorted_set 0 1 REV",
				"ZRANGE sorted_set 0 1 WITHSCORES",
				"ZRANGE sorted_set 0 1 REV WITHSCORES"},
			expected: []interface{}{
				int64(2),
				[]string{"mem1", "mem2"},
				[]string{"mem2", "mem1"},
				[]string{"mem1", "1", "mem2", "2"},
				[]string{"mem2", "2", "mem1", "1"}},
		},
	}

	runTestcases(t, client, testCases)
}

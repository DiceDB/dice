// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"errors"
	"testing"
)

func TestHGET(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name:     "Get Value for Field stored at Hash Key",
			commands: []string{"HSET k f 1", "HGET k f"},
			expected: []interface{}{1, "1"},
		},
		{
			name:     "Get Hash Field on non-hash Key",
			commands: []string{"SET key f", "HGET key f"},
			expected: []interface{}{"OK",
				errors.New("wrongtype operation against a key holding the wrong kind of value"),
			},
		},
		{
			name:     "Get Hash Key with no Field argument",
			commands: []string{"HGET k"},
			expected: []interface{}{
				errors.New("wrong number of arguments for 'HGET' command"),
			},
		},
		{
			name:     "Get Hash Field Value for Non-existent Field",
			commands: []string{"HSET k1 f 1", "HGET k1 non_existent_field"},
			expected: []interface{}{1, nil},
		},
		{
			name:     "Get Hash Field Value for Non-existent Key",
			commands: []string{"HGET non_existent_key f"},
			expected: []interface{}{nil},
		},
		{
			name:     "Get Hash Field Value for multiple Fields stored at Hash Key",
			commands: []string{"HSET k f1 v1 f2 v2", "HGET k f1", "HGET k f2"},
			expected: []interface{}{2, "v1", "v2"},
		},
	}
	runTestcases(t, client, testCases)
}

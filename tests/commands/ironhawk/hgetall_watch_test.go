// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"errors"
	"testing"

	"google.golang.org/protobuf/types/known/structpb"
)

func TestHGETALLWATCH(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name:     "HGETALL.WATCH Hash Key with Initial Values",
			commands: []string{"HSET k f1 v1 f2 v2", "HGETALL.WATCH k"},
			expected: []interface{}{2,
				[]*structpb.Value{
					structpb.NewStringValue("f1"),
					structpb.NewStringValue("v1"),
					structpb.NewStringValue("f2"),
					structpb.NewStringValue("v2"),
				},
			},
		},
		{
			name:     "HGETALL.WATCH Empty Hash",
			commands: []string{"HGETALL.WATCH nonexistent_key"},
			expected: []interface{}{
				[]*structpb.Value{},
			},
		},
		{
			name:     "HGETALL.WATCH with no key argument",
			commands: []string{"HGETALL.WATCH"},
			expected: []interface{}{
				errors.New("wrong number of arguments for 'HGETALL.WATCH' command"),
			},
		},
		{
			name:     "HGETALL.WATCH with too many arguments",
			commands: []string{"HGETALL.WATCH key extra_arg"},
			expected: []interface{}{
				errors.New("wrong number of arguments for 'HGETALL.WATCH' command"),
			},
		},
		{
			name:     "HGETALL.WATCH with non-hash key",
			commands: []string{"SET key 5", "HGETALL.WATCH key"},
			expected: []interface{}{"OK",
				errors.New("wrongtype operation against a key holding the wrong kind of value"),
			},
		},
	}

	runTestcases(t, client, testCases)
}

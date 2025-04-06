// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"errors"
	"fmt"
	"testing"

	"google.golang.org/protobuf/types/known/structpb"
)

// Function that generates HSET with a count of field-value pairs followed by HGETALL
func generateLargeHashCommands(count int) []string {
	commands := []string{"HSET large_k"}
	for i := 0; i < count; i++ {
		commands[0] += fmt.Sprintf(" f%d v%d", i, i)
	}
	commands = append(commands, "HGETALL large_k")
	return commands
}

// Function that generates expected result with a count of field-value pairs
func generateLargeHashExpectedResult(count int) []interface{} {
	result := []interface{}{count}
	values := []*structpb.Value{}
	for i := 0; i < count; i++ {
		values = append(values,
			structpb.NewStringValue(fmt.Sprintf("f%d", i)),
			structpb.NewStringValue(fmt.Sprintf("v%d", i)))
	}
	result = append(result, values)
	return result
}

func TestHGETALL(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name:     "Get Value for Field stored at Hash Key",
			commands: []string{"HSET k f1 v1 f2 v2", "HGETALL k"},
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
			name:     "HGETALL with no key argument",
			commands: []string{"HGETALL"},
			expected: []interface{}{
				errors.New("wrong number of arguments for 'HGETALL' command"),
			},
		},
		{
			name:     "HGETALL with too many arguments",
			commands: []string{"HGETALL key extra_arg"},
			expected: []interface{}{
				errors.New("wrong number of arguments for 'HGETALL' command"),
			},
		},
		{
			name:     "HGETALL with non hash key",
			commands: []string{"SET key 5", "HGETALL key"},
			expected: []interface{}{"OK",
				errors.New("wrongtype operation against a key holding the wrong kind of value"),
			},
		},
		{
			name:     "HGETALL on empty hash",
			commands: []string{"HSET k f3 v3 f4, v4", "HGETALL new_k"},
			expected: []interface{}{2,
				[]*structpb.Value{
					structpb.NewStringValue("f1"),
					structpb.NewStringValue(""),
				},
			},
		},
		{
			name:     "HGETALL preserves field insertion order",
			commands: []string{"HSET k1 f3 v3 f1 v1 f2 v2", "HGETALL k1"},
			expected: []interface{}{3,
				[]*structpb.Value{
					structpb.NewStringValue("f3"),
					structpb.NewStringValue("v3"),
					structpb.NewStringValue("f1"),
					structpb.NewStringValue("v1"),
					structpb.NewStringValue("f2"),
					structpb.NewStringValue("v2"),
				},
			},
		},
		{
			name:     "HGETALL with very large hash",
			commands: generateLargeHashCommands(1000),
			expected: generateLargeHashExpectedResult(1000),
		},
	}
	runTestcases(t, client, testCases)
}

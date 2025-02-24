// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/dicedb/dicedb-go/wire"
)

func TestDel(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []struct {
		name     string
		commands []string
		expected []any
	}{
		{
			name:     "DEL with set key",
			commands: []string{"SET k1 v1", "DEL k1", "GET k1"},
			expected: []interface{}{
				&wire.Response_VStr{VStr: "OK"},
				&wire.Response_VInt{VInt: 1},
				&wire.Response_VNil{VNil: true},
			},
		},
		{
			// TODO: 3rd and 4th command should be together but delete on multi shard isn't there as of now
			name:     "DEL with multiple keys",
			commands: []string{"SET k1 v1", "SET k2 v2", "DEL k1", "DEL k2", "GET k1", "GET k2"},
			expected: []interface{}{
				&wire.Response_VStr{VStr: "OK"},
				&wire.Response_VStr{VStr: "OK"},
				&wire.Response_VInt{VInt: 1},
				&wire.Response_VInt{VInt: 1},
				&wire.Response_VNil{VNil: true},
				&wire.Response_VNil{VNil: true},
			},
		},
		{
			name:     "DEL with key not set",
			commands: []string{"GET k3", "DEL k3"},
			expected: []interface{}{
				&wire.Response_VNil{VNil: true},
				&wire.Response_VInt{VInt: 0},
			},
		},
		{
			name:     "DEL with no keys or arguments",
			commands: []string{"DEL"},
			expected: []any{
				"wrong number of arguments for 'DEL' command",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result := client.FireString(cmd)

				var resultValue any = result.GetValue()
				if result.Err != "" {
					resultValue = result.Err
				}
				
				assert.Equal(t, tc.expected[i], resultValue, "Value mismatch for cmd %s", cmd)
			}
		})
	}
}

// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"strings"
	"testing"

	"github.com/dicedb/dicedb-go/wire"
	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []struct {
		name     string
		commands []string
		expected []any
	}{
		{
			name: "Get with expiration",
			commands: []string{
				"SET k v",
				"GET k",
			},
			expected: []any{
				&wire.Response_VStr{VStr: "OK"},
				&wire.Response_VStr{VStr: "v"},
			},
		},
		{
			name: "Get with non existent key",
			commands: []string{
				"GET nek",
			},
			expected: []any{
				&wire.Response_VNil{VNil: true},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result := client.Fire(&wire.Command{
					Cmd:  strings.Split(cmd, " ")[0],
					Args: strings.Split(cmd, " ")[1:],
				})
				assert.Equal(t, tc.expected[i], result.GetValue(), "Value mismatch for cmd %s", cmd)
			}
		})
	}
}

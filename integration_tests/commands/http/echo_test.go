// This file is part of DiceDB.
// Copyright (C) 2025  DiceDB (dicedb.io).
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEchoHttp(t *testing.T) {

	exec := NewHTTPCommandExecutor()

	testCases := []struct {
		name     string
		commands []HTTPCommand
		expected []interface{}
	}{
		{
			name: "ECHO with invalid number of arguments",
			commands: []HTTPCommand{
				{Command: "ECHO", Body: map[string]interface{}{"key": "", "value": ""}},
			},
			expected: []interface{}{"ERR wrong number of arguments for 'echo' command"},
		},
		{
			name: "ECHO with one argument",
			commands: []HTTPCommand{
				{Command: "ECHO", Body: map[string]interface{}{"value": "hello world"}}, // Providing one argument "hello world"
			},
			expected: []interface{}{"hello world"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}
}

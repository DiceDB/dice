// This file is part of DiceDB.
// Copyright (C) 2025DiceDB (dicedb.io).
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

func TestCommandHelp(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []TestCase{
		{
			name: "Command help should not support any argument",
			commands: []HTTPCommand{
				{Command: "COMMAND/HELP", Body: map[string]interface{}{"key": ""}},
			},
			expected: []interface{}{"ERR wrong number of arguments for 'command|help' command"},
		},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)
				assert.Equal(t, tc.expected[i], result)

			}

		})
	}
}

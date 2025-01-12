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

package resp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDBSIZE(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	testCases := []struct {
		name     string
		setup    []string
		commands []string
		expected []interface{}
		delay    []time.Duration
		cleanUp  []string
	}{
		{
			name:     "DBSIZE",
			setup:    []string{"FLUSHDB", "MSET k1 v1 k2 v2 k3 v3"},
			commands: []string{"DBSIZE"},
			expected: []interface{}{int64(3)},
			delay:    []time.Duration{0},
			cleanUp:  []string{"DEL k1 k2 k3"},
		},
		{
			name:     "DBSIZE with repeative keys in MSET/SET",
			setup:    []string{"MSET k1 v1 k2 v2 k3 v3 k1 v3", "SET k2 v22"},
			commands: []string{"DBSIZE"},
			expected: []interface{}{int64(3)},
			delay:    []time.Duration{0},
			cleanUp:  []string{"DEL k1 k2 k3"},
		},
		{
			name:     "DBSIZE with expired keys",
			setup:    []string{"MSET k1 v1 k2 v2 k3 v3", "SET k3 v3 ex 1"},
			commands: []string{"DBSIZE", "DBSIZE"},
			expected: []interface{}{int64(3), int64(2)},
			delay:    []time.Duration{0, 2 * time.Second},
			cleanUp:  []string{"DEL k1 k2 k3"},
		},
		{
			name:     "DBSIZE with deleted keys",
			setup:    []string{"MSET k1 v1 k2 v2 k3 v3"},
			commands: []string{"DBSIZE", "DEL k1 k2", "DBSIZE"},
			expected: []interface{}{int64(3), int64(2), int64(1)},
			delay:    []time.Duration{0, 0, 0},
			cleanUp:  []string{"DEL k1 k2 k3"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			for _, cmd := range tc.setup {
				result := FireCommand(conn, cmd)
				assert.Equal(t, "OK", result, "Setup Failed")
			}

			for i, cmd := range tc.commands {
				if tc.delay[i] > 0 {
					time.Sleep(tc.delay[i])
				}
				result := FireCommand(conn, cmd)
				assert.Equal(t, tc.expected[i], result)
			}

			for _, cmd := range tc.cleanUp {
				FireCommand(conn, cmd)
			}
		})
	}
}

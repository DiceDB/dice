// This file is part of DiceDB.
// Copyright (C) 2024 DiceDB (dicedb.io).
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

func TestExists(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	testCases := []struct {
		name     string
		command  []string
		expected []interface{}
		delay    []time.Duration
	}{
		{
			name:     "Test EXISTS command",
			command:  []string{"SET key value", "EXISTS key", "EXISTS key2"},
			expected: []interface{}{"OK", int64(1), int64(0)},
			delay:    []time.Duration{0, 0, 0},
		},
		{
			name:     "Test EXISTS command with multiple keys",
			command:  []string{"SET key value", "SET key2 value2", "EXISTS key key2 key3", "EXISTS key key2 key3 key4", "DEL key", "EXISTS key key2 key3 key4"},
			expected: []interface{}{"OK", "OK", int64(2), int64(2), int64(1), int64(1)},
			delay:    []time.Duration{0, 0, 0, 0, 0, 0},
		},
		{
			name:     "Test EXISTS an expired key",
			command:  []string{"SET key value ex 1", "EXISTS key", "EXISTS key"},
			expected: []interface{}{"OK", int64(1), int64(0)},
			delay:    []time.Duration{0, 0, 2 * time.Second},
		},
		{
			name:     "Test EXISTS with multiple keys and expired key",
			command:  []string{"SET key value ex 2", "SET key2 value2", "SET key3 value3", "EXISTS key key2 key3", "EXISTS key key2 key3"},
			expected: []interface{}{"OK", "OK", "OK", int64(3), int64(2)},
			delay:    []time.Duration{0, 0, 0, 0, 2 * time.Second},
		},
	}
	for _, tcase := range testCases {
		t.Run(tcase.name, func(t *testing.T) {
			// deleteTestKeys([]string{"key", "key2", "key3", "key4"}, store)
			FireCommand(conn, "DEL key")
			FireCommand(conn, "DEL key2")
			FireCommand(conn, "DEL key3")
			FireCommand(conn, "DEL key4")

			for i := 0; i < len(tcase.command); i++ {
				if tcase.delay[i] > 0 {
					time.Sleep(tcase.delay[i])
				}
				cmd := tcase.command[i]
				out := tcase.expected[i]
				assert.Equal(t, out, FireCommand(conn, cmd), "Value mismatch for cmd %s\n.", cmd)
			}
		})
	}
}

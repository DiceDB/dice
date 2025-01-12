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

	"github.com/stretchr/testify/assert"
)

var renameKeysTestCases = []struct {
	name     string
	inCmd    []string
	expected []interface{}
}{
	{
		name:     "Set key and Rename key",
		inCmd:    []string{"set sourceKey hello", "get sourceKey", "rename sourceKey destKey", "get destKey", "get sourceKey"},
		expected: []interface{}{"OK", "hello", "OK", "hello", "(nil)"},
	},
	{
		name:     "same key for source and destination on Rename",
		inCmd:    []string{"set Key hello", "get Key", "rename Key Key", "get Key"},
		expected: []interface{}{"OK", "hello", "OK", "hello"},
	},
	{
		name:     "If source key doesn't exists",
		inCmd:    []string{"rename unknownKey Key"},
		expected: []interface{}{"ERR no such key"},
	},
	{
		name:     "If source key doesn't exists and renaming the same key to the same key",
		inCmd:    []string{"rename unknownKey unknownKey"},
		expected: []interface{}{"ERR no such key"},
	},
	{
		name:     "If destination Key already presents",
		inCmd:    []string{"set destinationKey world", "set newKey hello", "rename newKey destinationKey", "get newKey", "get destinationKey"},
		expected: []interface{}{"OK", "OK", "OK", "(nil)", "hello"},
	},
}

func TestCommandRename(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	for _, tc := range renameKeysTestCases {
		t.Run(tc.name, func(t *testing.T) {
			// deleteTestKeys([]string{"k", "k1", "k2"}, store)
			FireCommand(conn, "DEL k1")
			FireCommand(conn, "DEL k2")
			FireCommand(conn, "DEL 3")
			for i, cmd := range tc.inCmd {
				result := FireCommand(conn, cmd)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}
}

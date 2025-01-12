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

var getKeysTestCases = []struct {
	name     string
	inCmd    string
	expected interface{}
}{
	{"Set command", "set 1 2 3 4", []interface{}{"1"}},
	{"Get command", "get key", []interface{}{"key"}},
	{"TTL command", "ttl key", []interface{}{"key"}},
	{"Del command", "del 1 2 3 4 5 6", []interface{}{"1", "2", "3", "4", "5", "6"}},
	// TODO: Fix this for multi shard support
	//{"MSET command", "MSET key1 val1 key2 val2", []interface{}{"key1", "key2"}},
	{"Expire command", "expire key time extra", []interface{}{"key"}},
	{"Ping command", "ping", "ERR the command has no key arguments"},
	{"Invalid Get command", "get", "ERR invalid number of arguments specified for command"},
	{"Abort command", "abort", "ERR the command has no key arguments"},
	{"Invalid command", "NotValidCommand", "ERR invalid command specified"},
	{"Wrong number of arguments", "", "ERR wrong number of arguments for 'command|getkeys' command"},
}

func TestCommandGetKeys(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	for _, tc := range getKeysTestCases {
		t.Run(tc.name, func(t *testing.T) {
			result := FireCommand(conn, "COMMAND GETKEYS "+tc.inCmd)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func BenchmarkGetKeysMatch(b *testing.B) {
	conn := getLocalConnection()
	defer conn.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tc := range getKeysTestCases {
			FireCommand(conn, "COMMAND GETKEYS "+tc.inCmd)
		}
	}
}

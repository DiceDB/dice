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

var getInfoTestCases = []struct {
	name     string
	inCmd    string
	expected interface{}
}{
	{"Set command", "SET", []interface{}{[]interface{}{"set", int64(-3), int64(1), int64(0), int64(0), []any{}}}},
	{"Get command", "GET", []interface{}{[]interface{}{"get", int64(2), int64(1), int64(0), int64(0), []any{}}}},
	{"Ping command", "PING", []interface{}{[]interface{}{"ping", int64(-1), int64(0), int64(0), int64(0), []any{}}}},
	{"Invalid command", "INVALID_CMD", []interface{}{string("(nil)")}},
	{"Combination of valid and Invalid command", "SET INVALID_CMD", []interface{}{
		[]interface{}{"set", int64(-3), int64(1), int64(0), int64(0), []any{}},
		string("(nil)"),
	}},
	{"Combination of multiple valid commands", "SET GET", []interface{}{
		[]interface{}{"set", int64(-3), int64(1), int64(0), int64(0), []any{}},
		[]interface{}{"get", int64(2), int64(1), int64(0), int64(0), []any{}},
	}},
}

func TestCommandInfo(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	for _, tc := range getInfoTestCases {
		t.Run(tc.name, func(t *testing.T) {
			result := FireCommand(conn, "COMMAND INFO "+tc.inCmd)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func BenchmarkCommandInfo(b *testing.B) {
	conn := getLocalConnection()
	defer conn.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tc := range getInfoTestCases {
			FireCommand(conn, "COMMAND INFO "+tc.inCmd)
		}
	}
}

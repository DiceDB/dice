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

package testutils

import (
	"strings"
	"unicode/utf8"
)

func ParseCommand(cmd string) []string {
	var args []string
	var builder strings.Builder
	var inQuotes bool

	flushBuilder := func() {
		if builder.Len() > 0 {
			args = append(args, builder.String())
			builder.Reset()
		}
	}

	for cmd != "" {
		r, size := utf8.DecodeRuneInString(cmd)
		switch {
		case r == utf8.RuneError && size == 1:
			// Invalid UTF-8 sequence, treat each byte as a separate character
			builder.WriteByte(cmd[0])
			cmd = cmd[1:]
		case r == ' ' && !inQuotes:
			flushBuilder()
			cmd = cmd[size:]
		case r == '"':
			inQuotes = !inQuotes
			builder.WriteRune(r)
			cmd = cmd[size:]
		default:
			builder.WriteString(cmd[:size])
			cmd = cmd[size:]
		}
	}

	flushBuilder()

	// Remove quotes from each argument
	for i, arg := range args {
		args[i] = trimQuotes(arg)
	}

	return args
}

func trimQuotes(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}

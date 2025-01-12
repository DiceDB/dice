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

package eval

import (
	"testing"

	"github.com/dicedb/dice/internal/object"

	"github.com/dicedb/dice/internal/server/utils"
)

// TestDeduceType tests the deduceType function using table-driven tests.
func TestDeduceType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantType object.ObjectType
		wantEnc  uint8
	}{
		{
			name:     "Integer string",
			input:    "123",
			wantType: object.ObjTypeInt,
		},
		{
			name:     "Short string",
			input:    "short string",
			wantType: object.ObjTypeString,
		},
		{
			name:     "Long string",
			input:    "this is a very long string that exceeds the maximum length for EMBSTR encoding",
			wantType: object.ObjTypeString,
		},
		{
			name:     "Empty string",
			input:    utils.EmptyStr,
			wantType: object.ObjTypeString,
		},
		{
			name:     "Boundary length string",
			input:    "this string is exactly forty-four characters long",
			wantType: object.ObjTypeString,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, gotType := getRawStringOrInt(tt.input)
			if gotType != tt.wantType {
				t.Errorf("deduceType(%q) = (%v), want (%v)", tt.input, gotType, tt.wantType)
			}
		})
	}
}

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

package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetJsonFieldType(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		wantType string
	}{
		{
			name:     "string test",
			input:    "123",
			wantType: "string",
		},
		{
			name:     "integer test",
			input:    1,
			wantType: "integer",
		},
		{
			name:     "float test",
			input:    1.1,
			wantType: "number",
		},
		{
			name:     "boolean test",
			input:    true,
			wantType: "boolean",
		},
		{
			name:     "nil test",
			input:    nil,
			wantType: "null",
		},
		{
			name:     "array test",
			input:    []interface{}{"string"},
			wantType: "array",
		},
		{
			name:     "object test",
			input:    map[string]interface{}{},
			wantType: "object",
		},
		{
			name:     "unknown test",
			input:    struct{}{},
			wantType: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetJSONFieldType(tt.input)
			assert.Equal(t, tt.wantType, result)
		})
	}
}

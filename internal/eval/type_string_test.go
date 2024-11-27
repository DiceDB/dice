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
		wantType uint8
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
			gotType := deduceType(tt.input)
			if gotType != tt.wantType {
				t.Errorf("deduceType(%q) = (%v), want (%v)", tt.input, gotType, tt.wantType)
			}
		})
	}
}

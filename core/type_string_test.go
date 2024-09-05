package core

import (
	"testing"

	"github.com/dicedb/dice/internal/constants"
	"github.com/dicedb/dice/internal/store"
)

// TestDeduceTypeEncoding tests the deduceTypeEncoding function using table-driven tests.
func TestDeduceTypeEncoding(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantType uint8
		wantEnc  uint8
	}{
		{
			name:     "Integer string",
			input:    "123",
			wantType: store.ObjTypeInt,
			wantEnc:  store.ObjEncodingInt,
		},
		{
			name:     "Short string",
			input:    "short string",
			wantType: store.ObjTypeString,
			wantEnc:  store.ObjEncodingEmbStr,
		},
		{
			name:     "Long string",
			input:    "this is a very long string that exceeds the maximum length for EMBSTR encoding",
			wantType: store.ObjTypeString,
			wantEnc:  store.ObjEncodingRaw,
		},
		{
			name:     "Empty string",
			input:    constants.EmptyStr,
			wantType: store.ObjTypeString,
			wantEnc:  store.ObjEncodingEmbStr,
		},
		{
			name:     "Boundary length string",
			input:    "this string is exactly forty-four characters long",
			wantType: store.ObjTypeString,
			wantEnc:  store.ObjEncodingRaw,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, gotEnc := deduceTypeEncoding(tt.input)
			if gotType != tt.wantType || gotEnc != tt.wantEnc {
				t.Errorf("deduceTypeEncoding(%q) = (%v, %v), want (%v, %v)", tt.input, gotType, gotEnc, tt.wantType, tt.wantEnc)
			}
		})
	}
}

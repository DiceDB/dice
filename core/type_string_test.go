package core

import "testing"

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
			wantType: OBJ_TYPE_STRING,
			wantEnc:  OBJ_ENCODING_INT,
		},
		{
			name:     "Short string",
			input:    "short string",
			wantType: OBJ_TYPE_STRING,
			wantEnc:  OBJ_ENCODING_EMBSTR,
		},
		{
			name:     "Long string",
			input:    "this is a very long string that exceeds the maximum length for EMBSTR encoding",
			wantType: OBJ_TYPE_STRING,
			wantEnc:  OBJ_ENCODING_RAW,
		},
		{
			name:     "Empty string",
			input:    "",
			wantType: OBJ_TYPE_STRING,
			wantEnc:  OBJ_ENCODING_EMBSTR,
		},
		{
			name:     "Boundary length string",
			input:    "this string is exactly forty-four characters long",
			wantType: OBJ_TYPE_STRING,
			wantEnc:  OBJ_ENCODING_RAW,
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

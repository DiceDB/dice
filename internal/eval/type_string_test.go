package eval

import (
	"github.com/dicedb/dice/internal/object"
	"testing"

	"github.com/dicedb/dice/internal/server/utils"
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
			wantType: object.ObjTypeInt,
			wantEnc:  object.ObjEncodingInt,
		},
		{
			name:     "Short string",
			input:    "short string",
			wantType: object.ObjTypeString,
			wantEnc:  object.ObjEncodingEmbStr,
		},
		{
			name:     "Long string",
			input:    "this is a very long string that exceeds the maximum length for EMBSTR encoding",
			wantType: object.ObjTypeString,
			wantEnc:  object.ObjEncodingRaw,
		},
		{
			name:     "Empty string",
			input:    utils.EmptyStr,
			wantType: object.ObjTypeString,
			wantEnc:  object.ObjEncodingEmbStr,
		},
		{
			name:     "Boundary length string",
			input:    "this string is exactly forty-four characters long",
			wantType: object.ObjTypeString,
			wantEnc:  object.ObjEncodingRaw,
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

func TestContainsSpacesNewlinesOrSpecialChars(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"NoSpecialChars123", false},
		{"HelloWorld123", false},
		{"1234567890", false},
		{"", false},
		{"₹₹", true},
		{"Hello, World!", true},
		{"Hello\nWorld", true},
		{"Hello_World", true},
		{"\tTabbedText", true},
		{"NormalText!", true},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := containsSpacesNewlinesOrSpecialChars(test.input)
			if result != test.expected {
				t.Errorf("For input '%s', expected %v but got %v", test.input, test.expected, result)
			}
		})
	}
}

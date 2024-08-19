package utils

import (
	"gotest.tools/v3/assert"
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

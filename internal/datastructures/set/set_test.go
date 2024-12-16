package set

import (
	"math"
	"strconv"
	"testing"
)

var testCases = []struct {
	name   string
	items  []string
	result int
}{
	{
		name:   "int8",
		items:  []string{"1", "2", "3"},
		result: EncodingInt8,
	},
	{
		name:   "int16",
		items:  []string{"1", "2", "3", "32767"},
		result: EncodingInt16,
	},
	{
		name:   "int32",
		items:  []string{"1", "2", "3", "32767", "2147483647"},
		result: EncodingInt32,
	},
	{
		name:   "int64",
		items:  []string{"1", "2", "3", "32767", "2147483647", "9223372036854775807"},
		result: EncodingInt64,
	},
	{
		name:   "float32",
		items:  []string{"1.1", "2.2", "3.3"},
		result: EncodingFloat32,
	},
	{
		name:   "float64",
		items:  []string{"1.1", "2.2", "3.3", strconv.FormatFloat(math.MaxFloat64, 'f', -1, 64)},
		result: EncodingFloat64,
	},
	{
		name:   "string",
		items:  []string{"a", "b", "c"},
		result: EncodingString,
	},
	{
		name:   "mixed",
		items:  []string{"1", "2", "3", "a", "b", "c"},
		result: EncodingString,
	},
	{
		name:   "mixed with float",
		items:  []string{"1", "2", "3", "1.1", "2.2", "3.3"},
		result: EncodingString,
	},
	{
		name:   "mixed with float and int",
		items:  []string{"1", "2", "3", "1.1", "2.2", "3.3", "32767"},
		result: EncodingString,
	},
}

func TestEncodingFromItems(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := EncodingFromItems(tc.items)
			if result != tc.result {
				t.Errorf("expected %v, got %v", tc.result, result)
			}
		})
	}
}

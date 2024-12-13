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

func TestNewTypedSetFromItems(t *testing.T) {
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			dsObject := NewTypedSetFromItems(test.items)
			// check all items
			for _, item := range test.items {
				switch test.result {
				case EncodingInt8:
					i, _ := strconv.ParseInt(item, 10, 8)
					if !dsObject.(*TypedSet[int8]).Contains(int8(i)) {
						t.Errorf("expected %v to be in set", int8(i))
					}
				case EncodingInt16:
					i, _ := strconv.ParseInt(item, 10, 16)
					if !dsObject.(*TypedSet[int16]).Contains(int16(i)) {
						t.Errorf("expected %v to be in set", int16(i))
					}
				case EncodingInt32:
					i, _ := strconv.ParseInt(item, 10, 32)
					if !dsObject.(*TypedSet[int32]).Contains(int32(i)) {
						t.Errorf("expected %v to be in set", int32(i))
					}
				case EncodingInt64:
					i, _ := strconv.ParseInt(item, 10, 64)
					if !dsObject.(*TypedSet[int64]).Contains(int64(i)) {
						t.Errorf("expected %v to be in set", int64(i))
					}
				case EncodingFloat32:
					i, _ := strconv.ParseFloat(item, 32)
					if !dsObject.(*TypedSet[float32]).Contains(float32(i)) {
						t.Errorf("expected %v to be in set", float32(i))
					}
				case EncodingFloat64:
					i, _ := strconv.ParseFloat(item, 64)
					if !dsObject.(*TypedSet[float64]).Contains(float64(i)) {
						t.Errorf("expected %v to be in set", float64(i))
					}
				case EncodingString:
					if !dsObject.(*TypedSet[string]).Contains(item) {
						t.Errorf("expected %v to be in set", item)
					}

				}
			}
		})
	}
}

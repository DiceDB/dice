package set

import (
	"cmp"
	"strconv"

	ds "github.com/dicedb/dice/internal/datastructures"
)

const (
	EncodingInt8 = iota
	EncodingInt16
	EncodingInt32
	EncodingInt64

	EncodingFloat32
	EncodingFloat64

	EncodingString
)

var SignedIntEncodings = []int{EncodingInt64, EncodingInt32, EncodingInt16, EncodingInt8}
var FloatEncodings = []int{EncodingFloat64, EncodingFloat32}

type constraint interface {
	cmp.Ordered
}

var _ ds.DSInterface = &TypedSet[int8]{}
var _ ds.DSInterface = &TypedSet[int16]{}
var _ ds.DSInterface = &TypedSet[int32]{}
var _ ds.DSInterface = &TypedSet[int64]{}
var _ ds.DSInterface = &TypedSet[float32]{}
var _ ds.DSInterface = &TypedSet[float64]{}
var _ ds.DSInterface = &TypedSet[string]{}

type TypedSet[T constraint] struct {
	ds.BaseDataStructure[ds.DSInterface]
	Value map[T]struct{}
}

func IsTypeTypedSet(obj ds.DSInterface) bool {
	switch any(obj).(type) {
	case *TypedSet[int8], *TypedSet[int16], *TypedSet[int32], *TypedSet[int64], *TypedSet[float32], *TypedSet[float64], *TypedSet[string]:
		return true
	default:
		return false
	}
}
func NewTypedSet[T constraint](encoding int) ds.DSInterface {
	return &TypedSet[T]{
		Value: make(map[T]struct{}),
	}
}

func (s *TypedSet[T]) Serialize() []byte {
	// add length of the set
	b := make([]byte, 0)
	b = append(b, byte(len(s.Value)))
	for key := range s.Value {
		if intKey, ok := any(key).(int64); ok {
			b = append(b, byte(intKey))
		} else if strKey, ok := any(key).(string); ok {
			b = append(b, []byte(strKey)...)
		}
	}
	return b
}

func DeduceEncodingFromItems(et map[int]struct{}) int {
	// If there is only one encoding, return it.
	if len(et) == 1 {
		for encoding := range et {
			return encoding
		}
	}
	// if there are string items, return string encoding.
	if _, ok := (et)[EncodingString]; ok {
		return EncodingString
	}

	// if there are encodings of just signed integers, return the biggest one.
	// 01 bit 0: signed integer
	// 10 bit 1: float
	bitmask := 0
	for encoding := range et {
		if _, ok := (et)[encoding]; ok {
			switch encoding {
			case EncodingString:
				continue
			case EncodingFloat64, EncodingFloat32:
				bitmask |= 2
			case EncodingInt64, EncodingInt32, EncodingInt16, EncodingInt8:
				bitmask |= 1
			}
		}
	}

	if bitmask == 1 {
		// return the biggest signed integer encoding.
		for _, encoding := range SignedIntEncodings {
			if _, ok := (et)[encoding]; ok {
				return encoding
			}
		}
	} else if bitmask == 2 {
		// return the biggest float encoding.
		for _, encoding := range FloatEncodings {
			if _, ok := (et)[encoding]; ok {
				return encoding
			}
		}
	}
	return EncodingString
}

func NewTypedSetFromEncoding(encoding int) ds.DSInterface {
	switch encoding {
	case EncodingInt8:
		return NewTypedSet[int8](encoding)
	case EncodingInt16:
		return NewTypedSet[int16](encoding)
	case EncodingInt32:
		return NewTypedSet[int32](encoding)
	case EncodingInt64:
		return NewTypedSet[int64](encoding)
	case EncodingFloat32:
		return NewTypedSet[float32](encoding)
	case EncodingFloat64:
		return NewTypedSet[float64](encoding)
	default:
		return NewTypedSet[string](encoding)
	}
}

func NewTypedSetFromEncodingAndItems(items []string, encoding int) ds.DSInterface {
	set := NewTypedSetFromEncoding(encoding)
	switch encoding {
	case EncodingInt8:
		for _, item := range items {
			i, _ := strconv.ParseInt(item, 10, 8)
			set.(*TypedSet[int8]).Value[int8(i)] = struct{}{}
		}
	case EncodingInt16:
		for _, item := range items {
			i, _ := strconv.ParseInt(item, 10, 16)
			set.(*TypedSet[int16]).Value[int16(i)] = struct{}{}
		}
	case EncodingInt32:
		for _, item := range items {
			i, _ := strconv.ParseInt(item, 10, 32)
			set.(*TypedSet[int32]).Value[int32(i)] = struct{}{}
		}
	case EncodingInt64:
		for _, item := range items {
			i, _ := strconv.ParseInt(item, 10, 64)
			set.(*TypedSet[int64]).Value[int64(i)] = struct{}{}
		}
	case EncodingFloat32:
		for _, item := range items {
			i, _ := strconv.ParseFloat(item, 32)
			set.(*TypedSet[float32]).Value[float32(i)] = struct{}{}
		}
	case EncodingFloat64:
		for _, item := range items {
			i, _ := strconv.ParseFloat(item, 64)
			set.(*TypedSet[float64]).Value[float64(i)] = struct{}{}
		}
	default:
		for _, item := range items {
			set.(*TypedSet[string]).Value[item] = struct{}{}
		}
	}
	return set
}

func EncodingFromItems(items []string) int {
	et := make(map[int]struct{})
	for _, item := range items {
		encoding := ds.GetElementType(item)
		if encoding == EncodingString {
			return EncodingString
		}
		et[encoding] = struct{}{}
	}
	return DeduceEncodingFromItems(et)
}

func NewTypedSetFromItems(items []string) ds.DSInterface {
	encoding := EncodingFromItems(items)
	return NewTypedSetFromEncodingAndItems(items, encoding)
}

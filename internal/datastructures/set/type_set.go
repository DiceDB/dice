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

type TypedSetInterface[T constraint] interface {
	ds.DSInterface
	All() []T
	Serialize() []byte
	Size() int
	Add(item T) bool
	Remove(item T)
	Contains(item T) bool
}

type TypedSet[T constraint] struct {
	ds.BaseDataStructure[ds.DSInterface]
	value map[T]struct{}
}

func NewTypedSet[T constraint](encoding int) ds.DSInterface {
	return &TypedSet[T]{
		value: make(map[T]struct{}),
		BaseDataStructure: ds.BaseDataStructure[ds.DSInterface]{
			Encoding: encoding,
		},
	}
}

func (s *TypedSet[T]) All() []T {
	result := make([]T, 0, len(s.value))
	for key := range s.value {
		result = append(result, key)
	}
	return result
}

func (s *TypedSet[T]) Serialize() []byte {
	// add length of the set
	b := make([]byte, 0)
	b = append(b, byte(len(s.value)))
	for key := range s.value {
		if intKey, ok := any(key).(int64); ok {
			b = append(b, byte(intKey))
		} else if strKey, ok := any(key).(string); ok {
			b = append(b, []byte(strKey)...)
		}
	}
	return b
}

func (s *TypedSet[T]) Size() int {
	return len(s.value)
}

func (s *TypedSet[T]) Add(item T) bool {
	if _, ok := s.value[item]; ok {
		return false
	}
	s.value[item] = struct{}{}
	return true
}

func (s *TypedSet[T]) Remove(item T) {
	delete(s.value, item)
}

func (s *TypedSet[T]) Contains(item T) bool {
	_, ok := s.value[item]
	return ok
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

func NewTypedSetFromEncodingAndItems(items []string, encoding int) (ds.DSInterface, int) {
	set := NewTypedSetFromEncoding(encoding)
	switch encoding {
	case EncodingInt8:
		for _, item := range items {
			i, _ := strconv.ParseInt(item, 10, 8)
			set.(*TypedSet[int8]).Add(int8(i))
		}
	case EncodingInt16:
		for _, item := range items {
			i, _ := strconv.ParseInt(item, 10, 16)
			set.(*TypedSet[int16]).Add(int16(i))
		}
	case EncodingInt32:
		for _, item := range items {
			i, _ := strconv.ParseInt(item, 10, 32)
			set.(*TypedSet[int32]).Add(int32(i))
		}
	case EncodingInt64:
		for _, item := range items {
			i, _ := strconv.ParseInt(item, 10, 64)
			set.(*TypedSet[int64]).Add(int64(i))
		}
	case EncodingFloat32:
		for _, item := range items {
			i, _ := strconv.ParseFloat(item, 32)
			set.(*TypedSet[float32]).Add(float32(i))
		}
	case EncodingFloat64:
		for _, item := range items {
			i, _ := strconv.ParseFloat(item, 64)
			set.(*TypedSet[float64]).Add(float64(i))
		}
	default:
		for _, item := range items {
			set.(*TypedSet[string]).Add(item)
		}
	}
	return set, encoding
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

func NewTypedSetFromItems(items []string) (ds.DSInterface, int) {
	encoding := EncodingFromItems(items)
	return NewTypedSetFromEncodingAndItems(items, encoding)
}

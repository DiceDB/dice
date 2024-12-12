package sds

import (
	"cmp"
	"errors"
	"math"
	"strconv"

	ds "github.com/dicedb/dice/internal/datastructures"
)

// This constraint defines the list of GoLang types that can be used as a value in the SDS Data Structure
type constraint interface {
	cmp.Ordered | []byte
}

var (
	// Ensure SDS[int] implements DSInterface
	_ ds.DSInterface = &SDS[int]{}
	_ SDSInterface   = &SDS[int]{}
)

type SDSInterface interface {
	ds.DSInterface
	Set(val string) error
	Get() string
}

func GetIfTypeSDS(ds ds.DSInterface) (SDSInterface, bool) {
	sds, ok := ds.(SDSInterface)
	return sds, ok
}

// SDS is similar to the sds data-structure in Redis
type SDS[T constraint] struct {
	ds.BaseDataStructure[ds.DSInterface]
	value T
}

func NewString(s string) ds.DSInterface {
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		switch {
		case i >= math.MinInt8 && i <= math.MaxInt8:
			return &SDS[int8]{value: int8(i)}
		case i >= math.MinInt16 && i <= math.MaxInt16:
			return &SDS[int16]{value: int16(i)}
		case i >= math.MinInt32 && i <= math.MaxInt32:
			return &SDS[int32]{value: int32(i)}
		default:
			return &SDS[int64]{value: i}
		}
	}

	if u, err := strconv.ParseUint(s, 10, 64); err == nil {
		switch {
		case u <= math.MaxUint8:
			return &SDS[uint8]{value: uint8(u)}
		case u <= math.MaxUint16:
			return &SDS[uint16]{value: uint16(u)}
		case u <= math.MaxUint32:
			return &SDS[uint32]{value: uint32(u)}
		default:
			return &SDS[uint64]{value: u}
		}
	}

	if f, err := strconv.ParseFloat(s, 64); err == nil {
		if f >= -math.MaxFloat32 && f <= math.MaxFloat32 {
			return &SDS[float32]{value: float32(f)}
		}
		return &SDS[float64]{value: f}
	}

	return &SDS[[]byte]{value: []byte(s)}
}

func (s *SDS[T]) Get() string {
	switch v := any(s.value).(type) {
	case int8:
		if val, ok := any(v).(int8); ok {
			return strconv.FormatInt(int64(val), 10)
		}
	case int16:
		if val, ok := any(v).(int16); ok {
			return strconv.FormatInt(int64(val), 10)
		}
	case int32:
		if val, ok := any(v).(int32); ok {
			return strconv.FormatInt(int64(val), 10)
		}
	case int64:
		if val, ok := any(v).(int64); ok {
			return strconv.FormatInt(val, 10)
		}
	case uint8:
		if val, ok := any(v).(uint8); ok {
			return strconv.FormatUint(uint64(val), 10)
		}
	case uint16:
		if val, ok := any(v).(uint16); ok {
			return strconv.FormatUint(uint64(val), 10)
		}
	case uint32:
		if val, ok := any(v).(uint32); ok {
			return strconv.FormatUint(uint64(val), 10)
		}
	case uint64:
		if val, ok := any(v).(uint64); ok {
			return strconv.FormatUint(val, 10)
		}
	case float32:
		if val, ok := any(v).(float32); ok {
			return strconv.FormatFloat(float64(val), 'f', -1, 32)
		}
	case float64:
		if val, ok := any(v).(float64); ok {
			return strconv.FormatFloat(val, 'f', -1, 64)
		}
	case []byte:
		if val, ok := any(v).([]byte); ok {
			return string(val)
		}
	default:
		// Handle unexpected types or return a default value
		return ""
	}

	// If none of the cases match or type assertion fails
	return ""
}

func (s *SDS[T]) Set(val string) error {
	var zero T

	switch any(zero).(type) {
	case int8:
		if i, err := strconv.ParseInt(val, 10, 64); err == nil && i >= math.MinInt8 && i <= math.MaxInt8 {
			s.value = any(int8(i)).(T)
			return nil
		}
	case int16:
		if i, err := strconv.ParseInt(val, 10, 64); err == nil && i >= math.MinInt16 && i <= math.MaxInt16 {
			s.value = any(int16(i)).(T)
			return nil
		}
	case int32:
		if i, err := strconv.ParseInt(val, 10, 64); err == nil && i >= math.MinInt32 && i <= math.MaxInt32 {
			s.value = any(int32(i)).(T)
			return nil
		}
	case int64:
		if i, err := strconv.ParseInt(val, 10, 64); err == nil {
			s.value = any(i).(T)
			return nil
		}
	case uint8:
		if u, err := strconv.ParseUint(val, 10, 64); err == nil && u <= math.MaxUint8 {
			s.value = any(uint8(u)).(T)
			return nil
		}
	case uint16:
		if u, err := strconv.ParseUint(val, 10, 64); err == nil && u <= math.MaxUint16 {
			s.value = any(uint16(u)).(T)
			return nil
		}
	case uint32:
		if u, err := strconv.ParseUint(val, 10, 64); err == nil && u <= math.MaxUint32 {
			s.value = any(uint32(u)).(T)
			return nil
		}
	case uint64:
		if u, err := strconv.ParseUint(val, 10, 64); err == nil {
			s.value = any(u).(T)
			return nil
		}
	case float32:
		if f, err := strconv.ParseFloat(val, 64); err == nil && f >= -math.MaxFloat32 && f <= math.MaxFloat32 {
			s.value = any(float32(f)).(T)
			return nil
		}
	case float64:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			s.value = any(f).(T)
			return nil
		}
	case []byte:
		s.value = any([]byte(val)).(T)
		return nil
	}

	return errors.New("value out of range or invalid type")
}

func (s *SDS[T]) Serialize() []byte {
	return []byte(s.Get())
}

func (s *SDS[T]) Size() int {
	return len(s.Get())
}

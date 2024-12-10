package datastructures

import (
	"cmp"
	"fmt"
	"math"
	"strconv"
)

type constraint interface {
	cmp.Ordered | []byte
}

// toString converts the value to a string
func ToString[T constraint](v T) string {
	switch v := any(v).(type) {
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
		return ""
	}
	return ""
}

func GetElementType(s string) int {
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		switch {
		case i >= math.MinInt8 && i <= math.MaxInt8:
			return EncodingInt8
		case i >= math.MinInt16 && i <= math.MaxInt16:
			return EncodingInt16
		case i >= math.MinInt32 && i <= math.MaxInt32:
			return EncodingInt32
		default:
			return EncodingInt64
		}
	}

	if f, err := strconv.ParseFloat(s, 32); err == nil {
		if f >= -math.MaxFloat32 && f <= math.MaxFloat32 {
			return EncodingFloat32
		}
	}

	if f, err := strconv.ParseFloat(s, 64); err == nil {
		if f >= -math.MaxFloat64 && f <= math.MaxFloat64 {
			return EncodingFloat64
		}
	}
	return EncodingString
}

func parseStringToType(s string, encoding int) (any, error) {
	if GetElementType(s) != encoding {
		return nil, fmt.Errorf("invalid type")
	}
	switch encoding {
	case EncodingInt8:
		i, err := strconv.ParseInt(s, 10, 8)
		if err != nil {
			return nil, err
		}
		return int8(i), nil
	case EncodingInt16:
		i, err := strconv.ParseInt(s, 10, 16)
		if err != nil {
			return nil, err
		}
		return int16(i), nil
	case EncodingInt32:
		i, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			return nil, err
		}
		return int32(i), nil
	case EncodingInt64:
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return nil, err
		}
		return int64(i), nil
	case EncodingFloat32:
		i, err := strconv.ParseFloat(s, 32)
		if err != nil {
			return nil, err
		}
		return float32(i), nil
	case EncodingFloat64:
		i, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, err
		}
		return float64(i), nil
	default:
		return s, nil
	}
}

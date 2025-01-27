package sds

import (
	"strconv"
)

// toString converts the value to a string
func toString[T constraint](v T) string {
	switch v := any(v).(type) {
	case uint8, uint16, uint32:
		return strconv.FormatUint(v.(uint64), 10)
	case int8, int16, int32:
		return strconv.FormatInt(v.(int64), 10)
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case []byte:
		return string(v)
	default:
		return ""
	}
}

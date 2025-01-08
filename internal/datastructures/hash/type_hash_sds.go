package hash

import (
	"cmp"
	"errors"
	"math"
	"strconv"

	"github.com/dicedb/dice/internal/server/utils"
)

type ItemConstraint interface {
	cmp.Ordered
}

type HashItemInterface interface {
	Set(val string) error
	Get() (string, bool)
	Expiry() int64
	SetExpiry(expiry int64)
	Expired() bool
}

type HashItem[T ItemConstraint] struct {
	expiry int64
	Value  T
}

func NewHashItem(value string, expiry int64) HashItemInterface {
	if i, err := strconv.ParseInt(value, 10, 64); err == nil {
		switch {
		case i >= math.MinInt8 && i <= math.MaxInt8:
			return &HashItem[int8]{Value: int8(i), expiry: expiry}
		case i >= math.MinInt16 && i <= math.MaxInt16:
			return &HashItem[int16]{Value: int16(i), expiry: expiry}
		case i >= math.MinInt32 && i <= math.MaxInt32:
			return &HashItem[int32]{Value: int32(i), expiry: expiry}
		default:
			return &HashItem[int64]{Value: i, expiry: expiry}
		}
	}
	if u, err := strconv.ParseUint(value, 10, 64); err == nil {
		switch {
		case u <= math.MaxUint8:
			return &HashItem[uint8]{Value: uint8(u), expiry: expiry}
		case u <= math.MaxUint16:
			return &HashItem[uint16]{Value: uint16(u), expiry: expiry}
		case u <= math.MaxUint32:
			return &HashItem[uint32]{Value: uint32(u), expiry: expiry}
		default:
			return &HashItem[uint64]{Value: u, expiry: expiry}
		}
	}
	if f, err := strconv.ParseFloat(value, 64); err == nil {
		if f >= -math.MaxFloat32 && f <= math.MaxFloat32 {
			return &HashItem[float32]{Value: float32(f), expiry: expiry}
		}
		return &HashItem[float64]{Value: f, expiry: expiry}
	}
	return &HashItem[string]{Value: value, expiry: expiry}
}

func (i *HashItem[T]) Set(val string) error {
	var zero T
	switch any(zero).(type) {
	case int8:
		if val, err := strconv.ParseInt(val, 10, 8); err == nil {
			i.Value = any(int8(val)).(T)
			return nil
		}
	case int16:
		if val, err := strconv.ParseInt(val, 10, 16); err == nil {
			i.Value = any(int16(val)).(T)
			return nil
		}
	case int32:
		if val, err := strconv.ParseInt(val, 10, 32); err == nil {
			i.Value = any(int32(val)).(T)
			return nil
		}
	case int64:
		if val, err := strconv.ParseInt(val, 10, 64); err == nil {
			i.Value = any(val).(T)
			return nil
		}
	case uint8:
		if val, err := strconv.ParseUint(val, 10, 8); err == nil {
			i.Value = any(uint8(val)).(T)
			return nil
		}
	case uint16:
		if val, err := strconv.ParseUint(val, 10, 16); err == nil {
			i.Value = any(uint16(val)).(T)
			return nil
		}
	case uint32:
		if val, err := strconv.ParseUint(val, 10, 32); err == nil {
			i.Value = any(uint32(val)).(T)
			return nil
		}
	case uint64:
		if val, err := strconv.ParseUint(val, 10, 64); err == nil {
			i.Value = any(val).(T)
			return nil
		}
	case float32:
		if val, err := strconv.ParseFloat(val, 32); err == nil {
			i.Value = any(float32(val)).(T)
			return nil
		}
	case float64:
		if val, err := strconv.ParseFloat(val, 64); err == nil {
			i.Value = any(val).(T)
			return nil
		}
	case []byte:
		i.Value = any([]byte(val)).(T)
		return nil
	case string:
		i.Value = any(val).(T)
		return nil
	}
	return errors.New("value out of range or invalid type")
}

// func (i *HashItem[T]) Get() (string, bool) {
// 	if i.Expired() {
// 		return "", false
// 	}
// 	switch any(i.Value).(type) {
// 	case int8, int16, int32, int64, uint8, uint16, uint32, uint64:
// 		return strconv.Itoa(any(i.Value).(int)), true
// 	case float32, float64:
// 		return strconv.FormatFloat(float64(any(i.Value).(float64)), 'f', -1, 64), true
// 	case []byte:
// 		return string(any(i.Value).([]byte)), true
// 	case string:
// 		return any(i.Value).(string), true
// 	}
// 	return "", false
// }

func (i *HashItem[T]) Get() (string, bool) {

	if i.Expired() {
		return "", false
	}
	switch any(i.Value).(type) {
	case string:
		return any(i.Value).(string), true
	case int8:
		return strconv.FormatInt(int64(any(i.Value).(int8)), 10), true
	case int16:
		return strconv.FormatInt(int64(any(i.Value).(int16)), 10), true
	case int32:
		return strconv.FormatInt(int64(any(i.Value).(int32)), 10), true
	case int64:
		return strconv.FormatInt(any(i.Value).(int64), 10), true
	case uint8:
		return strconv.FormatUint(uint64(any(i.Value).(uint8)), 10), true
	case uint16:
		return strconv.FormatUint(uint64(any(i.Value).(uint16)), 10), true
	case uint32:
		return strconv.FormatUint(uint64(any(i.Value).(uint32)), 10), true
	case uint64:
		return strconv.FormatUint(any(i.Value).(uint64), 10), true
	case float32:
		return strconv.FormatFloat(float64(any(i.Value).(float32)), 'f', -1, 32), true
	case float64:
		return strconv.FormatFloat(any(i.Value).(float64), 'f', -1, 64), true
	case []byte:
		return string(any(i.Value).([]byte)), true
	}
	return "", false
}

func (i *HashItem[T]) Expiry() int64 {
	return i.expiry
}

func (i *HashItem[T]) SetExpiry(expiry int64) {
	if expiry > 0 {
		expiry += utils.GetCurrentTime().UnixMilli()
	}
	i.expiry = expiry
}

func (i *HashItem[T]) Expired() bool {
	if i.expiry == -1 {
		return false
	}
	return utils.GetCurrentTime().UnixMilli() >= i.expiry
}

package eval

import (
	"strconv"

	diceerrors "github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/object"
)

type String struct {
	Value string
	Type  uint8
}

func NewString(value string) *String {
	return &String{
		Value: value,
		Type:  object.ObjTypeString,
	}
}

func (s *String) Serialize() []byte {
	return []byte{}
}

func deduceType(v string) (o uint8) {
	// Check if the value has leading zero
	if len(v) > 1 && v[0] == '0' {
		// If so, treat as string
		return object.ObjTypeString
	}
	if _, err := strconv.ParseInt(v, 10, 64); err == nil {
		return object.ObjTypeInt
	}
	return object.ObjTypeString
}

func getRawStringOrInt(value string) (interface{}, uint8) {
	intValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil { // value is not an integer, hence a string
		return value, object.ObjTypeString
	}
	return intValue, object.ObjTypeInt // value is an integer
}

// Function to convert the value to a string for concatenation or manipulation
func convertValueToString(obj *object.Obj, oType uint8) (string, error) {
	var currentValueStr string

	switch oType {
	case object.ObjTypeInt:
		// Convert int64 to string for concatenation
		currentValueStr = strconv.FormatInt(obj.Value.(int64), 10)
	case object.ObjTypeString:
		// Use the string value directly
		currentValueStr = obj.Value.(string)
	case object.ObjTypeByteArray:
		val, ok := obj.Value.(*ByteArray)
		if !ok {
			return "", diceerrors.ErrWrongTypeOperation
		}
		currentValueStr = string(val.data)
	default:
		return "", diceerrors.ErrWrongTypeOperation
	}

	return currentValueStr, nil
}

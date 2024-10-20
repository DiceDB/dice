package eval

import (
	"strconv"

	diceerrors "github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/object"

	dstore "github.com/dicedb/dice/internal/object"
)

// Similar to
// tryObjectEncoding function in Redis
func deduceTypeEncoding(v string) (o, e uint8) {
	// Check if the value has leading zero
	if len(v) > 1 && v[0] == '0' {
		// If so, treat as string
		return object.ObjTypeString, object.ObjEncodingRaw
	}

	if _, err := strconv.ParseInt(v, 10, 64); err == nil {
		return dstore.ObjTypeInt, dstore.ObjEncodingInt
	}
	if len(v) <= 44 {
		return dstore.ObjTypeString, dstore.ObjEncodingEmbStr
	}
	return dstore.ObjTypeString, dstore.ObjEncodingRaw
}

// Function to handle converting the value based on the encoding type
func storeValueWithEncoding(value string, oEnc uint8) (interface{}, []byte) {
	var returnValue interface{}

	// treat as string if value has leading zero
	if len(value) > 1 && value[0] == '0' {
		// If so, treat as string
		return value, nil
	}

	switch oEnc {
	case object.ObjEncodingInt:
		intValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, diceerrors.NewErrWithFormattedMessage(diceerrors.WrongTypeErr)
		}
		returnValue = intValue
	case object.ObjEncodingEmbStr, object.ObjEncodingRaw:
		returnValue = value
	default:
		return nil, diceerrors.NewErrWithFormattedMessage(diceerrors.WrongTypeErr)
	}

	return returnValue, nil
}

// Function to convert the value to a string for concatenation or manipulation
func convertValueToString(obj *object.Obj, oEnc uint8) (string, []byte) {
	var currentValueStr string

	switch oEnc {
	case object.ObjEncodingInt:
		// Convert int64 to string for concatenation
		currentValueStr = strconv.FormatInt(obj.Value.(int64), 10)
	case object.ObjEncodingEmbStr, object.ObjEncodingRaw:
		// Use the string value directly
		currentValueStr = obj.Value.(string)
	default:
		return "", diceerrors.NewErrWithFormattedMessage(diceerrors.WrongTypeErr)
	}

	return currentValueStr, nil
}

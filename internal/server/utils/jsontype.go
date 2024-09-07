package utils

import (
	"github.com/dicedb/dice/internal/constants"
)

func GetJSONFieldType(v interface{}) string {
	switch v.(type) {
	case map[string]interface{}:
		return constants.ObjectType
	case []interface{}:
		return constants.ArrayType
	case string:
		return constants.StringType
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return constants.IntegerType
	case float32, float64:
		return constants.NumberType
	case bool:
		return constants.BooleanType
	case nil:
		return constants.NullType
	default:
		return constants.UnknownType
	}
}

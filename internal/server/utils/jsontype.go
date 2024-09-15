package utils

func GetJSONFieldType(v interface{}) string {
	switch v.(type) {
	case map[string]interface{}:
		return ObjectType
	case []interface{}:
		return ArrayType
	case string:
		return StringType
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return IntegerType
	case float32, float64:
		return NumberType
	case bool:
		return BooleanType
	case nil:
		return NullType
	default:
		return UnknownType
	}
}

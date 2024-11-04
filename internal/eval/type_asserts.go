package eval

// IsInt64 checks if the variable is of type int64.
func IsInt64(v interface{}) bool {
	_, ok := v.(int64)
	return ok
}

// IsString checks if the variable is of type string.
func IsString(v interface{}) bool {
	_, ok := v.(string)
	return ok
}

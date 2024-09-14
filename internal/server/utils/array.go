package utils

func IsArray(data any) bool {
	_, ok := data.([]any)
	return ok
}

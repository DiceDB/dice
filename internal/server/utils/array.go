package utils

import "reflect"

func IsArray(data any) bool {
	kind := reflect.TypeOf(data).Kind()
	return kind == reflect.Array || kind == reflect.Slice
}

package utils

func IsFloatToIntPossible(value float64) (int, bool) {
	intValue := int64(value)
	if value == float64(intValue) {
		return int(intValue), true
	}
	return 0, false
}

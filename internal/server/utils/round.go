package utils

import "math"

// RoundToDecimals rounds a float64 or float32 to a specified number of decimal places.
func RoundToDecimals[T float32 | float64](num T, decimals int) T {
	pow := math.Pow(10, float64(decimals))
	return T(math.Round(float64(num)*pow) / pow)
}

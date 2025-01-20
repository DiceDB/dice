// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package utils

import "math"

// RoundToDecimals rounds a float64 or float32 to a specified number of decimal places.
func RoundToDecimals[T float32 | float64](num T, decimals int) T {
	pow := math.Pow(10, float64(decimals))
	return T(math.Round(float64(num)*pow) / pow)
}

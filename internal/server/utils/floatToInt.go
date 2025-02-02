// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package utils

func IsFloatToIntPossible(value float64) (int, bool) {
	intValue := int64(value)
	if value == float64(intValue) {
		return int(intValue), true
	}
	return 0, false
}

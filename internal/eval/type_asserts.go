// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

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

// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package utils

import "reflect"

func IsArray(data any) bool {
	kind := reflect.TypeOf(data).Kind()
	return kind == reflect.Array || kind == reflect.Slice
}

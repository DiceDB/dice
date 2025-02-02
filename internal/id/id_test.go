// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package id

import (
	"testing"
)

func BenchmarkNextID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ExpandID(NextID())
	}
}

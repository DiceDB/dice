// This file is part of DiceDB.
// Copyright (C) 2025  DiceDB (dicedb.io).
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package dencoding_test

import (
	"math"
	"testing"

	"github.com/dicedb/dice/internal/dencoding"
	"github.com/stretchr/testify/assert"
)

func BenchmarkEncodeDecodeInt(b *testing.B) {
	// Benchmark the performance of encoding and decoding int64 values
	for i := 0; i < b.N; i++ {
		value := int64(i % math.MaxInt64)
		encoded := dencoding.EncodeInt(value)
		decoded := dencoding.DecodeInt(encoded)
		assert.Equal(b, value, decoded, "Encode-Decode round trip failed")
	}
}

func BenchmarkEncodeUIntConcurrent(b *testing.B) {
	// Benchmark the performance of encoding uint64 values concurrently
	b.RunParallel(func(pb *testing.PB) {
		i := uint64(0)
		for pb.Next() {
			dencoding.EncodeUInt(i)
			i++
		}
	})
}

func BenchmarkDecodeUIntConcurrent(b *testing.B) {
	// Benchmark the performance of encoding uint64 values concurrently
	encoded := dencoding.EncodeUInt(math.MaxUint64)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			dencoding.DecodeUInt(encoded)
		}
	})
}

func BenchmarkMinMaxEncodeDecodeInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var value int64
		switch {
		case i%3 == 0:
			value = int64(i % math.MaxInt64) // positive numbers
		case i%3 == 1:
			value = -int64(i % math.MaxInt64) // negative numbers
		default:
			value = math.MaxInt64 - int64(i%1000) // numbers close to MaxInt64
			if i%2000 > 1000 {
				value = math.MinInt64 + int64(i%1000) // numbers close to MinInt64
			}
		}

		encoded := dencoding.EncodeInt(value)
		decoded := dencoding.DecodeInt(encoded)
		if decoded != value {
			b.Errorf("DecodeInt(%v) = %d; want %d", encoded, decoded, value)
		}
	}
}

package id

import (
	"testing"
)

func BenchmarkNextID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ExpandID(NextID())
	}
}

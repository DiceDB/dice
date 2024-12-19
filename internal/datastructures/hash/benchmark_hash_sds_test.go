package hash

import (
	"fmt"
	"testing"
)

func BenchmarkHashSDSSetInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewHashItem("1", 0)
	}
}

func BenchmarkHashSDSSetString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewHashItem(fmt.Sprintf("val%d", i), 0)
	}
}

func BenchmarkHashSDSGetInt(b *testing.B) {
	item := NewHashItem("1", 0)
	for i := 0; i < b.N; i++ {
		item.Get()
	}
}

func BenchmarkHashSDSGetString(b *testing.B) {
	item := NewHashItem("val", 0)
	for i := 0; i < b.N; i++ {
		item.Get()
	}
}

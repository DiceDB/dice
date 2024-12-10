package set

import (
	"fmt"
	"testing"
)

func BenchmarkBaseInt8Set(b *testing.B) {
	items := []int8{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	b.ResetTimer()
	set := map[int8]struct{}{}
	for i := 0; i < b.N; i++ {
		for _, item := range items {
			set[item] = struct{}{}
		}
	}
}

func BenchmarkBaseOnlyStringSet(b *testing.B) {
	set := map[string]struct{}{}
	for i := 0; i < b.N; i++ {
		set[fmt.Sprintf("%d", i)] = struct{}{}
	}
}

func BenchmarkNewInt8SetFromStrings(b *testing.B) {
	items := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewTypedSetFromItems(items)
	}
}

func BenchmarkNewInt16SetFromStrings(b *testing.B) {
	items := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "12312"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewTypedSetFromItems(items)
	}
}

func BenchmarkNewStringSet(b *testing.B) {
	items := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewTypedSetFromItems(items)
	}
}

func BenchmarkAddInt8Set(b *testing.B) {
	items := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}
	set, _ := NewTypedSetFromItems(items)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		set.(*TypedSet[int8]).Add(int8(i))
	}
}

func BenchmarkAddInt16Set(b *testing.B) {
	items := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "257"}
	set, _ := NewTypedSetFromItems(items)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		set.(*TypedSet[int16]).Add(int16(i))
	}
}

func BenchmarkAddStringSet(b *testing.B) {
	items := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	set, _ := NewTypedSetFromItems(items)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		set.(*TypedSet[string]).Add(fmt.Sprintf("%d", i))
	}
}

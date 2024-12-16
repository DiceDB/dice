package set

import (
	"fmt"
	"testing"
)

func newInt8BaseSetFromItems(items []int8) map[int8]struct{} {
	set := map[int8]struct{}{}
	for _, item := range items {
		set[item] = struct{}{}
	}
	return set
}

func newStringBaseSetFromItems(items []string) map[string]struct{} {
	set := map[string]struct{}{}
	for _, item := range items {
		set[item] = struct{}{}
	}
	return set
}
func BenchmarkBaseInt8Set(b *testing.B) {
	items := []int8{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		newInt8BaseSetFromItems(items)
	}
}

func BenchmarkBaseOnlyStringSet(b *testing.B) {
	items := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		newStringBaseSetFromItems(items)
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
	set := NewTypedSetFromItems(items)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		set.(*TypedSet[int8]).Value[int8(i)] = struct{}{}
	}
}

func BenchmarkAddInt16Set(b *testing.B) {
	items := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "257"}
	set := NewTypedSetFromItems(items)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		set.(*TypedSet[int16]).Value[int16(i)] = struct{}{}
	}
}

func BenchmarkAddStringSet(b *testing.B) {
	items := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	set := NewTypedSetFromItems(items)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		set.(*TypedSet[string]).Value[fmt.Sprintf("%d", i)] = struct{}{}
	}
}

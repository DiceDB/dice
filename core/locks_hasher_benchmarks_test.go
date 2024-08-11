package core

import (
	"testing"
)

func BenchmarkNewLockHasher(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewLockHasher()
	}
}

func BenchmarkGetHashKey(b *testing.B) {
	lockHasher := NewLockHasher()
	keys := []string{"test1", "test2", "test3", "test4", "test5"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = lockHasher.GetHashKey(keys[i%len(keys)])
	}
}

func BenchmarkGetHashKeyEmpty(b *testing.B) {
	lockHasher := NewLockHasher()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = lockHasher.GetHashKey("")
	}
}

func BenchmarkGetStore(b *testing.B) {
	lockHasher := NewLockHasher()
	keys := []string{"test1", "test2", "test3", "test4", "test5"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = lockHasher.GetStore(keys[i%len(keys)])
	}
}

func BenchmarkGetStoreEmpty(b *testing.B) {
	lockHasher := NewLockHasher()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = lockHasher.GetStore("")
	}
}

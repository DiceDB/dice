package hash

import (
	"fmt"
	"testing"

	"github.com/dicedb/dice/internal/server/utils"
)

func BenchmarkBaseMapInt(b *testing.B) {
	m := make(map[string]string)
	expiry := make(map[string]int64)
	for i := 0; i < b.N; i++ {
		m[fmt.Sprintf("key%d", i)] = fmt.Sprintf("%d", i)
		expiry[fmt.Sprintf("key%d", i)] = 0
	}
}
func BenchmarkHashSetInt(b *testing.B) {
	h := NewHash()
	hash, ok := h.(*Hash)
	if !ok {
		b.Fail()
	}
	for i := 0; i < b.N; i++ {
		hash.Add(fmt.Sprintf("key%d", i), fmt.Sprintf("%d", i), 0)
	}
}

func BenchmarkBaseMapString(b *testing.B) {
	m := make(map[string]string)
	expiry := make(map[string]int64)
	for i := 0; i < b.N; i++ {
		m[fmt.Sprintf("key%d", i)] = fmt.Sprintf("val%d", i)
		expiry[fmt.Sprintf("key%d", i)] = 0
	}
}

func BenchmarkHashSetString(b *testing.B) {
	h := NewHash()
	hash, ok := h.(*Hash)
	if !ok {
		b.Fail()
	}
	for i := 0; i < b.N; i++ {
		hash.Add(fmt.Sprintf("key%d", i), fmt.Sprintf("val%d", i), 0)
	}
}

func BenchmarkBaseMapGet(b *testing.B) {
	m := make(map[string]string)
	expiry := make(map[string]int64)
	for i := 0; i < b.N; i++ {
		m[fmt.Sprintf("key%d", i)] = fmt.Sprintf("%d", i)
		expiry[fmt.Sprintf("key%d", i)] = 0
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		expireMs, _ := expiry[fmt.Sprintf("key%d", i)]
		if expireMs < utils.GetCurrentTime().UnixMilli() {
			delete(m, fmt.Sprintf("key%d", i))
			delete(expiry, fmt.Sprintf("key%d", i))
			continue
		}
		_, _ = m[fmt.Sprintf("key%d", i)]
	}
}
func BenchmarkHashGet(b *testing.B) {
	h := NewHash()
	hash, ok := h.(*Hash)
	if !ok {
		b.Fail()
	}
	for i := 0; i < b.N; i++ {
		hash.Add(fmt.Sprintf("key%d", i), fmt.Sprintf("%d", i), 0)
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		hash.Get(fmt.Sprintf("key%d", i))
	}
}

func BenchmarkBaseMapExists(b *testing.B) {
	m := make(map[string]string)
	expiry := make(map[string]int64)
	for i := 0; i < b.N; i++ {
		m[fmt.Sprintf("key%d", i)] = fmt.Sprintf("%d", i)
		expiry[fmt.Sprintf("key%d", i)] = 0
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		expireMs, _ := expiry[fmt.Sprintf("key%d", i)]
		if expireMs < utils.GetCurrentTime().UnixMilli() {
			delete(m, fmt.Sprintf("key%d", i))
			delete(expiry, fmt.Sprintf("key%d", i))
			continue
		}
		_, _ = m[fmt.Sprintf("key%d", i)]
	}
}
func BenchmarkHashExists(b *testing.B) {
	h := NewHash()
	hash, ok := h.(*Hash)
	if !ok {
		b.Fail()
	}
	for i := 0; i < b.N; i++ {
		hash.Add(fmt.Sprintf("key%d", i), fmt.Sprintf("%d", i), 0)
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		hash.Exists(fmt.Sprintf("key%d", i))
	}
}

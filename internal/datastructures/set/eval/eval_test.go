package eval

import (
	"fmt"
	"testing"

	dstore "github.com/dicedb/dice/internal/store"
)

func BenchmarkEvalSadd(b *testing.B) {
	store := dstore.NewStore(nil, nil, nil)

	for i := 0; i < b.N; i++ {
		evalSADD([]string{"key", fmt.Sprintf("%d", i)}, store)
		b.ReportAllocs()
	}
}

func BenchmarkEvalSaddIntWithExistingSet(b *testing.B) {
	store := dstore.NewStore(nil, nil, nil)
	key_values := make([]string, 1001)
	key_values = append(key_values, "key")
	for i := 0; i < 1000; i++ {
		key_values = append(key_values, fmt.Sprintf("value%d", i))
	}
	evalSADD(key_values, store)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		evalSADD([]string{"key", fmt.Sprintf("%d", i)}, store)
	}
	b.ReportAllocs()
}

func BenchmarkEvalSaddWithExistingSet(b *testing.B) {
	store := dstore.NewStore(nil, nil, nil)
	key_values := make([]string, 1001)
	key_values = append(key_values, "key")
	for i := 0; i < 1000; i++ {
		key_values = append(key_values, fmt.Sprintf("value%d", i))
	}
	evalSADD(key_values, store)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		evalSADD([]string{"key", fmt.Sprintf("value%d", i)}, store)
	}
	b.ReportAllocs()
}

func BenchmarkEvalSMemberInt(b *testing.B) {
	store := dstore.NewStore(nil, nil, nil)
	for i := 0; i < 1000; i++ {
		evalSADD([]string{"key", fmt.Sprintf("%d", i)}, store)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		evalSMEMBERS([]string{"key"}, store)
	}
	b.ReportAllocs()
}

func BenchmarkEvalSMembersStr(b *testing.B) {
	store := dstore.NewStore(nil, nil, nil)
	for i := 0; i < 1000; i++ {
		evalSADD([]string{"key", fmt.Sprintf("value%d", i)}, store)
	}
	for i := 0; i < b.N; i++ {
		evalSMEMBERS([]string{"key"}, store)
	}
	b.ReportAllocs()
}

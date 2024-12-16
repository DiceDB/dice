package eval

import (
	"fmt"
	"testing"

	dstore "github.com/dicedb/dice/internal/store"
)

func BenchmarkEvalSaddInt8(b *testing.B) {
	store := dstore.NewStore(nil, nil, nil)
	evalSADD([]string{"key", "1"}, store)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x := i % 128
		evalSADD([]string{"key", fmt.Sprintf("%d", x)}, store)
	}
	b.ReportAllocs()
}

func BenchmarkEvalSaddInt32(b *testing.B) {
	store := dstore.NewStore(nil, nil, nil)
	evalSADD([]string{"key", "2147483647"}, store)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		evalSADD([]string{"key", fmt.Sprintf("%d", i)}, store)
	}
	b.ReportAllocs()
}

func BenchmarkEvalSADDInt64(b *testing.B) {
	store := dstore.NewStore(nil, nil, nil)
	evalSADD([]string{"key", "9223372036854775807"}, store)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		evalSADD([]string{"key", fmt.Sprintf("%d", i)}, store)
	}
	b.ReportAllocs()
}

func BenchmarkEvalSaddIntWithExistingSet(b *testing.B) {
	store := dstore.NewStore(nil, nil, nil)
	key_values := []string{}
	key_values = append(key_values, "key")
	for i := 0; i < 1000; i++ {
		key_values = append(key_values, fmt.Sprintf("value%d", i))
	}
	b.ResetTimer()
	b.ReportAllocs()
	evalSADD(key_values, store)
	for i := 0; i < b.N; i++ {
		evalSADD([]string{"key", fmt.Sprintf("%d", i)}, store)
	}
}

func BenchmarkEvalSADDWithInt32ExistingSet(b *testing.B) {
	store := dstore.NewStore(nil, nil, nil)
	key_values := []string{}
	key_values = append(key_values, "key")
	for i := 0; i < 1000; i++ {
		key_values = append(key_values, fmt.Sprintf("%d", i))
	}
	key_values = append(key_values, "2147483647")
	evalSADD(key_values, store)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		evalSADD([]string{"key", fmt.Sprintf("%d", i)}, store)
	}
}

func BenchmarkEvalSaddWithExistingSet(b *testing.B) {
	store := dstore.NewStore(nil, nil, nil)
	key_values := []string{}
	key_values = append(key_values, "key")
	for i := 0; i < 1000; i++ {
		key_values = append(key_values, fmt.Sprintf("value%d", i))
	}
	evalSADD(key_values, store)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		evalSADD([]string{"key", fmt.Sprintf("value%d", i)}, store)
	}
}

func BenchmarkEvalSMemberInt(b *testing.B) {
	store := dstore.NewStore(nil, nil, nil)
	for i := 0; i < 100; i++ {
		evalSADD([]string{"key", fmt.Sprintf("%d", i)}, store)
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		evalSMEMBERS([]string{"key"}, store)
	}
}

func BenchmarkEvalSMembersStr(b *testing.B) {
	store := dstore.NewStore(nil, nil, nil)
	for i := 0; i < 100; i++ {
		evalSADD([]string{"key", fmt.Sprintf("value%d", i)}, store)
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		evalSMEMBERS([]string{"key"}, store)
	}
}

func BenchmarkEvalSCARDInt8(b *testing.B) {
	store := dstore.NewStore(nil, nil, nil)
	for i := 0; i < 100; i++ {
		evalSADD([]string{"key", fmt.Sprintf("%d", i)}, store)
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		evalSCARD([]string{"key"}, store)
	}
}

func BenchmarkEvalSCARDStr(b *testing.B) {
	store := dstore.NewStore(nil, nil, nil)
	for i := 0; i < 100; i++ {
		evalSADD([]string{"key", fmt.Sprintf("value%d", i)}, store)
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		evalSCARD([]string{"key"}, store)
	}
}

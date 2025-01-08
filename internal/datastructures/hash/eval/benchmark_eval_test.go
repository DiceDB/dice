package eval

import (
	"fmt"
	"math"
	"testing"
	"unsafe"

	dstore "github.com/dicedb/dice/internal/store"
)

func calculateStringSize(s string) uintptr {
	// Fixed overhead for the string structure
	overhead := unsafe.Sizeof(s) // Typically 16 bytes on a 64-bit system

	// Size of the actual string content
	contentSize := uintptr(len(s))

	return overhead + contentSize
}

func BenchmarkEvalHSETString(b *testing.B) {
	// Create a new hash
	store := dstore.NewStore(nil, nil, nil)
	for i := 0; i < b.N; i++ {
		evalHSET([]string{"key", fmt.Sprintf("field%d", i), fmt.Sprintf("value%d", i)}, store)
	}
}

func BenchmarkEvalHSETInt(b *testing.B) {
	// Create a new hash
	store := dstore.NewStore(nil, nil, nil)
	for i := 0; i < b.N; i++ {
		evalHSET([]string{"key", fmt.Sprintf("field%d", i), fmt.Sprintf("%d", math.MaxInt)}, store)
	}
	// obj := store.Get("key")
	// hash, ok := hash.GetIfTypeHash(obj)

	// if !ok {
	// 	b.Fatalf("Error getting hash")
	// }
	// totalMem := 0
	// actaulSize := len(hash.Value)
	// for _, v := range hash.Value {
	// 	x, _ := v.Get()
	// 	totalMem += int(calculateStringSize(x))
	// }
	// b.Logf("Total memory: %d", totalMem)
	// b.Logf("Actual memory: %d", actaulSize)

}

func BenchmarkEvalHGETString(b *testing.B) {
	// Create a new hash
	store := dstore.NewStore(nil, nil, nil)
	for i := 0; i < b.N; i++ {
		evalHSET([]string{"key", fmt.Sprintf("field%d", i), fmt.Sprintf("value%d", i)}, store)
	}
}

func BenchmarkEvalHGETInt(b *testing.B) {
	// Create a new hash
	store := dstore.NewStore(nil, nil, nil)
	for i := 0; i < b.N; i++ {
		evalHSET([]string{"key", fmt.Sprintf("field%d", i), fmt.Sprintf("%d", i)}, store)
	}
}

func BenchmarkEvalHDELString(b *testing.B) {
	// Create a new hash
	store := dstore.NewStore(nil, nil, nil)
	for i := 0; i < b.N; i++ {
		evalHSET([]string{"key", fmt.Sprintf("field%d", i), fmt.Sprintf("value%d", i)}, store)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		evalHDEL([]string{"key", fmt.Sprintf("field%d", i)}, store)
	}
}

func BenchmarkEvalHDELInt(b *testing.B) {
	// Create a new hash
	store := dstore.NewStore(nil, nil, nil)
	for i := 0; i < b.N; i++ {
		evalHSET([]string{"key", fmt.Sprintf("field%d", i), fmt.Sprintf("%d", i)}, store)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		evalHDEL([]string{"key", fmt.Sprintf("field%d", i)}, store)
	}
}

func BenchmarkEvalHLENString(b *testing.B) {
	// Create a new hash
	store := dstore.NewStore(nil, nil, nil)
	for i := 0; i < b.N; i++ {
		evalHSET([]string{"key", fmt.Sprintf("field%d", i), fmt.Sprintf("value%d", i)}, store)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		evalHLEN([]string{"key"}, store)
	}
}

func BenchmarkEvalHLENInt(b *testing.B) {
	// Create a new hash
	store := dstore.NewStore(nil, nil, nil)
	for i := 0; i < b.N; i++ {
		evalHSET([]string{"key", fmt.Sprintf("field%d", i), fmt.Sprintf("%d", i)}, store)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		evalHLEN([]string{"key"}, store)
	}
}

func BenchmarkEvalHSTRLEN(b *testing.B) {
	// Create a new hash
	store := dstore.NewStore(nil, nil, nil)
	for i := 0; i < b.N; i++ {
		evalHSET([]string{"key", fmt.Sprintf("field%d", i), fmt.Sprintf("value%d", i)}, store)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		evalHSTRLEN([]string{"key", fmt.Sprintf("field%d", i)}, store)
	}
}

func BenchmarkEvalHGETALL(b *testing.B) {
	// Create a new hash
	store := dstore.NewStore(nil, nil, nil)
	for i := 0; i < b.N; i++ {
		evalHSET([]string{"key", fmt.Sprintf("field%d", i), fmt.Sprintf("value%d", i)}, store)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		evalHGETALL([]string{"key"}, store)
	}
}

func BenchmarkEvalHKEYS(b *testing.B) {
	// Create a new hash
	store := dstore.NewStore(nil, nil, nil)
	for i := 0; i < b.N; i++ {
		evalHSET([]string{"key", fmt.Sprintf("field%d", i), fmt.Sprintf("value%d", i)}, store)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		evalHKEYS([]string{"key"}, store)
	}
}

func BenchmarkEvalHVALS(b *testing.B) {
	// Create a new hash
	store := dstore.NewStore(nil, nil, nil)
	for i := 0; i < b.N; i++ {
		evalHSET([]string{"key", fmt.Sprintf("field%d", i), fmt.Sprintf("value%d", i)}, store)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		evalHVALS([]string{"key"}, store)
	}
}

func BenchmarkEvalHINCRBYInt(b *testing.B) {
	// Create a new hash
	store := dstore.NewStore(nil, nil, nil)
	for i := 0; i < b.N; i++ {
		evalHSET([]string{"key", fmt.Sprintf("field%d", i), fmt.Sprintf("%d", i)}, store)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		evalHINCRBY([]string{"key", fmt.Sprintf("field%d", i), "1"}, store)
	}
}

func BenchmarkEvalHINCRBYFloat(b *testing.B) {
	store := dstore.NewStore(nil, nil, nil)
	for i := 0; i < b.N; i++ {
		evalHSET([]string{"key", fmt.Sprintf("field%d", i), fmt.Sprintf("%d", i)}, store)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		evalHINCRBY([]string{"key", fmt.Sprintf("field%d", i), "1.1"}, store)
	}
}

package eval

import (
	"fmt"
	"testing"

	dstore "github.com/dicedb/dice/internal/store"
)

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
		evalHSET([]string{"key", fmt.Sprintf("field%d", i), fmt.Sprintf("%d", i)}, store)
	}
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

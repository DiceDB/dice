package eval

import (
	"fmt"
	"testing"

	set "github.com/dicedb/dice/internal/datastructures/set"
)

// This benchmark starts with a set of type int8 and
// tries to add a new element to the set for more bits
// than the current encoding. Hence the set should be
// converting as we go ahead and add more elements.
func BenchmarkTryToAddSet(b *testing.B) {
	items := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}
	set := set.NewTypedSetFromItems(items)
	for i := 0; i < b.N; i++ {
		tryAndAddToSet(set, fmt.Sprintf("%d", i))
	}
	b.ReportAllocs()
}

// This benchmark starts with a set of type int8 and
// tries to add a new element to the set for more bits
// than the current encoding. Hence the set should be
// converting as we go ahead and add more elements.
func BenchmarkTryToAddSet2(b *testing.B) {
	items := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}
	set := set.NewTypedSetFromItems(items)
	for i := 0; i < b.N; i++ {
		tryAndAddToSet2(set, fmt.Sprintf("%d", i))
	}
	b.ReportAllocs()
}

// This benchmark test when we add homogeneous elements
// We should not be observing any conversion in the set
func BenchmarkTryToAddSetInt32(b *testing.B) {
	items := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "2147483647"}
	set := set.NewTypedSetFromItems(items)
	for i := 0; i < b.N; i++ {
		tryAndAddToSet(set, fmt.Sprintf("%d", i))
	}
	b.ReportAllocs()
}

// This benchmark test when we add homogeneous elements
// We should not be observing any conversion in the set
func BenchmarkTryToAddSet2Int32(b *testing.B) {
	items := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "2147483647"}
	set := set.NewTypedSetFromItems(items)
	for i := 0; i < b.N; i++ {
		tryAndAddToSet2(set, fmt.Sprintf("%d", i))
	}
	b.ReportAllocs()
}

// This benchmark test when we add homogeneous elements
// We should not be observing any conversion in the set
func BenchmarkTryToAddSetString(b *testing.B) {
	items := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	set := set.NewTypedSetFromItems(items)
	for i := 0; i < b.N; i++ {
		tryAndAddToSet(set, fmt.Sprintf("value%d", i))
	}
	b.ReportAllocs()
}

// This benchmark test when we add homogeneous elements
// We should not be observing any conversion in the set
func BenchmarkTryToAddSet2String(b *testing.B) {
	items := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	set := set.NewTypedSetFromItems(items)
	for i := 0; i < b.N; i++ {
		tryAndAddToSet2(set, fmt.Sprintf("value%d", i))
	}
	b.ReportAllocs()
}

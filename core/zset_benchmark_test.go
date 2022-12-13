/*
 * Benchmark Suite for Sorted Set implementation for Dice.
 * The sorted set is backed by a SkipList and a dictionary.
 *
 * The implementation is inspired by the following open-source projects:
 * - https://github.com/liyiheng/zset/
 */

package core

import (
	"math/rand"
	"testing"
)

var z *ZSet

func init() {
	z = NewZSet()
}

func BenchmarkZSet_Add(b *testing.B) {
	b.StopTimer()
	// data initialization
	scores := make([]float64, b.N)
	IDs := make([]int64, b.N)
	for i := range IDs {
		scores[i] = rand.Float64() + float64(rand.Int31n(99))
		IDs[i] = int64(i) + 100000
	}
	// BCE
	_ = scores[:b.N]
	_ = IDs[:b.N]

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		z.AddOrUpdate(IDs[i], SCORE(scores[i]), nil)
	}
}

func BenchmarkZSet_GetRank(b *testing.B) {
	l := z.Length()
	for i := 0; i < b.N; i++ {
		z.FindRank(100000 + int64(i)%l)
	}
}

func BenchmarkZSet_GetDataByRank(b *testing.B) {
	l := z.Length()

	if l == 0 {
		return
	}

	for i := 0; i < b.N; i++ {
		z.GetByRank(i%int(l), true)
	}
}

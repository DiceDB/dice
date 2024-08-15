package core_test

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/dicedb/dice/core"
	"gotest.tools/v3/assert"
)

var randGenerator *rand.Rand

func init() {
	randSeed := time.Now().UnixNano()
	randGenerator = rand.New(rand.NewSource(randSeed))
	fmt.Printf("rand seed: %v", randSeed)
}

var runes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_!@#$%^&*()-=+[]\\;':,.<>/?~.|")

func randStr(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = runes[randGenerator.Intn(len(runes))]
	}
	return string(b)
}

func TestDeqEncodeEntryString(t *testing.T) {
	testCases := []string{
		randStr(1),                // min 6 bit string
		randStr(10),               // 6 bit string
		randStr((1 << 6) - 1),     // max 6 bit string
		randStr(1 << 6),           // min 12 bit string
		randStr(2024),             // 12 bit string
		randStr((1 << 12) - 1),    // max 12 bit string
		randStr(1 << 12),          // min 32 bit string
		randStr((1 << 20) - 1000), // 32 bit string
		// randStr((1 << 32) - 1),   // max 32 bit string, maybe too huge to test..

		"0",                    // min 7 bit uint
		"28",                   // 7 bit uint
		"127",                  // max 7 bit uint
		"-4096",                // min 13 bit int
		"2024",                 // + 13 bit int
		"-2024",                // - 13 bit int
		"4095",                 // max 13 bit int
		"-32768",               // min 16 bit int
		"15384",                // + 16 bit int
		"-15384",               // - 16 bit int
		"32767",                // max 16 bit int
		"-8388608",             // min 24 bit int
		"4193301",              // + 24 bit int
		"-4193301",             // - 24 bit int
		"8388607",              // max 24 bit int
		"-2147483648",          // min 32 bit int
		"1073731765",           // + 32 bit int
		"-1073731765",          // - 32 bit int
		"2147483647",           // max 32 bit int
		"-9223372036854775808", // min 64 bit int
		"4611686018427287903",  // + 64 bit int
		"-4611686018427287903", // - 64 bit int
		"9223372036854775807",  // max 64 bit int
	}

	for _, tc := range testCases {
		x, _ := core.DecodeDeqEntry(core.EncodeDeqEntry(tc))
		assert.Equal(t, tc, x)
	}
}

func dequeRPushIntStrMany(howmany int, deq core.DequeI, b *testing.B) {
	for i := 0; i < howmany; i++ {
		deq.RPush(strconv.FormatInt(int64(i), 10))
	}
}

func dequeLPushIntStrMany(howmany int, deq core.DequeI, b *testing.B) {
	for i := 0; i < howmany; i++ {
		deq.LPush(strconv.FormatInt(int64(i), 10))
	}
}

func BenchmarkBasicDequeRPush20(b *testing.B) {
	for n := 0; n < b.N; n++ {
		dequeRPushIntStrMany(20, core.NewBasicDeque(), b)
	}
}

func BenchmarkBasicDequeRPush200(b *testing.B) {
	for n := 0; n < b.N; n++ {
		dequeRPushIntStrMany(200, core.NewBasicDeque(), b)
	}
}

func BenchmarkBasicDequeRPush2000(b *testing.B) {
	for n := 0; n < b.N; n++ {
		dequeRPushIntStrMany(2000, core.NewBasicDeque(), b)
	}
}

func BenchmarkDequeRPush20(b *testing.B) {
	for n := 0; n < b.N; n++ {
		dequeRPushIntStrMany(20, core.NewDeque(), b)
	}
}

func BenchmarkDequeRPush200(b *testing.B) {
	for n := 0; n < b.N; n++ {
		dequeRPushIntStrMany(200, core.NewDeque(), b)
	}
}

func BenchmarkDequeRPush2000(b *testing.B) {
	for n := 0; n < b.N; n++ {
		dequeRPushIntStrMany(2000, core.NewDeque(), b)
	}
}

func BenchmarkDequeLPush20(b *testing.B) {
	for n := 0; n < b.N; n++ {
		dequeLPushIntStrMany(20, core.NewDeque(), b)
	}
}

func BenchmarkDequeLPush200(b *testing.B) {
	for n := 0; n < b.N; n++ {
		dequeLPushIntStrMany(200, core.NewDeque(), b)
	}
}

func BenchmarkDequeLPush2000(b *testing.B) {
	for n := 0; n < b.N; n++ {
		dequeLPushIntStrMany(2000, core.NewDeque(), b)
	}
}

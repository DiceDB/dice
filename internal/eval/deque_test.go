// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package eval_test

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/dicedb/dice/internal/eval"
	"github.com/stretchr/testify/assert"
)

var deqRandGenerator *rand.Rand

func deqTestInit() {
	randSeed := time.Now().UnixNano()
	deqRandGenerator = rand.New(rand.NewSource(randSeed))
	fmt.Printf("rand seed: %v", randSeed)
}

var deqRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_!@#$%^&*()-=+[]\\;':,.<>/?~.|")

func deqRandStr(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = deqRunes[deqRandGenerator.Intn(len(deqRunes))]
	}
	return string(b)
}

func TestDeqEncodeEntryString(t *testing.T) {
	deqTestInit()
	testCases := []string{
		deqRandStr(1),                // min 6 bit string
		deqRandStr(10),               // 6 bit string
		deqRandStr((1 << 6) - 1),     // max 6 bit string
		deqRandStr(1 << 6),           // min 12 bit string
		deqRandStr(2024),             // 12 bit string
		deqRandStr((1 << 12) - 1),    // max 12 bit string
		deqRandStr(1 << 12),          // min 32 bit string
		deqRandStr((1 << 20) - 1000), // 32 bit string
		// randStr((1 << 32) - 1),   // max 32 bit string, maybe too huge to test.

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
		x, _ := eval.DecodeDeqEntry(eval.EncodeDeqEntry(tc))
		assert.Equal(t, tc, x)
	}
}

func dequeRPushIntStrMany(howmany int, deq eval.DequeI) {
	for i := 0; i < howmany; i++ {
		deq.RPush(strconv.FormatInt(int64(i), 10))
	}
}

func dequeLPushIntStrMany(howmany int, deq eval.DequeI) {
	for i := 0; i < howmany; i++ {
		deq.LPush(strconv.FormatInt(int64(i), 10))
	}
}

func dequeLInsertIntStrMany(howMany int, beforeAfter string, deq eval.DequeI) {
	const pivot string = "10"
	const element string = "50"
	deq.LPush(pivot)
	for i := 0; i < howMany; i++ {
		deq.LInsert(pivot, element, beforeAfter)
	}
}

func BenchmarkBasicDequeLInsertBefore2000(b *testing.B) {
	for n := 0; n < b.N; n++ {
		dequeLInsertIntStrMany(2000, "before", eval.NewBasicDeque())
	}
}

func BenchmarkBasicDequeLInsertAfter2000(b *testing.B) {
	for n := 0; n < b.N; n++ {
		dequeLInsertIntStrMany(2000, "after", eval.NewBasicDeque())
	}
}

func BenchmarkDequeLInsertBefore2000(b *testing.B) {
	for n := 0; n < b.N; n++ {
		dequeLInsertIntStrMany(2000, "before", eval.NewDeque())
	}
}

func BenchmarkDequeLInsertAfter2000(b *testing.B) {
	for n := 0; n < b.N; n++ {
		dequeLInsertIntStrMany(2000, "after", eval.NewDeque())
	}
}

func BenchmarkBasicDequeRPush20(b *testing.B) {
	for n := 0; n < b.N; n++ {
		dequeRPushIntStrMany(20, eval.NewBasicDeque())
	}
}

func BenchmarkBasicDequeRPush200(b *testing.B) {
	for n := 0; n < b.N; n++ {
		dequeRPushIntStrMany(200, eval.NewBasicDeque())
	}
}

func BenchmarkBasicDequeRPush2000(b *testing.B) {
	for n := 0; n < b.N; n++ {
		dequeRPushIntStrMany(2000, eval.NewBasicDeque())
	}
}

func BenchmarkDequeRPush20(b *testing.B) {
	for n := 0; n < b.N; n++ {
		dequeRPushIntStrMany(20, eval.NewDeque())
	}
}

func BenchmarkDequeRPush200(b *testing.B) {
	for n := 0; n < b.N; n++ {
		dequeRPushIntStrMany(200, eval.NewDeque())
	}
}

func BenchmarkDequeRPush2000(b *testing.B) {
	for n := 0; n < b.N; n++ {
		dequeRPushIntStrMany(2000, eval.NewDeque())
	}
}

func BenchmarkDequeLPush20(b *testing.B) {
	for n := 0; n < b.N; n++ {
		dequeLPushIntStrMany(20, eval.NewDeque())
	}
}

func BenchmarkDequeLPush200(b *testing.B) {
	for n := 0; n < b.N; n++ {
		dequeLPushIntStrMany(200, eval.NewDeque())
	}
}

func BenchmarkDequeLPush2000(b *testing.B) {
	for n := 0; n < b.N; n++ {
		dequeLPushIntStrMany(2000, eval.NewDeque())
	}
}

func TestLRange(t *testing.T) {
	testCases := []struct {
		name           string
		dq             eval.DequeI
		input          []string
		expectedOutput []string
		start          int64
		stop           int64
	}{
		{"DequeWithStartStopPositiveAndInRange", eval.NewDeque(), []string{"a", "b", "c"}, []string{"c", "b", "a"}, 0, 2},
		{"DequeWhereStopIsOutOfRange", eval.NewDeque(), []string{"a", "b", "c"}, []string{"c", "b", "a"}, 0, 20},
		{"DequeWhereStartIsOutOfRange", eval.NewDeque(), []string{"a", "b", "c"}, []string{}, 10, 2},
		{"DequeWhereStartIsNegative", eval.NewDeque(), []string{"a", "b", "c"}, []string{"b", "a"}, -2, 2},
		{"DequeWhereStartIsNegativeOutOfRange", eval.NewDeque(), []string{"a", "b", "c"}, []string{"c", "b", "a"}, -20, 2},
		{"DequeWhereStopIsNegative", eval.NewDeque(), []string{"a", "b", "c"}, []string{"c", "b"}, 0, -2},
		{"DequeWhereStopIsNegativeOutOfRange", eval.NewDeque(), []string{"a", "b", "c"}, []string{}, 0, -4},
		{"DequeWhereStartGreaterThanStop", eval.NewDeque(), []string{"a", "b", "c"}, []string{}, 2, 0},
		{"BasicDequeWithStartStopPositiveAndInRange", eval.NewBasicDeque(), []string{"a", "b", "c"}, []string{"c", "b", "a"}, 0, 2},
		{"BasicDequeWhereStopIsOutOfRange", eval.NewBasicDeque(), []string{"a", "b", "c"}, []string{"c", "b", "a"}, 0, 20},
		{"BasicDequeWhereStartIsOutOfRange", eval.NewBasicDeque(), []string{"a", "b", "c"}, []string{}, 10, 2},
		{"BasicDequeWhereStartIsNegative", eval.NewBasicDeque(), []string{"a", "b", "c"}, []string{"b", "a"}, -2, 2},
		{"BasicDequeWhereStartIsNegativeOutOfRange", eval.NewBasicDeque(), []string{"a", "b", "c"}, []string{"c", "b", "a"}, -20, 2},
		{"BasicDequeWhereStopIsNegative", eval.NewBasicDeque(), []string{"a", "b", "c"}, []string{"c", "b"}, 0, -2},
		{"BasicDequeWhereStopIsNegativeOutOfRange", eval.NewBasicDeque(), []string{"a", "b", "c"}, []string{}, 0, -4},
		{"BasicDequeWhereStartGreaterThanStop", eval.NewBasicDeque(), []string{"a", "b", "c"}, []string{}, 2, 0},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, i := range tc.input {
				tc.dq.LPush(i)
			}
			output, _ := tc.dq.LRange(tc.start, tc.stop)
			assert.ElementsMatch(t, output, tc.expectedOutput)

		})
	}
}

func TestLInsertOnInvalidOperationTypeReturnsError(t *testing.T) {
	testCases := []struct {
		name string
		dq   eval.DequeI
	}{
		{"WithDeque", eval.NewDeque()},
		{"WithBasicDeque", eval.NewBasicDeque()},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.dq.LPush("a")
			tc.dq.LPush("b")
			tc.dq.LPush("c")
			newLen, err := tc.dq.LInsert("a", "x", "randomOperation")
			if err == nil || err.Error() != "syntax error" {
				t.Errorf("Expected error 'syntax error', got %v", err)
			}
			if newLen != -1 {
				t.Errorf("Expected int -1, got %v", newLen)
			}
		})
	}
}

func TestLInsertBasicDeque(t *testing.T) {
	dq := eval.NewBasicDeque()
	dq.RPush("a")
	dq.RPush("b")
	dq.RPush("c")
	testCases := []struct {
		name                  string
		pivotElement          string
		elementToBeInserted   string
		beforeAfter           string
		expectedOutput        int64
		expectedErr           error
		expectedElementsOrder []string
	}{
		{"InMiddleBefore", "b", "d", "before", 4, nil, []string{"a", "d", "b", "c"}},
		{"AtFrontBefore", "a", "e", "before", 5, nil, []string{"e", "a", "d", "b", "c"}},
		{"AtEndBefore", "c", "f", "before", 6, nil, []string{"e", "a", "d", "b", "f", "c"}},
		{"InMiddleAfter", "b", "g", "after", 7, nil, []string{"e", "a", "d", "b", "g", "f", "c"}},
		{"AtFrontAfter", "e", "h", "after", 8, nil, []string{"e", "h", "a", "d", "b", "g", "f", "c"}},
		{"AtEndAfter", "c", "i", "after", 9, nil, []string{"e", "h", "a", "d", "b", "g", "f", "c", "i"}},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := dq.LInsert(tc.pivotElement, tc.elementToBeInserted, tc.beforeAfter)
			if result != tc.expectedOutput {
				t.Errorf("Expected %v, got %v.", tc.expectedOutput, result)
			}
			if err != tc.expectedErr {
				t.Errorf("Expected error %v, got %v", tc.expectedErr, err)
			}
			iter := dq.NewIterator()
			for i, expected := range tc.expectedElementsOrder {
				val, err := iter.Next()
				if err != nil {
					t.Errorf("Error iterating deque: %v", err)
				}
				if strings.Compare(val, expected) != 0 {
					t.Errorf("Expected value %d to be '%s', got '%s'", i, expected, val)
				}
			}
		})
	}
}

type DequeLInsertFixture struct {
	dq                   *eval.Deque
	initialElements      []string
	elementsToBeInserted []string
}

func newDequeLInsertFixture() *DequeLInsertFixture {
	dq := eval.NewDeque()
	initElements := []string{deqRandStr(10), deqRandStr(100), deqRandStr(250), deqRandStr(150), deqRandStr(200)}
	for _, elem := range initElements {
		dq.LPush(elem)
	}
	return &DequeLInsertFixture{
		dq,
		initElements,
		[]string{deqRandStr(30), deqRandStr(50), deqRandStr(80), deqRandStr(130)},
	}
}

func TestDequeLInsertBefore(t *testing.T) {
	deqTestInit()
	fixture := newDequeLInsertFixture()
	testCases := []struct {
		name                  string
		pivotElement          string
		elementToBeInserted   string
		beforeAfter           string
		expectedOutput        int64
		expectedErr           error
		expectedElementsOrder []string
	}{
		{"WhenPivotInMiddleOfHeadNode",
			fixture.initialElements[3],
			fixture.elementsToBeInserted[0],
			"before",
			6,
			nil,
			[]string{fixture.initialElements[4], fixture.elementsToBeInserted[0], fixture.initialElements[3], fixture.initialElements[2], fixture.initialElements[1], fixture.initialElements[0]},
		},
		{"WhenPivotAtStartOfHeadNode",
			fixture.initialElements[4],
			fixture.elementsToBeInserted[1],
			"before",
			7,
			nil,
			[]string{fixture.elementsToBeInserted[1], fixture.initialElements[4], fixture.elementsToBeInserted[0], fixture.initialElements[3], fixture.initialElements[2], fixture.initialElements[1], fixture.initialElements[0]},
		},
		{"WhenPivotAtStartOfNonHeadNode",
			fixture.initialElements[2],
			fixture.elementsToBeInserted[2],
			"before",
			8,
			nil,
			[]string{fixture.elementsToBeInserted[1], fixture.initialElements[4], fixture.elementsToBeInserted[0], fixture.initialElements[3], fixture.elementsToBeInserted[2], fixture.initialElements[2], fixture.initialElements[1], fixture.initialElements[0]},
		},
		{"WhenPivotInMiddleOfNonHeadNode",
			fixture.initialElements[1],
			fixture.elementsToBeInserted[3],
			"before",
			9,
			nil,
			[]string{fixture.elementsToBeInserted[1], fixture.initialElements[4], fixture.elementsToBeInserted[0], fixture.initialElements[3], fixture.elementsToBeInserted[2], fixture.initialElements[2], fixture.elementsToBeInserted[3], fixture.initialElements[1], fixture.initialElements[0]},
		},
		{"WhenPivotDoesNotExist",
			"pivot",
			"element",
			"before",
			-1,
			nil,
			[]string{fixture.elementsToBeInserted[1], fixture.initialElements[4], fixture.elementsToBeInserted[0], fixture.initialElements[3], fixture.elementsToBeInserted[2], fixture.initialElements[2], fixture.elementsToBeInserted[3], fixture.initialElements[1], fixture.initialElements[0]},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := fixture.dq.LInsert(tc.pivotElement, tc.elementToBeInserted, tc.beforeAfter)
			if result != tc.expectedOutput {
				t.Errorf("Expected %v, got %v.", tc.expectedOutput, result)
			}
			if err != tc.expectedErr {
				t.Errorf("Expected error %v, got %v", tc.expectedErr, err)
			}
			iter := fixture.dq.NewIterator()
			for i, expected := range tc.expectedElementsOrder {
				val, err := iter.Next()
				if err != nil {
					t.Errorf("Error iterating deque: %v", err)
				}
				if strings.Compare(val, expected) != 0 {
					t.Errorf("Expected value %d to be '%s', got '%s'", i, expected, val)
				}
			}
		})
	}
}

func TestLInsertAfter(t *testing.T) {
	deqTestInit()
	fixture := newDequeLInsertFixture()
	testCases := []struct {
		name                  string
		pivotElement          string
		elementToBeInserted   string
		beforeAfter           string
		expectedOutput        int64
		expectedErr           error
		expectedElementsOrder []string
	}{
		{"WhenPivotInMiddleOfTailNode",
			fixture.initialElements[1],
			fixture.elementsToBeInserted[0],
			"after",
			6,
			nil,
			[]string{fixture.initialElements[4], fixture.initialElements[3], fixture.initialElements[2], fixture.initialElements[1], fixture.elementsToBeInserted[0], fixture.initialElements[0]},
		},
		{"WhenPivotAtEndOfTailNode",
			fixture.initialElements[0],
			fixture.elementsToBeInserted[1],
			"after",
			7,
			nil,
			[]string{fixture.initialElements[4], fixture.initialElements[3], fixture.initialElements[2], fixture.initialElements[1], fixture.elementsToBeInserted[0], fixture.initialElements[0], fixture.elementsToBeInserted[1]},
		},
		{"WhenPivotAtEndOfNonTailNode",
			fixture.initialElements[3],
			fixture.elementsToBeInserted[2],
			"after",
			8,
			nil,
			[]string{fixture.initialElements[4], fixture.initialElements[3], fixture.elementsToBeInserted[2], fixture.initialElements[2], fixture.initialElements[1], fixture.elementsToBeInserted[0], fixture.initialElements[0], fixture.elementsToBeInserted[1]},
		},
		{"WhenPivotInMiddleOfNonLastNode",
			fixture.initialElements[4],
			fixture.elementsToBeInserted[3],
			"after",
			9,
			nil,
			[]string{fixture.initialElements[4], fixture.elementsToBeInserted[3], fixture.initialElements[3], fixture.elementsToBeInserted[2], fixture.initialElements[2], fixture.initialElements[1], fixture.elementsToBeInserted[0], fixture.initialElements[0], fixture.elementsToBeInserted[1]},
		},
		{"WhenPivotDoesNotExist",
			"pivot",
			"element",
			"after",
			-1,
			nil,
			[]string{fixture.initialElements[4], fixture.elementsToBeInserted[3], fixture.initialElements[3], fixture.elementsToBeInserted[2], fixture.initialElements[2], fixture.initialElements[1], fixture.elementsToBeInserted[0], fixture.initialElements[0], fixture.elementsToBeInserted[1]},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := fixture.dq.LInsert(tc.pivotElement, tc.elementToBeInserted, tc.beforeAfter)
			if result != tc.expectedOutput {
				t.Errorf("Expected %v, got %v.", tc.expectedOutput, result)
			}
			if err != tc.expectedErr {
				t.Errorf("Expected error %v, got %v", tc.expectedErr, err)
			}
			iter := fixture.dq.NewIterator()
			for i, expected := range tc.expectedElementsOrder {
				val, err := iter.Next()
				if err != nil {
					t.Errorf("Error iterating deque: %v", err)
				}
				if strings.Compare(val, expected) != 0 {
					t.Errorf("Expected value %d to be '%s', got '%s'", i, expected, val)
				}
			}
		})
	}
}

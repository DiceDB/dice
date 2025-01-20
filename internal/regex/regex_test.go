// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package regex

import (
	"testing"

	"github.com/dicedb/dice/internal/server/utils"
)

func TestWildCardMatch(t *testing.T) {
	tests := []struct {
		pattern string
		key     string
		want    bool
	}{
		{"*", "anything", true},
		{"*", utils.EmptyStr, true},
		{"?", "a", true},
		{"?", utils.EmptyStr, false},
		{"a?", "ab", true},
		{"a?", "a", false},
		{"a*", "abc", true},
		{"a*", "a", true},
		{"a*b", "ab", true},
		{"a*b", "acb", true},
		{"a*b", "aXb", true},
		{"a*b", "abbb", true},
		{"a*b", "a", false},
		{"a*b", "abX", false},
		{"a*b*c", "abc", true},
		{"a*b*c", "aXbYc", true},
		{"a*b*c", "aXYbYZc", true},
		{"a*b*c", "abcX", false},
		{"a*b*c", "aXbYcZ", false},
		{"a?b", "aXb", true},
		{"a?b", "ab", false},
		{"a?b", "aXYb", false},
		{"a??b", "aXYb", true},
		{"a??b", "aXb", false},
		{"a?b*", "aXbY", true},
		{"a?b*", "aXb", true},
		{"a?b*", "ab", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.key, func(t *testing.T) {
			if got := WildCardMatch(tt.pattern, tt.key); got != tt.want {
				t.Errorf("WildcardMatch(%q, %q) = %v, want %v", tt.pattern, tt.key, got, tt.want)
			}
		})
	}
}

func BenchmarkWildCardMatch(b *testing.B) {
	testCases := []struct {
		pattern string
		key     string
	}{
		{"*", "anystringwillmatch"},
		{"?????", "exact"},
		{"a?c*d", "abcdefgd"},
		{"*test*", "thisIsATestString"},
		{"???*", "abcdefghijklmnop"},
		{"*a*b*c*", "111a222b333c444"},
		{"a?b*c??d", "acb123cxxd"},
		{"*a*b?c*d*", "111aaa222bxc333ddd444"},
		{"a*b*c*d*e*", "axbxxcxxxdxxxxe"},
		{"*a*b*c*d*e*", "11aa22bb33cc44dd55ee66"},
	}

	for _, tc := range testCases {
		b.Run(tc.pattern, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				WildCardMatch(tc.pattern, tc.key)
			}
		})
	}
}

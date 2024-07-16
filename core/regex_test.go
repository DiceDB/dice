package core_test

import (
	"testing"

	"github.com/dicedb/dice/core"
)

func TestRegexMatch(t *testing.T) {
	tests := []struct {
		pattern string
		key     string
		match   bool
	}{
		{"*", "anything", true},
		{"*", "", true},
		{"?", "a", true},
		{"?", "", false},
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
			match := core.RegexMatch(tt.pattern, tt.key)
			if match != tt.match {
				t.Errorf("RegexMatch(%q, %q) = %v; want %v", tt.pattern, tt.key, match, tt.match)
			}
		})
	}
}

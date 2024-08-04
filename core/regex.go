package core

import (
	"regexp"
	"strings"
)

func wildcardToRegex(pattern string) string {
	// Escape any regex special characters
	rePattern := regexp.QuoteMeta(pattern)
	// Replace escaped * and ? with their regex equivalents
	rePattern = strings.ReplaceAll(rePattern, `\*`, `.*`)
	rePattern = strings.ReplaceAll(rePattern, `\?`, `.`)
	// Add anchors to match the whole string
	return "^" + rePattern + "$"
}

// RegexMatch checks if the key matches the pattern using * and ? as wildcards
func RegexMatch(pattern, key string) bool {
	rePattern := wildcardToRegex(pattern)
	re := regexp.MustCompile(rePattern)
	return re.MatchString(key)
}

// WildCardMatch checks if the key matches the pattern using * and ? as wildcards using two pointer approach
func WildCardMatch(pattern, key string) bool {
	patternLen := len(pattern)
	keyLen := len(key)
	patternIndex := 0
	keyIndex := 0
	starIndex := -1
	kIndex := -1

	for keyIndex < keyLen {
		if patternIndex < patternLen && (pattern[patternIndex] == '?' || pattern[patternIndex] == key[keyIndex]) {
			patternIndex++
			keyIndex++
		} else if patternIndex < patternLen && pattern[patternIndex] == '*' {
			starIndex = patternIndex
			kIndex = keyIndex
			patternIndex++
		} else if starIndex != -1 {
			patternIndex = starIndex + 1
			kIndex++
			keyIndex = kIndex
		} else {
			return false
		}
	}

	for patternIndex < patternLen && pattern[patternIndex] == '*' {
		patternIndex++
	}

	return patternIndex == patternLen
}

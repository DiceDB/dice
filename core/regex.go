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

// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package regex

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

// This file is part of DiceDB.
// Copyright (C) 2024 DiceDB (dicedb.io).
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

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

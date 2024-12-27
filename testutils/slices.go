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

package testutils

// UnorderedEqual compares two slices of interfaces and returns true if they
// contain the same elements, regardless of order.
func UnorderedEqual(expected, actual interface{}) bool {
	expectedSlice, ok := expected.([]interface{})
	if !ok {
		return false
	}

	actualSlice, ok := actual.([]interface{})
	if !ok {
		return false
	}

	if len(expectedSlice) != len(actualSlice) {
		return false
	}

	for _, e := range expectedSlice {
		found := false
		for _, a := range actualSlice {
			if e == a {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// EqualByteSlice compares two byte slices and returns true if they are equal.
func EqualByteSlice(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

// EqualInt64Slice compares two int64 slices and returns true if they are equal.
func EqualInt64Slice(a, b []int64) bool {
	if len(a) != len(b) {
		return false
	}

	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

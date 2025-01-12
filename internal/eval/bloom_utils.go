// This file is part of DiceDB.
// Copyright (C) 2025DiceDB (dicedb.io).
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

package eval

// setBit sets the bit at index `b` to "1" in `buf`.
func setBit(buf []byte, b uint64) {
	idx, offset := b/8, 7-b%8
	if idx >= uint64(len(buf)) {
		return
	}

	buf[idx] |= 1 << offset
}

// isBitSet checks if the bit at index `b` is set to "1" or not in `buf`.
func isBitSet(buf []byte, b uint64) bool {
	idx, offset := b/8, 7-b%8
	if idx >= uint64(len(buf)) {
		return false
	}

	if buf[idx]&(1<<offset) == 1<<offset {
		return true
	}
	return false
}

// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

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

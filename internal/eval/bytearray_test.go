// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package eval

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMixedOperations(t *testing.T) {
	byteArray := NewByteArray(8)

	// Mixed operations
	byteArray.SetBit(0, true)
	byteArray.SetBit(1, true)
	byteArray.SetBit(2, false)
	byteArray.SetBit(3, true)
	byteArray.SetBit(4, true)

	assert.Equal(t, byteArray.BitCount(), 4, "Total set bits should be 4")

	byteArray.SetBit(0, false)
	byteArray.SetBit(1, false)

	assert.Equal(t, byteArray.BitCount(), 2, "Total set bits should be 2")
}

func TestByteArray(t *testing.T) {
	byteArray := NewByteArray(100) // Larger array size

	// Test SetBit and GetBit
	byteArray.SetBit(10, true)
	assert.Equal(t, byteArray.GetBit(10), true, "Bit at position 10 should be set to true")

	byteArray.SetBit(10, false)
	assert.Equal(t, byteArray.GetBit(10), false, "Bit at position 10 should be set to false")

	// Test BitCount with multiple bits
	byteArray.SetBit(10, true)
	byteArray.SetBit(15, true)
	byteArray.SetBit(100, true)
	byteArray.SetBit(200, true)
	byteArray.SetBit(300, true)
	assert.Equal(t, byteArray.BitCount(), 5, "Total set bits should be 5")

	byteArray.SetBit(15, false)
	assert.Equal(t, byteArray.BitCount(), 4, "Total set bits should be 4 after unsetting bit 15")

	// Test edge/boundary cases
	byteArray.SetBit(799, true)
	assert.Equal(t, byteArray.GetBit(799), true, "Bit at position 799 should be set to true")

	byteArray.SetBit(799, false)
	assert.Equal(t, byteArray.GetBit(799), false, "Bit at position 799 should be set to false")

	byteArray.SetBit(0, true)
	assert.Equal(t, byteArray.GetBit(0), true, "Bit at position 0 should be set to true")

	byteArray.SetBit(0, false)
	assert.Equal(t, byteArray.GetBit(0), false, "Bit at position 0 should be set to false")
}

func TestLargeByteArray(t *testing.T) {
	byteArray := NewByteArray(10000) // Even larger array size

	var setcount = 0
	// Set and test bits in various positions
	for i := 0; i < 10000*8; i += 1000 {
		byteArray.SetBit(i, true)
		assert.Equal(t, byteArray.GetBit(i), true, "Bit at position should be set to true")
		setcount++
	}

	// Test BitCount
	assert.Equal(t, byteArray.BitCount(), setcount, "Total set bits should match the number of set positions")

	// Unset and test bits
	for i := 0; i < 10000*8; i += 1000 {
		byteArray.SetBit(i, false)
		assert.Equal(t, byteArray.GetBit(i), false, "Bit at position should be set to false")
	}

	// Test BitCount after unsetting
	assert.Equal(t, byteArray.BitCount(), 0, "Total set bits should be 0 after unsetting all bits")
}

func BenchmarkLargeByteArray1(t *testing.B) {
	byteArray := NewByteArray(10000)

	for i := 0; i < 10000*8; i += 100 {
		byteArray.SetBit(i, true)
		assert.Equal(t, byteArray.GetBit(i), true, "Bit at position should be set to true")
		byteArray.BitCount()
	}
}

func BenchmarkLargeByteArray2(t *testing.B) {
	byteArray := NewByteArray(1000000)

	for i := 0; i < 1000000*8; i += 500 {
		byteArray.SetBit(i, true)
		assert.Equal(t, byteArray.GetBit(i), true, "Bit at position should be set to true")
		byteArray.BitCount()
	}
}

func TestReverseByte(t *testing.T) {
	byteArray := NewByteArray(1) // Larger array size

	byteArray.SetBit(2, true)
	byteArray.SetBit(4, true)

	reversedByte := reverseByte(byteArray.data[0])

	assert.Equal(t, reversedByte, byte(0b00010100), "Reversed byte should be 0b00010100")
}

func TestDeepCopy(t *testing.T) {
	original := NewByteArray(8)

	// Mixed operations
	original.SetBit(0, true)
	original.SetBit(1, false)
	original.SetBit(2, false)
	original.SetBit(3, true)

	// Create a deep deepCopy of the ByteArray
	deepCopy := original.DeepCopy()

	// Verify that the deepCopy is not nil
	if deepCopy == nil {
		t.Fatalf("DeepCopy returned nil, expected a valid deepCopy")
	}

	// Verify that the data slice is correctly copied
	assert.Equal(t, len(deepCopy.data), len(original.data), "ByteArray DeepCopy data length mismatch")

	for i := range original.data {
		assert.Equal(t, deepCopy.data[i], original.data[i], "ByteArray DeepCopy data element mismatch")
	}

	// Verify that the Length field is correctly copied
	assert.Equal(t, deepCopy.Length, original.Length, "ByteArray DeepCopy Length mismatch")

	// Modify the deepCopy's data to ensure it's independent of the original
	deepCopy.data[0] = 9
	assert.True(t, deepCopy.data[0] != original.data[0], "ByteArray DeepCopy did not create an independent deepCopy, original and deepCopy data are linked")

	// Modify the original's data to ensure it doesn't affect the deepCopy
	original.data[1] = 8
	assert.True(t, deepCopy.data[1] != original.data[1], "ByteArray DeepCopy did not create an independent deepCopy, original and deepCopy data are linked")
}

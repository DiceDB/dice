// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package dencoding_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/dicedb/dice/internal/dencoding"
	"github.com/dicedb/dice/testutils"
	"github.com/stretchr/testify/assert"
)

func TestDencodingUInt(t *testing.T) {
	t.Run("EncodeDecodeRoundTrip", func(t *testing.T) {
		// Test that encoding and then decoding an uint64 value returns the original value
		tests := generateUIntTestCases()
		for _, v := range tests {
			t.Run(fmt.Sprintf("value_%d", v), func(t *testing.T) {
				assert.Equal(t, v, dencoding.DecodeUInt(dencoding.EncodeUInt(v)), "Encode-Decode round trip failed")
			})
		}
	})

	t.Run("EncodedLength", func(t *testing.T) {
		// Test that encoding and then decoding an uint64 value returns the original value
		testCases := map[uint64]int{
			0:   1, // 0 should be encoded as 1 byte
			127: 1, // 127 should be encoded as 1 byte
			128: 2, // 128 should be encoded as 2 bytes
			129: 2, // 129 should be encoded as 2 bytes
		}
		for value, expectedLength := range testCases {
			t.Run(fmt.Sprintf("value_%d", value), func(t *testing.T) {
				encoded := dencoding.EncodeUInt(value)
				assert.Equal(t, expectedLength, len(encoded), "Unexpected encoded length")
			})
		}
	})

	t.Run("SpecificEncodings", func(t *testing.T) {
		// Test specific uint64 values are encoded to the expected byte sequences
		testCases := map[int][]byte{
			1:   {0b00000001},
			2:   {0b00000010},
			127: {0b01111111},
			128: {0b10000000, 0b00000001},
			129: {0b10000001, 0b00000001},
			130: {0b10000010, 0b00000001},
			131: {0b10000011, 0b00000001},
		}
		for value, expected := range testCases {
			t.Run(fmt.Sprintf("value_%d", value), func(t *testing.T) {
				encoded := dencoding.EncodeUInt(uint64(value))
				assert.True(t, testutils.EqualByteSlice(encoded, expected), "Unexpected encoding")
			})
		}
	})
}

func TestDencodingInt(t *testing.T) {
	t.Run("EncodeDecodeRoundTrip", func(t *testing.T) {
		// Test that encoding and then decoding an int64 value returns the original value
		tests := generateIntTestCases()
		for _, v := range tests {
			t.Run(fmt.Sprintf("value_%d", v), func(t *testing.T) {
				assert.Equal(t, v, dencoding.DecodeInt(dencoding.EncodeInt(v)), "Encode-Decode round trip failed")
			})
		}
	})

	t.Run("EncodedLength", func(t *testing.T) {
		// Test that the encoded length of int64 values is as expected
		testCases := map[int64]int{
			0:   1, // 0 should be encoded as 1 byte
			127: 2, // 127 should be encoded as 2 bytes
			128: 2, // 128 should be encoded as 2 bytes
			129: 2, // 129 should be encoded as 2 bytes
		}
		for value, expectedLength := range testCases {
			t.Run(fmt.Sprintf("value_%d", value), func(t *testing.T) {
				encoded := dencoding.EncodeInt(value)
				assert.Equal(t, expectedLength, len(encoded), "Unexpected encoded length")
			})
		}
	})

	t.Run("SpecificEncodings", func(t *testing.T) {
		// Test specific int64 values are encoded to the expected byte sequences
		testCases := map[int][]byte{
			0:    {0x00},
			1:    {0x02},
			-1:   {0x01},
			63:   {0x7E},
			-64:  {0x7F},
			64:   {0x80, 0x01},
			-65:  {0x81, 0x01},
			127:  {0xFE, 0x01},
			128:  {0x80, 0x02},
			-128: {0xFF, 0x01},
			-129: {0x81, 0x02},
		}
		for value, expected := range testCases {
			t.Run(fmt.Sprintf("value_%d", value), func(t *testing.T) {
				encoded := dencoding.EncodeInt(int64(value))
				assert.True(t, testutils.EqualByteSlice(encoded, expected), "Unexpected encoding")
			})
		}
	})
}

func TestEncodeDecodeInt64Min(t *testing.T) {
	// Test specific int64 values are encoded to the expected byte sequences
	value := int64(math.MinInt64)
	encoded := dencoding.EncodeInt(value)
	decoded := dencoding.DecodeInt(encoded)
	assert.Equal(t, value, decoded, "Encode-Decode failed for MinInt64")
}

func generateUIntTestCases() []uint64 {
	// Generate a set of uint64 test cases including edge cases
	var tests []uint64
	tests = append(tests, 0)
	for i := 0; i < 64; i++ {
		tests = append(tests, 1<<i, (1<<i)-1, (1<<i)+1)
	}
	return tests
}

func generateIntTestCases() []int64 {
	// Generate a set of int64 test cases including edge cases and negative values
	var tests []int64
	tests = append(tests, 0)
	for i := 0; i < 64; i++ {
		tests = append(tests, 1<<i, (1<<i)-1, (1<<i)+1, -1<<i, -(1<<i)-1, -(1<<i)+1)
	}
	return tests
}

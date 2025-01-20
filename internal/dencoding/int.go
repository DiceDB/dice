// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package dencoding

import "sync"

var BITMASK = []byte{
	0x01,
	0x03,
	0x07,
	0x0F,
	0x1F,
	0x3F,
	0x7F,
	0xFF,
}

// getLSB returns the least significant `n` bits from
// the byte value `x`.
func getLSB(x byte, n uint8) byte {
	if n > 8 {
		panic("can extract at max 8 bits from the number")
	}
	return x & BITMASK[n-1]
}

var bufPool = sync.Pool{
	New: func() any {
		return new([11]byte)
	},
}

var bitShifts = [10]uint8{7, 7, 7, 7, 7, 7, 7, 7, 7, 1}

// EncodeUInt encodes the unsigned 64-bit integer value into a varint and returns an array of bytes (little endian
// encoded)
func EncodeUInt(x uint64) []byte {
	var i int
	buf := bufPool.Get().(*[11]byte)
	for i = 0; i < len(bitShifts); i++ {
		buf[i] = getLSB(byte(x), bitShifts[i]) | 0b10000000 // marking the continuation bit
		x >>= bitShifts[i]
		if x == 0 {
			break
		}
	}
	buf[i] &= 0b01111111 // marking the termination bit
	newBuf := append(make([]byte, 0, i+1), buf[:i+1]...)
	bufPool.Put(buf)
	return newBuf
}

// DecodeUInt decodes the array of bytes and returns an unsigned 64-bit integer
func DecodeUInt(vint []byte) uint64 {
	var i int
	var v uint64 = 0
	for i = 0; i < len(vint); i++ {
		b := getLSB(vint[i], 7)
		v |= uint64(b) << (7 * i)
	}
	return v
}

// EncodeInt encodes the signed 64-bit integer value into a varint
// and returns an array of bytes (little endian encoded)
func EncodeInt(x int64) []byte {
	return EncodeUInt(uint64((x << 1) ^ (x >> 63)))
}

// DecodeInt decodes the array of bytes and returns a signed 64-bit integer
func DecodeInt(vint []byte) int64 {
	ux := DecodeUInt(vint)
	return int64((ux >> 1) ^ uint64((int64(ux)<<63)>>63))
}

// EncodeUIntRevInPlace encodes the uint64 value into a reversed-varint in place,
// produce the same result as slices.Reverse(EncodeUInt(x))
func EncodeUIntRevInPlace(x uint64, buf []byte) {
	var i int
	for i = 0; i < len(bitShifts); i++ {
		buf[len(buf)-i-1] = getLSB(byte(x), bitShifts[i]) | 0b10000000 // marking the continuation bit
		x >>= bitShifts[i]
		if x == 0 {
			break
		}
	}
	buf[0] &= 0b01111111 // marking the termination bit
}

// GetEncodeUIntSize returns the size in byte the encoded varint will take
func GetEncodeUIntSize(x uint64) uint64 {
	if x <= 127 {
		return 1
	} else if x < 16383 {
		return 2
	} else if x < 2097151 {
		return 3
	} else if x < 268435455 {
		return 4
	}
	return 5
}

// DecodeUIntRev decodes the varint encoded by EncodeUIntRev[InPlace]
func DecodeUIntRev(vint []byte) uint64 {
	var i int
	var v uint64 = 0
	for i = 0; i < len(vint); i++ {
		b := getLSB(vint[len(vint)-i-1], 7)
		v |= uint64(b) << (7 * i)
	}
	return v
}

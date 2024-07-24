package core

type ByteArray struct {
	data []byte
}

// NewByteArray initializes a new ByteArray with the given size
func NewByteArray(size int) *ByteArray {
	return &ByteArray{
		data: make([]byte, size),
	}
}

// SetBit sets the bit at the given position to the specified value
func (b *ByteArray) SetBit(pos int, value bool) {
	byteIndex := pos / 8
	bitIndex := uint(pos % 8)

	if value {
		b.data[byteIndex] |= (1 << bitIndex)
	} else {
		b.data[byteIndex] &^= (1 << bitIndex)
	}
}

// GetBit gets the bit at the given position
func (b *ByteArray) GetBit(pos int) bool {
	byteIndex := pos / 8
	bitIndex := uint(pos % 8)

	return (b.data[byteIndex] & (1 << bitIndex)) != 0
}

// BitCount counts the number of bits set to 1
func (b *ByteArray) BitCount() int {
	count := 0
	for _, byteVal := range b.data {
		count += int(popcount(byteVal))
	}
	return count
}

// population counting, counts the number of set bits in a byte
// Using: https://en.wikipedia.org/wiki/Hamming_weight
func popcount(x byte) byte {
	// pairing bits and counting them in pairs
	x = x - ((x >> 1) & 0x55)
	// counting bits in groups of four
	x = (x & 0x33) + ((x >> 2) & 0x33)
	// isolates the lower four bits
	// which now contain the total count of set bits in the original byte
	return (x + (x >> 4)) & 0x0F
}

// // countBits counts the number of bits set to 1 in a byte.
// func countBits(b byte) int {
// 	count := 0
// 	for b != 0 {
// 		count += int(b & 1)
// 		b >>= 1
// 	}
// 	return count
// }

// // byteToBinary converts a byte to its binary string representation
// func byteToBinary(b byte) []byte {

// 	// Convert integer to binary string
// 	binaryStr := strconv.FormatInt(int64(b), 2)

// 	// Prepend zeros if the length is less than 8
// 	for len(binaryStr) < 8 {
// 		binaryStr = "0" + binaryStr
// 	}

// 	// Convert binary string to a slice of bytes
// 	binaryBytes := []byte(binaryStr)
// 	return binaryBytes
// }

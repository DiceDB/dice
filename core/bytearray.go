package core

type ByteArray struct {
	data   []byte
	Length int64
}

// NewByteArray initializes a new ByteArray with the given size
func NewByteArray(size int) *ByteArray {
	return &ByteArray{
		data:   make([]byte, size),
		Length: int64(size),
	}
}

// SetBit sets the bit at the given position to the specified value
func (b *ByteArray) SetBit(pos int, value bool) {
	byteIndex := pos / 8
	bitIndex := 7 - uint(pos%8)

	if value {
		b.data[byteIndex] |= (1 << bitIndex)
	} else {
		b.data[byteIndex] &^= (1 << bitIndex)
	}
}

// GetBit gets the bit at the given position
func (b *ByteArray) GetBit(pos int) bool {
	byteIndex := pos / 8
	bitIndex := 7 - uint(pos%8)

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

func (b *ByteArray) IncreaseSize(increaseSizeTo int) *ByteArray {
	currentByteArray := b.data
	currentByteArraySize := len(currentByteArray)

	// Input is decreasing the size
	if currentByteArraySize >= increaseSizeTo {
		return b
	}

	sizeDifference := increaseSizeTo - currentByteArraySize
	currentByteArray = append(currentByteArray, make([]byte, sizeDifference)...)

	b.data = currentByteArray
	b.Length = int64(increaseSizeTo)

	return b
}

func (b *ByteArray) ResizeIfNecessary() *ByteArray {

	byteArrayLength := b.Length
	decreaseLengthBy := 0
	for i := byteArrayLength - 1; i >= 0; i-- {
		if b.data[i] == 0x0 {
			decreaseLengthBy++
		} else {
			break
		}
	}

	if decreaseLengthBy == 0 {
		return b
	}

	// Decrease the size of the slice to n elements
	// and create a new slice with reduced capacity
	capacityReducedSlice := make([]byte, byteArrayLength-int64(decreaseLengthBy))
	copy(capacityReducedSlice, b.data[:byteArrayLength-int64(decreaseLengthBy)])

	b.data = capacityReducedSlice
	b.Length = int64(len(capacityReducedSlice))

	return b
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

// reverseByte reverses the order of bits in a single byte.
func reverseByte(b byte) byte {
	var reversed byte = 0
	for i := 0; i < 8; i++ {
		reversed = (reversed << 1) | (b & 1)
		b >>= 1
	}
	return reversed
}

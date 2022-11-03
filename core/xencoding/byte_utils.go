package xencoding

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
	return byte(x) & BITMASK[n-1]
}

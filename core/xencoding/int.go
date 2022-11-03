package xencoding

// TOOD: not thread safe
var buf [11]byte
var bitShifts []uint8 = []uint8{7, 7, 7, 7, 7, 7, 7, 7, 7, 1}

// XEncodeInt encodes the unsigned 64 bit integer value into a varint
// and returns an array of bytes (little endian encoded)
func XEncodeUInt(x uint64) []byte {
	var i int = 0
	for i = 0; i < len(bitShifts); i++ {
		buf[i] = getLSB(byte(x), bitShifts[i]) | 0b10000000 // marking the continuation bit
		x = x >> bitShifts[i]
		if x == 0 {
			break
		}
	}
	buf[len(buf)-1] = buf[len(buf)-1] & 0b01111111 // marking the termination bit
	return append(make([]byte, 0, i+1), buf[:i+1]...)
}

// XDecodeInt decodes the array of bytes and returns an unsigned 64 bit integer
func XDecodeUInt(vint []byte) uint64 {
	var i int = 0
	var v uint64 = 0
	for i = 0; i < len(vint); i++ {
		b := getLSB(vint[i], 7)
		v = v | uint64(b)<<(7*i)
	}
	return v
}

package object

type Obj struct {
	TypeEncoding uint8
	// Redis allocates 24 bits to these bits, but we will use 32 bits because
	// golang does not support bitfields, and we need not make this super-complicated
	// by merging TypeEncoding + LastAccessedAt in one 32-bit integer.
	// But nonetheless, we can benchmark and see how that fares.
	// For now, we continue with 32-bit integer to Store the LastAccessedAt
	LastAccessedAt uint32
	Value          interface{}
}

var ObjTypeString uint8 = 0 << 4

var ObjEncodingRaw uint8 = 0
var ObjEncodingInt uint8 = 1
var ObjEncodingEmbStr uint8 = 8

var ObjTypeByteList uint8 = 1 << 4
var ObjEncodingDeque uint8 = 4

var ObjTypeBitSet uint8 = 2 << 4 // 00100000
var ObjEncodingBF uint8 = 2      // 00000010

var ObjTypeJSON uint8 = 3 << 4 // 00110000
var ObjEncodingJSON uint8 = 0

var ObjTypeByteArray uint8 = 4 << 4 // 01000000
var ObjEncodingByteArray uint8 = 4

var ObjTypeInt uint8 = 5 << 4 // 01010000

var ObjTypeSet uint8 = 6 << 4 // 01010000
var ObjEncodingSetInt uint8 = 11
var ObjEncodingSetStr uint8 = 12

var ObjEncodingHashMap uint8 = 6
var ObjTypeHashMap uint8 = 7 << 4

var ObjTypeSortedSet uint8 = 8 << 4
var ObjEncodingBTree uint8 = 8

var ObjTypeCountMinSketch uint8 = 9 << 4
var ObjEncodingMatrix uint8 = 9

func ExtractTypeEncoding(obj *Obj) (e1, e2 uint8) {
	return obj.TypeEncoding & 0b11110000, obj.TypeEncoding & 0b00001111
}

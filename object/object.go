package object

import (
	"time"
)

type Obj struct {
	TypeEncoding uint8
	// Redis allots 24 bits to these bits, but we will use 32 bits because
	// golang does not support bitfields and we need not make this super-complicated
	// by merging TypeEncoding + LastAccessedAt in one 32 bit integer.
	// But nonetheless, we can benchmark and see how that fares.
	// For now, we continue with 32 bit integer to store the LastAccessedAt
	LastAccessedAt uint32
	Value          interface{}
}

func NewObj(value interface{}, expDurationMs int64, oType uint8, oEnc uint8) *Obj {
	obj := &Obj{
		Value:          value,
		TypeEncoding:   oType | oEnc,
		LastAccessedAt: uint32(time.Now().Unix()) & 0x00FFFFFF,
	}
	if expDurationMs > 0 {
		expires.SetExpiry(obj, expDurationMs)
	}
	return obj
}

type DiceWorkerBuffer struct {
	Key   string
	Value *Obj
}

var OBJ_TYPE_STRING uint8 = 0 << 4

var OBJ_ENCODING_RAW uint8 = 0
var OBJ_ENCODING_INT uint8 = 1
var OBJ_ENCODING_EMBSTR uint8 = 8

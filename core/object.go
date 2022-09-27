package core

// TODO: change ExpiresAt it to LRU Bits as handled by Redis
type Obj struct {
	TypeEncoding uint8
	Value        interface{}
	ExpiresAt    int64
}

var OBJ_TYPE_STRING uint8 = 0 << 4

var OBJ_ENCODING_RAW uint8 = 0
var OBJ_ENCODING_INT uint8 = 1
var OBJ_ENCODING_EMBSTR uint8 = 8

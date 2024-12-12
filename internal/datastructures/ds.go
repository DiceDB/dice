package datastructures

import (
	"time"
)

// Define the types of data structures
const (
	ObjTypeString int = iota
	_                 // skip 1 and 2 to maintain compatibility
	_
	ObjTypeJSON
	ObjTypeByteArray
	ObjTypeInt
	ObjTypeSet
	ObjTypeHashMap
	ObjTypeSortedSet
	ObjTypeCountMinSketch
	ObjTypeBF
	ObjTypeDequeue
)

// Define the internal
const (
	EncodingInt8 int = iota
	EncodingInt16
	EncodingInt32
	EncodingInt64

	EncodingFloat32
	EncodingFloat64

	EncodingString
)

// DSInterface defines the common behavior for all data structures
type DSInterface interface {
	GetLastAccessedAt() uint32
	UpdateLastAccessedAt()
	Serialize() []byte
	Size() int
	GetType() int
	GetEncoding() int
}

type BaseDataStructure[T DSInterface] struct {
	Encoding       int
	lastAccessedAt uint32
	Type           int
}

func (b *BaseDataStructure[T]) GetLastAccessedAt() uint32 {
	return b.lastAccessedAt
}

func (b *BaseDataStructure[T]) UpdateLastAccessedAt() {
	b.lastAccessedAt = uint32(time.Now().Unix()) & 0x00FFFFFF
}

func (b *BaseDataStructure[T]) GetEncoding() int {
	return b.Encoding
}

func (b *BaseDataStructure[T]) GetType() int {
	return b.Type
}

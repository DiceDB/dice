package datastructures

import (
	"time"
)

const (
	EncodingInt8 = iota
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
}

type BaseDataStructure[T DSInterface] struct {
	Encoding       int
	lastAccessedAt uint32
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

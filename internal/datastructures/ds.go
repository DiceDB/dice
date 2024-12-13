package datastructures

import (
	"time"
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
}

type BaseDataStructure[T DSInterface] struct {
	lastAccessedAt uint32
}

func (b *BaseDataStructure[T]) GetLastAccessedAt() uint32 {
	return b.lastAccessedAt
}

func (b *BaseDataStructure[T]) UpdateLastAccessedAt() {
	b.lastAccessedAt = uint32(time.Now().Unix()) & 0x00FFFFFF
}

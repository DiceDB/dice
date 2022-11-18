package core

import (
	"unsafe"
)

const QueueRefMaxBuf int = 256

type QueueRef struct {
	qi *QueueInt
}

func NewQueueRef() *QueueRef {
	return &QueueRef{
		qi: NewQueueInt(),
	}
}

func (q *QueueRef) Size() int64 {
	return q.qi.Size()
}

// Insert inserts reference of the key in the QueueRef q.
// returns false if key does not exist
func (q *QueueRef) Insert(key string) bool {
	x, ok := keypool[key]
	if !ok {
		return false
	}
	q.qi.Insert(int64(uintptr(x)))
	return true
}

// Remove removes the reference from the queue q.
// returns nil if key does not exist in the store any more
func (q *QueueRef) Remove() (*Obj, error) {
	val, err := q.qi.Remove()
	if err != nil {
		return nil, err
	}
	ptr := unsafe.Pointer(uintptr(val))
	obj, ok := store[ptr]
	if !ok {
		return nil, nil
	}
	return obj, nil
}

func (q *QueueRef) Iterate(n int) []*Obj {
	return nil
}

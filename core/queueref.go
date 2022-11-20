package core

import (
	"unsafe"
)

type QueueRef struct {
	qi *QueueInt
}

type QueueElement struct {
	Key string
	Obj *Obj
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
// if the expired key is popped from the queue, we continue to pop until
// until we find one non-expired key
// TODO: test for expired keys
func (q *QueueRef) Remove() (*QueueElement, error) {
	for {
		val, err := q.qi.Remove()
		if err != nil {
			return nil, err
		}
		key := *((*string)(unsafe.Pointer(uintptr(val))))
		obj := Get(key)
		if obj != nil {
			return &QueueElement{key, obj}, nil
		}
	}
}

// Iterate iterates through the QueueRef
// it also filters out the keys that are expired
func (q *QueueRef) Iterate(n int) []*QueueElement {
	vals := q.qi.Iterate(n)
	elements := make([]*QueueElement, 0, len(vals))
	for _, val := range vals {
		key := *((*string)(unsafe.Pointer(uintptr(val))))
		obj := Get(key)
		if obj != nil {
			elements = append(elements, &QueueElement{key, obj})
		}
	}
	return elements
}

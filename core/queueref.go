package core

import (
	"sync"
	"unsafe"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/core/diceerrors"
)

var (
	QueueCount = 0
	muQueue    sync.Mutex
)

type QueueRef struct {
	qi *QueueInt
}

type QueueElement struct {
	Key string
	Obj *Obj
}

func NewQueueRef() (*QueueRef, error) {
	if QueueCount >= config.MaxQueues {
		return nil, diceerrors.NewErr("ERR maximum number of queues reached")
	}
	muQueue.Lock()
	QueueCount++
	muQueue.Unlock()
	return &QueueRef{
		qi: NewQueueInt(),
	}, nil
}

func (q *QueueRef) Size(store *Store) int64 {
	q.QueueRefCleanup(store)
	return q.qi.Size()
}

// Insert inserts reference of the key in the QueueRef q.
// returns false if key does not exist
func (q *QueueRef) Insert(key string, store *Store) bool {
	var x *string
	var ok bool
	if q.qi.Length >= int64(config.MaxQueueSize) {
		return false // Prevent inserting if the queue is at maximum capacity
	}
	withLocks(func() {
		x, ok = store.keypool[key]
	}, store, WithKeypoolRLock())

	if !ok {
		return false
	}
	value := int64(uintptr(unsafe.Pointer(x)))
	q.qi.Insert(value)
	return true
}

// Remove removes the reference from the queue q.
// returns error if queue is empty
// if the expired key is popped from the queue, we continue to pop until
// until we find one non-expired key
// TODO: test for expired keys
func (q *QueueRef) Remove(store *Store) (*QueueElement, error) {
	for {
		val, err := q.qi.Remove()
		if err != nil {
			return nil, err
		}
		key := *(*string)(unsafe.Pointer(uintptr(val)))
		obj := store.Get(key)
		if obj != nil {
			return &QueueElement{key, obj}, nil
		}
	}
}

// Iterate iterates through the QueueRef
// it also filters out the keys that are expired
func (q *QueueRef) Iterate(n int, store *Store) []*QueueElement {
	vals := q.qi.Iterate(n)
	elements := make([]*QueueElement, 0, len(vals))
	for _, val := range vals {
		key := *((*string)(unsafe.Pointer(uintptr(val))))
		obj := store.Get(key)
		if obj != nil {
			elements = append(elements, &QueueElement{key, obj})
		}
	}
	return elements
}

func (q *QueueRef) DeepCopy() *QueueRef {
	return &QueueRef{
		qi: q.qi.DeepCopy(),
	}
}

// Returns the length of the queue
func (q *QueueRef) Length(store *Store) int64 {
	q.QueueRefCleanup(store)
	return q.qi.Length
}

// QueueRefCleanup removes expired or deleted keys from the QueueRef to maintain the accuracy of QUEUEREFLEN operations.
// While this process ensures correctness, it can be computationally intensive, especially as
// the queue grows. To mitigate performance overhead, constraints such as limiting the queue size
// to 10,000 entries and the number of queues to 100 are imposed, ensuring that the cleanup
// remains efficient without degrading system performance
func (q *QueueRef) QueueRefCleanup(store *Store) {
	var validElements []int64
	for {
		val, err := q.qi.Remove()
		if err != nil {
			break // Break the loop if the queue is empty
		}

		key := *((*string)(unsafe.Pointer(uintptr(val))))
		obj := store.Get(key)
		if obj != nil {
			validElements = append(validElements, val)
		}
	}

	for _, val := range validElements {
		q.qi.Insert(val)
	}
}

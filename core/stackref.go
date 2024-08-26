package core

import (
	"errors"
	"sync"
	"unsafe"

	"github.com/dicedb/dice/config"
)

var (
	StackCount = 0
	muStack    sync.Mutex
)

type StackRef struct {
	si *StackInt
}

type StackElement struct {
	Key string
	Obj *Obj
}

func NewStackRef() (*StackRef, error) {
	muStack.Lock()
	defer muStack.Unlock()
	if StackCount >= config.MaxStacks {
		return nil, errors.New("ERR maximum number of stacks reached")
	}

	StackCount++

	return &StackRef{
		si: NewStackInt(),
	}, nil
}

func (s *StackRef) Size(store *Store) int64 {
	s.StackRefCleanup(store)
	return s.si.Size()
}

// Length returns the actual length of the stack
func (s *StackRef) Length(store *Store) int64 {
	s.StackRefCleanup(store)
	return s.si.Length
}

// Push pushes reference of the key in the StackRef s.
// returns false if key does not exist
func (s *StackRef) Push(key string, store *Store) bool {
	var x *string
	var ok bool

	if s.si.Length >= int64(config.MaxStackSize) {
		return false // Prevent pushing if the stack is at maximum capacity
	}

	withLocks(func() {
		x, ok = store.keypool[key]
	}, store, WithKeypoolRLock())

	if !ok {
		return false
	}

	value := int64(uintptr(unsafe.Pointer(x)))
	s.si.Push(value)
	return true
}

// Pop pops the reference from the stack s.
// returns error if stack is empty
// if the expired key is popped from the stack, we continue to pop until
// until we find one non-expired key
func (s *StackRef) Pop(store *Store) (*StackElement, error) {
	for {
		val, err := s.si.Pop()
		if err != nil {
			return nil, err
		}

		key := *(*string)(unsafe.Pointer(uintptr(val)))
		obj := store.Get(key)
		if obj != nil {
			return &StackElement{key, obj}, nil
		}
	}
}

// Iterate iterates through the StackRef
// it also filters out the keys that are expired
func (s *StackRef) Iterate(n int, store *Store) []*StackElement {
	vals := s.si.Iterate(n)
	elements := make([]*StackElement, 0, len(vals))
	for _, val := range vals {
		key := *(*string)(unsafe.Pointer(uintptr(val)))
		obj := store.Get(key)
		if obj != nil {
			elements = append(elements, &StackElement{key, obj})
		}
	}
	return elements
}

func (s *StackRef) DeepCopy() *StackRef {
	return &StackRef{
		si: s.si.DeepCopy(),
	}
}

// StackRefCleanup removes expired or deleted keys from the StackRef to maintain the accuracy of STACKREFLEN operations.
// While this process ensures correctness, it can be computationally intensive, especially as
// the stack grows. To mitigate performance overhead, constraints such as limiting the stack size
// to 10,000 entries and the number of stacks to 100 is imposed, ensuring that the cleanup
// remains efficient without degrading system performance.
func (s *StackRef) StackRefCleanup(store *Store) {
	var validElements []int64

	for {
		val, err := s.si.Pop()
		if err != nil {
			break // Break the loop if the stack is empty
		}

		key := *((*string)(unsafe.Pointer(uintptr(val))))
		obj := store.Get(key)
		if obj != nil {
			validElements = append(validElements, val)
		}
	}

	for i := len(validElements) - 1; i >= 0; i-- {
		s.si.Push(validElements[i])
	}
}

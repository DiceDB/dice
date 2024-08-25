package core

import (
	"unsafe"
)

type StackRef struct {
	si *StackInt
}

type StackElement struct {
	Key string
	Obj *Obj
}

func NewStackRef() *StackRef {
	return &StackRef{
		si: NewStackInt(),
	}
}

func (s *StackRef) Size() int64 {
	return s.si.Size()
}

// Push pushes reference of the key in the StackRef s.
// returns false if key does not exist
func (s *StackRef) Push(key string, store *Store) bool {
	var x *string
	var ok bool

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
// returns nil if key does not exist in the store any more
// if the expired key is popped from the stack, we continue to pop until
// until we find one non-expired key
// TODO: test for expired keys
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

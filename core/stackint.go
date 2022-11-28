package core

import (
	"errors"
	"unsafe"

	"github.com/dicedb/dice/core/dencoding"
)

const StackIntMaxBuf int = 256

var ErrStackEmpty = errors.New("stack is empty")

// Represents a stack of integers
type StackInt struct {
	Length int64
	list   *byteList
}

// Returns a pointer to a newly allocated StackInt.
func NewStackInt() *StackInt {
	s := &StackInt{
		Length: 0,
		list:   newByteList(StackIntMaxBuf),
	}
	return s
}

func (s *StackInt) Size() int64 {
	return int64(unsafe.Sizeof(*s)) + s.list.size
}

// Push pushes the integer `x` in the the StackInt s.
func (s *StackInt) Push(x int64) {
	var xb []byte
	if x >= 0 {
		xb = dencoding.EncodeUInt(uint64(x))
	} else {
		// TODO: support negative integers
		panic("negative integers not supported yet")
	}

	var bn *byteListNode
	if s.list.head == nil {
		bn = s.list.newNode()
		s.list.append(bn)
	} else {
		bn = s.list.tail
	}

	// add new node in bytelist if space is insufficient
	if cap(bn.buf)-len(bn.buf) < len(xb) {
		bn = s.list.newNode()
		s.list.append(bn)
	}

	bn.buf = append(bn.buf, xb...)
	s.Length++
}

// Pop Pops an integer from the Stack s.
func (s *StackInt) Pop() (int64, error) {
	var val int64
	bn := s.list.tail

	if bn == nil || len(bn.buf) == 0 {
		return 0, ErrStackEmpty
	}

	// tailIdx represents the index of the terminating byte of the element
	// at the top of the stack.
	tailIdx := int64(len(bn.buf)) - 1

	var startIdx int64
	for startIdx = tailIdx - 1; startIdx >= 0; startIdx-- {
		// Current byte.
		b := bn.buf[startIdx]

		// Check if b is the terminating byte.
		if b&0b10000000 == 0 {
			// Place startIdx at the start of the last element.
			startIdx++
			break
		}
	}

	// For the first element in the byteListNode, startIdx will be, so we need to
	// manually set it to 0.
	if startIdx == -1 {
		startIdx = 0
	}

	val = int64(dencoding.DecodeUInt(bn.buf[startIdx : tailIdx+1]))

	bn.buf = bn.buf[:startIdx]

	if len(bn.buf) == 0 {
		s.list.delete(bn)
	}

	s.Length--
	return val, nil
}

// Iterate inserts the integer `x` in the the StackInt q
// through at max `n` elements.
// the function returns empty list for invalid `n`
func (s *StackInt) Iterate(n int) []int64 {
	if n <= 0 {
		return []int64{}
	}

	var vals []int64

	p := s.list.tail
	for p != nil {
		tailIdx := int64(len(p.buf)) - 1
		for startIdx := tailIdx - 1; startIdx >= 0; startIdx-- {
			b := p.buf[startIdx]

			// if b is the terminating byte
			if b&0b10000000 == 0 {
				// Place startIdx at the start of the element being processed.
				startIdx++
				vals = append(vals, int64(dencoding.DecodeUInt(p.buf[startIdx:tailIdx+1])))

				n--
				if n == 0 {
					return vals
				}

				// Place tailIdx at the end of the next element in the stack.
				tailIdx = startIdx - 1
				// Place startIdx at the end of the next element in the stack, it will
				// automatically be decremented by 1 byte in the next iteration.
				startIdx = tailIdx
			}
		}

		// The first element in the byteListNode won't be processed in the loop
		// above.
		vals = append(vals, int64(dencoding.DecodeUInt(p.buf[:tailIdx+1])))
		p = p.prev
	}

	return vals
}

package core

import (
	"errors"
	"unsafe"

	"github.com/dicedb/dice/core/dencoding"
)

const QueueIntMaxBuf int = 256

var ErrQueueEmpty = errors.New("queue is empty")

type QueueInt struct {
	Length int64
	list   *byteList
}

func NewQueueInt() *QueueInt {
	q := &QueueInt{
		Length: 0,
		list:   newByteList(QueueIntMaxBuf),
	}
	return q
}

func (q *QueueInt) Size() int64 {
	return int64(unsafe.Sizeof(*q)) + q.list.size
}

// Insert inserts the integer `x` in the the QueueInt q.
func (q *QueueInt) Insert(x int64) {
	var xb []byte
	if x >= 0 {
		xb = dencoding.EncodeUInt(uint64(x))
	} else {
		// TODO: support negative integers
		panic("negative integers not supported yet")
	}

	var bn *byteListNode
	if q.list.head == nil {
		bn = q.list.newNode()
		q.list.append(bn)
	} else {
		bn = q.list.tail
	}

	// add new node in bytelist is space is insufficient
	if cap(bn.buf)-len(bn.buf) < len(xb) {
		bn = q.list.newNode()
		q.list.append(bn)
	}

	bn.buf = append(bn.buf, xb...)
	q.Length++
}

// Remove removes the integer from the queue q.
func (q *QueueInt) Remove() (int64, error) {
	var val int64
	bn := q.list.head
	if bn == nil || len(bn.buf) == 0 {
		return 0, ErrQueueEmpty
	}
	var i int
	var tbuf []byte = make([]byte, 11)
	var tbufIdx int = 0
	for i = 0; i < len(bn.buf); i++ {
		b := bn.buf[i]
		tbuf[tbufIdx] = b
		tbufIdx++

		// if b is the terminating byte
		if b&0b10000000 == 0 {
			val = int64(dencoding.DecodeUInt(tbuf[:tbufIdx]))
			tbufIdx = 0
			break
		}
	}
	bn.buf = bn.buf[i+1:]
	if len(bn.buf) == 0 {
		q.list.delete(bn)
	}

	q.Length--
	return val, nil
}

// Iterate inserts the integer `x` in the the QueueInt q
// through at max `n` elements.
// the function returns empty list for invalid `n`
func (q *QueueInt) Iterate(n int) []int64 {
	if n <= 0 {
		return []int64{}
	}

	var vals []int64

	p := q.list.head
	for p != nil {
		var tbuf []byte = make([]byte, 11)
		var tbufIdx int = 0
		for i := 0; i < len(p.buf); i++ {
			b := p.buf[i]
			tbuf[tbufIdx] = b
			tbufIdx++
			// if b is the terminating byte
			if b&0b10000000 == 0 {
				// TODO: set the index instead of append
				// needs benchmarking
				vals = append(vals, int64(dencoding.DecodeUInt(tbuf[:tbufIdx])))
				tbufIdx = 0

				n--
				if n == 0 {
					return vals
				}
			}
		}
		p = p.next
	}

	return vals
}

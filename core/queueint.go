package core

import (
	"errors"

	"github.com/dicedb/dice/core/xencoding"
)

const QueueIntMaxBuf int = 32

type QueueInt struct {
	Length int64
	List   *byteList
}

func NewQueueInt() *QueueInt {
	return &QueueInt{
		Length: 0,
		List:   NewByteList(QueueIntMaxBuf),
	}
}

// QueueIntInsert inserts the integer `x` in the the QueueInt q.
func (q *QueueInt) QueueIntInsert(x int64) {
	var xb []byte
	if x >= 0 {
		xb = xencoding.XEncodeUInt(uint64(x))
	} else {
		// TODO: support negative integers
		panic("negative integers not supported yet")
	}

	var bn *byteListNode
	if q.List.head == nil {
		bn = q.List.NewNode()
		q.List.Append(bn)
	} else {
		bn = q.List.tail
	}

	// add new node in bytelist is space is insufficient
	if cap(bn.buf)-len(bn.buf) < len(xb) {
		bn = q.List.NewNode()
		q.List.Append(bn)
	}

	bn.buf = append(bn.buf, xb...)
	q.Length++
}

// QueueIntRemove removes the integer from the queue q.
func (q *QueueInt) QueueIntRemove() (int64, error) {
	var val int64
	bn := q.List.head
	if bn == nil || len(bn.buf) == 0 {
		return 0, errors.New("queueint is empty")
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
			val = int64(xencoding.XDecodeUInt(tbuf[:tbufIdx]))
			tbufIdx = 0
			break
		}
	}
	bn.buf = bn.buf[i+1:]
	if len(bn.buf) == 0 {
		q.List.Delete(bn)
	}

	q.Length--
	return val, nil
}

// QueueIntInsert inserts the integer `x` in the the QueueInt q.
func (q *QueueInt) QueueIntIterate() []int64 {
	var vals []int64

	p := q.List.head
	for p != nil {
		var tbuf []byte = make([]byte, 11)
		var tbufIdx int = 0
		for i := 0; i < len(p.buf); i++ {
			b := p.buf[i]
			tbuf[tbufIdx] = b
			tbufIdx++
			// if b is the terminating byte
			if b&0b10000000 == 0 {
				vals = append(vals, int64(xencoding.XDecodeUInt(tbuf[:tbufIdx])))
				tbufIdx = 0
			}
		}
		p = p.next
	}

	return vals
}

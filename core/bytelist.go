package core

import (
	"unsafe"
)

var byteListNodeSize int64 = 0

func init() {
	var n byteListNode
	byteListNodeSize = int64(unsafe.Sizeof(n))
}

// ByteList is the linkedlist implementation
// where each node holds the byte array
type byteList struct {
	bufLen int
	size   int64
	head   *byteListNode
	tail   *byteListNode
}

type byteListNode struct {
	buf  []byte
	next *byteListNode
	prev *byteListNode
}

func newByteList(bufLen int) *byteList {
	return &byteList{
		head:   nil,
		tail:   nil,
		bufLen: bufLen,
	}
}

func (b *byteList) newNode() *byteListNode {
	bn := &byteListNode{
		buf: make([]byte, 0, b.bufLen),
	}
	b.size += byteListNodeSize
	return bn
}

func (b *byteList) append(bn *byteListNode) {
	bn.prev = b.tail
	if b.tail != nil {
		b.tail.next = bn
	}
	b.tail = bn
	if b.head == nil {
		b.head = bn
	}
}

func (b *byteList) prepend(bn *byteListNode) {
	bn.next = b.head
	if b.head != nil {
		b.head.prev = bn
	}
	b.head = bn
	if b.tail == nil {
		b.tail = bn
	}
}

func (b *byteList) delete(bn *byteListNode) {
	if bn == b.head {
		b.head = bn.next
	}

	if bn == b.tail {
		b.tail = bn.prev
	}

	if bn.prev != nil {
		bn.prev.next = bn.next
	}

	if bn.next != nil {
		bn.next.prev = bn.prev
	}

	b.size -= byteListNodeSize
}

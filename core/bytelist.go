package core

// ByteList is the linkedlist implementation
// where each node holds the byte array
type byteList struct {
	bufLen int
	head   *byteListNode
	tail   *byteListNode
}

type byteListNode struct {
	buf  []byte
	next *byteListNode
	prev *byteListNode
}

func NewByteList(bufLen int) *byteList {
	return &byteList{
		head:   nil,
		tail:   nil,
		bufLen: bufLen,
	}
}

func (b *byteList) NewNode() *byteListNode {
	return &byteListNode{
		buf: make([]byte, 0, b.bufLen),
	}
}

func (b *byteList) Append(bn *byteListNode) {
	bn.prev = b.tail
	if b.tail != nil {
		b.tail.next = bn
	}
	b.tail = bn
	if b.head == nil {
		b.head = bn
	}
}

func (b *byteList) Prepend(bn *byteListNode) {
	bn.next = b.head
	if b.head != nil {
		b.head.prev = bn
	}
	b.head = bn
	if b.tail == nil {
		b.tail = bn
	}
}

func (b *byteList) Delete(bn *byteListNode) {
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
}

// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package eval

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

func (b *byteList) newNodeWithCapacity(capacity int) *byteListNode {
	bn := &byteListNode{
		buf: make([]byte, 0, capacity),
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

// DeepCopy creates a deep copy of the byteList.
func (b *byteList) DeepCopy() *byteList {
	if b == nil {
		return nil
	}

	// Create a new byteList instance using newByteList
	copyList := newByteList(b.bufLen)
	copyList.size = b.size

	// Copy the nodes recursively starting from the head
	if b.head != nil {
		copyList.head = b.head.deepCopyNode(nil)
	}

	// Set the tail to the last node in the copied list
	currentNode := copyList.head
	for currentNode != nil {
		if currentNode.next == nil {
			copyList.tail = currentNode
		}
		currentNode = currentNode.next
	}

	return copyList
}

func (node *byteListNode) deepCopyNode(prevCopy *byteListNode) *byteListNode {
	if node == nil {
		return nil
	}

	// Create a copy of the current node
	copyNode := &byteListNode{
		buf:  append([]byte(nil), node.buf...),
		prev: prevCopy,
	}

	// Recursively copy the next node
	if node.next != nil {
		copyNode.next = node.next.deepCopyNode(copyNode)
	}

	return copyNode
}

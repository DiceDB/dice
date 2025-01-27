// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package eval

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func newNode(bl *byteList, b byte) *byteListNode {
	bn := bl.newNode()
	bn.buf = append(bn.buf, b)
	return bn
}

func toByteArray(bl *byteList) []byte {
	var res []byte
	p := bl.head
	for p != nil {
		res = append(res, p.buf[0])
		p = p.next
	}
	return res
}

func getNode(bl *byteList, b byte) *byteListNode {
	p := bl.head
	for p != nil {
		if p.buf[0] == b {
			return p
		}
		p = p.next
	}
	return nil
}

type tcase struct {
	op  string
	val []byte
}

func TestByteList(t *testing.T) {
	bl := newByteList(1)
	for _, tc := range []tcase{
		{"a1", []byte{byte('1')}},
		{"a2", []byte{byte('1'), byte('2')}},
		{"p0", []byte{byte('0'), byte('1'), byte('2')}},
		{"a3", []byte{byte('0'), byte('1'), byte('2'), byte('3')}},
		{"p4", []byte{byte('4'), byte('0'), byte('1'), byte('2'), byte('3')}},
		{"d0", []byte{byte('4'), byte('1'), byte('2'), byte('3')}},
		{"d4", []byte{byte('1'), byte('2'), byte('3')}},
		{"d3", []byte{byte('1'), byte('2')}},
		{"d1", []byte{byte('2')}},
		{"d2", []byte{}},
	} {
		switch tc.op[0] {
		case 'a':
			bl.append(newNode(bl, tc.op[1]))
		case 'p':
			bl.prepend(newNode(bl, tc.op[1]))
		case 'd':
			bl.delete(getNode(bl, tc.op[1]))
		}

		r := toByteArray(bl)
		if !bytes.Equal(r, tc.val) {
			t.Errorf("bytelist test failed. should have been %v but found %v", tc.val, r)
		}
	}
}

func TestByteListDeepCopy(t *testing.T) {
	// Create an original byteList using newByteList
	original := newByteList(4)
	original.size = 8

	node1 := original.newNode()
	node1.buf = append(node1.buf, []byte{1, 2, 3, 4}...)

	node2 := original.newNode()
	node2.buf = append(node2.buf, []byte{5, 6, 7, 8}...)

	node1.next = node2
	node2.prev = node1
	original.head = node1
	original.tail = node2

	deepCopy := original.DeepCopy()

	// Verify that the deepCopy has the same bufLen and size
	assert.Equal(t, deepCopy.bufLen, original.bufLen, "bufLen should be the same")
	assert.Equal(t, deepCopy.size, original.size, "size should be the same")

	// Verify that the head node data is correctly copied
	assert.Equal(t, deepCopy.head.buf[0], original.head.buf[0], "head node buffer should be the same")

	// Verify that changes to the deepCopy do not affect the original
	deepCopy.head.buf[0] = 9
	assert.True(t, original.head.buf[0] != deepCopy.head.buf[0], "Original and deepCopy head buffer should not be linked")

	// Verify that changes to the original do not affect the deepCopy
	original.head.buf[1] = 8
	assert.True(t, original.head.buf[1] != deepCopy.head.buf[1], "Original and deepCopy head buffer should not be linked")
}

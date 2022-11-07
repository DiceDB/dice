package core

import (
	"bytes"
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

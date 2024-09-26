package eval

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/dicedb/dice/internal/dencoding"
)

var ErrDequeEmpty = errors.New("deque is empty")

type DequeI interface {
	LPush(string)
	RPush(string)
	LPop() (string, error)
	RPop() (string, error)
	LInsert(string, string, string) (int64, error)
}

var _ DequeI = (*DequeBasic)(nil)

type DequeBasic struct {
	Length int64
	buf    []byte
}

func NewBasicDeque() *DequeBasic {
	l := &DequeBasic{
		Length: 0,
	}
	return l
}

// LPush pushes `x` into the left side of the Deque.
func (q *DequeBasic) LPush(x string) {
	// enc + data + backlen
	xb := EncodeDeqEntry(x)

	if cap(q.buf)-len(q.buf) < len(xb) {
		newArr := make([]byte, len(q.buf)+len(xb), (len(q.buf)+len(xb))*2)
		copy(newArr[len(xb):], q.buf)
		copy(newArr, xb)
		q.buf = newArr
	} else {
		q.buf = q.buf[:len(xb)+len(q.buf)]
		copy(q.buf[len(xb):], q.buf)
		copy(q.buf, xb)
	}

	q.Length++
}

// RPush pushes `x` into the right side of the Deque.
func (q *DequeBasic) RPush(x string) {
	// enc + data + backlen
	xb := EncodeDeqEntry(x)
	q.buf = append(q.buf, xb...)
	q.Length++
}

// RPop pops an element from the right side of the Deque.
func (q *DequeBasic) RPop() (string, error) {
	if q.Length == 0 {
		return "", ErrDequeEmpty
	}

	backlenStartIdx := len(q.buf) - 1
	for q.buf[backlenStartIdx]&0x80 != 0 {
		backlenStartIdx--
	}
	backlen := dencoding.DecodeUIntRev(q.buf[backlenStartIdx:])

	entryStartIdx := backlenStartIdx - int(backlen)
	x, _ := DecodeDeqEntry(q.buf[entryStartIdx:backlenStartIdx])

	q.buf = q.buf[:entryStartIdx]
	q.Length--

	return x, nil
}

// LPop pops an element from the left side of the Deque.
func (q *DequeBasic) LPop() (string, error) {
	if q.Length == 0 {
		return "", ErrDequeEmpty
	}

	x, entryLen := DecodeDeqEntry(q.buf)

	copy(q.buf, q.buf[entryLen:])
	q.buf = q.buf[:len(q.buf)-entryLen]
	q.Length--

	return x, nil
}

func (q *DequeBasic) insertElementAfterIndex(element string, idx int) {
	// enc + data + backlen
	xb := EncodeDeqEntry(element)
	xbLen := len(xb)

	if cap(q.buf)-len(q.buf) < xbLen {
		newArr := make([]byte, len(q.buf)+xbLen, (len(q.buf)+xbLen)*2)
		copy(newArr[xbLen+idx:], q.buf[idx:])
		copy(newArr[:idx], q.buf[:idx])
		copy(newArr[idx:idx+xbLen], xb)
		q.buf = newArr
	} else {
		q.buf = q.buf[:xbLen+len(q.buf)]
		copy(q.buf[xbLen+idx:], q.buf[idx:])
		copy(q.buf[:idx], q.buf[:idx])
		copy(q.buf[idx:idx+xbLen], xb)
	}
	q.Length++
}

func (q *DequeBasic) insertBeforeAfterPivot(element, beforeAfter string, pivotIndexStart int, qIterator *DequeBasicIterator) {
	if pivotIndexStart == 0 && beforeAfter == Before {
		q.LPush(element)
		return
	}
	if !qIterator.HasNext() && beforeAfter == After {
		q.RPush(element)
		return
	}
	idx := pivotIndexStart
	if beforeAfter == After {
		idx = qIterator.bufIndex
	}
	q.insertElementAfterIndex(element, idx)
}

func (q *DequeBasic) LInsert(pivot, element, beforeAfter string) (int64, error) {
	// Check if the deque is empty.
	if q.Length == 0 {
		return -1, nil
	}
	if beforeAfter != Before && beforeAfter != After {
		return -1, errors.New("invalid before/after argument")
	}

	qIterator := q.NewIterator()
	for qIterator.HasNext() {
		pivotIndexStart := qIterator.bufIndex
		if x, _ := qIterator.Next(); x == pivot {
			q.insertBeforeAfterPivot(element, beforeAfter, pivotIndexStart, qIterator)
			return q.Length, nil
		}
	}
	return -1, nil
}

type DequeBasicIterator struct {
	deque             *DequeBasic
	elementsTraversed int64
	bufIndex          int
}

func (q *DequeBasic) NewIterator() *DequeBasicIterator {
	return &DequeBasicIterator{
		deque:             q,
		elementsTraversed: 0,
		bufIndex:          0,
	}
}

func (i *DequeBasicIterator) HasNext() bool {
	return i.elementsTraversed < i.deque.Length
}

func (i *DequeBasicIterator) Next() (string, error) {
	if !i.HasNext() {
		return "", fmt.Errorf("iterator exhausted")
	}
	x, entryLen := DecodeDeqEntry(i.deque.buf[i.bufIndex:])
	i.bufIndex += entryLen
	i.elementsTraversed++
	return x, nil
}

const (
	minDequeNodeSize = 256
	Before           = "before"
	After            = "after"
)

var _ DequeI = (*Deque)(nil)

type Deque struct {
	Length  int64
	list    *byteList
	leftIdx int
}

func NewDeque() *Deque {
	return &Deque{
		Length:  0,
		list:    newByteList(minDequeNodeSize),
		leftIdx: 0,
	}
}

func (q *Deque) LPush(x string) {
	// enc + data + backlen
	entrySize := int(GetEncodeDeqEntrySize(x))
	head := q.list.head

	if q.leftIdx >= entrySize {
		q.leftIdx -= entrySize
		EncodeDeqEntryInPlace(x, head.buf[q.leftIdx:q.leftIdx+entrySize])
	} else if q.leftIdx > 0 {
		newBuf := make([]byte, entrySize, entrySize+minDequeNodeSize-q.leftIdx)
		EncodeDeqEntryInPlace(x, newBuf[0:entrySize])
		newBuf = append(newBuf, head.buf[q.leftIdx:]...)
		head.buf = newBuf
		q.leftIdx = 0
	} else {
		if entrySize > minDequeNodeSize {
			head = q.list.newNodeWithCapacity(entrySize)
		} else {
			head = q.list.newNode()
		}
		q.list.prepend(head)
		head.buf = head.buf[:cap(head.buf)]
		q.leftIdx = len(head.buf) - entrySize
		EncodeDeqEntryInPlace(x, head.buf[q.leftIdx:])
	}

	q.Length++
}

func (q *Deque) RPush(x string) {
	// enc + data + backlen
	entrySize := int(GetEncodeDeqEntrySize(x))
	tail := q.list.tail
	if tail == nil || len(tail.buf) == cap(tail.buf) {
		if entrySize > minDequeNodeSize {
			tail = q.list.newNodeWithCapacity(entrySize)
		} else {
			tail = q.list.newNode()
		}
		q.list.append(tail)
		tail.buf = tail.buf[:entrySize]
		EncodeDeqEntryInPlace(x, tail.buf[:entrySize])
	} else if cap(tail.buf)-len(tail.buf) < entrySize {
		newBuf := make([]byte, len(tail.buf)+entrySize)
		copy(newBuf, tail.buf)
		EncodeDeqEntryInPlace(x, newBuf[len(tail.buf):])
		tail.buf = newBuf
	} else {
		oriLen := len(tail.buf)
		tail.buf = tail.buf[:oriLen+entrySize]
		EncodeDeqEntryInPlace(x, tail.buf[oriLen:])
	}

	q.Length++
}

func (q *Deque) LPop() (string, error) {
	if q.Length == 0 {
		return "", ErrDequeEmpty
	}

	head := q.list.head
	x, entryLen := DecodeDeqEntry(head.buf[q.leftIdx:])

	q.leftIdx += entryLen
	if q.leftIdx == len(head.buf) {
		q.list.delete(head)
		q.leftIdx = 0
	}
	q.Length--

	return x, nil
}

func (q *Deque) RPop() (string, error) {
	if q.Length == 0 {
		return "", ErrDequeEmpty
	}

	tail := q.list.tail
	backlenStartIdx := len(tail.buf) - 1
	for tail.buf[backlenStartIdx]&0x80 != 0 {
		backlenStartIdx--
	}
	backlen := dencoding.DecodeUIntRev(tail.buf[backlenStartIdx:])

	entryStartIdx := backlenStartIdx - int(backlen)
	x, _ := DecodeDeqEntry(tail.buf[entryStartIdx:backlenStartIdx])

	tail.buf = tail.buf[:entryStartIdx]
	if len(tail.buf) == 0 {
		q.list.delete(tail)
		if q.list.tail == nil {
			q.leftIdx = 0
		}
	}
	q.Length--

	return x, nil
}

func (q *Deque) breakPivotNodeAndInsertAfter(qIterator *DequeIterator, pivotNode *byteListNode, element string) *byteListNode {
	newNode := q.list.newNode()
	if pivotNode.next != nil {
		pivotNode.next.prev = newNode
	}
	newNode.next = pivotNode.next
	pivotNode.next = nil
	newNode.buf = append([]byte{}, pivotNode.buf[qIterator.BufIndex:]...)
	pivotNode.buf = pivotNode.buf[:qIterator.BufIndex]
	q.list.tail = pivotNode
	q.RPush(element)
	newNode.prev = q.list.tail
	q.list.tail.next = newNode
	return newNode
}

func (q *Deque) insertAfterPivotNodeHelper(element string, qIterator *DequeIterator, pivotNode *byteListNode) {
	prevTail := q.list.tail
	if qIterator.BufIndex == 0 {
		pivotNode.next = nil
		q.list.tail = pivotNode
		q.RPush(element)
		q.list.tail.next = qIterator.CurrentNode
		qIterator.CurrentNode.prev = q.list.tail
		q.list.tail = prevTail
	} else {
		newNode := q.breakPivotNodeAndInsertAfter(qIterator, pivotNode, element)
		if newNode.next == nil {
			q.list.tail = newNode
		} else {
			q.list.tail = prevTail
		}
	}
}

func (q *Deque) insertAfterPivotNode(element string, qIterator *DequeIterator, pivotNode *byteListNode) {
	if qIterator.ElementsTraversed == q.Length {
		// Element needs to be inserted at the end of the Deque.
		q.RPush(element)
	} else {
		// Element needs to be inserted b/w 2 nodes in the Deque.
		q.insertAfterPivotNodeHelper(element, qIterator, pivotNode)
	}
}

func (q *Deque) getNewNodeWithElement(element string) *byteListNode {
	elementEntryLen := int(GetEncodeDeqEntrySize(element))
	elementNode := q.list.newNodeWithCapacity(elementEntryLen)
	elementNode.buf = elementNode.buf[:elementEntryLen]
	EncodeDeqEntryInPlace(element, elementNode.buf[:elementEntryLen])
	return elementNode
}

func (q *Deque) breakPivotNodeAndInsertBefore(qIterator *DequeIterator, pivotNode, elementNode *byteListNode, pivotEntryLen, leftIdx int) *byteListNode {
	bufIndex := qIterator.BufIndex
	if bufIndex == 0 {
		bufIndex = len(pivotNode.buf)
	}
	newNode := q.list.newNodeWithCapacity(bufIndex - pivotEntryLen - leftIdx)
	newNode.buf = append([]byte{}, pivotNode.buf[leftIdx:bufIndex-pivotEntryLen]...)
	pivotNode.buf = pivotNode.buf[bufIndex-pivotEntryLen:]
	if pivotNode.prev != nil {
		pivotNode.prev.next = newNode
	}
	newNode.prev = pivotNode.prev
	newNode.next = elementNode
	elementNode.prev = newNode
	elementNode.next = pivotNode
	pivotNode.prev = elementNode
	return newNode
}

func (q *Deque) updateHeadLInsert(newNode, prevHead *byteListNode, prevLeftIdx int) {
	if newNode.prev == nil {
		q.list.head = newNode
		q.leftIdx = 0
	} else {
		q.list.head = prevHead
		q.leftIdx = prevLeftIdx
	}
}

func (q *Deque) insertBeforePivotHelper(pivot, element string, qIterator *DequeIterator, pivotNode *byteListNode) {
	pivotEntryLen := int(GetEncodeDeqEntrySize(pivot))
	prevHead := q.list.head
	prevLeftIdx := q.leftIdx
	leftIdx := q.leftIdx
	if pivotNode.prev != nil {
		leftIdx = 0
	}
	elementNode := q.getNewNodeWithElement(element)
	if qIterator.BufIndex == pivotEntryLen || (qIterator.BufIndex == 0 && (len(pivotNode.buf) == pivotEntryLen)) {
		// No need to break the pivotNode into two nodes when the pivot element is the first element in the buffer.
		prevNode := pivotNode.prev
		prevNode.next = elementNode
		pivotNode.prev = elementNode
		elementNode.next = pivotNode
		elementNode.prev = prevNode
	} else {
		newNode := q.breakPivotNodeAndInsertBefore(qIterator, pivotNode, elementNode, pivotEntryLen, leftIdx)
		q.updateHeadLInsert(newNode, prevHead, prevLeftIdx)
	}
	q.Length++
}

func (q *Deque) insertBeforePivotNode(pivot, element string, qIterator *DequeIterator, pivotNode *byteListNode) {
	if qIterator.ElementsTraversed == 1 {
		// Element needs to be inserted at the front of the Deque.
		q.LPush(element)
	} else {
		q.insertBeforePivotHelper(pivot, element, qIterator, pivotNode)
	}
}

func (q *Deque) LInsert(pivot, element, beforeAfter string) (int64, error) {
	// Check if the deque is empty.
	if q.Length == 0 {
		return -1, nil
	}
	if beforeAfter != Before && beforeAfter != After {
		return -1, errors.New("invalid before/after argument")
	}

	qIterator := q.NewIterator()
	for qIterator.HasNext() {
		pivotNode := qIterator.CurrentNode
		if x, _ := qIterator.Next(); x == pivot {
			switch beforeAfter {
			case Before:
				q.insertBeforePivotNode(pivot, element, qIterator, pivotNode)
			case After:
				q.insertAfterPivotNode(element, qIterator, pivotNode)
			}
			return q.Length, nil
		}
	}
	return -1, nil
}

type DequeIterator struct {
	deque             *Deque
	CurrentNode       *byteListNode
	ElementsTraversed int64
	BufIndex          int
}

func (q *Deque) NewIterator() *DequeIterator {
	return &DequeIterator{
		deque:             q,
		CurrentNode:       q.list.head,
		ElementsTraversed: 0,
		BufIndex:          q.leftIdx,
	}
}

func (i *DequeIterator) HasNext() bool {
	return i.ElementsTraversed < i.deque.Length
}

func (i *DequeIterator) Next() (string, error) {
	if !i.HasNext() {
		return "", fmt.Errorf("iterator exhausted")
	}
	x, entryLen := DecodeDeqEntry(i.CurrentNode.buf[i.BufIndex:])
	i.BufIndex += entryLen
	if i.BufIndex == len(i.CurrentNode.buf) {
		i.CurrentNode = i.CurrentNode.next
		i.BufIndex = 0
	}
	i.ElementsTraversed++
	return x, nil
}

// *************************** deque entry encode/decode ***************************

// EncodeDeqEntry encodes `x` into an entry of Deque. An entry will be encoded as [enc + data + backlen].
// References: lpEncodeString, lpEncodeIntegerGetType in redis implementation.
func EncodeDeqEntry(x string) []byte {
	if len(x) >= 21 {
		return EncodeDeqStr(x)
	}
	v, err := strconv.ParseInt(x, 10, 64)
	if err != nil {
		return EncodeDeqStr(x)
	}
	return EncodeDeqInt(v)
}

func EncodeDeqStr(x string) []byte {
	var buf []byte
	var backlen uint64
	strLen := uint64(len(x))
	if strLen <= 63 {
		// 6 bit string
		backlen = 1 + strLen
		backlenSize := dencoding.GetEncodeUIntSize(backlen)
		buf = make([]byte, 1, backlen+backlenSize)
		buf[0] = 0x80 | byte(strLen)
	} else if strLen <= 4095 {
		// 12 bit string
		backlen = 2 + strLen
		backlenSize := dencoding.GetEncodeUIntSize(backlen)
		buf = make([]byte, 2, backlen+backlenSize)
		buf[0] = 0xE0 | byte(strLen>>8)
		buf[1] = byte(strLen)
	} else {
		// 32 bit string
		backlen = 5 + strLen
		backlenSize := dencoding.GetEncodeUIntSize(backlen)
		buf = make([]byte, 5, backlen+backlenSize)
		buf[0] = 0xF0
		buf[1] = byte(strLen)
		buf[2] = byte(strLen >> 8)
		buf[3] = byte(strLen >> 16)
		buf[4] = byte(strLen >> 24)
	}
	buf = append(buf, x...)
	buf = buf[:cap(buf)]
	dencoding.EncodeUIntRevInPlace(backlen, buf[backlen:])
	return buf
}

func EncodeDeqInt(v int64) []byte {
	var buf []byte
	if 0 <= v && v <= 127 {
		// 7 bit uint
		buf = make([]byte, 2)
		buf[0] = byte(v)
	} else if v >= -4096 && v <= 4095 {
		// 13 bit int
		buf = make([]byte, 3)
		buf[0], buf[1] = byte(0xC0|((v>>8)&0x1F)), byte(v)
	} else if v >= -32768 && v <= 32767 {
		// 16 bit int
		buf = make([]byte, 4)
		buf[0], buf[1], buf[2] = byte(0xF1), byte(v), byte(v>>8)
	} else if v >= -8388608 && v <= 8388607 {
		// 24 bit int
		buf = make([]byte, 5)
		buf[0], buf[1], buf[2], buf[3] = byte(0xF2), byte(v), byte(v>>8), byte(v>>16)
	} else if v >= -2147483648 && v <= 2147483647 {
		// 32 bit int
		buf = make([]byte, 6)
		buf[0], buf[1], buf[2], buf[3], buf[4] = byte(0xF3), byte(v), byte(v>>8), byte(v>>16), byte(v>>24)
	} else {
		// 64 bit int
		buf = make([]byte, 10)
		buf[0], buf[1], buf[2], buf[3], buf[4], buf[5], buf[6], buf[7], buf[8] =
			byte(0xF4), byte(v), byte(v>>8), byte(v>>16), byte(v>>24), byte(v>>32), byte(v>>40), byte(v>>48), byte(v>>56)
	}
	buf[len(buf)-1] = byte(len(buf) - 1)
	return buf
}

// EncodeDeqEntryInPlace encodes `x` into an entry of Deque in place.
// References: lpEncodeString, lpEncodeIntegerGetType in redis implementation.
func EncodeDeqEntryInPlace(x string, buf []byte) {
	if len(x) >= 21 {
		EncodeDeqStrInPlace(x, buf)
	} else {
		v, err := strconv.ParseInt(x, 10, 64)
		if err != nil {
			EncodeDeqStrInPlace(x, buf)
		} else {
			EncodeDeqIntInPlace(v, buf)
		}
	}
}

func EncodeDeqStrInPlace(x string, buf []byte) {
	var enclen uint64
	strLen := uint64(len(x))
	if strLen <= 63 {
		// 6 bit string
		enclen = 1
		buf[0] = 0x80 | byte(strLen)
	} else if strLen <= 4095 {
		// 12 bit string
		enclen = 2
		buf[0] = 0xE0 | byte(strLen>>8)
		buf[1] = byte(strLen)
	} else {
		// 32 bit string
		enclen = 5
		buf[0] = 0xF0
		buf[1] = byte(strLen)
		buf[2] = byte(strLen >> 8)
		buf[3] = byte(strLen >> 16)
		buf[4] = byte(strLen >> 24)
	}
	copy(buf[enclen:], x)
	dencoding.EncodeUIntRevInPlace(enclen+strLen, buf[enclen+strLen:])
}

func EncodeDeqIntInPlace(v int64, buf []byte) {
	if 0 <= v && v <= 127 {
		// 7 bit uint
		buf[0] = byte(v)
	} else if v >= -4096 && v <= 4095 {
		// 13 bit int
		buf[0], buf[1] = byte(0xC0|((v>>8)&0x1F)), byte(v)
	} else if v >= -32768 && v <= 32767 {
		// 16 bit int
		buf[0], buf[1], buf[2] = byte(0xF1), byte(v), byte(v>>8)
	} else if v >= -8388608 && v <= 8388607 {
		// 24 bit int
		buf[0], buf[1], buf[2], buf[3] = byte(0xF2), byte(v), byte(v>>8), byte(v>>16)
	} else if v >= -2147483648 && v <= 2147483647 {
		// 32 bit int
		buf[0], buf[1], buf[2], buf[3], buf[4] = byte(0xF3), byte(v), byte(v>>8), byte(v>>16), byte(v>>24)
	} else {
		// 64 bit int
		buf[0], buf[1], buf[2], buf[3], buf[4], buf[5], buf[6], buf[7], buf[8] =
			byte(0xF4), byte(v), byte(v>>8), byte(v>>16), byte(v>>24), byte(v>>32), byte(v>>40), byte(v>>48), byte(v>>56)
	}
	buf[len(buf)-1] = byte(len(buf) - 1)
}

// GetEncodeDeqEntrySize returns the number of bytes the encoded `x` will take.
func GetEncodeDeqEntrySize(x string) uint64 {
	v, err := strconv.ParseInt(x, 10, 64)
	if err != nil {
		return GetEncodeDeqStrSize(x)
	}
	return GetEncodeDeqIntSize(v)
}

func GetEncodeDeqStrSize(x string) uint64 {
	var enclen uint64
	strLen := uint64(len(x))
	if strLen <= 63 {
		enclen = 1
	} else if strLen <= 4095 {
		enclen = 2
	} else {
		enclen = 5
	}
	return enclen + strLen + dencoding.GetEncodeUIntSize(enclen+strLen)
}

func GetEncodeDeqIntSize(v int64) uint64 {
	if 0 <= v && v <= 127 {
		return 2
	} else if v >= -4096 && v <= 4095 {
		return 3
	} else if v >= -32768 && v <= 32767 {
		return 4
	} else if v >= -8388608 && v <= 8388607 {
		return 5
	} else if v >= -2147483648 && v <= 2147483647 {
		return 6
	}
	return 10
}

// DecodeDeqEntry decode `xb` started from index 0, returns the decoded `x` and the
// overall length of [enc + data + backlen].
// References: lpEncodeString, lpEncodeIntegerGetType in redis implementation.
// TODO possible optimizations:
// 1. return the string with the underlying array of `xb` to save memory usage?
// 2. replace strconv with more efficient/memory-saving implementation
func DecodeDeqEntry(xb []byte) (x string, entryLen int) {
	var val int64
	var bit int
	if xb[0]&0x80 == 0 {
		// 7 bit uint
		val = int64(xb[0] & 0x7F)
		bit = 8
		entryLen = 2
	} else if xb[0]&0xE0 == 0xC0 {
		// 13 bit int
		val = (int64(xb[0]&0x1F) << 8) | int64(xb[1])
		bit = 13
		entryLen = 3
	} else if xb[0]&0xFF == 0xF1 {
		// 16 bit int
		val = int64(xb[1]) | int64(xb[2])<<8
		bit = 16
		entryLen = 4
	} else if xb[0]&0xFF == 0xF2 {
		// 24 bit int
		val = int64(xb[1]) | int64(xb[2])<<8 | int64(xb[3])<<16
		bit = 24
		entryLen = 5
	} else if xb[0]&0xFF == 0xF3 {
		// 32 bit int
		val = int64(xb[1]) | int64(xb[2])<<8 | int64(xb[3])<<16 | int64(xb[4])<<24
		bit = 32
		entryLen = 6
	} else if xb[0]&0xFF == 0xF4 {
		// 64 bit int
		val = int64(xb[1]) | int64(xb[2])<<8 | int64(xb[3])<<16 | int64(xb[4])<<24 | int64(xb[5])<<32 | int64(xb[6])<<40 | int64(xb[7])<<48 | int64(xb[8])<<56
		bit = 64
		entryLen = 10
	} else if xb[0]&0xC0 == 0x80 {
		// 6 bit string
		strLen := xb[0] & 0x3F
		backlenlen := dencoding.GetEncodeUIntSize(uint64(1 + strLen))
		return string(xb[1 : 1+strLen]), 1 + int(strLen) + int(backlenlen)
	} else if xb[0]&0xF0 == 0xE0 {
		// 12 bit string
		strLen := (int64(xb[0]&0xF) << 8) | int64(xb[1])
		backlenLen := dencoding.GetEncodeUIntSize(uint64(2 + strLen))
		return string(xb[2 : 2+strLen]), 2 + int(strLen) + int(backlenLen)
	} else if xb[0]&0xFF == 0xF0 {
		// 32 bit string
		strLen := int64(xb[1]) | (int64(xb[2]) << 8) | (int64(xb[3]) << 16) | (int64(xb[4]) << 24)
		backlenlen := dencoding.GetEncodeUIntSize(uint64(5 + strLen))
		return string(xb[5 : 5+strLen]), 5 + int(strLen) + int(backlenlen)
	} else {
		// for recognizing badly encoding case instead of panicking
		val = 12345678900000000 + int64(xb[0])
		bit = 64
	}

	val <<= 64 - bit
	val >>= 64 - bit
	return strconv.FormatInt(val, 10), entryLen
}

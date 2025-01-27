// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package eval

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/dicedb/dice/internal/object"

	dstore "github.com/dicedb/dice/internal/store"
)

type ByteArray struct {
	data   []byte
	Length int64
}

// NewByteArray initializes a new ByteArray with the given size
func NewByteArray(size int) *ByteArray {
	return &ByteArray{
		data:   make([]byte, size),
		Length: int64(size),
	}
}

func NewByteArrayFromObj(obj *object.Obj) (*ByteArray, error) {
	b, err := getValueAsByteSlice(obj)
	if err != nil {
		return nil, err
	}

	return &ByteArray{
		data:   b,
		Length: int64(len(b)),
	}, nil
}

func getValueAsByteSlice(obj *object.Obj) ([]byte, error) {
	oType := obj.Type
	switch oType {
	case object.ObjTypeInt:
		return []byte(strconv.FormatInt(obj.Value.(int64), 10)), nil
	case object.ObjTypeString:
		return getStringValueAsByteSlice(obj)
	// TODO: Have this case as SETBIT stores values encoded as byte arrays. Need to investigate this further.
	case object.ObjTypeByteArray:
		return getByteArrayValueAsByteSlice(obj)
	default:
		return nil, fmt.Errorf("ERR unsopported type")
	}
}

func getStringValueAsByteSlice(obj *object.Obj) ([]byte, error) {
	switch obj.Type {
	case object.ObjTypeInt:
		intVal, ok := obj.Value.(int64)
		if !ok {
			return nil, errors.New("expected integer value but got another type")
		}

		return []byte(strconv.FormatInt(intVal, 10)), nil
	case object.ObjTypeString:
		strVal, ok := obj.Value.(string)
		if !ok {
			return nil, errors.New("expected string value but got another type")
		}

		return []byte(strVal), nil
	default:
		return nil, fmt.Errorf("unsupported type type: %d", obj.Type)
	}
}

func getByteArrayValueAsByteSlice(obj *object.Obj) ([]byte, error) {
	byteArray, ok := obj.Value.(*ByteArray)
	if !ok {
		return nil, errors.New("expected byte array value but got another type")
	}

	return byteArray.data, nil
}

// ByteSliceToObj converts a byte slice to an Obj of the specified type and encoding
func ByteSliceToObj(store *dstore.Store, oldObj *object.Obj, b []byte, objType object.ObjectType) (*object.Obj, error) {
	switch objType {
	case object.ObjTypeInt:
		return ByteSliceToIntObj(store, oldObj, b)
	case object.ObjTypeString:
		return ByteSliceToStringObj(store, oldObj, b)
	case object.ObjTypeByteArray:
		return ByteSliceToByteArrayObj(store, oldObj, b)
	default:
		return nil, fmt.Errorf("unsupported object type")
	}
}

// ByteSliceToIntObj converts a byte slice to an Obj with an integer value
func ByteSliceToIntObj(store *dstore.Store, oldObj *object.Obj, b []byte) (*object.Obj, error) {
	intVal, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		return store.NewObj(string(b), -1, object.ObjTypeString), nil
	}
	return store.NewObj(intVal, -1, object.ObjTypeInt), nil
}

// ByteSliceToStringObj converts a byte slice to an Obj with a string value
func ByteSliceToStringObj(store *dstore.Store, oldObj *object.Obj, b []byte) (*object.Obj, error) {
	return store.NewObj(string(b), -1, object.ObjTypeString), nil
}

// ByteSliceToByteArrayObj converts a byte slice to an Obj with a ByteArray value
func ByteSliceToByteArrayObj(store *dstore.Store, oldObj *object.Obj, b []byte) (*object.Obj, error) {
	byteValue := &ByteArray{
		data:   b,
		Length: int64(len(b)),
	}
	return store.NewObj(byteValue, -1, object.ObjTypeByteArray), nil
}

// SetBit sets the bit at the given position to the specified value
func (b *ByteArray) SetBit(pos int, value bool) {
	byteIndex := pos / 8
	bitIndex := 7 - uint(pos%8)

	if value {
		b.data[byteIndex] |= 1 << bitIndex
	} else {
		b.data[byteIndex] &^= 1 << bitIndex
	}
}

// GetBit gets the bit at the given position
func (b *ByteArray) GetBit(pos int) bool {
	byteIndex := pos / 8
	bitIndex := 7 - uint(pos%8)

	return (b.data[byteIndex] & (1 << bitIndex)) != 0
}

// BitCount counts the number of bits set to 1
func (b *ByteArray) BitCount() int {
	count := 0
	for _, byteVal := range b.data {
		count += int(popcount(byteVal))
	}
	return count
}

func (b *ByteArray) IncreaseSize(increaseSizeTo int) *ByteArray {
	currentByteArray := b.data
	currentByteArraySize := len(currentByteArray)

	// Input is decreasing the size
	if currentByteArraySize >= increaseSizeTo {
		return b
	}

	sizeDifference := increaseSizeTo - currentByteArraySize
	currentByteArray = append(currentByteArray, make([]byte, sizeDifference)...)

	b.data = currentByteArray
	b.Length = int64(increaseSizeTo)

	return b
}

func (b *ByteArray) ResizeIfNecessary() *ByteArray {
	byteArrayLength := b.Length
	decreaseLengthBy := 0
	for i := byteArrayLength - 1; i >= 0; i-- {
		if b.data[i] == 0x0 {
			decreaseLengthBy++
		} else {
			break
		}
	}

	if decreaseLengthBy == 0 {
		return b
	}

	// Decrease the size of the slice to n elements
	// and create a new slice with reduced capacity
	capacityReducedSlice := make([]byte, byteArrayLength-int64(decreaseLengthBy))
	copy(capacityReducedSlice, b.data[:byteArrayLength-int64(decreaseLengthBy)])

	b.data = capacityReducedSlice
	b.Length = int64(len(capacityReducedSlice))

	return b
}

// DeepCopy creates a deep copy of the ByteArray
func (b *ByteArray) DeepCopy() *ByteArray {
	if b == nil {
		return nil
	}

	copyArray := NewByteArray(int(b.Length))

	// Copy the data from the original to the new ByteArray
	copy(copyArray.data, b.data)
	return copyArray
}

func (b *ByteArray) getBits(offset, width int, signed bool) int64 {
	extraBits := 0
	if offset+width > int(b.Length)*8 {
		// If bits exceed the current data size, we will pad the result with zeros for the missing bits.
		extraBits = offset + width - int(b.Length)*8
	}
	var value int64
	for i := 0; i < width-extraBits; i++ {
		value <<= 1
		byteIndex := (offset + i) / 8
		bitIndex := 7 - ((offset + i) % 8)
		if b.data[byteIndex]&(1<<bitIndex) != 0 {
			value |= 1 << 0
		}
	}
	value <<= int64(extraBits)
	if signed && (value&(1<<(width-1)) != 0) {
		value -= 1 << width
	}
	return value
}

func (b *ByteArray) setBits(offset, width int, value int64) {
	if offset+width > int(b.Length)*8 {
		newSize := (offset + width + 7) / 8
		b.IncreaseSize(newSize)
	}
	for i := 0; i < width; i++ {
		byteIndex := (offset + i) / 8
		bitIndex := (offset + i) % 8
		if value&(1<<i) != 0 {
			b.data[byteIndex] |= 1 << bitIndex
		} else {
			b.data[byteIndex] &^= 1 << bitIndex
		}
	}
}

// Increment value at a specific bitfield and handle overflow.
func (b *ByteArray) incrByBits(offset, width int, increment int64, overflow string, signed bool) (int64, error) {
	if offset+width > int(b.Length)*8 {
		newSize := (offset + width + 7) / 8
		b.IncreaseSize(newSize)
	}

	value := b.getBits(offset, width, signed)
	newValue := value + increment

	var maxVal, minVal int64
	if signed {
		maxVal = int64(1<<(width-1) - 1)
		minVal = int64(-1 << (width - 1))
	} else {
		maxVal = int64(1<<width - 1)
		minVal = 0
	}

	switch overflow {
	case WRAP:
		if signed {
			rangeSize := maxVal - minVal + 1
			newValue = ((newValue-minVal)%rangeSize+rangeSize)%rangeSize + minVal
		} else {
			newValue %= maxVal + 1
		}
	case SAT:
		// Handle saturation
		if newValue > maxVal {
			newValue = maxVal
		} else if newValue < minVal {
			newValue = minVal
		}
	case FAIL:
		// Handle failure on overflow
		if newValue > maxVal || newValue < minVal {
			return value, errors.New("overflow detected")
		}
	default:
		return value, errors.New("invalid overflow type")
	}

	b.setBits(offset, width, newValue)
	return newValue, nil
}

// population counting, counts the number of set bits in a byte
// Using: https://en.wikipedia.org/wiki/Hamming_weight
func popcount(x byte) byte {
	// pairing bits and counting them in pairs
	x -= (x >> 1) & 0x55
	// counting bits in groups of four
	x = (x & 0x33) + ((x >> 2) & 0x33)
	// isolates the lower four bits
	// which now contain the total count of set bits in the original byte
	return (x + (x >> 4)) & 0x0F
}

// reverseByte reverses the order of bits in a single byte.

//nolint:unused
func reverseByte(b byte) byte {
	var reversed byte = 0
	for i := 0; i < 8; i++ {
		reversed = (reversed << 1) | (b & 1)
		b >>= 1
	}
	return reversed
}

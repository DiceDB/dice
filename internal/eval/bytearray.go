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
	oType, oEnc := object.ExtractTypeEncoding(obj)
	switch oType {
	case object.ObjTypeInt:
		return []byte(strconv.FormatInt(obj.Value.(int64), 10)), nil
	case object.ObjTypeString:
		return getStringValueAsByteSlice(obj, oEnc)
	// TODO: Have this case as SETBIT stores values encoded as byte arrays. Need to investigate this further.
	case object.ObjTypeByteArray:
		return getByteArrayValueAsByteSlice(obj)
	default:
		return nil, fmt.Errorf("ERR unsopported type")
	}
}

func getStringValueAsByteSlice(obj *object.Obj, oEnc uint8) ([]byte, error) {
	switch oEnc {
	case object.ObjEncodingInt:
		intVal, ok := obj.Value.(int64)
		if !ok {
			return nil, errors.New("expected integer value but got another type")
		}

		return []byte(strconv.FormatInt(intVal, 10)), nil
	case object.ObjEncodingEmbStr, object.ObjEncodingRaw:
		strVal, ok := obj.Value.(string)
		if !ok {
			return nil, errors.New("expected string value but got another type")
		}

		return []byte(strVal), nil
	default:
		return nil, fmt.Errorf("unsupported encoding type: %d", oEnc)
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
func ByteSliceToObj(store *dstore.Store, oldObj *object.Obj, b []byte, objType, encoding uint8) (*object.Obj, error) {
	switch objType {
	case object.ObjTypeInt:
		return ByteSliceToIntObj(store, oldObj, b)
	case object.ObjTypeString:
		return ByteSliceToStringObj(store, oldObj, b, encoding)
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
		return nil, fmt.Errorf("failed to parse byte slice to int: %v", err)
	}
	return store.NewObj(intVal, -1, object.ObjTypeInt, object.ObjEncodingInt), nil
}

// ByteSliceToStringObj converts a byte slice to an Obj with a string value
func ByteSliceToStringObj(store *dstore.Store, oldObj *object.Obj, b []byte, encoding uint8) (*object.Obj, error) {
	switch encoding {
	case object.ObjEncodingInt:
		return ByteSliceToIntObj(store, oldObj, b)
	case object.ObjEncodingEmbStr, object.ObjEncodingRaw:
		return store.NewObj(string(b), -1, object.ObjTypeString, object.ObjEncodingEmbStr), nil
	default:
		return nil, fmt.Errorf("unsupported encoding type")
	}
}

// ByteSliceToByteArrayObj converts a byte slice to an Obj with a ByteArray value
func ByteSliceToByteArrayObj(store *dstore.Store, oldObj *object.Obj, b []byte) (*object.Obj, error) {
	byteValue := &ByteArray{
		data:   b,
		Length: int64(len(b)),
	}
	return store.NewObj(byteValue, -1, object.ObjTypeByteArray, object.ObjEncodingByteArray), nil
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

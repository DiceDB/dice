// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package eval

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"hash/crc64"

	"github.com/dicedb/dice/internal/eval/sortedset"
	"github.com/dicedb/dice/internal/object"
)

func rdbDeserialize(data []byte) (*object.Obj, error) {
	if len(data) < 3 {
		return nil, errors.New("insufficient data for deserialization")
	}
	var value interface{}
	var err error
	var valueRaw interface{}

	buf := bytes.NewReader(data)
	_, err = buf.ReadByte()
	if err != nil {
		return nil, err
	}
	_oType, err := buf.ReadByte()
	if err != nil {
		return nil, err
	}

	objType := object.ObjectType(_oType)
	switch objType {
	case object.ObjTypeString:
		value, err = readString(buf)
	case object.ObjTypeInt: // Integer type
		value, err = readInt(buf)
	case object.ObjTypeSet: // Set type
		value, err = readSet(buf)
	case object.ObjTypeJSON: // JSON type
		valueRaw, err = readString(buf)
		if err := json.Unmarshal([]byte(valueRaw.(string)), &value); err != nil {
			return nil, err
		}
	case object.ObjTypeByteArray: // Byte array type
		valueRaw, err = readInt(buf)
		if err != nil {
			return nil, err
		}
		byteArray := &ByteArray{
			Length: valueRaw.(int64),
			data:   make([]byte, valueRaw.(int64)),
		}
		if _, err := buf.Read(byteArray.data); err != nil {
			return nil, err
		}
		value = byteArray
	case object.ObjTypeDequeue: // Byte list type (Deque)
		value, err = DeserializeDeque(buf)
	case object.ObjTypeBF: // Bloom filter type
		value, err = DeserializeBloom(buf)
	case object.ObjTypeSortedSet:
		value, err = sortedset.DeserializeSortedSet(buf)
	case object.ObjTypeCountMinSketch:
		value, err = DeserializeCMS(buf)
	default:
		return nil, errors.New("unsupported object type")
	}
	if err != nil {
		return nil, err
	}
	return &object.Obj{Type: objType, Value: value}, nil
}

func readString(buf *bytes.Reader) (interface{}, error) {
	var strLen uint32
	if err := binary.Read(buf, binary.BigEndian, &strLen); err != nil {
		return nil, err
	}

	strBytes := make([]byte, strLen)
	if _, err := buf.Read(strBytes); err != nil {
		return nil, err
	}

	return string(strBytes), nil
}

func readInt(buf *bytes.Reader) (interface{}, error) {
	var intVal int64
	if err := binary.Read(buf, binary.BigEndian, &intVal); err != nil {
		return nil, err
	}

	return intVal, nil
}

func readSet(buf *bytes.Reader) (interface{}, error) {
	var strLen uint64
	if err := binary.Read(buf, binary.BigEndian, &strLen); err != nil {
		return nil, err
	}
	setItems := make(map[string]struct{})
	for i := 0; i < int(strLen); i++ {
		value, err := readString(buf)
		if err != nil {
			return nil, err
		}
		setItems[value.(string)] = struct{}{}
	}
	return setItems, nil
}

func rdbSerialize(obj *object.Obj) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(0x09)
	buf.WriteByte(byte(obj.Type))
	switch obj.Type {
	case object.ObjTypeString:
		str, ok := obj.Value.(string)
		if !ok {
			return nil, errors.New("invalid string value")
		}
		if err := writeString(&buf, str); err != nil {
			return nil, err
		}

	case object.ObjTypeInt:
		intVal, ok := obj.Value.(int64)
		if !ok {
			return nil, errors.New("invalid integer value")
		}
		writeInt(&buf, intVal)
	case object.ObjTypeSet:
		setItems, ok := obj.Value.(map[string]struct{})
		if !ok {
			return nil, errors.New("invalid set value")
		}
		if err := writeSet(&buf, setItems); err != nil {
			return nil, err
		}
	case object.ObjTypeJSON:
		jsonValue, err := json.Marshal(obj.Value)
		if err != nil {
			return nil, err
		}
		if err := writeString(&buf, string(jsonValue)); err != nil {
			return nil, err
		}
	case object.ObjTypeByteArray:
		byteArray, ok := obj.Value.(*ByteArray)
		if !ok {
			return nil, errors.New("invalid byte array value")
		}
		writeInt(&buf, byteArray.Length)
		buf.Write(byteArray.data)
	case object.ObjTypeDequeue:
		deque, ok := obj.Value.(*Deque)
		if !ok {
			return nil, errors.New("invalid byte list value")
		}
		if err := deque.Serialize(&buf); err != nil {
			return nil, err
		}
	case object.ObjTypeBF:
		bitSet, ok := obj.Value.(*Bloom)
		if !ok {
			return nil, errors.New("invalid bloom filter value")
		}
		if err := bitSet.Serialize(&buf); err != nil {
			return nil, err
		}
	case object.ObjTypeSortedSet:
		sortedSet, ok := obj.Value.(*sortedset.Set)
		if !ok {
			return nil, errors.New("invalid sorted set value")
		}
		if err := sortedSet.Serialize(&buf); err != nil {
			return nil, err
		}
	case object.ObjTypeCountMinSketch:
		cms, ok := obj.Value.(*CountMinSketch)
		if !ok {
			return nil, errors.New("invalid countminsketch value")
		}
		if err := cms.serialize(&buf); err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("unsupported object type")
	}

	buf.WriteByte(0xFF) // End marker
	return appendChecksum(buf.Bytes()), nil
}

func writeString(buf *bytes.Buffer, str string) error {
	strLen := uint32(len(str))
	if err := binary.Write(buf, binary.BigEndian, strLen); err != nil {
		return err
	}
	buf.WriteString(str)
	return nil
}

func writeInt(buf *bytes.Buffer, intVal int64) {
	tempBuf := make([]byte, 8)
	binary.BigEndian.PutUint64(tempBuf, uint64(intVal))
	buf.Write(tempBuf)
}

func writeSet(buf *bytes.Buffer, setItems map[string]struct{}) error {
	setLen := uint64(len(setItems))
	if err := binary.Write(buf, binary.BigEndian, setLen); err != nil {
		return err
	}
	for item := range setItems {
		if err := writeString(buf, item); err != nil {
			return err
		}
	}
	return nil
}
func appendChecksum(data []byte) []byte {
	checksum := crc64.Checksum(data, crc64.MakeTable(crc64.ECMA))
	checksumBuf := make([]byte, 8)
	binary.BigEndian.PutUint64(checksumBuf, checksum)
	return append(data, checksumBuf...)
}

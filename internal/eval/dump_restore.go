package eval

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc64"

	"github.com/dicedb/dice/internal/eval/sortedset"
	"github.com/dicedb/dice/internal/object"
)

func rdbDeserialize(data []byte) (*object.Obj, error) {
	if len(data) < 3 {
		return nil, errors.New("insufficient data for deserialization")
	}
	buf := bytes.NewReader(data)
	_, err := buf.ReadByte()
	if err != nil {
		return nil, err
	}
	objType, err := buf.ReadByte()
	if err != nil {
		return nil, err
	}
	switch objType {
	case object.ObjTypeString:
		value, err := readString(buf)
		if err != nil {
			return nil, err
		}
		return &object.Obj{Type: objType, Value: value}, nil
	case object.ObjTypeInt: // Integer type
		value, err := readInt(buf)
		if err != nil {
			return nil, err
		}
		return &object.Obj{Type: objType, Value: value}, nil
	case object.ObjTypeSet: // Set type
		value, err := readSet(buf)
		if err != nil {
			return nil, err
		}
		return &object.Obj{Type: objType, Value: value}, nil
	case object.ObjTypeJSON: // JSON type
		value, err := readString(buf)
		if err != nil {
			return nil, err
		}
		var jsonValue interface{}
		if err := json.Unmarshal([]byte(value.(string)), &jsonValue); err != nil {
			return nil, err
		}
		return &object.Obj{Type: objType, Value: jsonValue}, nil
	case object.ObjTypeByteArray: // Byte array type
		value, err := readInt(buf)
		if err != nil {
			return nil, err
		}
		byteArray := &ByteArray{
			Length: value.(int64),
			data:   make([]byte, value.(int64)),
		}
		if _, err := buf.Read(byteArray.data); err != nil {
			return nil, err
		}
		return &object.Obj{Type: objType, Value: byteArray}, nil
	case object.ObjTypeDequeue:
		value, err := readInt(buf)
		if err != nil {
			return nil, err
		}
		byteList := newByteList(value.(int))
		for {
			node := byteList.newNode()
			n, err := buf.Read(node.buf)
			if err != nil {
				return nil, err
			}
			if n == 0 {
				break
			}
			byteList.append(node)
		}
		return &object.Obj{Type: objType, Value: byteList}, nil
	case object.ObjTypeBF: // Bloom filter type
		bloom, err := DeserializeBloom(buf)
		if err != nil {
			return nil, err
		}
		return &object.Obj{Type: objType, Value: bloom}, nil
	case object.ObjTypeSortedSet:
		ss, err := sortedset.DeserializeSortedSet(buf)
		if err != nil {
			return nil, err
		}
		return &object.Obj{Type: objType, Value: ss}, nil
	case object.ObjTypeCountMinSketch:
		cms, err := DeserializeCMS(buf)
		if err != nil {
			return nil, err
		}
		return &object.Obj{Type: objType, Value: cms}, nil
	default:
		return nil, errors.New("unsupported object type")
	}
}

func readString(buf *bytes.Reader) (interface{}, error) {
	var strLen uint32
	if err := binary.Read(buf, binary.BigEndian, &strLen); err != nil {
		return nil, err
	}
	fmt.Println("strLen", strLen)

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
		fmt.Println("value", value)
		setItems[value.(string)] = struct{}{}
	}
	return setItems, nil
}

func rdbSerialize(obj *object.Obj) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(0x09)

	switch obj.Type {
	case object.ObjTypeString:
		str, ok := obj.Value.(string)
		if !ok {
			return nil, errors.New("invalid string value")
		}
		buf.WriteByte(obj.Type)
		if err := writeString(&buf, str); err != nil {
			return nil, err
		}

	case object.ObjTypeInt:
		intVal, ok := obj.Value.(int64)
		if !ok {
			return nil, errors.New("invalid integer value")
		}
		buf.WriteByte(obj.Type)
		writeInt(&buf, intVal)
	case object.ObjTypeSet:
		setItems, ok := obj.Value.(map[string]struct{})
		if !ok {
			return nil, errors.New("invalid set value")
		}
		buf.WriteByte(obj.Type)
		writeSet(&buf, setItems)
	case object.ObjTypeJSON:
		jsonValue, err := json.Marshal(obj.Value)
		if err != nil {
			return nil, err
		}
		buf.WriteByte(obj.Type)
		if err := writeString(&buf, string(jsonValue)); err != nil {
			return nil, err
		}
	case object.ObjTypeByteArray:
		byteArray, ok := obj.Value.(*ByteArray)
		if !ok {
			return nil, errors.New("invalid byte array value")
		}
		buf.WriteByte(obj.Type)
		writeInt(&buf, byteArray.Length)
		buf.Write(byteArray.data)
	case object.ObjTypeDequeue:
		byteList, ok := obj.Value.(*byteList)
		if !ok {
			return nil, errors.New("invalid byte list value")
		}
		buf.WriteByte(obj.Type)
		writeInt(&buf, byteList.size)
		for node := byteList.head; node != nil; node = node.next {
			buf.Write(node.buf)
		}
	case object.ObjTypeBF:
		bitSet, ok := obj.Value.(*Bloom)
		if !ok {
			return nil, errors.New("invalid bloom filter value")
		}
		buf.WriteByte(obj.Type)
		if err := bitSet.Serialize(&buf); err != nil {
			return nil, err
		}
	case object.ObjTypeSortedSet:
		sortedSet, ok := obj.Value.(*sortedset.Set)
		if !ok {
			return nil, errors.New("invalid sorted set value")
		}
		buf.WriteByte(obj.Type)
		if err := sortedSet.Serialize(&buf); err != nil {
			return nil, err
		}
	case object.ObjTypeCountMinSketch:
		cms, ok := obj.Value.(*CountMinSketch)
		if !ok {
			return nil, errors.New("invalid countminsketch value")
		}
		buf.WriteByte(obj.Type)
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
		writeString(buf, item)
	}
	return nil
}
func appendChecksum(data []byte) []byte {
	checksum := crc64.Checksum(data, crc64.MakeTable(crc64.ECMA))
	checksumBuf := make([]byte, 8)
	binary.BigEndian.PutUint64(checksumBuf, checksum)
	return append(data, checksumBuf...)
}

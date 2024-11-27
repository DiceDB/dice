package eval

import (
	"bytes"
	"encoding/binary"
	"errors"
	"hash/crc64"

	"github.com/dicedb/dice/internal/object"
)

func rdbDeserialize(data []byte) (*object.Obj, error) {
	if len(data) < 3 {
		return nil, errors.New("insufficient data for deserialization")
	}
	objType := data[1]
	switch objType {
	case 0x00:
		return readString(data[2:])
	case 0xC0: // Integer type
		return readInt(data[2:])
	default:
		return nil, errors.New("unsupported object type")
	}
}

func readString(data []byte) (*object.Obj, error) {
	buf := bytes.NewReader(data)
	var strLen uint32
	if err := binary.Read(buf, binary.BigEndian, &strLen); err != nil {
		return nil, err
	}

	strBytes := make([]byte, strLen)
	if _, err := buf.Read(strBytes); err != nil {
		return nil, err
	}

	return &object.Obj{Type: object.ObjTypeString, Value: string(strBytes)}, nil
}

func readInt(data []byte) (*object.Obj, error) {
	var intVal int64
	if err := binary.Read(bytes.NewReader(data), binary.BigEndian, &intVal); err != nil {
		return nil, err
	}

	return &object.Obj{Type: object.ObjTypeInt, Value: intVal}, nil
}

func rdbSerialize(obj *object.Obj) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(0x09)

	switch object.GetType(obj.Type) {
	case object.ObjTypeString:
		str, ok := obj.Value.(string)
		if !ok {
			return nil, errors.New("invalid string value")
		}
		buf.WriteByte(0x00)
		if err := writeString(&buf, str); err != nil {
			return nil, err
		}

	case object.ObjTypeInt:
		intVal, ok := obj.Value.(int64)
		if !ok {
			return nil, errors.New("invalid integer value")
		}
		buf.WriteByte(0xC0)
		writeInt(&buf, intVal)

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

func appendChecksum(data []byte) []byte {
	checksum := crc64.Checksum(data, crc64.MakeTable(crc64.ECMA))
	checksumBuf := make([]byte, 8)
	binary.BigEndian.PutUint64(checksumBuf, checksum)
	return append(data, checksumBuf...)
}

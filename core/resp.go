package core

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
)

func readLength(buf *bytes.Buffer) (int64, error) {
	s, err := readStringUntilSr(buf)
	if err != nil {
		return 0, err
	}

	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}

	return v, nil
}

func readStringUntilSr(buf *bytes.Buffer) (string, error) {
	s, err := buf.ReadString('\r')
	if err != nil {
		return "", err
	}
	// increamenting to skip `\n`
	if _, err := buf.ReadByte(); err != nil {
		return "", err
	}
	return s[:len(s)-1], nil
}

// reads a RESP encoded simple string from data and returns
// the string and the error
// the function internally manipulates the buffer pointer
// and keepts it at a point where the subsequent value can be read from.
func readSimpleString(c io.ReadWriter, buf *bytes.Buffer) (string, error) {
	return readStringUntilSr(buf)
}

// reads a RESP encoded error from data and returns
// the error string and the error
// the function internally manipulates the buffer pointer
// and keepts it at a point where the subsequent value can be read from.
func readError(c io.ReadWriter, buf *bytes.Buffer) (string, error) {
	return readStringUntilSr(buf)
}

// reads a RESP encoded integer from data and returns
// the intger value and the error
// the function internally manipulates the buffer pointer
// and keepts it at a point where the subsequent value can be read from.
func readInt64(c io.ReadWriter, buf *bytes.Buffer) (int64, error) {
	s, err := readStringUntilSr(buf)
	if err != nil {
		return 0, err
	}

	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}

	return v, nil
}

// reads a RESP encoded string from data and returns
// the string, and the error
// the function internally manipulates the buffer pointer
// and keepts it at a point where the subsequent value can be read from.
func readBulkString(c io.ReadWriter, buf *bytes.Buffer) (string, error) {
	len, err := readLength(buf)
	if err != nil {
		return "", err
	}

	var bytesRem int64 = len + 2 // 2 for \r\n
	bytesRem = bytesRem - int64(buf.Len())
	for bytesRem > 0 {
		tbuf := make([]byte, bytesRem)
		n, err := c.Read(tbuf)
		if err != nil {
			return "", nil
		}
		buf.Write(tbuf[:n])
		bytesRem = bytesRem - int64(n)
	}

	bulkStr := make([]byte, len)
	if _, err := buf.Read(bulkStr); err != nil {
		return "", err
	}

	// moving buffer pointer by 2 for \r and \n
	if _, err := buf.ReadByte(); err != nil {
		return "", err
	}
	if _, err := buf.ReadByte(); err != nil {
		return "", err
	}

	// reading `len` bytes as string
	return string(bulkStr), nil
}

// reads a RESP encoded array from data and returns
// the array, the delta, and the error
// the function internally manipulates the buffer pointer
// and keepts it at a point where the subsequent value can be read from.
func readArray(c io.ReadWriter, buf *bytes.Buffer, rp *RESPParser) (interface{}, error) {
	count, err := readLength(buf)
	if err != nil {
		return nil, err
	}

	var elems []interface{} = make([]interface{}, count)
	for i := range elems {
		elem, err := rp.DecodeOne()
		if err != nil {
			return nil, err
		}
		elems[i] = elem
	}
	return elems, nil
}

func encodeString(v string) []byte {
	return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(v), v))
}

func Encode(value interface{}, isSimple bool) []byte {
	switch v := value.(type) {
	case string:
		if isSimple {
			return []byte(fmt.Sprintf("+%s\r\n", v))
		}
		return encodeString(v)
	case int, int8, int16, int32, int64:
		return []byte(fmt.Sprintf(":%d\r\n", v))
	case []string:
		var b []byte
		buf := bytes.NewBuffer(b)
		for _, b := range value.([]string) {
			buf.Write(encodeString(b))
		}
		return []byte(fmt.Sprintf("*%d\r\n%s", len(v), buf.Bytes()))
	case []*Obj:
		var b []byte
		buf := bytes.NewBuffer(b)
		for _, b := range value.([]*Obj) {
			buf.Write(Encode(b.Value, false))
		}
		return []byte(fmt.Sprintf("*%d\r\n%s", len(v), buf.Bytes()))
	case []interface{}:
		var b []byte
		buf := bytes.NewBuffer(b)
		for _, b := range value.([]string) {
			buf.Write(Encode(b, false))
		}
		return []byte(fmt.Sprintf("*%d\r\n%s", len(v), buf.Bytes()))
	case *QueueElement:
		var b []byte
		buf := bytes.NewBuffer(b)
		qe := value.(*QueueElement)
		buf.Write(Encode(qe.Key, false))
		buf.Write(Encode(qe.Obj.Value, false))
		return []byte(fmt.Sprintf("*2\r\n%s", buf.Bytes()))
	case []*QueueElement:
		var b []byte
		buf := bytes.NewBuffer(b)
		elements := value.([]*QueueElement)
		for _, qe := range elements {
			buf.Write([]byte("*2\r\n"))
			buf.Write(Encode(qe.Key, false))
			buf.Write(Encode(qe.Obj.Value, false))
		}
		return []byte(fmt.Sprintf("*%d\r\n%s", len(elements), buf.Bytes()))
	case *StackElement:
		var b []byte
		buf := bytes.NewBuffer(b)
		se := value.(*StackElement)
		buf.Write(Encode(se.Key, false))
		buf.Write(Encode(se.Obj.Value, false))
		return []byte(fmt.Sprintf("*2\r\n%s", buf.Bytes()))
	case []*StackElement:
		var b []byte
		buf := bytes.NewBuffer(b)
		elements := value.([]*StackElement)
		for _, se := range elements {
			buf.Write([]byte("*2\r\n"))
			buf.Write(Encode(se.Key, false))
			buf.Write(Encode(se.Obj.Value, false))
		}
		return []byte(fmt.Sprintf("*%d\r\n%s", len(elements), buf.Bytes()))
	case []int64:
		var b []byte
		buf := bytes.NewBuffer(b)
		for _, b := range value.([]int64) {
			buf.Write(Encode(b, false))
		}
		return []byte(fmt.Sprintf("*%d\r\n%s", len(v), buf.Bytes()))
	case error:
		return []byte(fmt.Sprintf("-%s\r\n", v))
	case WatchEvent:
		var b []byte
		buf := bytes.NewBuffer(b)
		we := value.(WatchEvent)
		buf.Write(Encode(fmt.Sprintf("key:%s", we.Key), false))
		buf.Write(Encode(fmt.Sprintf("op:%s", we.Operation), false))
		buf.Write(Encode(we.Value.Value, false))
		return []byte(fmt.Sprintf("*3\r\n%s", buf.Bytes()))
	case []DSQLQueryResultRow:
		var b []byte
		buf := bytes.NewBuffer(b)
		for _, row := range value.([]DSQLQueryResultRow) {
			buf.Write([]byte("*2\r\n"))
			buf.Write(Encode(row.Key, false))
			buf.Write(Encode(row.Value.Value, false))
		}
		return []byte(fmt.Sprintf("*%d\r\n%s", len(v), buf.Bytes()))
	default:
		fmt.Printf("Unsupported type: %T\n", v)
		return RESP_NIL
	}
}

package clientio

import (
	"bytes"
	"fmt"
	"io"
	"strconv"

	"github.com/dicedb/dice/internal/object"

	"github.com/dicedb/dice/internal/sql"

	"github.com/dicedb/dice/internal/server/utils"
	dstore "github.com/dicedb/dice/internal/store"
)

var RespNIL = []byte("$-1\r\n")
var RespOK = []byte("+OK\r\n")
var RespQueued = []byte("+QUEUED\r\n")
var RespZero = []byte(":0\r\n")
var RespOne = []byte(":1\r\n")
var RespMinusOne = []byte(":-1\r\n")
var RespMinusTwo = []byte(":-2\r\n")
var RespEmptyArray = []byte("*0\r\n")

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
		return utils.EmptyStr, err
	}
	// increamenting to skip `\n`
	if _, err := buf.ReadByte(); err != nil {
		return utils.EmptyStr, err
	}
	return s[:len(s)-1], nil
}

// reads a RESP encoded simple string from data and returns
// the string and the error
// the function internally manipulates the buffer pointer
// and keepts it at a point where the subsequent value can be read from.
func readSimpleString(buf *bytes.Buffer) (string, error) {
	return readStringUntilSr(buf)
}

// reads a RESP encoded error from data and returns
// the error string and the error
// the function internally manipulates the buffer pointer
// and keepts it at a point where the subsequent value can be read from.
func readError(buf *bytes.Buffer) (string, error) {
	return readStringUntilSr(buf)
}

// reads a RESP encoded integer from data and returns
// the intger value and the error
// the function internally manipulates the buffer pointer
// and keepts it at a point where the subsequent value can be read from.
func readInt64(buf *bytes.Buffer) (int64, error) {
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
	l, err := readLength(buf)
	if err != nil {
		return utils.EmptyStr, err
	}

	// handling RespNIL case
	if l == -1 {
		return "(nil)", nil
	}

	bytesRem := l + 2 // 2 for \r\n
	bytesRem -= int64(buf.Len())
	for bytesRem > 0 {
		tbuf := make([]byte, bytesRem)
		n, err := c.Read(tbuf)
		if err != nil {
			return utils.EmptyStr, nil
		}
		buf.Write(tbuf[:n])
		bytesRem -= int64(n)
	}

	bulkStr := make([]byte, l)
	if _, err := buf.Read(bulkStr); err != nil {
		return utils.EmptyStr, err
	}

	// moving buffer pointer by 2 for \r and \n
	if _, err := buf.ReadByte(); err != nil {
		return utils.EmptyStr, err
	}
	if _, err := buf.ReadByte(); err != nil {
		return utils.EmptyStr, err
	}

	// reading `len` bytes as string
	return string(bulkStr), nil
}

// reads a RESP encoded array from data and returns
// the array, the delta, and the error
// the function internally manipulates the buffer pointer
// and keepts it at a point where the subsequent value can be read from.
func readArray(buf *bytes.Buffer, rp *RESPParser) (interface{}, error) {
	count, err := readLength(buf)
	if err != nil {
		return nil, err
	}

	elems := make([]interface{}, count)
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
	// Use a type switch to determine the type of the provided value and encode accordingly.
	switch v := value.(type) {
	// Temporary case to maintain backwards compatibility.
	// This case handles byte slices ([]byte) directly, allowing existing functionality
	// that relies on byte slice inputs to continue working without modifications.
	// It serves as a transitional measure and should be revisited for removal
	// once all commands are migrated.
	case []byte:
		return v // Return the byte slice as-is.

	case string:
		// If isSimple is true, format the string in a simple RESP format.
		if isSimple {
			return []byte(fmt.Sprintf("+%s\r\n", v)) // Prefix with '+' for simple response.
		}
		return encodeString(v) // Use detailed encoding for the string.

	// Handle numeric types (int, uint, etc.) by formatting them as RESP integers.
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return []byte(fmt.Sprintf(":%d\r\n", v)) // Prefix with ':' for RESP integers.

	// Handle floating-point types similarly to integers.
	case float32, float64:
		return []byte(fmt.Sprintf(":%v\r\n", v)) // Format as RESP float.

	// Handle slices of strings.
	case []string:
		var b []byte
		buf := bytes.NewBuffer(b) // Create a buffer to accumulate encoded strings.
		for _, b := range value.([]string) {
			buf.Write(encodeString(b)) // Encode each string and write to the buffer.
		}
		return []byte(fmt.Sprintf("*%d\r\n%s", len(v), buf.Bytes())) // Return the encoded response.

	// Handle slices of custom objects (Obj).
	case []*object.Obj:
		var b []byte
		buf := bytes.NewBuffer(b) // Create a buffer to accumulate encoded objects.
		for _, b := range value.([]*object.Obj) {
			buf.Write(Encode(b.Value, false)) // Encode each object’s value and write to the buffer.
		}
		return []byte(fmt.Sprintf("*%d\r\n%s", len(v), buf.Bytes())) // Return the encoded response.

	// Handle slices of interfaces.
	case []interface{}:
		var b []byte
		buf := bytes.NewBuffer(b) // Create a buffer for accumulating encoded values.
		for _, elem := range v {
			buf.Write(Encode(elem, false)) // Encode each element and write to the buffer.
		}
		return []byte(fmt.Sprintf("*%d\r\n%s", len(v), buf.Bytes())) // Return the encoded response.

	// Handle slices of int64.
	case []int64:
		var b []byte
		buf := bytes.NewBuffer(b) // Create a buffer for accumulating encoded values.
		for _, b := range value.([]int64) {
			buf.Write(Encode(b, false)) // Encode each int64 and write to the buffer.
		}
		return []byte(fmt.Sprintf("*%d\r\n%s", len(v), buf.Bytes())) // Return the encoded response.

	// Handle error type by formatting it as a RESP error.
	case error:
		return []byte(fmt.Sprintf("-%s\r\n", v)) // Prefix with '-' for RESP error format.

	// Handle custom watch event struct.
	case dstore.WatchEvent:
		var b []byte
		buf := bytes.NewBuffer(b) // Create a buffer for accumulating encoded values.
		we := value.(dstore.WatchEvent)
		buf.Write(Encode(fmt.Sprintf("key:%s", we.Key), false))      // Encode the key field.
		buf.Write(Encode(fmt.Sprintf("op:%s", we.Operation), false)) // Encode the operation field.
		return []byte(fmt.Sprintf("*2\r\n%s", buf.Bytes()))          // Return the encoded response.

	// Handle slices of SQL query result rows.
	case []sql.QueryResultRow:
		var b []byte
		buf := bytes.NewBuffer(b) // Create a buffer for accumulating encoded rows.
		for _, row := range value.([]sql.QueryResultRow) {
			buf.WriteString("*2\r\n")                 // Start a new array for each row.
			buf.Write(Encode(row.Key, false))         // Encode the row key.
			buf.Write(Encode(row.Value.Value, false)) // Encode the row value.
		}
		return []byte(fmt.Sprintf("*%d\r\n%s", len(v), buf.Bytes())) // Return the encoded response.

	// Handle map[string]bool and return a nil response indicating unsupported types.
	case map[string]bool:
		return RespNIL // Return nil response for unsupported type.

	// For all other unsupported types, return a nil response.
	default:
		return RespNIL // Return nil response for unsupported types.
	}
}

package core

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/dicedb/dice/config"
)

type RESPParser struct {
	c    io.ReadWriter
	buf  *bytes.Buffer
	tbuf []byte
}

func NewRESPParser(c io.ReadWriter) *RESPParser {
	return NewRESPParserWithBytes(c, []byte{})
}

func NewRESPParserWithBytes(c io.ReadWriter, initBytes []byte) *RESPParser {
	var b []byte
	var buf *bytes.Buffer = bytes.NewBuffer(b)
	buf.Write(initBytes)
	return &RESPParser{
		c:   c,
		buf: buf,
		// assigning temporary buffer to read 512 bytes in one shot
		// and reading them in a loop until we have all the data
		// we want.
		// note: the size 512 is arbitrarily chosen, and we can put
		// a decent thought into deciding the optimal value (in case it affects the perf)
		tbuf: make([]byte, config.IOBufferLength),
	}
}

func (rp *RESPParser) DecodeOne() (interface{}, error) {
	// keep reading the bytes from the buffer until the first \r is found
	// note: there may be extra bytes read over the socket post \r
	// but that is fine.
	for {
		n, err := rp.c.Read(rp.tbuf)
		// this condition needs to be explicitly added to ensure
		// we break the loop if we read `0` bytes from the socket
		// implying end of the input
		if n <= 0 {
			break
		}
		rp.buf.Write(rp.tbuf[:n])
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		if bytes.Contains(rp.tbuf, []byte{'\r', '\n'}) {
			break
		}

		if rp.buf.Len() > config.IOBufferLengthMAX {
			return nil, fmt.Errorf("input too long. max input can be %d bytes", config.IOBufferLengthMAX)
		}
	}

	b, err := rp.buf.ReadByte()
	if err != nil {
		return nil, err
	}

	switch b {
	case '+':
		return readSimpleString(rp.c, rp.buf)
	case '-':
		return readError(rp.c, rp.buf)
	case ':':
		return readInt64(rp.c, rp.buf)
	case '$':
		return readBulkString(rp.c, rp.buf)
	case '*':
		return readArray(rp.c, rp.buf, rp)
	}

	// this also captures the Cross Protocol Scripting attack.
	// Since we do not support simple strings, anything that does
	// not start with any of the above special chars will be a potential
	// attack.
	// Details: https://bou.ke/blog/hacking-developers/
	log.Println("possible cross protocol scripting attack detected. dropping the request.")
	return nil, errors.New("possible cross protocol scripting attack detected")
}

func (rp *RESPParser) DecodeMultiple() ([]interface{}, error) {
	var values []interface{} = make([]interface{}, 0)
	for {
		value, err := rp.DecodeOne()
		if err != nil {
			return nil, err
		}
		values = append(values, value)
		if rp.buf.Len() == 0 {
			break
		}
	}
	return values, nil
}

package clientio

import (
	"bytes"
	"errors"
	"io"
	"log"
	"net"
	"os"
	"syscall"

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
	buf := bytes.NewBuffer(b)
	buf.Write(initBytes)
	return &RESPParser{
		c:   c,
		buf: buf,
		// assigning temporary buffer to read 512 bytes in one shot
		// and reading them in a loop until we have all the data
		// we want.
		// note: the size 512 is arbitrarily chosen, and we can put
		// a decent thought into deciding the optimal value (in case it affects the perf)
		tbuf: make([]byte, config.DiceConfig.Network.IOBufferLength),
	}
}

func (rp *RESPParser) DecodeOne() (interface{}, error) {
	// Read data until we find \r\n or hit an error/EOF
	//revive:disable:empty-block
	//revive:disable:empty-lines
	for {
		// 1. Check if the accumulated buffer (rp.buf) contains \r\n before reading
		// more data. This ensures the function does not hang if \r\n is split
		// across multiple reads.
		// 2. Check presence of \r\n in the accumulated buffer (rp.buf.Bytes())
		// rather than the temporary buffer (rp.tbuf). This ensures that data read
		// across multiple iterations of the loop is correctly considered.
		if rp.buf.Len() > 0 && bytes.Contains(rp.buf.Bytes(), []byte{'\r', '\n'}) {
			break
		}

		n, err := rp.c.Read(rp.tbuf)

		if n > 0 {
			rp.buf.Write(rp.tbuf[:n])
		}

		if err != nil {
			// If we have read some data and hit EOF, break to allow the remaining
			// data to be processed.
			if err == io.EOF && rp.buf.Len() > 0 {
				break
			}

			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				return nil, errors.New("connection timeout")
			}

			if errors.Is(err, syscall.EBADF) || errors.Is(err, os.ErrClosed) {
				return nil, errors.New("connection closed locally")
			}

			if errors.Is(err, syscall.ECONNRESET) {
				return nil, errors.New("connection reset")
			}

			return nil, err
		}

		// Handle the case where no data is read but no error is returned
		if n == 0 {
			// This can happen if the connection is closed on the client side but not properly detected
			return nil, net.ErrClosed
		}
	}

	b, err := rp.buf.ReadByte()
	if err != nil {
		return nil, err
	}

	switch b {
	case '+':
		return readSimpleString(rp.buf)
	case '-':
		return readError(rp.buf)
	case ':':
		return readInt64(rp.buf)
	case '$':
		return readBulkString(rp.c, rp.buf)
	case '*':
		return readArray(rp.buf, rp)
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
	values := make([]interface{}, 0)
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

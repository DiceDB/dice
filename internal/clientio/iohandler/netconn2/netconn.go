// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package netconn2

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/internal/clientio"
	"github.com/dicedb/dice/internal/clientio/iohandler"
	"github.com/dicedb/dice/internal/cmd"
	"github.com/dicedb/dice/wire"
	"google.golang.org/protobuf/proto"
)

const (
	maxRequestSize = 32 * 1024 * 1024 // 32 MB
	ioBufferSize   = 16 * 1024        // 16 KB
	idleTimeout    = 30 * time.Minute
)

var (
	ErrRequestTooLarge = errors.New("request too large")
	ErrIdleTimeout     = errors.New("connection idle timeout")
	ErrorClosed        = errors.New("connection closed")
)

// Pre-allocate the response array
// WARN: Do not change the ordering of the array elements
// It is strictly mapped to internal/eval/results.go enum.
var respArray = [][]byte{
	clientio.RespNIL,
	clientio.RespOK,
	clientio.RespQueued,
	clientio.RespZero,
	clientio.RespOne,
	clientio.RespMinusOne,
	clientio.RespMinusTwo,
	clientio.RespEmptyArray,
}

// IOHandler handles I/O operations for a network connection
type IOHandler struct {
	fd   int
	file *os.File
	conn net.Conn
}

var _ iohandler.IOHandler = (*IOHandler)(nil)

// NewIOHandler creates a new IOHandler from a file descriptor
func NewIOHandler(clientFD int) (*IOHandler, error) {
	file := os.NewFile(uintptr(clientFD), "client-connection")
	if file == nil {
		return nil, fmt.Errorf("failed to create file from file descriptor")
	}

	var conn net.Conn
	defer func() {
		// Only close the file if we haven't successfully created a net.Conn
		if conn == nil {
			file.Close()
		}
	}()

	var err error
	conn, err = net.FileConn(file)
	if err != nil {
		return nil, fmt.Errorf("failed to create net.Conn from file descriptor: %w", err)
	}

	if tcpConn, ok := conn.(*net.TCPConn); ok {
		if err := tcpConn.SetNoDelay(true); err != nil {
			return nil, fmt.Errorf("failed to set TCP_NODELAY: %w", err)
		}
		if err := tcpConn.SetKeepAlive(true); err != nil {
			return nil, fmt.Errorf("failed to set keepalive: %w", err)
		}
		if err := tcpConn.SetKeepAlivePeriod(time.Duration(config.KeepAlive) * time.Second); err != nil {
			return nil, fmt.Errorf("failed to set keepalive period: %w", err)
		}
	}

	return &IOHandler{
		fd:   clientFD,
		file: file,
		conn: conn,
	}, nil
}

func NewIOHandlerWithConn(conn net.Conn) *IOHandler {
	return &IOHandler{
		conn: conn,
	}
}

// ReadRequest reads data from the network connection
func (h *IOHandler) Read(ctx context.Context) ([]byte, error) {
	return nil, nil
}

// ReadRequest reads data from the network connection
func (h *IOHandler) ReadSync() (*cmd.Cmd, error) {
	var result []byte
	reader := bufio.NewReaderSize(h.conn, ioBufferSize)
	buf := make([]byte, ioBufferSize)

	for {
		n, err := reader.Read(buf)
		if n > 0 {
			if len(result)+n > maxRequestSize {
				return nil, fmt.Errorf("request too large")
			}

			result = append(result, buf[:n]...)
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		if n < len(buf) {
			break
		}
	}

	if len(result) == 0 {
		return nil, io.EOF
	}

	c := &wire.Command{}
	if err := proto.Unmarshal(result, c); err != nil {
		return nil, fmt.Errorf("failed to unmarshal command: %w", err)
	}
	return &cmd.Cmd{C: c}, nil
}

func (h *IOHandler) Write(ctx context.Context, r interface{}) error {
	return nil
}

func (h *IOHandler) WriteSync(ctx context.Context, r *cmd.CmdRes) error {
	var b []byte
	var err error

	if b, err = proto.Marshal(r.R); err != nil {
		return err
	}

	if _, err := h.conn.Write(b); err != nil {
		return err
	}

	return nil
}

// Close underlying network connection
func (h *IOHandler) Close() error {
	var err error
	if h.conn != nil {
		err = errors.Join(err, h.conn.Close())
	}
	if h.file != nil {
		err = errors.Join(err, h.file.Close())
	}

	return err
}

// handleResponse processes the incoming response from a client and returns the corresponding
// RESP formatted byte array based on the response content.
//
// The function takes an interface{} as input, attempts to assert it as a byte slice. If successful,
// it checks the content of the byte slice against predefined RESP responses using the `bytes.Contains`
// function. If a match is found, it returns the associated byte array response. If no match is found
// or if the input cannot be converted to a byte slice, the function returns nil.
//
// This function is designed to handle various response scenarios, such as:
// - $-1: Represents a nil response.
// - +OK: Indicates a successful command execution.
// - +QUEUED: Signifies that a command has been queued.
// - :0, :1, :-1, :-2: Represents integer values in RESP format.
// - *0: Represents an empty array in RESP format.
//
// Note: The use of `bytes.Contains` is to check if the provided response matches any of the
// predefined RESP responses, making it flexible in handling responses that might include
// additional content beyond the expected response format.
func HandlePredefinedResponse(response interface{}) []byte {
	if val, ok := response.(clientio.RespType); ok {
		return respArray[val]
	}

	return nil
}

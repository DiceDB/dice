// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package netconn

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/dicedb/dice/internal/clientio"
	"github.com/dicedb/dice/internal/clientio/iohandler"
)

const (
	maxRequestSize  = 32 * 1024 * 1024 // 32 MB, Redis max request size is 512MB
	ioBufferSize    = 16 * 1024        // 16 KB
	idleTimeout     = 30 * time.Minute
	writeTimeout    = 10 * time.Second
	keepAlivePeriod = 30 * time.Second
)

var (
	ErrRequestTooLarge = errors.New("request too large")
	ErrIdleTimeout     = errors.New("connection idle timeout")
	ErrorClosed        = errors.New("connection closed")
	errReadDeadline    = errors.New("error setting read deadline")
	errReadRequest     = errors.New("error reading request")
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
	fd       int
	file     *os.File
	conn     net.Conn
	reader   *bufio.Reader
	writer   *bufio.Writer
	readPool *sync.Pool
}

var _ iohandler.IOHandler = (*IOHandler)(nil)

// NewIOHandler creates a new IOHandler from a file descriptor
func NewIOHandler(clientFD int) (*IOHandler, error) {
	file := os.NewFile(uintptr(clientFD), "client-connection")
	if file == nil {
		return nil, fmt.Errorf("failed to create file from file descriptor")
	}

	// Ensure the file is closed if we exit this function with an error
	var conn net.Conn
	defer func() {
		if conn == nil {
			// Only close the file if we haven't successfully created a net.Conn
			err := file.Close()
			if err != nil {
				slog.Warn("Error closing file in NewIOHandler:", slog.Any("error", err))
			}
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
		if err := tcpConn.SetKeepAlivePeriod(keepAlivePeriod); err != nil {
			return nil, fmt.Errorf("failed to set keepalive period: %w", err)
		}
	}

	return &IOHandler{
		fd:     clientFD,
		file:   file,
		conn:   conn,
		reader: bufio.NewReaderSize(conn, ioBufferSize),
		writer: bufio.NewWriterSize(conn, ioBufferSize),
		readPool: &sync.Pool{
			New: func() interface{} {
				b := make([]byte, ioBufferSize)
				return &b // Return pointer to avoid interface conversion allocation
			},
		},
	}, nil
}

func NewIOHandlerWithConn(conn net.Conn) *IOHandler {
	return &IOHandler{
		conn:   conn,
		reader: bufio.NewReader(conn),
		writer: bufio.NewWriter(conn),
		readPool: &sync.Pool{
			New: func() interface{} {
				b := make([]byte, ioBufferSize)
				return &b
			},
		},
	}
}

func (h *IOHandler) FileDescriptor() int {
	return h.fd
}

// ReadRequest reads data from the network connection
func (h *IOHandler) Read(ctx context.Context) ([]byte, error) {
	// Get pointer from pool and dereference
	buf := *h.readPool.Get().(*[]byte)
	defer h.readPool.Put(&buf) // Put pointer back into pool

	var result []byte

	for {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
			err := h.conn.SetReadDeadline(time.Now().Add(idleTimeout))
			if err != nil {
				return nil, errReadDeadline
			}

			n, err := h.reader.Read(buf)
			if n > 0 {
				// Check if adding this chunk would exceed max request size
				if len(result)+n > maxRequestSize {
					return nil, ErrRequestTooLarge
				}

				result = append(result, buf[:n]...)
			}

			if err != nil {
				return h.handleReadError(err, result)
			}

			// If we've read less than the buffer size, we've likely got all the data
			if n < len(buf) {
				return result, nil
			}
		}
	}
}

// handleReadError handles various read errors and returns appropriate response
func (h *IOHandler) handleReadError(err error, data []byte) ([]byte, error) {
	switch {
	case errors.Is(err, syscall.EAGAIN), errors.Is(err, syscall.EWOULDBLOCK):
		return data, nil
	case errors.Is(err, io.EOF):
		if len(data) > 0 {
			return data, nil
		}
		return nil, io.EOF
	case errors.Is(err, net.ErrClosed), errors.Is(err, syscall.EPIPE), errors.Is(err, syscall.ECONNRESET):
		slog.Debug("Connection closed", slog.Any("error", err))
		cErr := h.Close()
		if cErr != nil {
			slog.Warn("Error closing connection", slog.Any("error", errors.Join(err, cErr)))
		}
		return nil, ErrorClosed
	case errors.Is(err, syscall.ETIMEDOUT):
		slog.Info("Connection idle timeout", slog.Any("error", err))
		cerr := h.Close()
		if cerr != nil {
			slog.Warn("Error closing connection", slog.Any("error", errors.Join(err, cerr)))
		}
		return nil, ErrIdleTimeout
	default:
		return nil, fmt.Errorf("%w: %v", errReadRequest, err)
	}
}

// WriteResponse writes the response back to the network connection
func (h *IOHandler) Write(ctx context.Context, response interface{}) error {
	// Process the incoming response by calling the handleResponse function.
	// This function checks the response against known RESP formatted values
	// and returns the corresponding byte array representation. The result
	// is assigned to the resp variable.
	resp := HandlePredefinedResponse(response)

	// Check if the processed response (resp) is not nil.
	// If it is not nil, this means incoming response was not
	// matched to any predefined RESP responses,
	// and we proceed to encode the original response using
	// the clientio.Encode function. This function converts the
	// response into the desired format based on the specified
	// isBlkEnc encoding flag, which indicates whether the
	// response should be encoded in a block format.
	if resp == nil {
		resp = clientio.Encode(response, true)
	}

	deadline := time.Now().Add(writeTimeout)
	if err := h.conn.SetWriteDeadline(deadline); err != nil {
		slog.Warn("error setting write deadline", slog.Any("error", err))
	}

	errChan := make(chan error, 1)
	go func() {
		defer close(errChan)

		var err error
		if _, err = h.writer.Write(resp); err == nil {
			err = h.writer.Flush()
		}

		errChan <- err
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err, ok := <-errChan:
		if !ok {
			slog.Warn("write operation failed: error channel closed unexpectedly")
		}

		if err != nil {
			if errors.Is(err, net.ErrClosed) || errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ECONNRESET) {
				if cErr := h.Close(); cErr != nil {
					err = errors.Join(err, cErr)
				}

				slog.Debug("connection closed", slog.Any("error", err))
				return err
			}

			return fmt.Errorf("error writing response: %w", err)
		}
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
// RESP (REdis Serialization Protocol) formatted byte array based on the response content.
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

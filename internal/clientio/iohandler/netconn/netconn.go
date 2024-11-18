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
	"syscall"
	"time"

	"github.com/dicedb/dice/internal/clientio"
	"github.com/dicedb/dice/internal/clientio/iohandler"
)

const (
	maxRequestSize = 512 * 1024 // 512 KB
	readBufferSize = 4 * 1024   // 4 KB
	idleTimeout    = 10 * time.Minute
)

var (
	ErrRequestTooLarge = errors.New("request too large")
	ErrIdleTimeout     = errors.New("connection idle timeout")
	ErrorClosed        = errors.New("connection closed")
)

// IOHandler handles I/O operations for a network connection
type IOHandler struct {
	fd     int
	file   *os.File
	conn   net.Conn
	reader *bufio.Reader
	writer *bufio.Writer
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

	return &IOHandler{
		fd:     clientFD,
		file:   file,
		conn:   conn,
		reader: bufio.NewReader(conn),
		writer: bufio.NewWriter(conn),
	}, nil
}

func NewIOHandlerWithConn(conn net.Conn) *IOHandler {
	return &IOHandler{
		conn:   conn,
		reader: bufio.NewReader(conn),
		writer: bufio.NewWriter(conn),
	}
}

func (h *IOHandler) FileDescriptor() int {
	return h.fd
}

// ReadRequest reads data from the network connection
func (h *IOHandler) Read(ctx context.Context) ([]byte, error) {
	var data []byte
	buf := make([]byte, readBufferSize)

	for {
		select {
		case <-ctx.Done():
			return data, ctx.Err()
		default:
			err := h.conn.SetReadDeadline(time.Now().Add(idleTimeout))
			if err != nil {
				return nil, fmt.Errorf("error setting read deadline: %w", err)
			}

			n, err := h.reader.Read(buf)
			if n > 0 {
				data = append(data, buf[:n]...)
			}
			if err != nil {
				switch {
				case errors.Is(err, syscall.EAGAIN), errors.Is(err, syscall.EWOULDBLOCK), errors.Is(err, io.EOF):
					// No more data to read at this time
					return data, nil
				case errors.Is(err, net.ErrClosed), errors.Is(err, syscall.EPIPE), errors.Is(err, syscall.ECONNRESET):
					slog.Debug("Connection closed", slog.Any("error", err))
					cerr := h.Close()
					if cerr != nil {
						slog.Warn("Error closing connection", slog.Any("error", errors.Join(err, cerr)))
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
					slog.Error("Error reading from connection", slog.Any("error", err))
					return nil, fmt.Errorf("error reading request: %w", err)
				}
			}

			if len(data) > maxRequestSize {
				slog.Warn("Request too large", slog.Any("size", len(data)))
				return nil, ErrRequestTooLarge
			}

			// If we've read less than the buffer size, we've likely got all the data
			if n < len(buf) {
				return data, nil
			}
		}
	}
}

// WriteResponse writes the response back to the network connection
func (h *IOHandler) Write(ctx context.Context, response interface{}) error {
	errChan := make(chan error, 1)

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

	go func(errChan chan error) {
		_, err := h.writer.Write(resp)
		if err == nil {
			err = h.writer.Flush()
		}

		errChan <- err
	}(errChan)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errChan:
		if err != nil {
			if errors.Is(err, net.ErrClosed) || errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ECONNRESET) {
				cerr := h.Close()
				if cerr != nil {
					err = errors.Join(err, cerr)
				}

				slog.Debug("Connection closed", slog.Any("error", err))
				return err
			}

			return fmt.Errorf("error writing response: %w", err)
		}
	}

	return nil
}

// Close underlying network connection
func (h *IOHandler) Close() error {
	slog.Info("Closing connection")
	return errors.Join(h.conn.Close(), h.file.Close())
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
	// WARN: Do not change the ordering of the array elements
	// It is strictly mapped to internal/eval/results.go enum.
	respArr := [][]byte{
		clientio.RespNIL,        // Represents a RESP Nil Bulk String, which indicates a null value.
		clientio.RespOK,         // Represents a RESP Simple String with value "OK".
		clientio.RespQueued,     // Represents a Simple String indicating that a command has been queued.
		clientio.RespZero,       // Represents a RESP Integer with value 0.
		clientio.RespOne,        // Represents a RESP Integer with value 1.
		clientio.RespMinusOne,   // Represents a RESP Integer with value -1.
		clientio.RespMinusTwo,   // Represents a RESP Integer with value -2.
		clientio.RespEmptyArray, // Represents an empty RESP Array.
	}

	switch val := response.(type) {
	case clientio.RespType:
		return respArr[val]
	default:
		return nil
	}
}

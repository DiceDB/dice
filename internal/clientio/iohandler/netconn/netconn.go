package netconn

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/charmbracelet/log"
	"github.com/dicedb/dice/internal/clientio/iohandler"
	"io"
	"net"
	"os"
	"time"
)

const (
	maxRequestSize = 512 * 1024 // 512 KB
	readBufferSize = 4 * 1024   // 4 KB
	idleTimeout    = 10 * time.Minute
)

var (
	ErrRequestTooLarge = errors.New("request too large")
	ErrIdleTimeout     = errors.New("connection idle timeout")
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
				log.Errorf("Error closing file in NewIOHandler: %v", err)
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
func (h *IOHandler) ReadRequest(ctx context.Context) ([]byte, error) {
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
				if errors.Is(err, io.EOF) {
					// encountered EOF, return the data we have so far
					return data, nil
				}

				var netErr net.Error
				if errors.As(err, &netErr) && netErr.Timeout() {
					return nil, ErrIdleTimeout
				}

				return data, fmt.Errorf("error reading request: %w", err)
			}

			if len(data) > maxRequestSize {
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
func (h *IOHandler) WriteResponse(ctx context.Context, response []byte) error {
	writeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	errChan := make(chan error, 1)

	go func() {
		_, err := h.writer.Write(response)
		if err == nil {
			err = h.writer.Flush()
		}

		errChan <- err
	}()

	select {
	case <-writeCtx.Done():
		return writeCtx.Err()
	case err := <-errChan:
		if err != nil {
			return fmt.Errorf("error writing response: %w", err)
		}
		return nil
	}
}

// Close underlying network connection
func (h *IOHandler) Close() error {
	return errors.Join(h.conn.Close(), h.file.Close())
}

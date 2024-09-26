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
	logger *slog.Logger
}

var _ iohandler.IOHandler = (*IOHandler)(nil)

// NewIOHandler creates a new IOHandler from a file descriptor
func NewIOHandler(clientFD int, logger *slog.Logger) (*IOHandler, error) {
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
				logger.Warn("Error closing file in NewIOHandler:", slog.Any("error", err))
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
		logger: logger,
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
					h.logger.Error("Connection closed", slog.Any("error", err))
					cerr := h.Close()
					if cerr != nil {
						h.logger.Warn("Error closing connection", slog.Any("error", errors.Join(err, cerr)))
					}
					return nil, ErrorClosed
				case errors.Is(err, syscall.ETIMEDOUT):
					h.logger.Info("Connection idle timeout", slog.Any("error", err))
					cerr := h.Close()
					if cerr != nil {
						h.logger.Warn("Error closing connection", slog.Any("error", errors.Join(err, cerr)))
					}
					return nil, ErrIdleTimeout
				default:
					h.logger.Error("Error reading from connection", slog.Any("error", err))
					return nil, fmt.Errorf("error reading request: %w", err)
				}
			}

			if len(data) > maxRequestSize {
				h.logger.Warn("Request too large", slog.Any("size", len(data)))
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
func (h *IOHandler) Write(ctx context.Context, response []byte) error {
	errChan := make(chan error, 1)

	go func(errChan chan error) {
		_, err := h.writer.Write(response)
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

				h.logger.Error("Connection closed", slog.Any("error", err))
				return err
			}

			return fmt.Errorf("error writing response: %w", err)
		}
	}

	return nil
}

// Close underlying network connection
func (h *IOHandler) Close() error {
	h.logger.Info("Closing connection")
	return errors.Join(h.conn.Close(), h.file.Close())
}

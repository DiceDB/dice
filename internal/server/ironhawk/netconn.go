// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

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
	"github.com/dicedb/dicedb-go/wire"
	"google.golang.org/protobuf/proto"
)

var (
	ErrRequestTooLarge = errors.New("request too large")
	ErrIdleTimeout     = errors.New("connection idle timeout")
	ErrorClosed        = errors.New("connection closed")
)

// IOHandler handles I/O operations for a network connection
type IOHandler struct {
	fd   int
	file *os.File
	conn net.Conn
}

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
func (h *IOHandler) ReadSync() (*wire.Command, error) {
	var result []byte
	reader := bufio.NewReaderSize(h.conn, config.IoBufferSize)
	buf := make([]byte, config.IoBufferSize)

	for {
		n, err := reader.Read(buf)
		if n > 0 {
			if len(result)+n > config.MaxRequestSize {
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
	return c, nil
}

func (h *IOHandler) Write(ctx context.Context, r interface{}) error {
	return nil
}

func (h *IOHandler) WriteSync(ctx context.Context, r *wire.Result) error {
	var b []byte
	var err error

	if b, err = proto.Marshal(r); err != nil {
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

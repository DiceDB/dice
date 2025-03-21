// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"context"
	"encoding/binary"
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

const (
	headerSize = 4
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
	headerBuffer := make([]byte, headerSize)
	if _, err := io.ReadFull(h.conn, headerBuffer); err != nil {
		return nil, fmt.Errorf("failed to read message header: %w", err)
	}

	messageSize := binary.BigEndian.Uint32(headerBuffer)
	if messageSize == 0 {
		return nil, fmt.Errorf("invalid message size: 0 bytes")
	}
	if messageSize > uint32(config.MaxRequestSize) {
		return nil, fmt.Errorf("message too large: %d bytes (max: %d)", messageSize, config.MaxRequestSize)
	}

	messageBuffer := make([]byte, messageSize)
	if _, err := io.ReadFull(h.conn, messageBuffer); err != nil {
		return nil, fmt.Errorf("failed to read message into buffer: %w", err)
	}

	cmd := &wire.Command{}
	if err := proto.Unmarshal(messageBuffer, cmd); err != nil {
		return nil, fmt.Errorf("failed to unmarshal command: %w", err)
	}

	return cmd, nil
}

func (h *IOHandler) Write(ctx context.Context, r interface{}) error {
	return nil
}

func (h *IOHandler) WriteSync(ctx context.Context, response *wire.Response) error {
	message, err := proto.Marshal(response)
	if err != nil {
		return err
	}

	messageSize := len(message)
	messageBuffer := make([]byte, headerSize+messageSize)
	binary.BigEndian.PutUint32(messageBuffer[:headerSize], uint32(messageSize))
	copy(messageBuffer[headerSize:], message)

	if _, err := h.conn.Write(messageBuffer); err != nil {
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

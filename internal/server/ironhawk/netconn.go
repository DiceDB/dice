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
	Ctx  context.Context
}

// NewIOHandler creates a new IOHandler from a file descriptor
func NewIOHandler(ctx context.Context, clientFD int) (*IOHandler, error) {
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
		Ctx:  ctx,
	}, nil
}

func NewIOHandlerWithConn(ctx context.Context, conn net.Conn) *IOHandler {
	return &IOHandler{
		conn: conn,
		Ctx:  ctx,
	}
}

// ReadRequest reads data from the network connection
func (h *IOHandler) Read(ctx context.Context) ([]byte, error) {
	return nil, nil
}

// ReadRequest reads data from the network connection
func (h *IOHandler) ReadSync() (*wire.Command, error) {
	resultChan := make(chan []byte)
	errorChan := make(chan error)

	go h.readDataFromBuffer(resultChan, errorChan)

	select {
	case <-h.Ctx.Done():
		return nil, io.EOF
	case result := <-resultChan:
		if len(result) == 0 {
			return nil, io.EOF
		}
		c := &wire.Command{}
		if err := proto.Unmarshal(result, c); err != nil {
			return nil, fmt.Errorf("failed to unmarshal command: %w", err)
		}
		return c, nil
	case err := <-errorChan:
		return nil, err
	}
}

func (h *IOHandler) readDataFromBuffer(resultChan chan<- []byte, errorChan chan<- error) {
	defer close(resultChan)
	defer close(errorChan)
	reader := bufio.NewReaderSize(h.conn, config.IoBufferSize)

	for {
		select {
		case <-h.Ctx.Done():
			errorChan <- io.EOF
			return
		default:
			var result []byte
			buf := make([]byte, config.IoBufferSize)
			n, err := reader.Read(buf)
			if n > 0 {
				if len(result)+n > config.MaxRequestSize {
					errorChan <- fmt.Errorf("request too large")
					return
				}

				result = append(result, buf[:n]...)
			}
			if err != nil {
				if err == io.EOF {
					resultChan <- result
					return
				}
				errorChan <- err
			}
			if n < len(buf) {
				resultChan <- result
				return
			}
		}
	}
}

func (h *IOHandler) Write(r interface{}) error {
	return nil
}

func (h *IOHandler) WriteSync(r *wire.Response) error {
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

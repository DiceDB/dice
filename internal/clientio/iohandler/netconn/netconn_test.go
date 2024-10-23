package netconn

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockConn implements net.Conn interface for testing
type mockConn struct {
	readData  []byte
	writeData bytes.Buffer
	readErr   error
	writeErr  error
	closed    bool
	mu        sync.Mutex
}

func (m *mockConn) Read(b []byte) (n int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.readErr != nil {
		return 0, m.readErr
	}
	if len(m.readData) == 0 {
		return 0, io.EOF
	}
	n = copy(b, m.readData)
	m.readData = m.readData[n:]
	return n, nil
}

func (m *mockConn) Write(b []byte) (n int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.writeErr != nil {
		return 0, m.writeErr
	}
	return m.writeData.Write(b)
}

func (m *mockConn) Close() error {
	m.closed = true
	return nil
}
func (m *mockConn) LocalAddr() net.Addr                { return nil }
func (m *mockConn) RemoteAddr() net.Addr               { return nil }
func (m *mockConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error { return nil }

func (m *mockConn) File() (*os.File, error) {
	return &os.File{}, nil
}

func TestNetConnIOHandler(t *testing.T) {
	tests := []struct {
		name             string
		readData         []byte
		readErr          error
		writeErr         error
		ctxTimeout       time.Duration
		response         []byte
		expectedRead     []byte
		expectedWrite    []byte
		expectedReadErr  error
		expectedWriteErr error
	}{
		{
			name:          "Simple read and write",
			readData:      []byte("Hello, World!\r\n"),
			expectedRead:  []byte("Hello, World!\r\n"),
			expectedWrite: []byte("Response\r\n"),
		},
		{
			name:            "Read error",
			readErr:         errors.New("read error"),
			expectedReadErr: errors.New("error reading request: read error"),
			expectedWrite:   []byte("Response\r\n"),
		},
		{
			name:             "Write error",
			readData:         []byte("Hello, World!\r\n"),
			expectedRead:     []byte("Hello, World!\r\n"),
			writeErr:         errors.New("write error"),
			response:         []byte("Hello, World!\r\n"),
			expectedWriteErr: errors.New("error writing response: write error"),
		},
		{
			name:          "Large data read",
			readData:      bytes.Repeat([]byte("a"), 1000),
			expectedRead:  bytes.Repeat([]byte("a"), 1000),
			expectedWrite: []byte("Response\r\n"),
		},
		{
			name:          "Empty read",
			readData:      []byte{},
			expectedRead:  []byte(nil),
			expectedWrite: []byte("Response\r\n"),
		},
		{
			name:          "Read with multiple chunks",
			readData:      []byte("Hello\r\nWorld\r\n"),
			expectedRead:  []byte("Hello\r\nWorld\r\n"),
			expectedWrite: []byte("Response\r\n"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockConn{
				readData: tt.readData,
				readErr:  tt.readErr,
				writeErr: tt.writeErr,
			}

			handler := &IOHandler{
				conn:   mock,
				reader: bufio.NewReaderSize(mock, 512),
				writer: bufio.NewWriterSize(mock, 1024),
			}

			ctx := context.Background()
			if tt.ctxTimeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, tt.ctxTimeout)
				defer cancel()
			}

			// Test ReadRequest
			data, err := handler.Read(ctx)
			if tt.expectedReadErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedReadErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRead, data)
			}

			// Test WriteResponse
			if tt.response == nil {
				tt.response = []byte("Response\r\n")
			}

			err = handler.Write(ctx, tt.response)

			if tt.expectedWriteErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedWriteErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedWrite, mock.writeData.Bytes())
			}
		})
	}
}

func TestNewNetConnIOHandler(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() (int, func(), error)
		expectedErr error
	}{
		{
			name: "Closed file descriptor",
			setup: func() (int, func(), error) {
				f, err := os.CreateTemp("", "test")
				if err != nil {
					return 0, nil, err
				}
				fd := int(f.Fd())
				f.Close() // Close immediately to create a closed fd
				return fd, func() {}, nil
			},
			expectedErr: errors.New("failed to create net.Conn"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fd, cleanup, err := tt.setup()
			require.NoError(t, err, "Setup failed")
			defer cleanup()

			handler, err := NewIOHandler(fd)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Nil(t, handler)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, handler)
				assert.NotNil(t, handler.conn)
				assert.NotNil(t, handler.reader)
				assert.NotNil(t, handler.writer)

				// Test if the created handler can perform basic I/O
				testData := []byte("Hello, World!")
				go func() {
					_, err := handler.conn.(io.Writer).Write(testData)
					assert.NoError(t, err)
				}()

				readData, err := handler.Read(context.Background())
				assert.NoError(t, err)
				assert.Equal(t, testData, readData)
			}
		})
	}
}

func TestNewNetConnIOHandler_RealNetwork(t *testing.T) { // More of an integration test
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err, "Failed to create listener")
	defer listener.Close()

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			t.Errorf("Failed to accept connection: %v", err)
			return
		}
		defer conn.Close()

		_, err = conn.Write([]byte("Hello, World!"))
		if err != nil {
			t.Errorf("Failed to write to connection: %v", err)
		}
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	require.NoError(t, err, "Failed to dial")

	tcpConn, ok := conn.(*net.TCPConn)
	require.True(t, ok, "Not a TCP connection")

	file, err := tcpConn.File()
	require.NoError(t, err, "Failed to get file from connection")

	fd := int(file.Fd())

	handler, err := NewIOHandler(fd)
	require.NoError(t, err, "Failed to create IOHandler")

	testData := []byte("Hello, World!")
	readData, err := handler.Read(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, testData, readData)

	err = handler.Close()
	assert.NoError(t, err)

	file.Close()
	conn.Close()
}

package server

import (
	"fmt"
	"net"
	"testing"
)

func TestNewNetConn(t *testing.T) {
	// Create a temporary network connection
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("failed to listen on a temporary port: %v", err)
	}
	defer l.Close()

	// Accept a connection
	go func() {
		_, err := l.Accept()
		if err != nil {
			t.Errorf("failed to accept connection: %v", err)
		}
	}()

	// Dial the server
	conn, err := net.Dial("tcp", l.Addr().String())
	if err != nil {
		t.Fatalf("failed to dial the server: %v", err)
	}
	defer conn.Close()

	// Get the file descriptor from the connection
	fd, err := getFDFromConn(conn)
	if err != nil {
		t.Fatalf("failed to get file descriptor from connection: %v", err)
	}

	// Create a NetConn instance
	nc := NewNetConn(fd)

	// Test writing to the connection
	msg := "hello"
	_, err = nc.Write([]byte(msg))
	if err != nil {
		t.Fatalf("failed to write to connection: %v", err)
	}

	// Test reading from the connection
	buffer := make([]byte, len(msg))
	_, err = nc.Read(buffer)
	if err != nil {
		t.Fatalf("failed to read from connection: %v", err)
	}
	if string(buffer) != msg {
		t.Errorf("expected %q, got %q", msg, string(buffer))
	}

	// Test closing the connection
	err = nc.Close()
	if err != nil {
		t.Fatalf("failed to close connection: %v", err)
	}

	// Test that no operations can be performed after closing
	_, err = nc.Write([]byte("test"))
	if err == nil {
		t.Errorf("expected error when writing after closing, got nil")
	}

	_, err = nc.Read(buffer)
	if err == nil {
		t.Errorf("expected error when reading after closing, got nil")
	}
}

// Helper function to extract the file descriptor from a net.Conn
func getFDFromConn(c net.Conn) (int, error) {
	tc, ok := c.(*net.TCPConn)
	if !ok {
		return -1, fmt.Errorf("failed to cast net.Conn to *net.TCPConn")
	}
	f, err := tc.File()
	if err != nil {
		return -1, err
	}
	return int(f.Fd()), nil
}

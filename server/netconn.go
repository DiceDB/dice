package server

import (
	"net"
	"syscall"
	"time"
)

// NetConn is a custom implementation of net.Conn.
type NetConn struct {
	fd int
}

// NewNetConn creates a new NetConn instance from a file descriptor.
func NewNetConn(fd int) *NetConn {
	return &NetConn{fd: fd}
}

// Read reads data from the connection.
func (nc *NetConn) Read(b []byte) (int, error) {
	n, err := syscall.Read(nc.fd, b)
	if err != nil {
		return n, err
	}
	return n, nil
}

// Write writes data to the connection.
func (nc *NetConn) Write(b []byte) (int, error) {
	n, err := syscall.Write(nc.fd, b)
	if err != nil {
		return n, err
	}
	return n, nil
}

// Close closes the connection.
func (nc *NetConn) Close() error {
	return syscall.Close(nc.fd)
}

// LocalAddr returns the local network address.
func (nc *NetConn) LocalAddr() net.Addr {
	return &net.IPAddr{}
}

// RemoteAddr returns the remote network address.
func (nc *NetConn) RemoteAddr() net.Addr {
	return &net.IPAddr{}
}

// SetDeadline sets the read and write deadlines for the connection.
func (nc *NetConn) SetDeadline(t time.Time) error {
	return nil
}

// SetReadDeadline sets the deadline for future read operations.
func (nc *NetConn) SetReadDeadline(t time.Time) error {
	return nil
}

// SetWriteDeadline sets the deadline for future write operations.
func (nc *NetConn) SetWriteDeadline(t time.Time) error {
	return nil
}

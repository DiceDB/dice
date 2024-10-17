package server

import (
	"errors"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
)

const (
	UnixScheme = "unix"
	TCPScheme  = "tcp"
	DiceScheme = "dice"
)

type Address struct {
	Kind string // allowed values: unix, tcp, dice
	Host string // e.g: "0.0.0.0", "::1"
	Port int    // e.g., 6379
	Path string // a valid unix socket path. eg: /tmp/dice.sock
}

func (a *Address) IsSocketAvail() bool {
	_, err := os.Stat(a.Path)
	return errors.Is(err, os.ErrNotExist)
}

func ParseAddress(address string) (*Address, error) {
	// Prefer tcp sockets
	if !strings.Contains(address, "://") {
		address = "tcp://" + address
	}
	u, err := url.Parse(address)
	if err != nil {
		return nil, err
	}

	switch u.Scheme {
	case UnixScheme:
		return &Address{
			Kind: "unix",
			Host: "",
			Port: -1,
			Path: u.Path,
		}, nil
	case TCPScheme, DiceScheme:
		host, portStr, err := net.SplitHostPort(u.Host)
		if err != nil {
			return nil, err
		}
		if host == "" {
			host = "0.0.0.0"
		}
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, err
		}
		return &Address{
			Kind: "dice",
			Host: host,
			Port: port,
			Path: "",
		}, nil
	}
	return nil, errors.New("invalid address supplied")
}

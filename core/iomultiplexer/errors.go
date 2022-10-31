package iomultiplexer

import "errors"

var (
	// ErrInvalidMaxClients is returned when the maxClients is less than 0
	ErrInvalidMaxClients = errors.New("invalid max clients")
)

// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package iomultiplexer

import (
	"fmt"
	"syscall"
	"time"

	"github.com/dicedb/dice/config"
)

// Epoll implements the IOMultiplexer interface for Linux-based systems
type Epoll struct {
	// fd stores the file descriptor of the epoll instance
	fd int
	// ePollEvents acts as a buffer for the events returned by the EpollWait syscall
	ePollEvents []syscall.EpollEvent
	// diceEvents stores the events after they are converted to the generic Event type
	// and is returned to the caller
	diceEvents []Event
}

// New creates a new Epoll instance
func New() (*Epoll, error) {
	if config.Config.MaxClients == 0 {
		return nil, ErrInvalidMaxClients
	}

	fd, err := syscall.EpollCreate1(0)
	if err != nil {
		return nil, err
	}

	return &Epoll{
		fd:          fd,
		ePollEvents: make([]syscall.EpollEvent, config.Config.MaxClients),
		diceEvents:  make([]Event, config.Config.MaxClients),
	}, nil
}

// Subscribe subscribes to the given event
func (ep *Epoll) Subscribe(event Event) error {
	nativeEvent := event.toNative()
	if err := syscall.EpollCtl(ep.fd, syscall.EPOLL_CTL_ADD, event.Fd, &nativeEvent); err != nil {
		return fmt.Errorf("epoll subscribe: %w", err)
	}
	return nil
}

// Poll polls for all the subscribed events simultaneously
// and returns all the events that were triggered
// It blocks until at least one event is triggered or the timeout is reached
func (ep *Epoll) Poll(timeout time.Duration) ([]Event, error) {
	nEvents, err := syscall.EpollWait(ep.fd, ep.ePollEvents, newTime(timeout))
	if err != nil {
		return nil, fmt.Errorf("epoll poll: %w", err)
	}

	for i := 0; i < nEvents; i++ {
		ep.diceEvents[i] = newEvent(ep.ePollEvents[i])
	}

	return ep.diceEvents[:nEvents], nil
}

// Close closes the Epoll instance
func (ep *Epoll) Close() error {
	return syscall.Close(ep.fd)
}

// This file is part of DiceDB.
// Copyright (C) 2025  DiceDB (dicedb.io).
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package iomultiplexer

import (
	"fmt"
	"syscall"
	"time"
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
func New(maxClients int32) (*Epoll, error) {
	if maxClients < 0 {
		return nil, ErrInvalidMaxClients
	}

	fd, err := syscall.EpollCreate1(0)
	if err != nil {
		return nil, err
	}

	return &Epoll{
		fd:          fd,
		ePollEvents: make([]syscall.EpollEvent, maxClients),
		diceEvents:  make([]Event, maxClients),
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

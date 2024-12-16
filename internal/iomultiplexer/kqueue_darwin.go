// This file is part of DiceDB.
// Copyright (C) 2024 DiceDB (dicedb.io).
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

// KQueue implements the IOMultiplexer interface for Darwin-based systems
type KQueue struct {
	// fd stores the file descriptor of the kqueue instance
	fd int
	// kQEvents acts as a buffer for the events returned by the Kevent syscall
	kQEvents []syscall.Kevent_t
	// diceEvents stores the events after they are converted to the generic Event type
	// and is returned to the caller
	diceEvents []Event
}

// New creates a new KQueue instance
func New(maxClients int32) (*KQueue, error) {
	if maxClients < 0 {
		return nil, ErrInvalidMaxClients
	}

	fd, err := syscall.Kqueue()
	if err != nil {
		return nil, err
	}

	return &KQueue{
		fd:         fd,
		kQEvents:   make([]syscall.Kevent_t, maxClients),
		diceEvents: make([]Event, maxClients),
	}, nil
}

// Subscribe subscribes to the given event
func (kq *KQueue) Subscribe(event Event) error {
	if subscribed, err := syscall.Kevent(kq.fd, []syscall.Kevent_t{event.toNative(syscall.EV_ADD)}, nil, nil); err != nil || subscribed == -1 {
		return fmt.Errorf("kqueue subscribe: %w", err)
	}
	return nil
}

// Poll polls for all the subscribed events simultaneously
// and returns all the events that were triggered
// It blocks until at least one event is triggered or the timeout is reached
func (kq *KQueue) Poll(timeout time.Duration) ([]Event, error) {
	nEvents, err := syscall.Kevent(kq.fd, nil, kq.kQEvents, newTime(timeout))
	if err != nil {
		return nil, fmt.Errorf("kqueue poll: %w", err)
	}

	for i := 0; i < nEvents; i++ {
		kq.diceEvents[i] = newEvent(kq.kQEvents[i])
	}

	return kq.diceEvents[:nEvents], nil
}

// Close closes the KQueue instance
func (kq *KQueue) Close() error {
	return syscall.Close(kq.fd)
}

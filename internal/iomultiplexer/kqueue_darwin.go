// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package iomultiplexer

import (
	"fmt"
	"syscall"
	"time"

	"github.com/dicedb/dice/config"
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
func New() (*KQueue, error) {
	if config.Config.MaxClients < 0 {
		return nil, ErrInvalidMaxClients
	}

	fd, err := syscall.Kqueue()
	if err != nil {
		return nil, err
	}

	return &KQueue{
		fd:         fd,
		kQEvents:   make([]syscall.Kevent_t, config.Config.MaxClients),
		diceEvents: make([]Event, config.Config.MaxClients),
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

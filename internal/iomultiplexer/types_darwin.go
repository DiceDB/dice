// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package iomultiplexer

import (
	"syscall"
	"time"
)

// newTime converts the given time.Duration to Darwin's timespec struct
func newTime(t time.Duration) *syscall.Timespec {
	if t < 0 {
		return nil
	}

	return &syscall.Timespec{
		Nsec: int64(t),
	}
}

// toNative converts the given generic Event to Darwin's Kevent_t struct
func (e Event) toNative(flags uint16) syscall.Kevent_t {
	return syscall.Kevent_t{
		Ident:  uint64(e.Fd),
		Filter: e.Op.toNative(),
		Flags:  flags,
	}
}

// newEvent converts the given Darwin's Kevent_t struct to the generic Event type
func newEvent(kEvent syscall.Kevent_t) Event {
	return Event{
		Fd: int(kEvent.Ident),
		Op: newOperations(kEvent.Filter),
	}
}

// toNative converts the given generic Operations to Darwin's filter type
func (op Operations) toNative() int16 {
	native := int16(0)

	if op&OpRead != 0 {
		native |= syscall.EVFILT_READ
	}
	if op&OpWrite != 0 {
		native |= syscall.EVFILT_WRITE
	}

	return native
}

// newOperations converts the given Darwin's filter type to the generic Operations type
func newOperations(filter int16) Operations {
	op := Operations(0)

	if filter&syscall.EVFILT_READ != 0 {
		op |= OpRead
	}
	if filter&syscall.EVFILT_WRITE != 0 {
		op |= OpWrite
	}

	return op
}

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
	"syscall"
	"time"
)

// newTime converts the given time.Duration to Linux's ms in int
func newTime(t time.Duration) int {
	if t < 0 {
		return -1
	}

	return int(t / time.Millisecond)
}

// toNative converts the given generic Event to Linux's EpollEvent struct
func (e Event) toNative() syscall.EpollEvent {
	return syscall.EpollEvent{
		Fd:     int32(e.Fd),
		Events: e.Op.toNative(),
	}
}

// newEvent converts the given Linux's EpollEvent struct to the generic Event type
func newEvent(ePEvent syscall.EpollEvent) Event {
	return Event{
		Fd: int(ePEvent.Fd),
		Op: newOperations(ePEvent.Events),
	}
}

// toNative converts the given generic Operations to Linux's EpollEvent type
func (op Operations) toNative() uint32 {
	native := uint32(0)

	if op&OpRead != 0 {
		native |= syscall.EPOLLIN
	}
	if op&OpWrite != 0 {
		native |= syscall.EPOLLOUT
	}

	return native
}

// newOperations converts the given Linux's EpollEvent type to the generic Operations type
func newOperations(events uint32) Operations {
	op := Operations(0)

	if events&syscall.EPOLLIN != 0 {
		op |= OpRead
	}
	if events&syscall.EPOLLOUT != 0 {
		op |= OpWrite
	}

	return op
}

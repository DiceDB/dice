// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package iomultiplexer

// Event is a platform independent representation of a subscribe event
// For linux platform, this is translated to syscall.EpollEvent
// For darwin platform, this is translated to syscall.Kevent_t
type Event struct {
	// Fd denotes the file descriptor
	Fd int
	// Op denotes the operations on file descriptor that are to be monitored
	Op Operations
}

// Operations is a platform independent representation of the operations
// that need to be monitored on a file descriptor
type Operations uint32

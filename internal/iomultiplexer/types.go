// This file is part of DiceDB.
// Copyright (C) 2025DiceDB (dicedb.io).
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
